package backend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
	ipahttp "AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/keychain"
	"AppleVault/src/backend/internal/ipatool/util/operatingsystem"
	"github.com/byteness/keyring"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/schollz/progressbar/v3"
)

const (
	ipatoolConfigDirectory = ".ipatool"
	ipatoolCookieFile      = "cookies"
	ipatoolKeychainService = "ipatool-auth.service"
)

// localMachine keeps credentials and cookies inside AppleVault's data directory
// instead of depending on ipatool's command-line environment overrides.
type localMachine struct {
	home string
}

func (localMachine) MacAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("读取网络接口失败: %w", err)
	}
	for _, item := range interfaces {
		if address := item.HardwareAddr.String(); address != "" {
			return address, nil
		}
	}
	return "", errors.New("未找到有效的网卡地址")
}

func (m localMachine) HomeDirectory() string { return m.home }

func (localMachine) ReadPassword(int) ([]byte, error) {
	return nil, errors.New("AppleVault 不支持命令行密码输入")
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
		a.emitLog(fmt.Sprintf("[网络代理] 已启用: %s", proxyAddress))
	} else {
		transport.Proxy = nil
	}

	osAdapter := operatingsystem.New()
	return appstore.NewAppStore(appstore.Args{
		CookieJar:       ipahttp.CookieJar(jar),
		Keychain:        keychain.New(keychain.Args{Keyring: ring, Label: serviceName}),
		Machine:         localMachine{home: a.getDataDir()},
		OperatingSystem: osAdapter,
		Context:         ctx,
		Transport:       transport,
	}), nil
}

func refreshAccount(store appstore.AppStore, account appstore.Account) (appstore.Account, error) {
	result, err := store.Login(appstore.LoginInput{
		Email:    account.Email,
		Password: account.Password,
	})
	if err != nil {
		return appstore.Account{}, err
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
		return zero, loginErr
	}
	return operation(account)
}

func resolveStoreApp(store appstore.AppStore, account appstore.Account, bundleID string, appID int64, platform appstore.Platform) (appstore.App, error) {
	if bundleID == "" {
		if appID <= 0 {
			return appstore.App{}, errors.New("必须提供 Bundle ID 或 App ID")
		}
		return appstore.App{ID: appID}, nil
	}
	result, err := store.Lookup(appstore.LookupInput{Account: account, BundleID: bundleID, Platform: platform})
	if err != nil {
		return appstore.App{}, err
	}
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
	Speed   int64
}

func (a *App) downloadFromStore(ctx context.Context, accountID, bundleID string, appID int64, versionID, outputPath, platformValue string, acquireLicense bool, onProgress func(downloadProgress)) (DownloadResult, error) {
	store, err := a.newAppStoreForAccount(ctx, accountID)
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
	if onProgress != nil {
		go func() {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					state := bar.State()
					onProgress(downloadProgress{
						Percent: int(state.CurrentPercent * 100),
						Current: int64(state.CurrentBytes),
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
	defer close(done)

	purchased := false
	result, err := withAccountRetry(store, func(account appstore.Account) (appstore.DownloadOutput, error) {
		item, err := resolveStoreApp(store, account, bundleID, appID, platform)
		if err != nil {
			return appstore.DownloadOutput{}, err
		}

		output, err := store.Download(appstore.DownloadInput{
			Account: account, App: item, OutputPath: outputPath,
			Progress: bar, ExternalVersionID: versionID, Platform: platform,
		})
		if errors.Is(err, appstore.ErrLicenseRequired) && acquireLicense {
			purchaseErr := store.Purchase(appstore.PurchaseInput{Account: account, App: item})
			if purchaseErr != nil && !errors.Is(purchaseErr, appstore.ErrLicenseAlreadyExists) {
				return appstore.DownloadOutput{}, purchaseErr
			}
			purchased = !errors.Is(purchaseErr, appstore.ErrLicenseAlreadyExists)
			output, err = store.Download(appstore.DownloadInput{
				Account: account, App: item, OutputPath: outputPath,
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
