package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	settings   Settings
	settingsMu sync.RWMutex
	cancelMu   sync.Mutex
	cancelFn   context.CancelFunc

	purchaseCancelMu     sync.Mutex
	purchaseCancelers    map[uint64]context.CancelFunc
	nextPurchaseCancelID uint64

	tasksMu     sync.RWMutex
	tasks       []*DownloadTask
	taskCancels map[string]context.CancelFunc

	accountMu          sync.RWMutex
	cachedAccountEmail string
}

func NewApp() *App {
	app := &App{
		settings: Settings{
			KeychainPassphrase: "123456",
			DefaultPlatform:    "iphone",
			EnableProxy:        true,
			ProxyUrl:           "http://127.0.0.1:10808",
		},
		tasks:             make([]*DownloadTask, 0),
		taskCancels:       make(map[string]context.CancelFunc),
		purchaseCancelers: make(map[uint64]context.CancelFunc),
	}
	app.settings.DefaultDownloadDir = app.getDownloadsDir()
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.ensureDataMigration()
	a.loadSettings()
	a.loadTasks()
	initWindowIcon()

	go func() {
		acc, err := a.GetAccountInfo()
		if err == nil && acc.Success && acc.Email != "" {
			a.accountMu.Lock()
			a.cachedAccountEmail = acc.Email
			a.accountMu.Unlock()
			a.settingsMu.Lock()
			a.settings.DefaultDownloadDir = a.getDownloadsDir()
			a.settingsMu.Unlock()
		}
	}()
}

func (a *App) getDataDir() string {
	// 1. 检查程序所在目录下的 data (release 模式)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		p := filepath.Join(dir, "data")
		lowerDir := strings.ToLower(dir)
		if !strings.Contains(lowerDir, "temp") && !strings.Contains(lowerDir, "tmp") {
			if fi, err := os.Stat(p); err == nil && fi.IsDir() {
				return p
			}
		}
	}
	// 2. 检查工作目录下的 build/bin/data 或 data (dev 模式)
	if cwd, err := os.Getwd(); err == nil {
		pBinData := filepath.Join(cwd, "build", "bin", "data")
		if fi, err := os.Stat(pBinData); err == nil && fi.IsDir() {
			return pBinData
		}
		pCwdData := filepath.Join(cwd, "data")
		if fi, err := os.Stat(pCwdData); err == nil && fi.IsDir() {
			return pCwdData
		}
	}
	// 3. 默认回退
	var baseDir string
	if exe, err := os.Executable(); err == nil {
		baseDir = filepath.Dir(exe)
	} else {
		baseDir, _ = os.Getwd()
	}
	dataDir := filepath.Join(baseDir, "data")
	_ = os.MkdirAll(dataDir, 0755)
	return dataDir
}

func sanitizeAccountDir(email string) string {
	clean := strings.TrimSpace(email)
	if clean == "" {
		return "default"
	}
	invalidChars := `\/:*?"<>|`
	for _, c := range invalidChars {
		clean = strings.ReplaceAll(clean, string(c), "_")
	}
	return clean
}

func (a *App) getAccountIdentifier() string {
	a.accountMu.RLock()
	email := a.cachedAccountEmail
	a.accountMu.RUnlock()

	if email != "" {
		return sanitizeAccountDir(email)
	}
	return "default"
}

func (a *App) getDownloadsDir() string {
	accountDir := a.getAccountIdentifier()
	downloadsDir := filepath.Join(a.getDataDir(), "downloads", accountDir)
	_ = os.MkdirAll(downloadsDir, 0755)
	return downloadsDir
}

func (a *App) getSettingsFilePath() string {
	return filepath.Join(a.getDataDir(), "settings.json")
}

func (a *App) getTasksFilePath() string {
	return filepath.Join(a.getDataDir(), "tasks.json")
}

func (a *App) ensureDataMigration() {
	dataDir := a.getDataDir()

	// 1. 迁移旧 APPDATA 下的 settings.json
	newSettings := filepath.Join(dataDir, "settings.json")
	if _, err := os.Stat(newSettings); os.IsNotExist(err) {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			oldSettings := filepath.Join(appData, "IPAToolGUI", "settings.json")
			if oldData, err := os.ReadFile(oldSettings); err == nil {
				_ = os.WriteFile(newSettings, oldData, 0644)
			}
		}
	}

	// 2. 迁移旧 ~/.ipatool 下的登录与钥匙串信息
	newIpatoolDir := filepath.Join(dataDir, ".ipatool")
	if _, err := os.Stat(newIpatoolDir); os.IsNotExist(err) {
		if home, err := os.UserHomeDir(); err == nil {
			oldIpatoolDir := filepath.Join(home, ".ipatool")
			if info, err := os.Stat(oldIpatoolDir); err == nil && info.IsDir() {
				copyDir(oldIpatoolDir, newIpatoolDir)
			}
		}
	}
}

