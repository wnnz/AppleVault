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
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                     context.Context
	dataDirOverride         string
	settings                Settings
	undecryptablePassphrase string
	settingsMu              sync.RWMutex
	cancelMu                sync.Mutex
	cancelFn                context.CancelFunc

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
	diagnosticMu       sync.Mutex
	diagnosticSecrets  []string

	storeMu            sync.Mutex
	storeSessions      map[storeSessionKey]appstore.AppStore
	lookupMu           sync.RWMutex
	lookupCache        map[string]cachedLookupEntry
	versionMetaMu      sync.Mutex
	versionMetaLoaded  bool
	versionMetaDataDir string
	versionMetaEntries map[string]cachedVersionMetadata
}

func NewApp() *App {
	app := &App{
		settings: Settings{
			DefaultPlatform: "iphone",
			EnableProxy:     true,
			ProxyUrl:        "http://127.0.0.1:10808",
		},
		tasks:              make([]*DownloadTask, 0),
		taskCancels:        make(map[string]context.CancelFunc),
		purchaseCancelers:  make(map[uint64]context.CancelFunc),
		storeSessions:      make(map[storeSessionKey]appstore.AppStore),
		lookupCache:        make(map[string]cachedLookupEntry),
		versionMetaEntries: make(map[string]cachedVersionMetadata),
	}
	app.settings.DefaultDownloadDir = app.getDownloadsDir()
	return app
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.ensureDataMigration()
	a.loadSettings()
	a.cleanupDiagnosticLogs(14*24*time.Hour, time.Now())
	a.initializeAccounts()
	a.settingsMu.Lock()
	a.settings.DefaultDownloadDir = a.getDownloadsDir()
	a.settingsMu.Unlock()
	a.loadTasks()
	a.versionMetaMu.Lock()
	a.ensureVersionMetadataCacheLoadedLocked()
	a.versionMetaMu.Unlock()
	a.writeDiagnostic("app.start", map[string]interface{}{"version": AppVersion, "go_version": goruntime.Version(), "http_proxy_environment": os.Getenv("HTTP_PROXY") != "", "https_proxy_environment": os.Getenv("HTTPS_PROXY") != ""})
	initWindowIcon()
}

func (a *App) getDataDir() string {
	if a.dataDirOverride != "" {
		return a.dataDirOverride
	}
	if goruntime.GOOS == "darwin" {
		return a.getDarwinDataDir()
	}
	// 1. 检查程序所在目录下的 data (release 模式)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		p := filepath.Join(dir, "data")
		if !isTemporaryExecutableDir(dir) {
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

func (a *App) getDarwinDataDir() string {
	exe, exeErr := os.Executable()
	inAppBundle := exeErr == nil && strings.Contains(filepath.ToSlash(exe), ".app/Contents/MacOS/")
	if inAppBundle {
		bundleParent := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(exe))))
		if !isTemporaryExecutableDir(exe) && bundleParent != "/Applications" {
			portableData := filepath.Join(bundleParent, "data")
			if fi, err := os.Stat(portableData); err == nil && fi.IsDir() {
				return portableData
			}
		}
	} else if exeErr == nil && !isTemporaryExecutableDir(exe) {
		exeData := filepath.Join(filepath.Dir(exe), "data")
		if fi, err := os.Stat(exeData); err == nil && fi.IsDir() {
			return exeData
		}
	}

	if !inAppBundle {
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
	}

	if configDir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(configDir) != "" {
		dataDir := filepath.Join(configDir, "AppleVault")
		_ = os.MkdirAll(dataDir, 0700)
		return dataDir
	}

	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		dataDir := filepath.Join(home, "Library", "Application Support", "AppleVault")
		_ = os.MkdirAll(dataDir, 0700)
		return dataDir
	}

	cwd, _ := os.Getwd()
	dataDir := filepath.Join(cwd, "data")
	_ = os.MkdirAll(dataDir, 0755)
	return dataDir
}

func isTemporaryExecutableDir(dir string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(dir), "/") {
		lower := strings.ToLower(segment)
		if lower == "temp" || lower == "tmp" || lower == "apptranslocation" || strings.HasPrefix(lower, "go-build") {
			return true
		}
	}
	return false
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

type persistedSettings struct {
	Settings
	ProtectedKeychainPassphrase string `json:"protectedKeychainPassphrase,omitempty"`
}

