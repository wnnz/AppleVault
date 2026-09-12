package backend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
)

const (
	accountsRegistryFile = "accounts.json"
	legacyAccountID      = "legacy"
)

type storedAccount struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Region   string `json:"region,omitempty"`
	Legacy   bool   `json:"legacy,omitempty"`
	LastUsed int64  `json:"lastUsed"`
}

type accountRegistry struct {
	ActiveAccountID string          `json:"activeAccountID,omitempty"`
	Accounts        []storedAccount `json:"accounts"`
}

func normalizeAccountEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func accountIDForEmail(email string) string {
	sum := sha256.Sum256([]byte(normalizeAccountEmail(email)))
	return hex.EncodeToString(sum[:16])
}

func regionFromStoreFront(storeFront string) string {
	region, err := appstore.CountryCodeFromStoreFront(storeFront)
	if err != nil {
		return ""
	}
	return region
}

func registryFromLegacyAccount(account appstore.Account, now int64) accountRegistry {
	if account.Email == "" {
		return accountRegistry{Accounts: make([]storedAccount, 0)}
	}
	return accountRegistry{
		ActiveAccountID: legacyAccountID,
		Accounts: []storedAccount{{
			ID: legacyAccountID, Name: account.Name, Email: account.Email,
			Region: regionFromStoreFront(account.StoreFront), Legacy: true, LastUsed: now,
		}},
	}
}

func mostRecentlyUsedAccountID(accounts []storedAccount) string {
	var selected string
	var mostRecent int64
	for _, item := range accounts {
		if selected == "" || item.LastUsed > mostRecent {
			selected = item.ID
			mostRecent = item.LastUsed
		}
	}
	return selected
}

func (a *App) accountsRegistryPath() string {
	return filepath.Join(a.getDataDir(), accountsRegistryFile)
}

func (a *App) initializeAccounts() {
	path := a.accountsRegistryPath()
	if data, err := os.ReadFile(path); err == nil {
		var registry accountRegistry
		if json.Unmarshal(data, &registry) == nil {
			a.accountMu.Lock()
			a.accounts = registry
			a.syncCachedAccountLocked()
			a.accountMu.Unlock()
			return
		}
	}

	registry := accountRegistry{Accounts: make([]storedAccount, 0)}
	store, err := a.newAppStoreWithStorage(context.Background(), legacyAccountID, true)
	if err == nil {
		if result, infoErr := store.AccountInfo(); infoErr == nil && result.Account.Email != "" {
			registry = registryFromLegacyAccount(result.Account, time.Now().UnixNano())
		}
	}

	a.accountMu.Lock()
	a.accounts = registry
	a.syncCachedAccountLocked()
	_ = a.saveAccountsLocked()
	a.accountMu.Unlock()
}

func (a *App) saveAccountsLocked() error {
	data, err := json.MarshalIndent(a.accounts, "", "  ")
	if err != nil {
		return err
	}
	path := a.accountsRegistryPath()
	tmp, err := os.CreateTemp(filepath.Dir(path), "accounts-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err == nil {
		return nil
	}
	// Windows does not replace an existing destination with os.Rename.
	backupPath := path + ".bak"
	_ = os.Remove(backupPath)
	if _, statErr := os.Stat(path); statErr == nil {
		if renameErr := os.Rename(path, backupPath); renameErr != nil {
			return err
		}
	}
	if renameErr := os.Rename(tmpPath, path); renameErr != nil {
		_ = os.Rename(backupPath, path)
		return renameErr
	}
	_ = os.Remove(backupPath)
	return nil
}

func (a *App) syncCachedAccountLocked() {
	a.cachedAccountID = ""
	a.cachedAccountEmail = ""
	for _, item := range a.accounts.Accounts {
		if item.ID == a.accounts.ActiveAccountID {
			a.cachedAccountID = item.ID
			a.cachedAccountEmail = item.Email
			return
		}
	}
	a.accounts.ActiveAccountID = ""
}