func copyDir(src, dst string) {
	_ = os.MkdirAll(dst, 0755)
	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(srcPath, dstPath)
		} else {
			data, err := os.ReadFile(srcPath)
			if err == nil {
				_ = os.WriteFile(dstPath, data, 0644)
			}
		}
	}
}

func (a *App) loadSettings() {
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()

	p := a.getSettingsFilePath()
	if data, err := os.ReadFile(p); err == nil {
		var s Settings
		if err := json.Unmarshal(data, &s); err == nil {
			a.settings = s
		}
	}

	if a.settings.ProxyUrl == "" {
		a.settings.ProxyUrl = "http://127.0.0.1:10808"
	}
	// 下载路径固定为 data/downloads，不可修改
	a.settings.DefaultDownloadDir = a.getDownloadsDir()
	if a.settings.DefaultPlatform == "" {
		a.settings.DefaultPlatform = "iphone"
	}

	a.settings.IpaToolPath = a.resolveIpaToolPath()
}

func (a *App) loadTasks() {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	p := a.getTasksFilePath()
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}

	var list []*DownloadTask
	if err := json.Unmarshal(data, &list); err == nil {
		for _, t := range list {
			if t.Status == "downloading" || t.Status == "pending" {
				t.Status = "canceled"
				t.Speed = "已中断"
			}
		}
		a.tasks = list
	}
}

func (a *App) saveTasksLocked() {
	p := a.getTasksFilePath()
	data, err := json.MarshalIndent(a.tasks, "", "  ")
	if err == nil {
		_ = os.WriteFile(p, data, 0644)
	}
}

