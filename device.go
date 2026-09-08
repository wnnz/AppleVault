package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// SelectIPA 打开文件选择器并返回用户选择的 IPA 绝对路径。
func (a *App) SelectIPA() (string, error) {
	return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择要安装的 IPA 文件",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "IPA 安装包 (*.ipa)", Pattern: "*.ipa"},
		},
	})
}

// ListDevices 返回当前通过 USB 或网络连接且能被 go-ios 识别的苹果设备。
func (a *App) ListDevices() ([]DeviceInfo, error) {
	ctx := a.startCommandContext()
	output, err := a.runIOSTool(ctx, "list", "--details")
	if err != nil {
		// 若获取详情失败（常见于锁屏休眠或会话超时），尝试 pair 握手激活并重试
		_, _ = a.runIOSTool(ctx, "pair")
		if retryOutput, retryErr := a.runIOSTool(ctx, "list", "--details"); retryErr == nil {
			output = retryOutput
			err = nil
		} else {
			// 若依然无法获取详情，尝试降级为基础设备列表（仅通过 USB 管道枚举 UDID）
			if fallbackOutput, fallbackErr := a.runIOSTool(ctx, "list"); fallbackErr == nil {
				output = fallbackOutput
				err = nil
			}
		}
	}
	if err != nil {
		return nil, friendlyDeviceToolError(err)
	}

	devices, err := parseDeviceList(output)
	if err != nil {
		return nil, fmt.Errorf("解析设备列表失败: %w", err)
	}
	if len(devices) == 0 {
		a.emitLog("未发现可用苹果设备，请确认设备已解锁并信任此电脑。")
	}
	return devices, nil
}

// InstallIPA 将指定 IPA 安装到目标设备。IPA 必须具有可供该设备使用的有效签名。
func (a *App) InstallIPA(ipaPath, udid string) (InstallResult, error) {
	absolutePath, err := validateIPAPath(ipaPath)
	if err != nil {
		return InstallResult{}, err
	}
	udid = strings.TrimSpace(udid)
	if udid == "" {
		return InstallResult{}, fmt.Errorf("请选择要安装的苹果设备")
	}

	a.emitLog(fmt.Sprintf("开始安装 IPA 到设备 %s: %s", udid, absolutePath))
	ctx := a.startCommandContext()
	output, err := a.runIOSTool(ctx, "install", "--path="+absolutePath, "--udid="+udid)
	if err != nil {
		return InstallResult{}, friendlyDeviceToolError(err)
	}

	a.emitLog("IPA 安装完成。")
	return InstallResult{
		Success: true,
		Message: "IPA 已成功安装到设备",
		Output:  strings.TrimSpace(output),
	}, nil
}

// resolveIOSToolPath 查找随应用放置在 tools 目录、bin 目录、同目录或系统已安装的 go-ios 设备工具。
func (a *App) resolveIOSToolPath() string {
	// 1. 同程序目录下的 tools / bin / 同级目录
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		pTools := filepath.Join(dir, "tools", "ios.exe")
		if _, err := os.Stat(pTools); err == nil {
			return pTools
		}
		pBin := filepath.Join(dir, "bin", "ios.exe")
		if _, err := os.Stat(pBin); err == nil {
			return pBin
		}
		p := filepath.Join(dir, "ios.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 2. 当前工作目录下的 tools / bin / 当前工作目录
	if cwd, err := os.Getwd(); err == nil {
		pTools := filepath.Join(cwd, "tools", "ios.exe")
		if _, err := os.Stat(pTools); err == nil {
			return pTools
		}
		pBin := filepath.Join(cwd, "bin", "ios.exe")
		if _, err := os.Stat(pBin); err == nil {
			return pBin
		}
		candidate := filepath.Join(cwd, "ios.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// 3. 系统 PATH
	if candidate, err := exec.LookPath("ios.exe"); err == nil {
		return candidate
	}
	return "ios.exe"
}

// runIOSTool 执行 go-ios 命令，并将标准输出和错误输出转发到现有日志面板。
func (a *App) runIOSTool(ctx context.Context, args ...string) (string, error) {
	executable := a.resolveIOSToolPath()
	if _, err := os.Stat(executable); err != nil {
		message := fmt.Sprintf("未找到设备安装工具 %s，请将 ios.exe 放在程序同目录下", executable)
		a.emitLog(message)
		return "", fmt.Errorf("%s", message)
	}

	displayArgs := make([]string, len(args))
	for index, arg := range args {
		if strings.ContainsAny(arg, " \t") {
			displayArgs[index] = fmt.Sprintf("%q", arg)
		} else {
			displayArgs[index] = arg
		}
	}
	a.emitLog(fmt.Sprintf("> ios %s", strings.Join(displayArgs, " ")))

	command := exec.CommandContext(ctx, executable, args...)
	setSysProcAttr(command)
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return "", err
	}
	if err := command.Start(); err != nil {
		return "", err
	}

	var stdoutLines []string
	var stderrLines []string
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutLines = append(stdoutLines, line)
			a.emitLog(line)
		}
	}()
	go func() {
		defer waitGroup.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrLines = append(stderrLines, line)
			a.emitLog("[go-ios] " + line)
		}
	}()

	waitGroup.Wait()
	commandErr := command.Wait()
	stdout := strings.Join(stdoutLines, "\n")
	stderr := strings.Join(stderrLines, "\n")
	if commandErr != nil {
		if strings.TrimSpace(stderr) != "" {
			return stdout, fmt.Errorf("%s", strings.TrimSpace(stderr))
		}
		return stdout, commandErr
	}
	return stdout, nil
}

