package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
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

	installMu     sync.Mutex
	installing    bool
	installCancel context.CancelFunc

	tokenRefreshMu  sync.Mutex
	tokenRefreshing map[string]*tokenRefreshCall

	tasksMu      sync.RWMutex
	tasks        []*DownloadTask
	taskCancels  map[string]context.CancelFunc
	taskDone     map[string]chan struct{}
	taskSeq      atomic.Uint64
	downloadSem  chan struct{}
	shuttingDown bool

	accountMu          sync.RWMutex
	accounts           accountRegistry
	cachedAccountID    string
	cachedAccountEmail string
	diagnosticMu       sync.Mutex
	diagnosticSecrets  []string
	recentLogsMu       sync.Mutex
	recentLogs         []string

	storeMu                  sync.Mutex
	storeSessions            map[storeSessionKey]appstore.AppStore
	lookupMu                 sync.RWMutex
	lookupCache              map[string]cachedLookupEntry
	versionMetaMu            sync.Mutex
	versionMetaLoaded        bool
	versionMetaDataDir       string
	versionMetaEntries       map[string]cachedVersionMetadata
	versionMetaBundleToAppID map[string]int64
	versionMetaDirty         bool
	versionMetaFlushTimer    *time.Timer
	versionMetaFlushDelay    time.Duration
	versionMetaWriteMu       sync.Mutex
}