func (a *App) resolveIpaToolPath() string {
	// 1. 同程序目录下的 tools 目录、bin 目录与根目录
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		pTools := filepath.Join(dir, "tools", "ipatool.exe")
		if _, err := os.Stat(pTools); err == nil {
			return pTools
		}
		pBin := filepath.Join(dir, "bin", "ipatool.exe")
		if _, err := os.Stat(pBin); err == nil {
			return pBin
		}
		p := filepath.Join(dir, "ipatool.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 2. 当前工作目录下的 tools、bin 或根目录
	if cwd, err := os.Getwd(); err == nil {
		pTools := filepath.Join(cwd, "tools", "ipatool.exe")
		if _, err := os.Stat(pTools); err == nil {
			return pTools
		}
		pBin := filepath.Join(cwd, "bin", "ipatool.exe")
		if _, err := os.Stat(pBin); err == nil {
			return pBin
		}
		p := filepath.Join(cwd, "ipatool.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 3. 系统 PATH
	if p, err := exec.LookPath("ipatool.exe"); err == nil {
		return p
	}

	return "ipatool.exe"
}

func (a *App) emitLog(message string) {
	if a.ctx != nil {
		timestamp := time.Now().Format("15:04:05")
		runtime.EventsEmit(a.ctx, "log", fmt.Sprintf("[%s] %s", timestamp, message))
	}
}

func (a *App) buildIpaToolCmd(ctx context.Context, exePath string, cmdArgs []string, enableProxy bool, proxyUrl string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, exePath, cmdArgs...)
	setSysProcAttr(cmd)

	dataDir := a.getDataDir()
	drive := filepath.VolumeName(dataDir)
	pathNoVolume := strings.TrimPrefix(dataDir, drive)
	if drive == "" {
		drive = "C:"
	}

	env := os.Environ()
	filtered := make([]string, 0, len(env)+12)
	for _, e := range env {
		upper := strings.ToUpper(e)
		if strings.HasPrefix(upper, "USERPROFILE=") ||
			strings.HasPrefix(upper, "HOME=") ||
			strings.HasPrefix(upper, "HOMEDRIVE=") ||
			strings.HasPrefix(upper, "HOMEPATH=") ||
			strings.HasPrefix(upper, "IPATOOL_DIRECTORY=") {
			continue
		}
		filtered = append(filtered, e)
	}

	ipatoolDir := filepath.Join(dataDir, ".ipatool")
	filtered = append(filtered,
		"USERPROFILE="+dataDir,
		"HOME="+dataDir,
		"HOMEDRIVE="+drive,
		"HOMEPATH="+pathNoVolume,
		"IPATOOL_DIRECTORY="+ipatoolDir,
	)

	if enableProxy && strings.TrimSpace(proxyUrl) != "" {
		proxy := strings.TrimSpace(proxyUrl)
		filtered = append(filtered,
			"HTTP_PROXY="+proxy,
			"HTTPS_PROXY="+proxy,
			"ALL_PROXY="+proxy,
			"http_proxy="+proxy,
			"https_proxy="+proxy,
			"all_proxy="+proxy,
		)
		a.emitLog(fmt.Sprintf("[网络代理] 已启用: %s", proxy))
	}

	cmd.Dir = dataDir
	cmd.Env = filtered
	return cmd
}

func (a *App) runIpaTool(ctx context.Context, args ...string) (string, error) {
	exePath := a.resolveIpaToolPath()
	if _, err := os.Stat(exePath); err != nil {
		errStr := fmt.Sprintf("未在程序同目录下找到 ipatool.exe！请确认存在于: %s", exePath)
		a.emitLog(errStr)
		return "", fmt.Errorf("%s", errStr)
	}

	cmdArgs := make([]string, 0, len(args)+4)
	cmdArgs = append(cmdArgs, args...)

	hasNonInteractive := false
	hasPassphrase := false
	for _, arg := range cmdArgs {
		if arg == "--non-interactive" {
			hasNonInteractive = true
		}
		if arg == "--keychain-passphrase" {
			hasPassphrase = true
		}
	}

	if !hasNonInteractive {
		cmdArgs = append(cmdArgs, "--non-interactive")
	}

	a.settingsMu.RLock()
	passphrase := a.settings.KeychainPassphrase
	enableProxy := a.settings.EnableProxy
	proxyUrl := a.settings.ProxyUrl
	a.settingsMu.RUnlock()

	if !hasPassphrase && passphrase != "" {
		cmdArgs = append(cmdArgs, "--keychain-passphrase", passphrase)
	}

	cmd := a.buildIpaToolCmd(ctx, exePath, cmdArgs, enableProxy, proxyUrl)

	// Mask passphrase in logs
	displayArgs := make([]string, len(cmdArgs))
	for i, arg := range cmdArgs {
		if i > 0 && cmdArgs[i-1] == "--keychain-passphrase" {
			displayArgs[i] = "******"
		} else if strings.Contains(arg, " ") {
			displayArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			displayArgs[i] = arg
		}
	}
	a.emitLog(fmt.Sprintf("> ipatool %s", strings.Join(displayArgs, " ")))

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var stdoutLines []string
	var stderrLines []string
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutLines = append(stdoutLines, line)
			a.emitLog(line)
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			stderrLines = append(stderrLines, line)
			a.emitLog("[ERR] " + line)
		}
	}()

	wg.Wait()
	cmdErr := cmd.Wait()

	stdout := strings.Join(stdoutLines, "\n")
	stderr := strings.Join(stderrLines, "\n")

	if cmdErr != nil {
		allOutput := stderr + "\n" + stdout
		// 1. 优先解析 JSON 格式错误: {"error": "..."} 或 {"message": "..."}
		for _, line := range strings.Split(allOutput, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
				var errObj struct {
					Error   string `json:"error"`
					Message string `json:"message"`
				}
				if err := json.Unmarshal([]byte(line), &errObj); err == nil {
					if errObj.Error != "" {
						return stdout, fmt.Errorf("%s", errObj.Error)
					}
					if errObj.Message != "" {
						return stdout, fmt.Errorf("%s", errObj.Message)
					}
				}
			}
		}

		// 2. 尝试 logfmt 格式 error="([^"]+)"
		re := regexp.MustCompile(`error="([^"]+)"`)
		if matches := re.FindStringSubmatch(allOutput); len(matches) > 1 {
			return stdout, fmt.Errorf("%s", matches[1])
		}
		if strings.TrimSpace(stderr) != "" {
			return stdout, fmt.Errorf("%s", stderr)
		}
		return stdout, cmdErr
	}

	return stdout, nil
}

