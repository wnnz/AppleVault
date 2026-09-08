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
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	output, err := a.runIOSTool(ctx, "list", "--details")
	if err != nil {
		// 若获取详情失败（常见于锁屏休眠或会话超时），尝试 pair 握手激活并重试
		_, _ = a.runIOSToolInternal(ctx, true, "pair")
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

	// 进一步并发查询每台设备的真实个性化名称（如 "王铮的iPad"、"张三的 iPhone"）
	var wg sync.WaitGroup
	for i := range devices {
		if devices[i].UDID == "" {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			devCtx, devCancel := context.WithTimeout(ctx, 3*time.Second)
			defer devCancel()
			if realName := a.queryDeviceName(devCtx, devices[idx].UDID); realName != "" {
				devices[idx].Name = realName
			}
		}(i)
	}
	wg.Wait()

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
	return a.runIOSToolInternal(ctx, false, args...)
}

// runIOSToolInternal 执行 go-ios 命令。当 silent 为 true 时，不向日志面板输出日常内部调用日志（如查询设备名）。
func (a *App) runIOSToolInternal(ctx context.Context, silent bool, args ...string) (string, error) {
	executable := a.resolveIOSToolPath()
	if _, err := os.Stat(executable); err != nil {
		message := fmt.Sprintf("未找到设备安装工具 %s，请将 ios.exe 放在程序同目录下", executable)
		if !silent {
			a.emitLog(message)
		}
		return "", fmt.Errorf("%s", message)
	}

	if !silent {
		displayArgs := make([]string, len(args))
		for index, arg := range args {
			if strings.ContainsAny(arg, " \t") {
				displayArgs[index] = fmt.Sprintf("%q", arg)
			} else {
				displayArgs[index] = arg
			}
		}
		a.emitLog(fmt.Sprintf("> ios %s", strings.Join(displayArgs, " ")))
	}

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
			if !silent {
				a.emitLog(line)
			}
		}
	}()
	go func() {
		defer waitGroup.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrLines = append(stderrLines, line)
			if !silent {
				a.emitLog("[go-ios] " + line)
			}
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

// queryDeviceName 通过 go-ios devicename 获取设备由用户自定义的真实名称（如 "王铮的iPad"）。
func (a *App) queryDeviceName(ctx context.Context, udid string) string {
	if strings.TrimSpace(udid) == "" {
		return ""
	}
	output, err := a.runIOSToolInternal(ctx, true, "devicename", "--udid="+udid)
	if err != nil {
		return ""
	}
	return parseDeviceName(output)
}

// parseDeviceName 解析 go-ios devicename 命令返回的 JSON 数据。
func parseDeviceName(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var payload struct {
			DeviceName string `json:"devicename"`
			Name       string `json:"name"`
		}
		if err := json.Unmarshal([]byte(line), &payload); err == nil {
			candidate := strings.TrimSpace(payload.DeviceName)
			if candidate == "" {
				candidate = strings.TrimSpace(payload.Name)
			}
			if candidate != "" && !strings.EqualFold(candidate, "iPhone OS") {
				return candidate
			}
		}
	}
	return ""
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

		// 提取设备名称：注意 Apple lockdown 协议中的 ProductName 通常是固定的 "iPhone OS"，不能误当作设备名称
		name := findJSONString(value, "deviceName", "name")
		if strings.EqualFold(name, "iPhone OS") {
			name = ""
		}
		if name == "" {
			productName := findJSONString(value, "productName")
			if productName != "" && !strings.EqualFold(productName, "iPhone OS") {
				name = productName
			}
		}

		rawProductType := findJSONString(value, "productType", "modelNumber", "hardwareModel")
		friendlyModel := friendlyDeviceModel(rawProductType)
		formattedProductType := formatProductType(rawProductType)

		if name == "" {
			name = friendlyModel
		}
		if name == "" {
			name = "Apple 设备"
		}

		devices = append(devices, DeviceInfo{
			UDID:           udid,
			Name:           name,
			ProductType:    formattedProductType,
			ProductVersion: findJSONString(value, "productVersion", "osVersion", "version"),
			ConnectionType: findJSONString(value, "connectionType", "connection"),
		})
	}

	sort.Slice(devices, func(i, j int) bool {
		return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
	})
	return devices, nil
}