func (a *App) activeAccountID() string {
	a.accountMu.RLock()
	defer a.accountMu.RUnlock()
	return a.cachedAccountID
}

func (a *App) accountByID(accountID string) (storedAccount, bool) {
	a.accountMu.RLock()
	defer a.accountMu.RUnlock()
	for _, item := range a.accounts.Accounts {
		if item.ID == accountID {
			return item, true
		}
	}
	return storedAccount{}, false
}

func (a *App) accountForEmail(email string) (storedAccount, bool) {
	normalized := normalizeAccountEmail(email)
	a.accountMu.RLock()
	defer a.accountMu.RUnlock()
	for _, item := range a.accounts.Accounts {
		if normalizeAccountEmail(item.Email) == normalized {
			return item, true
		}
	}
	return storedAccount{}, false
}

func accountInfo(item storedAccount, activeID string) AccountInfo {
	return AccountInfo{
		ID: item.ID, Name: item.Name, Email: item.Email, Region: item.Region,
		Active: item.ID == activeID, Success: true,
	}
}

func (a *App) GetAccounts() []AccountInfo {
	a.accountMu.RLock()
	activeID := a.accounts.ActiveAccountID
	items := append([]storedAccount(nil), a.accounts.Accounts...)
	a.accountMu.RUnlock()
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ID == activeID {
			return true
		}
		if items[j].ID == activeID {
			return false
		}
		return items[i].LastUsed > items[j].LastUsed
	})
	result := make([]AccountInfo, len(items))
	for index, item := range items {
		result[index] = accountInfo(item, activeID)
	}
	return result
}

func (a *App) RefreshAccount(accountID string) (AccountInfo, error) {
	item, ok := a.accountByID(accountID)
	if !ok {
		return AccountInfo{}, errors.New("账号不存在")
	}
	store, err := a.newAppStoreWithStorage(context.Background(), item.ID, item.Legacy)
	if err != nil {
		return AccountInfo{}, err
	}
	storedInfo, err := store.AccountInfo()
	if err != nil {
		return AccountInfo{}, fmt.Errorf("读取账号凭据失败: %w", err)
	}

	a.accountMu.Lock()
	defer a.accountMu.Unlock()
	for index := range a.accounts.Accounts {
		if a.accounts.Accounts[index].ID != accountID {
			continue
		}
		a.accounts.Accounts[index].Name = storedInfo.Account.Name
		a.accounts.Accounts[index].Email = storedInfo.Account.Email
		a.accounts.Accounts[index].Region = regionFromStoreFront(storedInfo.Account.StoreFront)
		if err := a.saveAccountsLocked(); err != nil {
			return AccountInfo{}, err
		}
		return accountInfo(a.accounts.Accounts[index], a.accounts.ActiveAccountID), nil
	}
	return AccountInfo{}, errors.New("账号不存在")
}

func (a *App) upsertAccount(account appstore.Account, storageID string, legacy bool) (AccountInfo, error) {
	region := regionFromStoreFront(account.StoreFront)
	a.accountMu.Lock()
	defer a.accountMu.Unlock()
	index := -1
	for i, item := range a.accounts.Accounts {
		if normalizeAccountEmail(item.Email) == normalizeAccountEmail(account.Email) {
			index = i
			break
		}
	}
	if index >= 0 {
		storageID = a.accounts.Accounts[index].ID
		legacy = a.accounts.Accounts[index].Legacy
		a.accounts.Accounts[index].Name = account.Name
		a.accounts.Accounts[index].Email = account.Email
		a.accounts.Accounts[index].Region = region
		a.accounts.Accounts[index].LastUsed = time.Now().UnixNano()
	} else {
		a.accounts.Accounts = append(a.accounts.Accounts, storedAccount{
			ID: storageID, Name: account.Name, Email: account.Email, Region: region,
			Legacy: legacy, LastUsed: time.Now().UnixNano(),
		})
	}
	a.accounts.ActiveAccountID = storageID
	a.syncCachedAccountLocked()
	if err := a.saveAccountsLocked(); err != nil {
		return AccountInfo{}, err
	}
	return accountInfo(storedAccount{ID: storageID, Name: account.Name, Email: account.Email, Region: region}, storageID), nil
}