func parseJSONFromOutput[T any](output string) (T, error) {
	var zero T
	lines := strings.Split(output, "\n")

	// 第一轮优先筛选：跳过纯日志行（只有 level/time/message 的 log 结构，或 level 为 error）
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			var rawMap map[string]interface{}
			if err := json.Unmarshal([]byte(line), &rawMap); err == nil {
				if level, ok := rawMap["level"].(string); ok && level == "error" {
					continue
				}
				hasOnlyLogKeys := true
				for k := range rawMap {
					if k != "level" && k != "time" && k != "message" {
						hasOnlyLogKeys = false
						break
					}
				}
				if hasOnlyLogKeys {
					continue
				}

				var result T
				if err := json.Unmarshal([]byte(line), &result); err == nil {
					return result, nil
				}
			}
		}
	}

	// 第二轮兜底：任意可解析为目标类型的 JSON
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			var result T
			if err := json.Unmarshal([]byte(line), &result); err == nil {
				return result, nil
			}
		}
	}

	var result T
	if err := json.Unmarshal([]byte(output), &result); err == nil {
		return result, nil
	}
	return zero, fmt.Errorf("failed to parse JSON from output")
}

func formatBytes(b int64) string {
	if b <= 0 {
		return "-"
	}
	const unit = 1024.0
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= int64(unit)
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), units[exp])
}

// --- Bound Methods for Wails Frontend ---

func (a *App) GetSettings() Settings {
	a.settingsMu.RLock()
	defer a.settingsMu.RUnlock()
	s := a.settings
	s.DefaultDownloadDir = a.getDownloadsDir()
	s.IpaToolPath = a.resolveIpaToolPath()
	return s
}

func (a *App) SaveSettings(s Settings) error {
	a.settingsMu.Lock()
	s.DefaultDownloadDir = a.getDownloadsDir()
	a.settings = s
	a.settings.IpaToolPath = a.resolveIpaToolPath()
	data, err := json.MarshalIndent(a.settings, "", "  ")
	a.settingsMu.Unlock()

	if err != nil {
		return err
	}
	return os.WriteFile(a.getSettingsFilePath(), data, 0644)
}

func (a *App) SelectDirectory(title, defaultDir string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
}

func (a *App) TestProxy(proxyUrl string) ProxyTestResult {
	if strings.TrimSpace(proxyUrl) == "" {
		return ProxyTestResult{Success: false, Message: "代理地址不能为空"}
	}

	pURL, err := url.Parse(proxyUrl)
	if err != nil {
		return ProxyTestResult{Success: false, Message: fmt.Sprintf("代理地址格式无效: %v", err)}
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(pURL),
		},
		Timeout: 6 * time.Second,
	}

	resp, err := client.Get("https://apple.com")
	if err != nil {
		return ProxyTestResult{Success: false, Message: fmt.Sprintf("连接失败: %v", err)}
	}
	defer resp.Body.Close()

	return ProxyTestResult{Success: true, Message: fmt.Sprintf("连接成功！HTTP 响应状态码: %s", resp.Status)}
}

func (a *App) OpenInExplorer(targetPath string) error {
	if strings.TrimSpace(targetPath) == "" {
		targetPath = a.getDownloadsDir()
	}
	if fi, err := os.Stat(targetPath); err == nil {
		if fi.IsDir() {
			return exec.Command("explorer", targetPath).Start()
		}
		return exec.Command("explorer", "/select,", targetPath).Start()
	}
	dir := filepath.Dir(targetPath)
	return exec.Command("explorer", dir).Start()
}

func (a *App) CancelRunningCommand() {
	a.cancelMu.Lock()
	cancel := a.cancelFn
	a.cancelFn = nil
	a.cancelMu.Unlock()

	if cancel != nil {
		cancel()
	}
	purchaseCancelled := a.cancelPurchaseCommands()
	cancelled := cancel != nil || purchaseCancelled
	if cancelled {
		a.emitLog("用户取消了当前操作。")
	}
}

func (a *App) cancelPurchaseCommands() bool {
	a.purchaseCancelMu.Lock()
	cancelers := make([]context.CancelFunc, 0, len(a.purchaseCancelers))
	for id, cancel := range a.purchaseCancelers {
		cancelers = append(cancelers, cancel)
		delete(a.purchaseCancelers, id)
	}
	a.purchaseCancelMu.Unlock()

	for _, cancel := range cancelers {
		cancel()
	}
	return len(cancelers) > 0
}

func (a *App) startPurchaseCommandContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	a.purchaseCancelMu.Lock()
	a.nextPurchaseCancelID++
	id := a.nextPurchaseCancelID
	a.purchaseCancelers[id] = cancel
	a.purchaseCancelMu.Unlock()

	cleanup := func() {
		a.purchaseCancelMu.Lock()
		delete(a.purchaseCancelers, id)
		a.purchaseCancelMu.Unlock()
	}
	return ctx, func() {
		cancel()
		cleanup()
	}
}

