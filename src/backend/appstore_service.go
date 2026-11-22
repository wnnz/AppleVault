package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
	ipahttp "AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/keychain"
	"AppleVault/src/backend/internal/ipatool/util/machine"
	"AppleVault/src/backend/internal/ipatool/util/operatingsystem"
	"github.com/byteness/keyring"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/schollz/progressbar/v3"
)

const (
	ipatoolConfigDirectory = ".ipatool"
	ipatoolCookieFile      = "cookies"
	ipatoolKeychainService = "ipatool-auth.service"
	lookupCacheTTL         = 30 * time.Minute
	maxLookupCacheItems    = 512
)

type storeSessionKey struct {
	accountID    string
	legacy       bool
	dataDir      string
	passphrase   string
	enableProxy  bool
	proxyAddress string
}

type cachedLookupEntry struct {
	app       appstore.App
	expiresAt time.Time
}

// cachingKeychain avoids repeated PBKDF2 decryption and redundant writes on file keyrings.
type cachingKeychain struct {
	inner keychain.Keychain
	mu    sync.Mutex
	items map[string][]byte
}

func newCachingKeychain(inner keychain.Keychain) *cachingKeychain {
	return &cachingKeychain{
		inner: inner,
		items: make(map[string][]byte),
	}
}

func (c *cachingKeychain) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, ok := c.items[key]; ok {
		return bytes.Clone(cached), nil
	}

	data, err := c.inner.Get(key)
	if err != nil {
		return nil, err
	}

	c.items[key] = bytes.Clone(data)
	return data, nil
}

func (c *cachingKeychain) Set(key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.items[key]; ok && bytes.Equal(existing, data) {
		return nil
	}

	if err := c.inner.Set(key, data); err != nil {
		return err
	}

	c.items[key] = bytes.Clone(data)
	return nil
}

func (c *cachingKeychain) Remove(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return c.inner.Remove(key)
}

// localMachine keeps credentials and cookies inside AppleVault's data directory
// instead of depending on ipatool's command-line environment overrides.
type localMachine struct {
	home     string
	identity machine.Machine
}

func (m localMachine) MacAddress() (string, error) {
	return m.identity.MacAddress()
}

func (m localMachine) HomeDirectory() string { return m.home }

func (localMachine) ReadPassword(int) ([]byte, error) {
	return nil, errors.New("AppleVault 不支持命令行密码输入")
}

func (a *App) currentStoreSessionKey(accountID string, legacy bool) storeSessionKey {
	a.settingsMu.RLock()
	passphrase := a.settings.KeychainPassphrase
	enableProxy := a.settings.EnableProxy
	proxyAddress := strings.TrimSpace(a.settings.ProxyUrl)
	a.settingsMu.RUnlock()

	return storeSessionKey{
		accountID:    accountID,
		legacy:       legacy,
		dataDir:      a.getDataDir(),
		passphrase:   passphrase,
		enableProxy:  enableProxy,
		proxyAddress: proxyAddress,
	}
}

func (a *App) sharedAppStore() (appstore.AppStore, error) {
	accountID := a.activeAccountID()
	if accountID == "" {
		return nil, errors.New("尚未登录 Apple ID")
	}
	return a.sharedAppStoreFor(accountID, "当前账号不存在")
}

func (a *App) sharedAppStoreForAccount(accountID string) (appstore.AppStore, error) {
	return a.sharedAppStoreFor(accountID, "下载任务所属账号不存在")
}

func (a *App) sharedAppStoreFor(accountID, notFoundMessage string) (appstore.AppStore, error) {
	item, ok := a.accountByID(accountID)
	if !ok {
		return nil, errors.New(notFoundMessage)
	}

	key := a.currentStoreSessionKey(item.ID, item.Legacy)
	a.storeMu.Lock()
	if cached, exists := a.storeSessions[key]; exists {
		a.storeMu.Unlock()
		return cached, nil
	}
	a.storeMu.Unlock()

	store, err := a.newAppStoreWithStorage(context.Background(), item.ID, item.Legacy)
	if err != nil {
		return nil, err
	}

	a.storeMu.Lock()
	if a.storeSessions == nil {
		a.storeSessions = make(map[storeSessionKey]appstore.AppStore)
	}
	for existingKey, existingStore := range a.storeSessions {
		if existingKey.accountID == item.ID && existingKey != key {
			if closer, ok := existingStore.(interface{ CloseIdleConnections() }); ok {
				closer.CloseIdleConnections()
			}
			delete(a.storeSessions, existingKey)
		}
	}
	if existingStore, exists := a.storeSessions[key]; exists {
		a.storeMu.Unlock()
		if closer, ok := store.(interface{ CloseIdleConnections() }); ok {
			closer.CloseIdleConnections()
		}
		return existingStore, nil
	}
	a.storeSessions[key] = store
	a.storeMu.Unlock()
	return store, nil
}