func (a *App) SwitchAccount(accountID string) (AccountInfo, error) {
	item, ok := a.accountByID(accountID)
	if !ok {
		return AccountInfo{}, errors.New("账号不存在")
	}
	store, err := a.newAppStoreWithStorage(context.Background(), item.ID, item.Legacy)
	if err != nil {
		return AccountInfo{}, err
	}
	storedInfo, err := store.AccountInfo()
	if err != nil {
		return AccountInfo{}, fmt.Errorf("读取账号凭据失败: %w", err)
	}

	a.CancelRunningCommand()
	a.accountMu.Lock()
	defer a.accountMu.Unlock()
	for index := range a.accounts.Accounts {
		if a.accounts.Accounts[index].ID != accountID {
			continue
		}
		a.accounts.Accounts[index].Name = storedInfo.Account.Name
		a.accounts.Accounts[index].Email = storedInfo.Account.Email
		a.accounts.Accounts[index].Region = regionFromStoreFront(storedInfo.Account.StoreFront)
		a.accounts.Accounts[index].LastUsed = time.Now().UnixNano()
		a.accounts.ActiveAccountID = accountID
		a.syncCachedAccountLocked()
		if err := a.saveAccountsLocked(); err != nil {
			return AccountInfo{}, err
		}
		return accountInfo(a.accounts.Accounts[index], accountID), nil
	}
	return AccountInfo{}, errors.New("账号不存在")
}

func (a *App) RemoveAccount(accountID string) (AccountInfo, error) {
	item, ok := a.accountByID(accountID)
	if !ok {
		return AccountInfo{}, errors.New("账号不存在")
	}
	store, err := a.newAppStoreWithStorage(context.Background(), item.ID, item.Legacy)
	if err != nil {
		return AccountInfo{}, err
	}
	if err = store.Revoke(); err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return AccountInfo{}, err
	}
	configDir, _ := a.accountStorage(item.ID, item.Legacy)
	if item.Legacy {
		_ = os.Remove(filepath.Join(configDir, ipatoolCookieFile))
	} else {
		_ = os.RemoveAll(configDir)
	}

	a.CancelRunningCommand()
	a.accountMu.Lock()
	filtered := make([]storedAccount, 0, len(a.accounts.Accounts)-1)
	for _, candidate := range a.accounts.Accounts {
		if candidate.ID != accountID {
			filtered = append(filtered, candidate)
		}
	}
	a.accounts.Accounts = filtered
	if a.accounts.ActiveAccountID == accountID {
		a.accounts.ActiveAccountID = mostRecentlyUsedAccountID(filtered)
	}
	a.syncCachedAccountLocked()
	err = a.saveAccountsLocked()
	activeID := a.accounts.ActiveAccountID
	var active storedAccount
	for _, candidate := range a.accounts.Accounts {
		if candidate.ID == activeID {
			active = candidate
			break
		}
	}
	a.accountMu.Unlock()
	if err != nil {
		return AccountInfo{}, err
	}
	if activeID == "" {
		return AccountInfo{}, nil
	}
	return accountInfo(active, activeID), nil
}

func (a *App) accountStorage(accountID string, legacy bool) (string, string) {
	if legacy {
		return filepath.Join(a.getDataDir(), ipatoolConfigDirectory), ipatoolKeychainService
	}
	return filepath.Join(a.getDataDir(), ipatoolConfigDirectory, "accounts", accountID),
		fmt.Sprintf("%s.%s", ipatoolKeychainService, accountID)
}