func (a *App) startCommandContext() context.Context {
	a.cancelMu.Lock()
	defer a.cancelMu.Unlock()
	if a.cancelFn != nil {
		a.cancelFn()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFn = cancel
	return ctx
}

func (a *App) GetAccountInfo() (AccountInfo, error) {
	ctx := a.startCommandContext()
	out, err := a.runIpaTool(ctx, "auth", "info", "--format", "json")
	if err != nil {
		return AccountInfo{}, err
	}
	acc, parseErr := parseJSONFromOutput[AccountInfo](out)
	if parseErr == nil && acc.Success && acc.Email != "" {
		a.accountMu.Lock()
		a.cachedAccountEmail = acc.Email
		a.accountMu.Unlock()

		a.settingsMu.Lock()
		a.settings.DefaultDownloadDir = a.getDownloadsDir()
		a.settingsMu.Unlock()
	}
	return acc, parseErr
}

func (a *App) Login(email, password, authCode string) (LoginResult, error) {
	ctx := a.startCommandContext()
	args := []string{"auth", "login", "-e", email, "-p", password, "--format", "json"}
	if strings.TrimSpace(authCode) != "" {
		args = append(args, "--auth-code", strings.TrimSpace(authCode))
	}

	out, err := a.runIpaTool(ctx, args...)

	allOutput := out
	if err != nil {
		allOutput += "\n" + err.Error()
	}

	requires2FA := strings.Contains(strings.ToLower(allOutput), "2fa") ||
		strings.Contains(strings.ToLower(allOutput), "auth code is required") ||
		strings.Contains(strings.ToLower(allOutput), "auth-code")

	if requires2FA {
		return LoginResult{
			Success:      false,
			Requires2FA:  true,
			ErrorMessage: "需要 Apple ID 双重认证验证码",
		}, nil
	}

	acc, parseErr := parseJSONFromOutput[AccountInfo](out)
	if parseErr == nil && acc.Success {
		a.accountMu.Lock()
		a.cachedAccountEmail = acc.Email
		a.accountMu.Unlock()

		a.settingsMu.Lock()
		a.settings.DefaultDownloadDir = a.getDownloadsDir()
		a.settingsMu.Unlock()

		return LoginResult{
			Success: true,
			Account: acc,
		}, nil
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = "登录未成功"
	}

	return LoginResult{
		Success:      false,
		ErrorMessage: errMsg,
	}, nil
}

func (a *App) Revoke() (bool, error) {
	ctx := a.startCommandContext()
	out, err := a.runIpaTool(ctx, "auth", "revoke", "--format", "json")
	if err != nil {
		return false, err
	}
	a.accountMu.Lock()
	a.cachedAccountEmail = ""
	a.accountMu.Unlock()

	a.settingsMu.Lock()
	a.settings.DefaultDownloadDir = a.getDownloadsDir()
	a.settingsMu.Unlock()

	var res map[string]interface{}
	if err := json.Unmarshal([]byte(out), &res); err == nil {
		if s, ok := res["success"].(bool); ok {
			return s, nil
		}
	}
	return true, nil
}

func (a *App) ClearKeychainCache() error {
	dataDir := a.getDataDir()
	p := filepath.Join(dataDir, ".ipatool")
	if _, err := os.Stat(p); err == nil {
		_ = os.RemoveAll(p)
	}
	if home, err := os.UserHomeDir(); err == nil {
		oldP := filepath.Join(home, ".ipatool")
		_ = os.RemoveAll(oldP)
	}
	return nil
}

func (a *App) Search(term string, limit int, platform string) (SearchResult, error) {
	ctx := a.startCommandContext()
	args := []string{"search", term, "-l", fmt.Sprintf("%d", limit), "--format", "json"}
	if strings.TrimSpace(platform) != "" {
		args = append(args, "--platform", platform)
	}

	out, err := a.runIpaTool(ctx, args...)
	if err != nil {
		return SearchResult{}, err
	}

	res, err := parseJSONFromOutput[SearchResult](out)
	if err != nil {
		return SearchResult{}, err
	}

	for i := range res.Apps {
		if res.Apps[i].Price == 0 {
			res.Apps[i].DisplayPrice = "免费"
		} else {
			res.Apps[i].DisplayPrice = fmt.Sprintf("¥%.2f", res.Apps[i].Price)
		}
	}

	return res, nil
}

func (a *App) ListVersions(bundleId string, appId int64) (VersionsResult, error) {
	args := []string{"list-versions", "--format", "json"}
	if bundleId != "" {
		args = append(args, "-b", bundleId)
	} else if appId > 0 {
		args = append(args, "-i", fmt.Sprintf("%d", appId))
	} else {
		return VersionsResult{}, fmt.Errorf("必须提供 Bundle ID 或 App ID")
	}

	var out string
	var err error
	maxAttempts := 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		ctx := a.startCommandContext()
		out, err = a.runIpaTool(ctx, args...)
		if err == nil {
			break
		}
		errStr := err.Error()
		if (strings.Contains(errStr, "HTTP 500") || strings.Contains(errStr, "empty or non-plist body")) && attempt < maxAttempts-1 {
			a.emitLog(fmt.Sprintf("[网络重试] 查询版本列表遇到 Apple 偶发 500 响应，正在自动重试 (%d/%d)...", attempt+1, maxAttempts))
			time.Sleep(time.Duration(400*(attempt+1)) * time.Millisecond)
			continue
		}
		return VersionsResult{}, err
	}

	res, err := parseJSONFromOutput[VersionsResult](out)
	if err != nil {
		return VersionsResult{}, err
	}

	// Reverse to show newest version first
	for i, j := 0, len(res.ExternalVersionIdentifiers)-1; i < j; i, j = i+1, j-1 {
		res.ExternalVersionIdentifiers[i], res.ExternalVersionIdentifiers[j] = res.ExternalVersionIdentifiers[j], res.ExternalVersionIdentifiers[i]
	}

	return res, nil
}