func (a *App) loadSettings() {
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()

	needsMigration := false
	p := a.getSettingsFilePath()
	if data, err := os.ReadFile(p); err == nil {
		var file persistedSettings
		if unmarshalErr := json.Unmarshal(data, &file); unmarshalErr == nil {
			s := file.Settings
			if file.ProtectedKeychainPassphrase != "" {
				plain, decryptErr := unprotectSecret(file.ProtectedKeychainPassphrase)
				if decryptErr == nil {
					s.KeychainPassphrase = plain
					a.undecryptablePassphrase = ""
				} else {
					s.KeychainPassphrase = ""
					a.undecryptablePassphrase = file.ProtectedKeychainPassphrase
					a.writeDiagnostic("settings.passphrase_unprotect_failed", map[string]interface{}{"error": decryptErr.Error()})
					a.emitLog("钥匙串密码无法在本机解密（数据目录可能来自其他电脑或用户），请在账号中心重新输入")
				}
			} else if s.KeychainPassphrase != "" {
				needsMigration = true
				a.undecryptablePassphrase = ""
			} else {
				a.undecryptablePassphrase = ""
			}
			a.settings = s
		} else {
			backupPath, backupErr := backupCorruptFile(p)
			fields := map[string]interface{}{"backup": backupPath, "error": unmarshalErr.Error()}
			if backupErr != nil {
				fields["backup_error"] = backupErr.Error()
			}
			a.writeDiagnostic("settings.corrupt", fields)
			if backupErr == nil {
				a.emitLog(fmt.Sprintf("设置文件损坏，已备份为 %s，已使用默认设置", filepath.Base(backupPath)))
			} else {
				a.emitLog("设置文件损坏，已使用默认设置")
			}
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
	if needsMigration {
		_ = a.saveSettingsLocked()
	}
	a.invalidateStoreSessions("")
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
	if unmarshalErr := json.Unmarshal(data, &list); unmarshalErr == nil {
		valid := make([]*DownloadTask, 0, len(list))
		for _, t := range list {
			if t == nil {
				continue
			}
			if t.Status == "downloading" || t.Status == "pending" {
				t.Status = "canceled"
				t.Speed = "已中断"
			}
			valid = append(valid, t)
		}
		a.tasks = valid
	} else {
		backupPath, backupErr := backupCorruptFile(p)
		fields := map[string]interface{}{"backup": backupPath, "error": unmarshalErr.Error()}
		if backupErr != nil {
			fields["backup_error"] = backupErr.Error()
		}
		a.writeDiagnostic("tasks.corrupt", fields)
		if backupErr == nil {
			a.emitLog(fmt.Sprintf("下载任务文件损坏，已备份为 %s", filepath.Base(backupPath)))
		} else {
			a.emitLog("下载任务文件损坏，已跳过加载")
		}
	}
}

func (a *App) saveTasksLocked() {
	p := a.getTasksFilePath()
	data, err := json.MarshalIndent(a.tasks, "", "  ")
	if err == nil {
		_ = writeFileAtomic(p, data, 0600)
	}
}

func (a *App) saveSettingsLocked() error {
	file := persistedSettings{Settings: a.settings}
	if a.settings.KeychainPassphrase != "" {
		if protected, err := protectSecret(a.settings.KeychainPassphrase); err == nil {
			file.KeychainPassphrase = ""
			file.ProtectedKeychainPassphrase = protected
		}
	} else if a.undecryptablePassphrase != "" {
		file.ProtectedKeychainPassphrase = a.undecryptablePassphrase
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(a.getSettingsFilePath(), data, 0600)
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
	units := []string{"KB", "MB", "GB", "TB"}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit && exp < len(units)-1; n /= unit {
		div *= int64(unit)
		exp++
	}
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), units[exp])
}

// --- Bound Methods for Wails Frontend ---

func (a *App) GetSettings() Settings {
	a.settingsMu.RLock()
	defer a.settingsMu.RUnlock()
	s := a.settings
	s.DefaultDownloadDir = a.getDownloadsDir()
	return s
}

func (a *App) SaveSettings(s Settings) error {
	a.settingsMu.Lock()
	// The keychain passphrase is managed explicitly from Account Center. Preserve it
	// when other settings are saved so a stale frontend form cannot invalidate credentials.
	s.KeychainPassphrase = a.settings.KeychainPassphrase
	s.ThemeHintShown = a.settings.ThemeHintShown
	s.DefaultDownloadDir = a.getDownloadsDir()
	a.settings = s
	err := a.saveSettingsLocked()
	a.settingsMu.Unlock()
	a.invalidateStoreSessions("")
	return err
}

func (a *App) GetThemeHintShown() bool {
	a.settingsMu.RLock()
	defer a.settingsMu.RUnlock()
	return a.settings.ThemeHintShown
}

func (a *App) MarkThemeHintShown() error {
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	a.settings.ThemeHintShown = true
	return a.saveSettingsLocked()
}

func (a *App) SetKeychainPassphrase(passphrase string) error {
	if passphrase == "" {
		return fmt.Errorf("钥匙串密码不能为空")
	}
	a.settingsMu.Lock()
	a.settings.KeychainPassphrase = passphrase
	a.undecryptablePassphrase = ""
	err := a.saveSettingsLocked()
	a.settingsMu.Unlock()
	a.invalidateStoreSessions("")
	return err
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

	fi, statErr := os.Stat(targetPath)
	switch goruntime.GOOS {
	case "darwin":
		if statErr == nil {
			if fi.IsDir() {
				return exec.Command("open", targetPath).Start()
			}
			return exec.Command("open", "-R", targetPath).Start()
		}
		return exec.Command("open", filepath.Dir(targetPath)).Start()
	case "windows":
		if statErr == nil {
			if fi.IsDir() {
				return exec.Command("explorer", targetPath).Start()
			}
			return exec.Command("explorer", "/select,"+targetPath).Start()
		}
		return exec.Command("explorer", filepath.Dir(targetPath)).Start()
	default:
		dir := targetPath
		if statErr != nil || !fi.IsDir() {
			dir = filepath.Dir(targetPath)
		}
		return exec.Command("xdg-open", dir).Start()
	}
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