// 常见 Apple 内部硬件型号映射表（Model Identifier -> 友好商业名称）
var appleDeviceModels = map[string]string{
	// iPad Pro
	"iPad16,3": "iPad Pro 11 英寸 (M4)",
	"iPad16,4": "iPad Pro 11 英寸 (M4)",
	"iPad16,5": "iPad Pro 13 英寸 (M4)",
	"iPad16,6": "iPad Pro 13 英寸 (M4)",
	"iPad14,3": "iPad Pro 11 英寸 (第 4 代)",
	"iPad14,4": "iPad Pro 11 英寸 (第 4 代)",
	"iPad14,5": "iPad Pro 12.9 英寸 (第 6 代)",
	"iPad14,6": "iPad Pro 12.9 英寸 (第 6 代)",
	"iPad13,4": "iPad Pro 11 英寸 (第 3 代)",
	"iPad13,5": "iPad Pro 11 英寸 (第 3 代)",
	"iPad13,6": "iPad Pro 11 英寸 (第 3 代)",
	"iPad13,7": "iPad Pro 11 英寸 (第 3 代)",
	"iPad13,8": "iPad Pro 12.9 英寸 (第 5 代)",
	"iPad13,9": "iPad Pro 12.9 英寸 (第 5 代)",
	"iPad13,10": "iPad Pro 12.9 英寸 (第 5 代)",
	"iPad13,11": "iPad Pro 12.9 英寸 (第 5 代)",
	"iPad8,9":   "iPad Pro 11 英寸 (第 2 代)",
	"iPad8,10":  "iPad Pro 11 英寸 (第 2 代)",
	"iPad8,11":  "iPad Pro 12.9 英寸 (第 4 代)",
	"iPad8,12":  "iPad Pro 12.9 英寸 (第 4 代)",
	"iPad8,1":   "iPad Pro 11 英寸",
	"iPad8,2":   "iPad Pro 11 英寸",
	"iPad8,3":   "iPad Pro 11 英寸",
	"iPad8,4":   "iPad Pro 11 英寸",
	"iPad8,5":   "iPad Pro 12.9 英寸 (第 3 代)",
	"iPad8,6":   "iPad Pro 12.9 英寸 (第 3 代)",
	"iPad8,7":   "iPad Pro 12.9 英寸 (第 3 代)",
	"iPad8,8":   "iPad Pro 12.9 英寸 (第 3 代)",

	// iPad Air
	"iPad16,1":  "iPad Air 11 英寸 (M2)",
	"iPad16,2":  "iPad Air 13 英寸 (M2)",
	"iPad13,16": "iPad Air (第 5 代)",
	"iPad13,17": "iPad Air (第 5 代)",
	"iPad13,1":  "iPad Air (第 4 代)",
	"iPad13,2":  "iPad Air (第 4 代)",
	"iPad11,3":  "iPad Air (第 3 代)",
	"iPad11,4":  "iPad Air (第 3 代)",

	// iPad mini
	"iPad14,1": "iPad mini (第 6 代)",
	"iPad14,2": "iPad mini (第 6 代)",
	"iPad11,1": "iPad mini (第 5 代)",
	"iPad11,2": "iPad mini (第 5 代)",

	// iPad (标准版)
	"iPad13,18": "iPad (第 10 代)",
	"iPad13,19": "iPad (第 10 代)",
	"iPad12,1":  "iPad (第 9 代)",
	"iPad12,2":  "iPad (第 9 代)",
	"iPad11,6":  "iPad (第 8 代)",
	"iPad11,7":  "iPad (第 8 代)",
	"iPad7,11":  "iPad (第 7 代)",
	"iPad7,12":  "iPad (第 7 代)",
	"iPad7,5":   "iPad (第 6 代)",
	"iPad7,6":   "iPad (第 6 代)",

	// iPhone 16 系列
	"iPhone17,1": "iPhone 16 Pro",
	"iPhone17,2": "iPhone 16 Pro Max",
	"iPhone17,3": "iPhone 16",
	"iPhone17,4": "iPhone 16 Plus",

	// iPhone 15 系列
	"iPhone16,1": "iPhone 15 Pro",
	"iPhone16,2": "iPhone 15 Pro Max",
	"iPhone15,4": "iPhone 15",
	"iPhone15,5": "iPhone 15 Plus",

	// iPhone 14 系列
	"iPhone15,2": "iPhone 14 Pro",
	"iPhone15,3": "iPhone 14 Pro Max",
	"iPhone14,7": "iPhone 14",
	"iPhone14,8": "iPhone 14 Plus",

	// iPhone 13 系列
	"iPhone14,2": "iPhone 13 Pro",
	"iPhone14,3": "iPhone 13 Pro Max",
	"iPhone14,4": "iPhone 13 mini",
	"iPhone14,5": "iPhone 13",

	// iPhone 12 系列
	"iPhone13,1": "iPhone 12 mini",
	"iPhone13,2": "iPhone 12",
	"iPhone13,3": "iPhone 12 Pro",
	"iPhone13,4": "iPhone 12 Pro Max",

	// iPhone 11 系列
	"iPhone12,1": "iPhone 11",
	"iPhone12,3": "iPhone 11 Pro",
	"iPhone12,5": "iPhone 11 Pro Max",

	// iPhone SE
	"iPhone14,6": "iPhone SE (第 3 代)",
	"iPhone12,8": "iPhone SE (第 2 代)",
	"iPhone8,4":  "iPhone SE (第 1 代)",

	// iPhone X / XS / XR
	"iPhone11,2": "iPhone XS",
	"iPhone11,4": "iPhone XS Max",
	"iPhone11,6": "iPhone XS Max",
	"iPhone11,8": "iPhone XR",
	"iPhone10,3": "iPhone X",
	"iPhone10,6": "iPhone X",

	// iPhone 8 / 7 / 6s
	"iPhone10,1": "iPhone 8",
	"iPhone10,4": "iPhone 8",
	"iPhone10,2": "iPhone 8 Plus",
	"iPhone10,5": "iPhone 8 Plus",
	"iPhone9,1":  "iPhone 7",
	"iPhone9,3":  "iPhone 7",
	"iPhone9,2":  "iPhone 7 Plus",
	"iPhone9,4":  "iPhone 7 Plus",
}