func (a *App) GetVersionMetadata(bundleId, versionId string, appId int64) (VersionMetadataResult, error) {
	ctx := a.startCommandContext()
	args := []string{"get-version-metadata", "--external-version-id", versionId, "--format", "json"}
	if bundleId != "" {
		args = append(args, "-b", bundleId)
	} else if appId > 0 {
		args = append(args, "-i", fmt.Sprintf("%d", appId))
	}

	out, err := a.runIpaTool(ctx, args...)
	if err != nil {
		return VersionMetadataResult{}, err
	}

	res, err := parseJSONFromOutput[VersionMetadataResult](out)
	if err != nil {
		return VersionMetadataResult{}, err
	}

	res.DisplayFileSize = formatBytes(res.FileSize)
	if strings.Contains(res.ReleaseDate, "T") {
		res.ReleaseDate = strings.Split(res.ReleaseDate, "T")[0]
	}
	if strings.HasPrefix(res.ReleaseDate, "0001") || res.ReleaseDate == "" {
		res.ReleaseDate = "-"
	}
	return res, nil
}

func (a *App) Download(bundleId string, appId int64, versionId, outputPath, platform string, purchase bool) (DownloadResult, error) {
	ctx := a.startCommandContext()
	args := []string{"download", "--format", "json"}

	if bundleId != "" {
		args = append(args, "-b", bundleId)
	} else if appId > 0 {
		args = append(args, "-i", fmt.Sprintf("%d", appId))
	}

	if strings.TrimSpace(versionId) != "" {
		args = append(args, "--external-version-id", versionId)
	}
	if strings.TrimSpace(outputPath) != "" {
		args = append(args, "-o", outputPath)
	} else {
		args = append(args, "-o", a.getDownloadsDir())
	}
	if strings.TrimSpace(platform) != "" {
		args = append(args, "--platform", platform)
	}
	if purchase {
		args = append(args, "--purchase")
	}

	out, err := a.runIpaTool(ctx, args...)
	if err != nil {
		return DownloadResult{}, err
	}

	return parseJSONFromOutput[DownloadResult](out)
}

func (a *App) Purchase(bundleId string) (PurchaseResult, error) {
	ctx := a.startCommandContext()
	out, err := a.runIpaTool(ctx, "purchase", "-b", bundleId, "--format", "json")
	if err != nil {
		return PurchaseResult{}, err
	}
	return parseJSONFromOutput[PurchaseResult](out)
}