// parseDeviceList 将 go-ios 的 JSON 输出转换为前端使用的稳定设备模型。
func parseDeviceList(output string) ([]DeviceInfo, error) {
	var envelope struct {
		DeviceList []json.RawMessage `json:"deviceList"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &envelope); err != nil {
		return nil, err
	}

	devices := make([]DeviceInfo, 0, len(envelope.DeviceList))
	seen := make(map[string]struct{}, len(envelope.DeviceList))
	for _, rawDevice := range envelope.DeviceList {
		var value any
		if err := json.Unmarshal(rawDevice, &value); err != nil {
			return nil, err
		}

		// 若直接为字符串（例如 ios list 输出的纯 UDID 数组）
		if udidStr, ok := value.(string); ok && strings.TrimSpace(udidStr) != "" {
			udid := strings.TrimSpace(udidStr)
			if _, exists := seen[udid]; !exists {
				seen[udid] = struct{}{}
				devices = append(devices, DeviceInfo{
					UDID:           udid,
					Name:           "已连接设备 (请点亮并解锁屏幕)",
					ProductType:    "iOS 设备",
					ProductVersion: "已锁定",
					ConnectionType: "USB",
				})
			}
			continue
		}

		udid := findJSONString(value, "serialNumber", "udid", "uniqueDeviceID", "identifier")
		if udid == "" {
			continue
		}
		if _, exists := seen[udid]; exists {
			continue
		}
		seen[udid] = struct{}{}

		name := findJSONString(value, "deviceName", "productName", "name")
		productType := findJSONString(value, "productType", "modelNumber", "hardwareModel")
		if name == "" {
			name = productType
		}
		if name == "" {
			name = "Apple 设备"
		}

		devices = append(devices, DeviceInfo{
			UDID:           udid,
			Name:           name,
			ProductType:    productType,
			ProductVersion: findJSONString(value, "productVersion", "osVersion", "version"),
			ConnectionType: findJSONString(value, "connectionType", "connection"),
		})
	}

	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
	})
	return devices, nil
}

// findJSONString 按候选字段优先级递归查找 JSON 字符串，兼容 go-ios 不同版本的嵌套结构。
func findJSONString(value any, candidateKeys ...string) string {
	for _, candidate := range candidateKeys {
		if result := findJSONStringByKey(value, candidate); result != "" {
			return result
		}
	}
	return ""
}

// findJSONStringByKey 在任意嵌套 JSON 值中查找指定字段的非空字符串值。
func findJSONStringByKey(value any, candidateKey string) string {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if strings.EqualFold(key, candidateKey) {
				if text, ok := child.(string); ok {
					return strings.TrimSpace(text)
				}
			}
		}
		for _, child := range typed {
			if result := findJSONStringByKey(child, candidateKey); result != "" {
				return result
			}
		}
	case []any:
		for _, child := range typed {
			if result := findJSONStringByKey(child, candidateKey); result != "" {
				return result
			}
		}
	}
	return ""
}

// validateIPAPath 校验 IPA 路径并返回规范化的绝对路径。
func validateIPAPath(ipaPath string) (string, error) {
	ipaPath = strings.TrimSpace(ipaPath)
	if ipaPath == "" {
		return "", fmt.Errorf("请选择要安装的 IPA 文件")
	}
	absolutePath, err := filepath.Abs(ipaPath)
	if err != nil {
		return "", fmt.Errorf("无法解析 IPA 路径: %w", err)
	}
	if !strings.EqualFold(filepath.Ext(absolutePath), ".ipa") {
		return "", fmt.Errorf("所选文件不是 IPA 安装包")
	}
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("无法访问 IPA 文件: %w", err)
	}
	if fileInfo.IsDir() {
		return "", fmt.Errorf("所选路径不是 IPA 文件")
	}
	return absolutePath, nil
}

// friendlyDeviceToolError 将常见的设备通信错误转换为可操作的中文提示。
func friendlyDeviceToolError(err error) error {
	message := strings.TrimSpace(err.Error())
	lowerMessage := strings.ToLower(message)
	switch {
	case strings.Contains(lowerMessage, "invalidhostid"),
		strings.Contains(lowerMessage, "startsession failed"),
		strings.Contains(lowerMessage, "not trusted"):
		return fmt.Errorf("设备尚未信任此电脑或配对失效。请解锁手机屏幕，若弹出「要信任此电脑吗？」请点击「信任」并输入锁屏密码后重试；若未弹出提示，可尝试重新拔插数据线")
	case strings.Contains(lowerMessage, "no device"), strings.Contains(lowerMessage, "device not found"):
		return fmt.Errorf("未找到目标设备，请确认设备已连接、解锁并信任此电脑")
	case strings.Contains(lowerMessage, "password protected"), strings.Contains(lowerMessage, "locked"):
		return fmt.Errorf("设备处于锁定状态，请点亮并解锁设备屏幕后重试")
	case strings.Contains(lowerMessage, "pair"), strings.Contains(lowerMessage, "trust"):
		return fmt.Errorf("设备尚未信任此电脑，请在设备上完成信任操作后重试")
	case strings.Contains(lowerMessage, "invalid signature"), strings.Contains(lowerMessage, "applicationverificationfailed"):
		return fmt.Errorf("IPA 签名无效或不适用于此设备: %s", message)
	case strings.Contains(lowerMessage, "exit status 1"):
		return fmt.Errorf("设备连接中断或已息屏锁屏。请点亮并解锁设备屏幕后重新刷新")
	default:
		return fmt.Errorf("%s", message)
	}
}