// friendlyDeviceModel 将硬件型号标识转换为商业中文名称。
func friendlyDeviceModel(productType string) string {
	productType = strings.TrimSpace(productType)
	if productType == "" {
		return "Apple 设备"
	}
	if name, ok := appleDeviceModels[productType]; ok {
		return name
	}
	if strings.HasPrefix(productType, "iPad") {
		return "iPad (" + productType + ")"
	}
	if strings.HasPrefix(productType, "iPhone") {
		return "iPhone (" + productType + ")"
	}
	if strings.HasPrefix(productType, "Watch") {
		return "Apple Watch (" + productType + ")"
	}
	if strings.HasPrefix(productType, "AppleTV") {
		return "Apple TV (" + productType + ")"
	}
	if strings.HasPrefix(productType, "iPod") {
		return "iPod touch (" + productType + ")"
	}
	return productType
}

// formatProductType 格式化型号标识，若存在商业名称则附带硬件标识（如 "iPad Air 13 英寸 (M2) (iPad16,2)"）。
func formatProductType(productType string) string {
	productType = strings.TrimSpace(productType)
	if productType == "" {
		return "Apple 设备"
	}
	model := friendlyDeviceModel(productType)
	if model != productType && !strings.Contains(model, productType) {
		return fmt.Sprintf("%s (%s)", model, productType)
	}
	return model
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