func (a *App) ListPurchases(page, limit int) (PurchasedResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	// ipatool 单次最多支持 100 条 (max results must not exceed 100)
	// 若请求超过 100 条 (如前端拉取名下全部已购记录)，按 100 每页自动分批拉取并合并
	if limit > 100 {
		var allApps []AppItem
		currentPage := 1
		totalCount := 0
		pageSize := 100

		for {
			ctx, cleanup := a.startPurchaseCommandContext()
			args := []string{"list-purchases", "-p", fmt.Sprintf("%d", currentPage), "-l", fmt.Sprintf("%d", pageSize), "--format", "json"}
			out, err := a.runIpaTool(ctx, args...)
			cleanup()
			if err != nil {
				return PurchasedResult{}, err
			}

			res, err := parseJSONFromOutput[PurchasedResult](out)
			if err != nil {
				return PurchasedResult{}, err
			}

			totalCount = res.TotalCount
			allApps = append(allApps, res.Apps...)

			if totalCount > 0 && len(allApps) >= totalCount {
				break
			}
			if len(allApps) >= limit || len(res.Apps) < pageSize {
				break
			}
			currentPage++
		}

		if len(allApps) > limit {
			allApps = allApps[:limit]
		}

		for i := range allApps {
			if allApps[i].Price == 0 {
				allApps[i].DisplayPrice = "免费"
			} else {
				allApps[i].DisplayPrice = fmt.Sprintf("¥%.2f", allApps[i].Price)
			}
		}

		return PurchasedResult{
			Count:      len(allApps),
			TotalCount: totalCount,
			Page:       1,
			Apps:       allApps,
		}, nil
	}

	ctx, cleanup := a.startPurchaseCommandContext()
	args := []string{"list-purchases", "-p", fmt.Sprintf("%d", page), "-l", fmt.Sprintf("%d", limit), "--format", "json"}

	out, err := a.runIpaTool(ctx, args...)
	cleanup()
	if err != nil {
		return PurchasedResult{}, err
	}

	res, err := parseJSONFromOutput[PurchasedResult](out)
	if err != nil {
		return PurchasedResult{}, err
	}

	for i := range res.Apps {
		if res.Apps[i].Price == 0 {
			res.Apps[i].DisplayPrice = "免费"
		} else {
			res.Apps[i].DisplayPrice = fmt.Sprintf("¥%.2f", res.Apps[i].Price)
		}
	}

	return res, nil
}

// --- Download Task Manager ---

func (a *App) AddDownloadTask(appName, bundleId string, appId int64, version, versionId, fileSize string) (*DownloadTask, error) {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	if strings.TrimSpace(appName) == "" {
		appName = bundleId
	}
	if strings.TrimSpace(version) == "" || version == "未查询" {
		version = "最新版"
	}

	task := &DownloadTask{
		ID:        id,
		AppName:   appName,
		BundleID:  bundleId,
		AppID:     appId,
		Version:   version,
		VersionID: versionId,
		FileSize:  fileSize,
		Status:    "pending",
		Speed:     "等待下载",
		Progress:  0,
		CreatedAt: time.Now().Format("15:04:05"),
	}

	a.tasks = append([]*DownloadTask{task}, a.tasks...)
	a.saveTasksLocked()
	go a.runDownloadTask(task)

	return task, nil
}

