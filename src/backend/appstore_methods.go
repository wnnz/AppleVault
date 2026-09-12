package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
)

func (a *App) GetAccountInfo() (AccountInfo, error) {
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return AccountInfo{}, err
	}
	result, err := store.AccountInfo()
	if err != nil {
		return AccountInfo{}, err
	}

	account := AccountInfo{Name: result.Account.Name, Email: result.Account.Email, Success: true}
	a.rememberAccount(account.Email)
	return account, nil
}

func (a *App) Login(email, password, authCode string) (LoginResult, error) {
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return LoginResult{ErrorMessage: err.Error()}, nil
	}

	result, err := store.Login(appstore.LoginInput{
		Email: email, Password: password, AuthCode: strings.TrimSpace(authCode),
	})
	if errors.Is(err, appstore.ErrAuthCodeRequired) {
		return LoginResult{Requires2FA: true, ErrorMessage: "需要 Apple ID 双重认证验证码"}, nil
	}
	if err != nil {
		return LoginResult{ErrorMessage: err.Error()}, nil
	}

	account := AccountInfo{Name: result.Account.Name, Email: result.Account.Email, Success: true}
	a.rememberAccount(account.Email)
	return LoginResult{Success: true, Account: account}, nil
}

func (a *App) Revoke() (bool, error) {
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return false, err
	}
	if err := store.Revoke(); err != nil {
		return false, err
	}
	a.rememberAccount("")
	return true, nil
}

func (a *App) ClearKeychainCache() error {
	configDirectory := filepath.Join(a.getDataDir(), ipatoolConfigDirectory)
	if err := os.RemoveAll(configDirectory); err != nil {
		return fmt.Errorf("清理凭据缓存失败: %w", err)
	}
	if home, err := os.UserHomeDir(); err == nil {
		legacyDirectory := filepath.Join(home, ipatoolConfigDirectory)
		if legacyDirectory != configDirectory {
			_ = os.RemoveAll(legacyDirectory)
		}
	}
	a.rememberAccount("")
	return nil
}

func (a *App) rememberAccount(email string) {
	a.accountMu.Lock()
	a.cachedAccountEmail = email
	a.accountMu.Unlock()

	a.settingsMu.Lock()
	a.settings.DefaultDownloadDir = a.getDownloadsDir()
	a.settingsMu.Unlock()
}

func (a *App) Search(term string, limit int, platformValue string) (SearchResult, error) {
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return SearchResult{}, err
	}
	platform, err := parseStorePlatform(platformValue)
	if err != nil {
		return SearchResult{}, err
	}

	result, err := withAccountRetry(store, func(account appstore.Account) (appstore.SearchOutput, error) {
		return store.Search(appstore.SearchInput{Account: account, Term: term, Limit: int64(limit), Platform: platform})
	})
	if err != nil {
		return SearchResult{}, err
	}
	return SearchResult{Count: result.Count, Apps: mapStoreApps(result.Results)}, nil
}

func (a *App) ListVersions(bundleID string, appID int64) (VersionsResult, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		ctx := a.startCommandContext()
		store, err := a.newAppStore(ctx)
		if err != nil {
			return VersionsResult{}, err
		}

		result, err := withAccountRetry(store, func(account appstore.Account) (appstore.ListVersionsOutput, error) {
			item, err := resolveStoreApp(store, account, bundleID, appID, "")
			if err != nil {
				return appstore.ListVersionsOutput{}, err
			}
			return store.ListVersions(appstore.ListVersionsInput{Account: account, App: item})
		})
		if err == nil {
			identifiers := result.ExternalVersionIdentifiers
			for left, right := 0, len(identifiers)-1; left < right; left, right = left+1, right-1 {
				identifiers[left], identifiers[right] = identifiers[right], identifiers[left]
			}
			return VersionsResult{BundleID: bundleID, ExternalVersionIdentifiers: identifiers, Success: true}, nil
		}

		lastErr = err
		if ctx.Err() != nil || !isRetryableAppleResponse(err) || attempt == 3 {
			break
		}
		a.emitLog(fmt.Sprintf("[网络重试] 查询版本列表响应异常，正在自动重试 (%d/3)...", attempt))
		time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
	}
	return VersionsResult{}, lastErr
}

