package backend

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
)

func (a *App) GetAccountInfo() (AccountInfo, error) {
	accountID := a.activeAccountID()
	if accountID == "" {
		return AccountInfo{}, nil
	}
	item, ok := a.accountByID(accountID)
	if !ok {
		return AccountInfo{}, nil
	}
	ctx, cleanup := a.startPurchaseCommandContext()
	defer cleanup()
	store, err := a.newAppStoreForAccount(ctx, accountID)
	if err != nil {
		return AccountInfo{}, err
	}
	result, err := store.AccountInfo()
	if err != nil {
		return AccountInfo{}, err
	}

	return a.upsertAccount(result.Account, item.ID, item.Legacy)
}

func (a *App) Login(email, password, authCode string) (output LoginResult, resultErr error) {
	a.addDiagnosticSecrets(email, password, authCode)
	a.writeDiagnostic("login.start", map[string]interface{}{"has_auth_code": strings.TrimSpace(authCode) != ""})
	defer func() {
		fields := map[string]interface{}{"success": output.Success, "requires_2fa": output.Requires2FA}
		if output.ErrorMessage != "" {
			fields["error"] = output.ErrorMessage
		}
		if resultErr != nil {
			fields["error"] = resultErr.Error()
		}
		a.writeDiagnostic("login.complete", fields)
	}()
	ctx := a.startCommandContext()
	store, storageID, legacy, err := a.newAppStoreForLogin(ctx, email)
	if err != nil {
		return LoginResult{ErrorMessage: err.Error()}, nil
	}

	result, err := store.Login(appstore.LoginInput{
		Email: email, Password: password, AuthCode: strings.TrimSpace(authCode),
	})
	if errors.Is(err, appstore.ErrAuthCodeRequired) {
		return LoginResult{Requires2FA: true, ErrorMessage: "Apple ID 或密码错误。若账号已开启双重认证且受信任设备收到了验证码，请在“双重认证验证码”栏中填入后重新登录"}, nil
	}
	if err != nil && authCode != "" && strings.Contains(err.Error(), "apple did not complete verification") {
		return LoginResult{ErrorMessage: "验证码错误或已过期，请在受信任设备上获取新的验证码后重新登录"}, nil
	}
	if err != nil {
		return LoginResult{ErrorMessage: authenticationErrorMessage(err)}, nil
	}

	account, err := a.upsertAccount(result.Account, storageID, legacy)
	if err != nil {
		return LoginResult{ErrorMessage: err.Error()}, nil
	}
	return LoginResult{Success: true, Account: account}, nil
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

func (a *App) ListVersions(bundleID string, appID int64) (output VersionsResult, resultErr error) {
	a.writeDiagnostic("history.start", map[string]interface{}{"bundle_id": bundleID, "app_id": appID})
	defer func() {
		fields := map[string]interface{}{"success": output.Success, "count": len(output.ExternalVersionIdentifiers)}
		if resultErr != nil {
			fields["error"] = resultErr.Error()
		}
		a.writeDiagnostic("history.complete", fields)
	}()
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
			return store.ListVersions(appstore.ListVersionsInput{Context: ctx, Account: account, App: item})
		})
		if err == nil {
			identifiers := result.ExternalVersionIdentifiers
			for left, right := 0, len(identifiers)-1; left < right; left, right = left+1, right-1 {
				identifiers[left], identifiers[right] = identifiers[right], identifiers[left]
			}
			return VersionsResult{BundleID: bundleID, ExternalVersionIdentifiers: identifiers, Success: true}, nil
		}

		lastErr = err
		a.writeDiagnostic("history.attempt_failed", map[string]interface{}{"attempt": attempt, "error": err.Error()})
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
	ctx, cleanup := a.startPurchaseCommandContext()
	defer cleanup()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return VersionMetadataResult{}, err
	}

	result, err := withAccountRetry(store, func(account appstore.Account) (appstore.GetVersionMetadataOutput, error) {
		item, err := resolveStoreApp(store, account, bundleID, appID, "")
		if err != nil {
			return appstore.GetVersionMetadataOutput{}, err
		}
		return store.GetVersionMetadata(appstore.GetVersionMetadataInput{Context: ctx, Account: account, App: item, VersionID: versionID})
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

func (a *App) Purchase(bundleID, platformValue string) (PurchaseResult, error) {
	resolvedPlatform := strings.TrimSpace(platformValue)
	if resolvedPlatform == "" {
		resolvedPlatform = a.currentDownloadPlatform()
	}
	platform, err := parseStorePlatform(resolvedPlatform)
	if err != nil {
		return PurchaseResult{}, err
	}
	ctx := a.startCommandContext()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return PurchaseResult{}, err
	}

	alreadyOwned := false
	_, err = withAccountRetry(store, func(account appstore.Account) (struct{}, error) {
		item, err := resolveStoreApp(store, account, bundleID, 0, platform)
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
	if limit > appstore.MaxOwnedAppsLimit {
		limit = appstore.MaxOwnedAppsLimit
	}
	ctx, cleanup := a.startPurchaseCommandContext()
	defer cleanup()
	store, err := a.newAppStore(ctx)
	if err != nil {
		return PurchasedResult{}, err
	}

	return listPurchasedPage(store, page, limit)
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
