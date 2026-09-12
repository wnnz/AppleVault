package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx             context.Context
	dataDirOverride string
	settings        Settings
	settingsMu      sync.RWMutex
	cancelMu        sync.Mutex
	cancelFn        context.CancelFunc

	purchaseCancelMu     sync.Mutex
	purchaseCancelers    map[uint64]context.CancelFunc
	nextPurchaseCancelID uint64

	tasksMu     sync.RWMutex
	tasks       []*DownloadTask
	taskCancels map[string]context.CancelFunc

	accountMu          sync.RWMutex
	accounts           accountRegistry
	cachedAccountID    string
	cachedAccountEmail string
}

func NewApp() *App {
	app := &App{
		settings: Settings{
			DefaultPlatform: "iphone",
			EnableProxy:     true,
			ProxyUrl:        "http://127.0.0.1:10808",
		},
		tasks:             make([]*DownloadTask, 0),
		taskCancels:       make(map[string]context.CancelFunc),
		purchaseCancelers: make(map[uint64]context.CancelFunc),
	}
	app.settings.DefaultDownloadDir = app.getDownloadsDir()
	return app
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.ensureDataMigration()
	a.loadSettings()
	a.initializeAccounts()
	a.settingsMu.Lock()
	a.settings.DefaultDownloadDir = a.getDownloadsDir()
	a.settingsMu.Unlock()
	a.loadTasks()
	initWindowIcon()
}

func (a *App) getDataDir() string {
	if a.dataDirOverride != "" {
		return a.dataDirOverride
	}
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

func (a *App) getDownloadsDirForAccount(accountID string) string {
	a.accountMu.RLock()
	email := ""
	for _, item := range a.accounts.Accounts {
		if item.ID == accountID {
			email = item.Email
			break
		}
	}
	a.accountMu.RUnlock()
	accountDir := sanitizeAccountDir(email)
	if accountDir == "default" && accountID != "" {
		accountDir = accountID
	}
	downloadsDir := filepath.Join(a.getDataDir(), "downloads", accountDir)
	_ = os.MkdirAll(downloadsDir, 0755)
	return downloadsDir
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

	a.settings.IpaToolPath = "内置 App Store 服务"
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

func (a *App) emitLog(message string) {
	if a.ctx != nil {
		timestamp := time.Now().Format("15:04:05")
		runtime.EventsEmit(a.ctx, "log", fmt.Sprintf("[%s] %s", timestamp, message))
	}
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
	s.IpaToolPath = "内置 App Store 服务"
	return s
}

func (a *App) SaveSettings(s Settings) error {
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	// The keychain passphrase is managed explicitly from Account Center. Preserve it
	// when other settings are saved so a stale frontend form cannot invalidate credentials.
	s.KeychainPassphrase = a.settings.KeychainPassphrase
	s.DefaultDownloadDir = a.getDownloadsDir()
	a.settings = s
	a.settings.IpaToolPath = "内置 App Store 服务"
	data, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.getSettingsFilePath(), data, 0644)
}

func (a *App) SetKeychainPassphrase(passphrase string) error {
	if passphrase == "" {
		return fmt.Errorf("钥匙串密码不能为空")
	}
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	a.settings.KeychainPassphrase = passphrase
	data, err := json.MarshalIndent(a.settings, "", "  ")
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
