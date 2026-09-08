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
}

func NewApp() *App {
	return &App{
		settings: Settings{
			KeychainPassphrase: "123456",
			DefaultDownloadDir: "D:\\Downloads",
			DefaultPlatform:    "iphone",
			EnableProxy:        true,
			ProxyUrl:           "http://127.0.0.1:10808",
		},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadSettings()
}

func (a *App) getSettingsFilePath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, ".config")
	}
	dir := filepath.Join(appData, "IPAToolGUI")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "settings.json")
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
	if a.settings.DefaultDownloadDir == "" {
		a.settings.DefaultDownloadDir = "D:\\Downloads"
	}
	if a.settings.DefaultPlatform == "" {
		a.settings.DefaultPlatform = "iphone"
	}

	a.settings.IpaToolPath = a.resolveIpaToolPath()
}

func (a *App) resolveIpaToolPath() string {
	// 1. Same directory as current executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		p := filepath.Join(dir, "ipatool.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 2. Current working directory
	if cwd, err := os.Getwd(); err == nil {
		p := filepath.Join(cwd, "ipatool.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 3. Known project directory
	devP := "D:\\Dev\\IPAToolGUI\\ipatool.exe"
	if _, err := os.Stat(devP); err == nil {
		return devP
	}

	return "ipatool.exe"
}

func (a *App) emitLog(message string) {
	if a.ctx != nil {
		timestamp := time.Now().Format("15:04:05")
		runtime.EventsEmit(a.ctx, "log", fmt.Sprintf("[%s] %s", timestamp, message))
	}
}

func (a *App) runIpaTool(ctx context.Context, args ...string) (string, error) {
	exePath := a.resolveIpaToolPath()
	if _, err := os.Stat(exePath); err != nil {
		errStr := fmt.Sprintf("未在程序同目录下找到 ipatool.exe！请确认存在于: %s", exePath)
		a.emitLog(errStr)
		return "", fmt.Errorf(errStr)
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

	cmd := exec.CommandContext(ctx, exePath, cmdArgs...)
	setSysProcAttr(cmd)

	// Configure Environment Variables
	cmd.Env = os.Environ()
	if enableProxy && strings.TrimSpace(proxyUrl) != "" {
		proxy := strings.TrimSpace(proxyUrl)
		cmd.Env = append(cmd.Env,
			"HTTP_PROXY="+proxy,
			"HTTPS_PROXY="+proxy,
			"ALL_PROXY="+proxy,
			"http_proxy="+proxy,
			"https_proxy="+proxy,
			"all_proxy="+proxy,
		)
		a.emitLog(fmt.Sprintf("[网络代理] 已启用: %s", proxy))
	}

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
		for scanner.Scan() {
			line := scanner.Text()
			stdoutLines = append(stdoutLines, line)
			a.emitLog(line)
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

	if cmdErr != nil {
		// Extract error from output
		allOutput := stderr + "\n" + stdout
		re := regexp.MustCompile(`error="([^"]+)"`)
		if matches := re.FindStringSubmatch(allOutput); len(matches) > 1 {
			return stdout, fmt.Errorf(matches[1])
		}
		if strings.TrimSpace(stderr) != "" {
			return stdout, fmt.Errorf(stderr)
		}
		return stdout, cmdErr
	}

	return stdout, nil
}

func parseJSONFromOutput[T any](output string) (T, error) {
	var zero T
	lines := strings.Split(output, "\n")
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
	s.IpaToolPath = a.resolveIpaToolPath()
	return s
}

func (a *App) SaveSettings(s Settings) error {
	a.settingsMu.Lock()
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
	defer a.cancelMu.Unlock()
	if a.cancelFn != nil {
		a.cancelFn()
		a.cancelFn = nil
		a.emitLog("用户取消了当前操作。")
	}
}

func (a *App) startCommandContext() context.Context {
	a.cancelMu.Lock()
	defer a.cancelMu.Unlock()
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
	return parseJSONFromOutput[AccountInfo](out)
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
	var res map[string]interface{}
	if err := json.Unmarshal([]byte(out), &res); err == nil {
		if s, ok := res["success"].(bool); ok {
			return s, nil
		}
	}
	return true, nil
}

func (a *App) ClearKeychainCache() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	p := filepath.Join(home, ".ipatool")
	if _, err := os.Stat(p); err == nil {
		return os.RemoveAll(p)
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
	ctx := a.startCommandContext()
	args := []string{"list-versions", "--format", "json"}
	if bundleId != "" {
		args = append(args, "-b", bundleId)
	} else if appId > 0 {
		args = append(args, "-i", fmt.Sprintf("%d", appId))
	} else {
		return VersionsResult{}, fmt.Errorf("必须提供 Bundle ID 或 App ID")
	}

	out, err := a.runIpaTool(ctx, args...)
	if err != nil {
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
	ctx := a.startCommandContext()
	args := []string{"list-purchases", "-p", fmt.Sprintf("%d", page), "-l", fmt.Sprintf("%d", limit), "--format", "json"}

	out, err := a.runIpaTool(ctx, args...)
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
