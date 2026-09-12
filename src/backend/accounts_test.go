package backend

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"AppleVault/src/backend/internal/ipatool/appstore"
)

func TestAccountIDNormalizesEmail(t *testing.T) {
	first := accountIDForEmail(" User@Example.COM ")
	second := accountIDForEmail("user@example.com")
	if first != second || len(first) != 32 {
		t.Fatalf("expected stable normalized account ID, got %q and %q", first, second)
	}
	if first == accountIDForEmail("other@example.com") {
		t.Fatal("different accounts must have different IDs")
	}
}

func TestRegionFromStoreFront(t *testing.T) {
	if got := regionFromStoreFront("143465-2,34"); got != "CN" {
		t.Fatalf("expected CN, got %q", got)
	}
	if got := regionFromStoreFront("unknown"); got != "" {
		t.Fatalf("unknown storefront should be empty, got %q", got)
	}
}

func TestLegacyAccountRegistration(t *testing.T) {
	registry := registryFromLegacyAccount(appstore.Account{
		Name: "Legacy", Email: "legacy@example.com", StoreFront: "143441-1,29",
	}, 42)
	if registry.ActiveAccountID != legacyAccountID || len(registry.Accounts) != 1 {
		t.Fatalf("unexpected legacy registry: %#v", registry)
	}
	item := registry.Accounts[0]
	if !item.Legacy || item.Region != "US" || item.LastUsed != 42 {
		t.Fatalf("unexpected legacy account: %#v", item)
	}
}

func TestAccountsOrderAndFallback(t *testing.T) {
	app := &App{accounts: accountRegistry{
		ActiveAccountID: "older",
		Accounts: []storedAccount{
			{ID: "newer", Email: "new@example.com", LastUsed: 20},
			{ID: "older", Email: "old@example.com", LastUsed: 10},
		},
	}}
	accounts := app.GetAccounts()
	if len(accounts) != 2 || accounts[0].ID != "older" || !accounts[0].Active || accounts[1].ID != "newer" {
		t.Fatalf("unexpected account order: %#v", accounts)
	}
	if got := mostRecentlyUsedAccountID(app.accounts.Accounts); got != "newer" {
		t.Fatalf("expected newer fallback, got %q", got)
	}
}

func TestAccountRegistryPersistenceAndStorageIsolation(t *testing.T) {
	app := &App{dataDirOverride: t.TempDir(), accounts: accountRegistry{
		ActiveAccountID: "one",
		Accounts:        []storedAccount{{ID: "one", Email: "one@example.com", LastUsed: 1}},
	}}
	app.accountMu.Lock()
	if err := app.saveAccountsLocked(); err != nil {
		app.accountMu.Unlock()
		t.Fatal(err)
	}
	app.accountMu.Unlock()
	app.accountMu.Lock()
	app.accounts.Accounts[0].Name = "Updated"
	if err := app.saveAccountsLocked(); err != nil {
		app.accountMu.Unlock()
		t.Fatal(err)
	}
	app.accountMu.Unlock()
	data, err := json.Marshal(app.accounts)
	if err != nil || len(data) == 0 {
		t.Fatal("registry should marshal")
	}
	loaded := &App{dataDirOverride: app.dataDirOverride}
	loaded.initializeAccounts()
	if loaded.activeAccountID() != "one" || len(loaded.GetAccounts()) != 1 || loaded.GetAccounts()[0].Name != "Updated" {
		t.Fatalf("unexpected loaded registry: %#v", loaded.accounts)
	}

	firstDir, firstService := app.accountStorage("one", false)
	secondDir, secondService := app.accountStorage("two", false)
	if firstDir == secondDir || firstService == secondService {
		t.Fatal("account credential storage must be isolated")
	}
	if filepath.Base(firstDir) != "one" {
		t.Fatalf("unexpected account directory: %q", firstDir)
	}
}

func TestDownloadTaskAccountIDIsBackwardCompatible(t *testing.T) {
	created := newDownloadTask("account-one", "new", "Example", "com.example", 1, "1.0", "100", "1 MB")
	if created.AccountID != "account-one" {
		t.Fatalf("new task should bind its account, got %q", created.AccountID)
	}

	var legacy DownloadTask
	if err := json.Unmarshal([]byte(`{"id":"legacy","status":"completed"}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.AccountID != "" {
		t.Fatalf("legacy task should have no account ID, got %q", legacy.AccountID)
	}
}