func (a *App) invalidateStoreSessions(accountID string) {
	var evicted []appstore.AppStore

	a.storeMu.Lock()
	for key, store := range a.storeSessions {
		if accountID == "" || key.accountID == accountID {
			evicted = append(evicted, store)
			delete(a.storeSessions, key)
		}
	}
	a.storeMu.Unlock()

	for _, store := range evicted {
		if closer, ok := store.(interface{ CloseIdleConnections() }); ok {
			closer.CloseIdleConnections()
		}
	}
}

func (a *App) newAppStore(ctx context.Context) (appstore.AppStore, error) {
	accountID := a.activeAccountID()
	if accountID == "" {
		return nil, errors.New("尚未登录 Apple ID")
	}
	item, ok := a.accountByID(accountID)
	if !ok {
		return nil, errors.New("当前账号不存在")
	}
	return a.newAppStoreWithStorage(ctx, item.ID, item.Legacy)
}

func (a *App) newAppStoreForAccount(ctx context.Context, accountID string) (appstore.AppStore, error) {
	item, ok := a.accountByID(accountID)
	if !ok {
		return nil, errors.New("下载任务所属账号不存在")
	}
	return a.newAppStoreWithStorage(ctx, item.ID, item.Legacy)
}

func (a *App) newAppStoreForLogin(ctx context.Context, email string) (appstore.AppStore, string, bool, error) {
	if item, ok := a.accountForEmail(email); ok {
		store, err := a.newAppStoreWithStorage(ctx, item.ID, item.Legacy)
		return store, item.ID, item.Legacy, err
	}
	accountID := accountIDForEmail(email)
	store, err := a.newAppStoreWithStorage(ctx, accountID, false)
	return store, accountID, false, err
}