func NewApp() *App {
	app := &App{
		settings: Settings{
			DefaultPlatform: "iphone",
			EnableProxy:     false,
			ProxyUrl:        "http://127.0.0.1:10808",
		},
		tasks:                    make([]*DownloadTask, 0),
		taskCancels:              make(map[string]context.CancelFunc),
		taskDone:                 make(map[string]chan struct{}),
		downloadSem:              make(chan struct{}, maxConcurrentDownloads),
		purchaseCancelers:        make(map[uint64]context.CancelFunc),
		tokenRefreshing:          make(map[string]*tokenRefreshCall),
		storeSessions:            make(map[storeSessionKey]appstore.AppStore),
		lookupCache:              make(map[string]cachedLookupEntry),
		versionMetaEntries:       make(map[string]cachedVersionMetadata),
		versionMetaBundleToAppID: make(map[string]int64),
		versionMetaFlushDelay:    defaultVersionMetadataFlushDelay,
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
	normalized := strings.ReplaceAll(dir, `\`, "/")
	for _, segment := range strings.Split(normalized, "/") {
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
				if secretProtectionSupported() {
					needsMigration = true
				}
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
		file, protected := a.preparePersistedSettingsLocked()
		if protected {
			if data, err := json.MarshalIndent(file, "", "  "); err == nil {
				_ = writeFileAtomic(a.getSettingsFilePath(), data, 0600)
			}
		}
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

func (a *App) preparePersistedSettingsLocked() (persistedSettings, bool) {
	file := persistedSettings{Settings: a.settings}
	if a.settings.KeychainPassphrase != "" {
		if protected, err := protectSecret(a.settings.KeychainPassphrase); err == nil && protected != "" {
			file.KeychainPassphrase = ""
			file.ProtectedKeychainPassphrase = protected
			return file, true
		}
	} else if a.undecryptablePassphrase != "" {
		file.ProtectedKeychainPassphrase = a.undecryptablePassphrase
	}
	return file, false
}

func (a *App) saveSettingsLocked() error {
	file, _ := a.preparePersistedSettingsLocked()
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(a.getSettingsFilePath(), data, 0600)
}

const maxRecentLogs = 100

func (a *App) emitLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("[%s] %s", timestamp, message)

	a.recentLogsMu.Lock()
	a.recentLogs = append(a.recentLogs, formatted)
	if len(a.recentLogs) > maxRecentLogs {
		trimmed := make([]string, maxRecentLogs)
		copy(trimmed, a.recentLogs[len(a.recentLogs)-maxRecentLogs:])
		a.recentLogs = trimmed
	}
	a.recentLogsMu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "log", formatted)
	}
}

func (a *App) GetRecentLogs() []string {
	a.recentLogsMu.Lock()
	defer a.recentLogsMu.Unlock()
	if len(a.recentLogs) == 0 {
		return []string{}
	}
	out := make([]string, len(a.recentLogs))
	copy(out, a.recentLogs)
	return out
}

func (a *App) GetRuntimeInfo() RuntimeInfo {
	return RuntimeInfo{
		OS:                 goruntime.GOOS,
		UsesSystemKeychain: goruntime.GOOS == "darwin",
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
	if s.EnableProxy && strings.TrimSpace(s.ProxyUrl) != "" {
		if _, err := parseAndValidateProxyURL(s.ProxyUrl); err != nil {
			return err
		}
	}
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

func parseAndValidateProxyURL(raw string) (*url.URL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("代理地址不能为空")
	}
	pURL, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("代理地址格式无效: %w", err)
	}
	switch strings.ToLower(pURL.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return nil, fmt.Errorf("代理地址格式无效: 不支持的协议 %q（支持 http/https/socks5）", pURL.Scheme)
	}
	if strings.TrimSpace(pURL.Hostname()) == "" {
		return nil, fmt.Errorf("代理地址格式无效: 缺少主机名（示例: http://127.0.0.1:10808）")
	}
	return pURL, nil
}

func (a *App) TestProxy(proxyUrl string) ProxyTestResult {
	pURL, err := parseAndValidateProxyURL(proxyUrl)
	if err != nil {
		return ProxyTestResult{Success: false, Message: err.Error()}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyURL(pURL)
	defer transport.CloseIdleConnections()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://apple.com", nil)
	if err != nil {
		return ProxyTestResult{Success: false, Message: fmt.Sprintf("创建测试请求失败: %v", err)}
	}

	resp, err := (&http.Client{Transport: transport, Timeout: 6 * time.Second}).Do(req)
	if err != nil {
		return ProxyTestResult{Success: false, Message: fmt.Sprintf("连接失败: %v", err)}
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 16<<10))
	_ = resp.Body.Close()

	return ProxyTestResult{Success: true, Message: fmt.Sprintf("连接成功！HTTP 响应状态码: %s", resp.Status)}
}

func (a *App) OpenInExplorer(targetPath string) error {
	if strings.TrimSpace(targetPath) == "" {
		targetPath = a.getDownloadsDir()
	}

	fi, statErr := os.Stat(targetPath)
	switch goruntime.GOOS {
	case "darwin":
		var cmd *exec.Cmd
		if statErr == nil {
			if fi.IsDir() {
				cmd = exec.Command("open", targetPath)
			} else {
				cmd = exec.Command("open", "-R", targetPath)
			}
		} else {
			cmd = exec.Command("open", filepath.Dir(targetPath))
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		go func() { _ = cmd.Wait() }()
		return nil
	case "windows":
		isDir := statErr == nil && fi.IsDir()
		return openExplorerWindows(targetPath, statErr, isDir)
	default:
		dir := targetPath
		if statErr != nil || !fi.IsDir() {
			dir = filepath.Dir(targetPath)
		}
		cmd := exec.Command("xdg-open", dir)
		if err := cmd.Start(); err != nil {
			return err
		}
		go func() { _ = cmd.Wait() }()
		return nil
	}
}

func (a *App) CancelRunningCommand() {
	if a.cancelStoreCommands() {
		a.emitLog("用户取消了当前操作。")
	}
}

func (a *App) cancelStoreCommands() bool {
	a.cancelMu.Lock()
	cancel := a.cancelFn
	a.cancelFn = nil
	a.cancelMu.Unlock()

	if cancel != nil {
		cancel()
	}
	purchaseCancelled := a.cancelPurchaseCommands()
	return cancel != nil || purchaseCancelled
}

func (a *App) startInstallCommandContext() (context.Context, func(), error) {
	a.installMu.Lock()
	defer a.installMu.Unlock()
	if a.installing {
		return nil, nil, fmt.Errorf("已有设备安装任务正在进行中，请等待完成或先取消")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	a.installing = true
	a.installCancel = cancel
	cleaned := false
	cleanup := func() {
		cancel()
		a.installMu.Lock()
		if !cleaned {
			cleaned = true
			a.installing = false
			a.installCancel = nil
		}
		a.installMu.Unlock()
	}
	return ctx, cleanup, nil
}

func (a *App) cancelInstallCommand() bool {
	a.installMu.Lock()
	cancel := a.installCancel
	a.installMu.Unlock()
	if cancel != nil {
		cancel()
		return true
	}
	return false
}

func (a *App) CancelInstallIPA() {
	if a.cancelInstallCommand() {
		a.emitLog("用户取消了 IPA 安装。")
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