func (a *App) runDownloadTask(task *DownloadTask) {
	ctx, cancel := context.WithCancel(context.Background())
	a.tasksMu.Lock()
	a.taskCancels[task.ID] = cancel
	task.Status = "downloading"
	task.Speed = "连接中..."
	a.saveTasksLocked()
	a.tasksMu.Unlock()

	a.emitTaskUpdated(task)

	defer func() {
		a.tasksMu.Lock()
		delete(a.taskCancels, task.ID)
		a.tasksMu.Unlock()
	}()

	a.settingsMu.RLock()
	platform := a.settings.DefaultPlatform
	passphrase := a.settings.KeychainPassphrase
	enableProxy := a.settings.EnableProxy
	proxyUrl := a.settings.ProxyUrl
	a.settingsMu.RUnlock()

	outDir := a.getDownloadsDir()

	args := []string{"download", "-b", task.BundleID, "--purchase", "--format", "json", "--non-interactive"}
	if task.VersionID != "" {
		args = append(args, "--external-version-id", task.VersionID)
	}
	if outDir != "" {
		args = append(args, "-o", outDir)
	}
	if platform != "" {
		args = append(args, "--platform", platform)
	}
	if passphrase != "" {
		args = append(args, "--keychain-passphrase", passphrase)
	}

	exePath := a.resolveIpaToolPath()
	cmd := a.buildIpaToolCmd(ctx, exePath, args, enableProxy, proxyUrl)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		a.failTask(task, err.Error())
		return
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		a.failTask(task, err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		a.failTask(task, err.Error())
		return
	}

	var stdoutLines []string
	var stderrLines []string
	var wg sync.WaitGroup
	wg.Add(2)

	progressRegex := regexp.MustCompile(`"event":"download-progress","current":(\d+),"total":(\d+),"speed":(\d+)`)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutLines = append(stdoutLines, line)
			a.emitLog(line)

			if matches := progressRegex.FindStringSubmatch(line); len(matches) > 3 {
				curr, _ := strconv.ParseInt(matches[1], 10, 64)
				tot, _ := strconv.ParseInt(matches[2], 10, 64)
				spd, _ := strconv.ParseInt(matches[3], 10, 64)

				a.tasksMu.Lock()
				task.CurrBytes = curr
				task.TotalBytes = tot
				if tot > 0 {
					task.Progress = int(float64(curr) / float64(tot) * 100)
				}
				if spd > 0 {
					task.Speed = formatBytes(spd) + "/s"
				} else {
					task.Speed = "处理中..."
				}
				a.tasksMu.Unlock()

				a.emitTaskUpdated(task)
			}
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrLines = append(stderrLines, line)
			a.emitLog("[ERR] " + line)
		}
	}()

	wg.Wait()
	cmdErr := cmd.Wait()

	stdout := strings.Join(stdoutLines, "\n")
	stderr := strings.Join(stderrLines, "\n")

	if ctx.Err() != nil {
		a.tasksMu.Lock()
		task.Status = "canceled"
		task.Speed = "已取消"
		a.saveTasksLocked()
		a.tasksMu.Unlock()
		a.emitTaskUpdated(task)
		return
	}

	if cmdErr != nil {
		allOutput := stderr + "\n" + stdout
		errMsg := cmdErr.Error()
		for _, line := range strings.Split(allOutput, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
				var errMap map[string]interface{}
				if json.Unmarshal([]byte(line), &errMap) == nil {
					if e, ok := errMap["error"].(string); ok && e != "" {
						errMsg = e
						break
					}
				}
			}
		}
		if errMsg == cmdErr.Error() {
			re := regexp.MustCompile(`error="([^"]+)"`)
			if matches := re.FindStringSubmatch(allOutput); len(matches) > 1 {
				errMsg = matches[1]
			}
		}
		a.failTask(task, errMsg)
		return
	}

	res, parseErr := parseJSONFromOutput[DownloadResult](stdout)
	a.tasksMu.Lock()
	task.Status = "completed"
	task.Progress = 100
	task.Speed = "已完成"
	if parseErr == nil && res.Output != "" {
		task.OutputPath = res.Output
	}
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	a.emitTaskUpdated(task)
}

func (a *App) failTask(task *DownloadTask, msg string) {
	a.tasksMu.Lock()
	task.Status = "error"
	task.ErrorMessage = msg
	task.Speed = "下载失败"
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	a.emitTaskUpdated(task)
}

func (a *App) emitTaskUpdated(task *DownloadTask) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-task-updated", task)
	}
}

func (a *App) GetDownloadTasks() []*DownloadTask {
	a.tasksMu.RLock()
	defer a.tasksMu.RUnlock()
	res := make([]*DownloadTask, len(a.tasks))
	copy(res, a.tasks)
	return res
}

func (a *App) CancelDownloadTask(id string) {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()
	if cancel, ok := a.taskCancels[id]; ok {
		cancel()
	}
	for _, t := range a.tasks {
		if t.ID == id && (t.Status == "downloading" || t.Status == "pending") {
			t.Status = "canceled"
			t.Speed = "已取消"
			a.saveTasksLocked()
			a.emitTaskUpdated(t)
			break
		}
	}
}

func (a *App) DeleteDownloadTask(id string) {
	a.tasksMu.Lock()
	defer a.tasksMu.Unlock()
	if cancel, ok := a.taskCancels[id]; ok {
		cancel()
		delete(a.taskCancels, id)
	}
	for i, t := range a.tasks {
		if t.ID == id {
			a.tasks = append(a.tasks[:i], a.tasks[i+1:]...)
			break
		}
	}
	a.saveTasksLocked()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-tasks-reload")
	}
}

func (a *App) ClearCompletedDownloadTasks() {
	a.tasksMu.Lock()
	active := make([]*DownloadTask, 0)
	for _, t := range a.tasks {
		if t.Status == "downloading" || t.Status == "pending" {
			active = append(active, t)
		}
	}
	a.tasks = active
	a.saveTasksLocked()
	a.tasksMu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download-tasks-reload")
	}
}