func (a *App) newAppStoreWithStorage(ctx context.Context, accountID string, legacy bool) (appstore.AppStore, error) {
	a.settingsMu.RLock()
	passphrase := a.settings.KeychainPassphrase
	enableProxy := a.settings.EnableProxy
	proxyAddress := strings.TrimSpace(a.settings.ProxyUrl)
	a.settingsMu.RUnlock()

	configDirectory, serviceName := a.accountStorage(accountID, legacy)
	if err := os.MkdirAll(configDirectory, 0700); err != nil {
		return nil, fmt.Errorf("创建 App Store 配置目录失败: %w", err)
	}

	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(configDirectory, ipatoolCookieFile),
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 Cookie 存储失败: %w", err)
	}

	ring, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.SecretServiceBackend,
			keyring.FileBackend,
		},
		ServiceName: serviceName,
		FileDir:     configDirectory,
		FilePasswordFunc: func(string) (string, error) {
			if passphrase == "" {
				return "", errors.New("钥匙串密码不能为空")
			}
			return passphrase, nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("初始化凭据存储失败: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if enableProxy && proxyAddress != "" {
		proxyURL, err := url.Parse(proxyAddress)
		if err != nil {
			return nil, fmt.Errorf("代理地址格式无效: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		a.emitLog(fmt.Sprintf("[网络代理] 已启用: %s", ipahttp.DiagnosticURL(proxyAddress)))
	} else {
		transport.Proxy = nil
	}

	osAdapter := operatingsystem.New()
	a.writeDiagnostic("store.config", map[string]interface{}{"account_id": accountID, "proxy_enabled": enableProxy && proxyAddress != "", "proxy_endpoint": ipahttp.DiagnosticURL(proxyAddress), "identity_strategy": "ipatool-2.8.1"})
	store := appstore.NewAppStore(appstore.Args{
		CookieJar:       ipahttp.CookieJar(jar),
		Keychain:        newCachingKeychain(keychain.New(keychain.Args{Keyring: ring, Label: serviceName})),
		Machine:         localMachine{home: a.getDataDir(), identity: machine.New(machine.Args{OS: osAdapter})},
		OperatingSystem: osAdapter,
		Context:         ctx,
		Transport:       transport,
	})
	if info, err := store.AccountInfo(); err == nil {
		a.addDiagnosticSecrets(info.Account.Email, info.Account.Password, info.Account.PasswordToken, info.Account.DirectoryServicesID)
	}
	return store, nil
}

func authenticationErrorMessage(err error) string {
	if errors.Is(err, appstore.ErrAuthenticationUnavailable) {
		status := ""
		var responseErr *ipahttp.UnexpectedResponseError
		if errors.As(err, &responseErr) {
			status = fmt.Sprintf("（HTTP %d）", responseErr.StatusCode)
		}
		return fmt.Sprintf("Apple 登录接口返回异常响应%s。请切换代理节点或网络后重试，或稍后再试。", status)
	}
	return err.Error()
}

func refreshAccount(store appstore.AppStore, account appstore.Account) (appstore.Account, error) {
	if strings.TrimSpace(account.Email) == "" || account.Password == "" {
		return appstore.Account{}, errors.New("缺少重新认证所需的凭据，请在账号中心重新登录")
	}
	result, err := store.Login(appstore.LoginInput{
		Email:    account.Email,
		Password: account.Password,
	})
	if err != nil {
		if errors.Is(err, appstore.ErrAuthCodeRequired) {
			return appstore.Account{}, fmt.Errorf("需要 Apple ID 双重认证，请在账号中心重新登录并输入验证码: %w", err)
		}
		if errors.Is(err, appstore.ErrAuthenticationUnavailable) {
			return appstore.Account{}, fmt.Errorf("Apple ID 重新认证失败，请切换代理节点或网络后重试，或稍后再试: %w", err)
		}
		return appstore.Account{}, fmt.Errorf("Apple ID 重新认证失败: %w", err)
	}
	return result.Account, nil
}

func withAccountRetry[T any](store appstore.AppStore, operation func(appstore.Account) (T, error)) (T, error) {
	var zero T
	info, err := store.AccountInfo()
	if err != nil {
		return zero, err
	}

	result, err := operation(info.Account)
	if !errors.Is(err, appstore.ErrPasswordTokenExpired) {
		return result, err
	}

	account, loginErr := refreshAccount(store, info.Account)
	if loginErr != nil {
		return zero, fmt.Errorf("Apple ID 登录已失效: %w", loginErr)
	}
	return operation(account)
}

func lookupBundleCacheKey(storeFront string, platform appstore.Platform, bundleID string) string {
	return fmt.Sprintf("b|%s|%s|%s", strings.ToLower(strings.TrimSpace(storeFront)), string(platform), strings.ToLower(strings.TrimSpace(bundleID)))
}

func lookupIDCacheKey(storeFront string, platform appstore.Platform, appID int64) string {
	return fmt.Sprintf("i|%s|%s|%d", strings.ToLower(strings.TrimSpace(storeFront)), string(platform), appID)
}

func (a *App) rememberLookupApp(storeFront string, platform appstore.Platform, item appstore.App) {
	if item.ID <= 0 {
		return
	}

	now := time.Now()
	expiresAt := now.Add(lookupCacheTTL)

	a.lookupMu.Lock()
	defer a.lookupMu.Unlock()

	if a.lookupCache == nil {
		a.lookupCache = make(map[string]cachedLookupEntry)
	}
	if len(a.lookupCache) > maxLookupCacheItems {
		for k, entry := range a.lookupCache {
			if !now.Before(entry.expiresAt) {
				delete(a.lookupCache, k)
			}
		}
		if len(a.lookupCache) > maxLookupCacheItems {
			clear(a.lookupCache)
		}
	}

	entry := cachedLookupEntry{app: item, expiresAt: expiresAt}
	a.lookupCache[lookupIDCacheKey(storeFront, platform, item.ID)] = entry
	if trimmedBundle := strings.TrimSpace(item.BundleID); trimmedBundle != "" {
		a.lookupCache[lookupBundleCacheKey(storeFront, platform, trimmedBundle)] = entry
	}
}

func (a *App) resolveStoreAppCached(
	ctx context.Context,
	store appstore.AppStore,
	account appstore.Account,
	bundleID string,
	appID int64,
	platform appstore.Platform,
	requireFullMetadata bool,
) (appstore.App, error) {
	trimmedBundleID := strings.TrimSpace(bundleID)
	if trimmedBundleID == "" && appID <= 0 {
		return appstore.App{}, errors.New("必须提供 Bundle ID 或 App ID")
	}

	now := time.Now()
	a.lookupMu.RLock()
	if trimmedBundleID != "" {
		if entry, ok := a.lookupCache[lookupBundleCacheKey(account.StoreFront, platform, trimmedBundleID)]; ok && now.Before(entry.expiresAt) {
			if appID <= 0 || entry.app.ID == appID {
				a.lookupMu.RUnlock()
				return entry.app, nil
			}
		}
	} else if appID > 0 {
		if entry, ok := a.lookupCache[lookupIDCacheKey(account.StoreFront, platform, appID)]; ok && now.Before(entry.expiresAt) {
			a.lookupMu.RUnlock()
			return entry.app, nil
		}
	}
	a.lookupMu.RUnlock()

	if trimmedBundleID == "" {
		return appstore.App{ID: appID}, nil
	}

	if !requireFullMetadata && appID > 0 {
		return appstore.App{ID: appID, BundleID: trimmedBundleID}, nil
	}

	result, err := store.Lookup(appstore.LookupInput{
		Context:  ctx,
		Account:  account,
		AppID:    appID,
		BundleID: trimmedBundleID,
		Platform: platform,
	})
	if err != nil {
		return appstore.App{}, err
	}

	a.rememberLookupApp(account.StoreFront, platform, result.App)
	return result.App, nil
}

func parseStorePlatform(value string) (appstore.Platform, error) {
	return appstore.ParsePlatform(strings.TrimSpace(value))
}

func mapStoreApp(item appstore.App) AppItem {
	purchaseDate := ""
	if !item.PurchaseDate.IsZero() {
		purchaseDate = item.PurchaseDate.Format(time.RFC3339)
	}
	displayPrice := "免费"
	if item.Price != 0 {
		displayPrice = fmt.Sprintf("¥%.2f", item.Price)
	}
	return AppItem{
		ID:           item.ID,
		BundleID:     item.BundleID,
		Name:         item.Name,
		Version:      item.Version,
		Price:        item.Price,
		PurchaseDate: purchaseDate,
		DisplayPrice: displayPrice,
	}
}

func mapStoreApps(items []appstore.App) []AppItem {
	result := make([]AppItem, len(items))
	for index, item := range items {
		result[index] = mapStoreApp(item)
	}
	return result
}

type downloadProgress struct {
	Percent int
	Current int64
	Total   int64
	Speed   int64
}

func (a *App) downloadFromStore(ctx context.Context, accountID, bundleID string, appID int64, versionID, outputPath, platformValue string, acquireLicense bool, onProgress func(downloadProgress)) (DownloadResult, error) {
	store, err := a.sharedAppStoreForAccount(accountID)
	if err != nil {
		return DownloadResult{}, err
	}
	platform, err := parseStorePlatform(platformValue)
	if err != nil {
		return DownloadResult{}, err
	}

	bar := progressbar.NewOptions64(1,
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetRenderBlankState(false),
	)
	done := make(chan struct{})
	var wg sync.WaitGroup
	if onProgress != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					state := bar.State()
					var total int64
					if state.CurrentPercent > 0 && !math.IsInf(state.CurrentPercent, 0) && state.CurrentBytes > 0 {
						total = int64(math.Round(state.CurrentBytes / state.CurrentPercent))
					}
					onProgress(downloadProgress{
						Percent: int(state.CurrentPercent * 100),
						Current: int64(state.CurrentBytes),
						Total:   total,
						Speed:   int64(state.KBsPerSecond * 1024),
					})
				case <-done:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	defer func() {
		close(done)
		wg.Wait()
	}()

	purchased := false
	result, err := withAccountRetry(store, func(account appstore.Account) (appstore.DownloadOutput, error) {
		item, err := a.resolveStoreAppCached(ctx, store, account, bundleID, appID, platform, true)
		if err != nil {
			return appstore.DownloadOutput{}, err
		}

		output, err := store.Download(appstore.DownloadInput{
			Context: ctx, Account: account, App: item, OutputPath: outputPath,
			Progress: bar, ExternalVersionID: versionID, Platform: platform,
		})
		if errors.Is(err, appstore.ErrLicenseRequired) && acquireLicense {
			purchaseErr := store.Purchase(appstore.PurchaseInput{Context: ctx, Account: account, App: item, Platform: platform})
			if purchaseErr != nil && !errors.Is(purchaseErr, appstore.ErrLicenseAlreadyExists) {
				return appstore.DownloadOutput{}, purchaseErr
			}
			purchased = !errors.Is(purchaseErr, appstore.ErrLicenseAlreadyExists)
			output, err = store.Download(appstore.DownloadInput{
				Context: ctx, Account: account, App: item, OutputPath: outputPath,
				Progress: bar, ExternalVersionID: versionID, Platform: platform,
			})
		}
		return output, err
	})
	if err != nil {
		return DownloadResult{}, err
	}
	if err := store.ReplicateSinf(appstore.ReplicateSinfInput{Sinfs: result.Sinfs, PackagePath: result.DestinationPath}); err != nil {
		return DownloadResult{}, err
	}
	return DownloadResult{Output: result.DestinationPath, Purchased: purchased, Success: true}, nil
}