func isRetryableAppleResponse(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "http 500") ||
		strings.Contains(message, "empty or non-plist body") ||
		strings.Contains(message, "invalid response")
}

func (a *App) GetVersionMetadata(bundleID, versionID string, appID int64) (VersionMetadataResult, error) {
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return VersionMetadataResult{}, err
	}

	result, err := withAccountRetry(store, func(account appstore.Account) (appstore.GetVersionMetadataOutput, error) {
		item, err := resolveStoreApp(store, account, bundleID, appID, "")
		if err != nil {
			return appstore.GetVersionMetadataOutput{}, err
		}
		return store.GetVersionMetadata(appstore.GetVersionMetadataInput{Account: account, App: item, VersionID: versionID})
	})
	if err != nil {
		return VersionMetadataResult{}, err
	}

	releaseDate := "-"
	if !result.ReleaseDate.IsZero() {
		releaseDate = result.ReleaseDate.Format("2006-01-02")
	}
	return VersionMetadataResult{
		ExternalVersionID: versionID,
		DisplayVersion:    result.DisplayVersion,
		ReleaseDate:       releaseDate,
		FileSize:          result.FileSize,
		DisplayFileSize:   formatBytes(result.FileSize),
		Success:           true,
	}, nil
}

func (a *App) Download(bundleID string, appID int64, versionID, outputPath, platform string, purchase bool) (DownloadResult, error) {
	ctx := a.startCommandContext()
	if strings.TrimSpace(outputPath) == "" {
		outputPath = a.getDownloadsDir()
	}
	return a.downloadFromStore(ctx, bundleID, appID, versionID, outputPath, platform, purchase, nil)
}

func (a *App) Purchase(bundleID string) (PurchaseResult, error) {
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return PurchaseResult{}, err
	}

	alreadyOwned := false
	_, err = withAccountRetry(store, func(account appstore.Account) (struct{}, error) {
		item, err := resolveStoreApp(store, account, bundleID, 0, "")
		if err != nil {
			return struct{}{}, err
		}
		err = store.Purchase(appstore.PurchaseInput{Account: account, App: item})
		if errors.Is(err, appstore.ErrLicenseAlreadyExists) {
			alreadyOwned = true
			return struct{}{}, nil
		}
		return struct{}{}, err
	})
	if err != nil {
		return PurchaseResult{}, err
	}
	return PurchaseResult{AlreadyOwned: alreadyOwned, Success: true}, nil
}

func (a *App) ListPurchases(page, limit int) (PurchasedResult, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	ctx, cleanup := a.startPurchaseCommandContext()
	defer cleanup()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return PurchasedResult{}, err
	}

	if limit <= appstore.MaxOwnedAppsLimit {
		return listPurchasedPage(store, page, limit)
	}

	allApps := make([]AppItem, 0, limit)
	totalCount := 0
	for currentPage := 1; len(allApps) < limit; currentPage++ {
		result, err := listPurchasedPage(store, currentPage, appstore.MaxOwnedAppsLimit)
		if err != nil {
			return PurchasedResult{}, err
		}
		totalCount = result.TotalCount
		allApps = append(allApps, result.Apps...)
		if len(result.Apps) < appstore.MaxOwnedAppsLimit || (totalCount > 0 && len(allApps) >= totalCount) {
			break
		}
	}
	if len(allApps) > limit {
		allApps = allApps[:limit]
	}
	return PurchasedResult{Count: len(allApps), TotalCount: totalCount, Page: 1, Apps: allApps}, nil
}

func listPurchasedPage(store appstore.AppStore, page, limit int) (PurchasedResult, error) {
	result, err := withAccountRetry(store, func(account appstore.Account) (appstore.OwnedAppsOutput, error) {
		return store.OwnedApps(appstore.OwnedAppsInput{Account: account, Page: page, Limit: limit})
	})
	if err != nil {
		return PurchasedResult{}, err
	}
	return PurchasedResult{
		Count: result.Count, TotalCount: result.TotalCount, Page: result.Page, Apps: mapStoreApps(result.Results),
	}, nil
}
