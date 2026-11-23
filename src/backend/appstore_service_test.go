package backend

import (
	"context"
	"errors"
	"fmt"
	gohttp "net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
	ipahttp "AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/util/machine"
)

type localIdentityTestMachine struct {
	machine.Machine
	calls int
}

func (m *localIdentityTestMachine) MacAddress() (string, error) {
	m.calls++
	return "02:01:02:03:04:05", nil
}

func TestLocalMachineDelegatesIdentityAndKeepsAppStorage(t *testing.T) {
	identity := &localIdentityTestMachine{}
	adapter := localMachine{home: t.TempDir(), identity: identity}
	for range 2 {
		address, err := adapter.MacAddress()
		if err != nil || address != "02:01:02:03:04:05" {
			t.Fatalf("identity = %q, %v", address, err)
		}
	}
	if identity.calls != 2 || adapter.HomeDirectory() != adapter.home {
		t.Fatal("local adapter must retain the same identity provider and app data directory")
	}
	if _, err := adapter.ReadPassword(0); err == nil {
		t.Fatal("desktop password input must remain noninteractive")
	}
}

func TestMapStoreApp(t *testing.T) {
	purchasedAt := time.Date(2026, time.September, 12, 8, 30, 0, 0, time.UTC)
	result := mapStoreApp(appstore.App{
		ID: 123, BundleID: "com.example.app", Name: "Example", Version: "2.0",
		Price: 1.5, Currency: "CNY", PurchaseDate: purchasedAt,
	})

	if result.ID != 123 || result.BundleID != "com.example.app" || result.DisplayPrice != "¥1.50" {
		t.Fatalf("unexpected mapped app: %#v", result)
	}
	if result.PurchaseDate != purchasedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected purchase date: %q", result.PurchaseDate)
	}

	usd := mapStoreApp(appstore.App{Price: 2.99, Currency: "USD"})
	if usd.DisplayPrice != "$2.99" {
		t.Fatalf("expected $2.99 for USD, got %q", usd.DisplayPrice)
	}

	formatted := mapStoreApp(appstore.App{Price: 18, Currency: "CNY", FormattedPrice: "HK$18.00"})
	if formatted.DisplayPrice != "HK$18.00" {
		t.Fatalf("expected FormattedPrice to take precedence, got %q", formatted.DisplayPrice)
	}

	free := mapStoreApp(appstore.App{Price: 0, Currency: "USD"})
	if free.DisplayPrice != "免费" {
		t.Fatalf("expected 免费 for price 0, got %q", free.DisplayPrice)
	}

	plain := mapStoreApp(appstore.App{Price: 1.5})
	if plain.DisplayPrice != "1.50" {
		t.Fatalf("expected 1.50 when currency omitted, got %q", plain.DisplayPrice)
	}
}

func TestParseStorePlatformAliases(t *testing.T) {
	platform, err := parseStorePlatform(" iOS ")
	if err != nil {
		t.Fatal(err)
	}
	if platform != appstore.PlatformIPhone {
		t.Fatalf("expected iphone platform, got %q", platform)
	}

	if _, err := parseStorePlatform("android"); err == nil {
		t.Fatal("expected unsupported platform to fail")
	}
}

func TestIsRetryableAppleResponse(t *testing.T) {
	for _, message := range []string{
		"unexpected response from Apple (HTTP 500)",
		"empty or non-plist body",
		"invalid response",
	} {
		if !isRetryableAppleResponse(assertionError(message)) {
			t.Fatalf("expected %q to be retryable", message)
		}
	}
	if isRetryableAppleResponse(assertionError("invalid credentials")) {
		t.Fatal("credentials error must not be retried as a transient response")
	}
}

func TestAuthenticationErrorMessageDistinguishesUnusableResponse(t *testing.T) {
	responseErr := appstore.NewErrorWithMetadata(fmt.Errorf("%w: %w", appstore.ErrAuthenticationUnavailable, &ipahttp.UnexpectedResponseError{StatusCode: 301, Reason: "missing Location"}), nil)
	message := authenticationErrorMessage(fmt.Errorf("request failed: %w", responseErr))
	if !strings.Contains(message, "切换代理节点或网络") || !strings.Contains(message, "HTTP 301") || strings.Contains(message, "missing Location") {
		t.Fatalf("missing actionable authentication diagnosis: %s", message)
	}
	credentialErr := errors.New("invalid credentials")
	if message := authenticationErrorMessage(credentialErr); message != credentialErr.Error() {
		t.Fatalf("real credential failure must retain its original message: %s", message)
	}
}

type assertionError string

func (e assertionError) Error() string { return string(e) }

type accountRetryTestStore struct {
	appstore.AppStore
	cached     appstore.Account
	refreshed  appstore.Account
	loginErr   error
	loginCalls int
	loginInput appstore.LoginInput
}

func (s *accountRetryTestStore) AccountInfo() (appstore.AccountInfoOutput, error) {
	return appstore.AccountInfoOutput{Account: s.cached}, nil
}

func (s *accountRetryTestStore) Login(input appstore.LoginInput) (appstore.LoginOutput, error) {
	s.loginCalls++
	s.loginInput = input
	return appstore.LoginOutput{Account: s.refreshed}, s.loginErr
}

func TestExpiredAccountRetriesWithFreshCredentials(t *testing.T) {
	store := &accountRetryTestStore{
		cached:    appstore.Account{Email: "test@example.com", Password: "password", PasswordToken: "expired"},
		refreshed: appstore.Account{Email: "test@example.com", PasswordToken: "fresh"},
	}
	var tokens []string
	result, err := withAccountRetry(store, func(account appstore.Account) (string, error) {
		tokens = append(tokens, account.PasswordToken)
		if account.PasswordToken == "expired" {
			return "", appstore.ErrPasswordTokenExpired
		}
		return account.PasswordToken, nil
	})
	if err != nil || result != "fresh" || len(tokens) != 2 || tokens[1] != "fresh" {
		t.Fatalf("expired credentials were not replaced: result=%q tokens=%v err=%v", result, tokens, err)
	}
	if store.loginCalls != 1 || store.loginInput.Email != store.cached.Email || store.loginInput.Password != store.cached.Password {
		t.Fatal("reauthentication must use the saved account's credentials exactly once")
	}
}

func TestFailedReauthenticationStopsQueryAndPreservesCause(t *testing.T) {
	for _, loginErr := range []error{appstore.ErrAuthCodeRequired, errors.New("Apple returned HTTP 404: Not Found")} {
		store := &accountRetryTestStore{
			cached: appstore.Account{Email: "test@example.com", Password: "password"}, loginErr: loginErr,
		}
		calls := 0
		_, err := withAccountRetry(store, func(appstore.Account) (string, error) {
			calls++
			return "", appstore.ErrPasswordTokenExpired
		})
		if calls != 1 || store.loginCalls != 1 || !errors.Is(err, loginErr) || !strings.Contains(err.Error(), "登录已失效") {
			t.Fatalf("failed reauthentication must preserve its cause without retrying the query: calls=%d err=%v", calls, err)
		}
	}
}

func TestRefreshAccountDoesNotAcceptCachedCredentialsWithoutLogin(t *testing.T) {
	store := &accountRetryTestStore{
		cached:   appstore.Account{Email: "test@example.com", Password: "password", PasswordToken: "expired"},
		loginErr: errors.New("HTTP 404: Not Found"),
	}
	account, err := refreshAccount(context.Background(), store, store.cached)
	if err == nil || account.PasswordToken != "" || store.loginCalls != 1 {
		t.Fatalf("local cached credentials must not count as a successful refresh: account token=%q err=%v", account.PasswordToken, err)
	}
	store.cached.Password = ""
	store.loginCalls = 0
	if _, err := refreshAccount(context.Background(), store, store.cached); err == nil || store.loginCalls != 0 {
		t.Fatal("missing credentials must ask for a new login without sending an empty password")
	}
}

type concurrentRefreshTestStore struct {
	appstore.AppStore
	mu         sync.Mutex
	cached     appstore.Account
	refreshed  appstore.Account
	loginCalls atomic.Int32
}

func (s *concurrentRefreshTestStore) AccountInfo() (appstore.AccountInfoOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return appstore.AccountInfoOutput{Account: s.cached}, nil
}

func (s *concurrentRefreshTestStore) Login(appstore.LoginInput) (appstore.LoginOutput, error) {
	s.loginCalls.Add(1)
	time.Sleep(35 * time.Millisecond)
	s.mu.Lock()
	s.cached = s.refreshed
	out := s.refreshed
	s.mu.Unlock()
	return appstore.LoginOutput{Account: out}, nil
}

func TestWithAccountRetryDeduplicatesConcurrentRefresh(t *testing.T) {
	app := NewApp()
	store := &concurrentRefreshTestStore{
		cached:    appstore.Account{Email: "user@example.com", Password: "secret", PasswordToken: "expired-token"},
		refreshed: appstore.Account{Email: "user@example.com", Password: "secret", PasswordToken: "fresh-token"},
	}

	const workers = 5
	start := make(chan struct{})
	results := make([]string, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			results[idx], errs[idx] = withAccountRetryContext(context.Background(), app, "acc-1", store, func(acc appstore.Account) (string, error) {
				if acc.PasswordToken != "fresh-token" {
					return "", appstore.ErrPasswordTokenExpired
				}
				return acc.PasswordToken, nil
			})
		}(i)
	}
	close(start)
	wg.Wait()

	if got := store.loginCalls.Load(); got != 1 {
		t.Fatalf("expected concurrent refresh to call Login once, got %d", got)
	}
	for i := 0; i < workers; i++ {
		if errs[i] != nil || results[i] != "fresh-token" {
			t.Fatalf("worker %d failed: result=%q err=%v", i, results[i], errs[i])
		}
	}
}

type pkgDownloadTestStore struct {
	appstore.AppStore
	destinationPath string
	replicateCalls  int
}

func (s *pkgDownloadTestStore) AccountInfo() (appstore.AccountInfoOutput, error) {
	return appstore.AccountInfoOutput{Account: appstore.Account{Email: "user@example.com", PasswordToken: "tok", StoreFront: "143441-1,29"}}, nil
}

func (s *pkgDownloadTestStore) Lookup(input appstore.LookupInput) (appstore.LookupOutput, error) {
	return appstore.LookupOutput{App: appstore.App{ID: input.AppID, BundleID: input.BundleID, Name: "TestApp"}}, nil
}

func (s *pkgDownloadTestStore) Download(appstore.DownloadInput) (appstore.DownloadOutput, error) {
	return appstore.DownloadOutput{
		DestinationPath: s.destinationPath,
		Sinfs:           []appstore.Sinf{{ID: 1, Data: []byte("sinf")}},
	}, nil
}

func (s *pkgDownloadTestStore) ReplicateSinf(appstore.ReplicateSinfInput) error {
	s.replicateCalls++
	return nil
}

func TestDownloadFromStoreSkipsReplicateSinfForPkg(t *testing.T) {
	dir := t.TempDir()
	app := NewApp()
	app.dataDirOverride = dir
	app.accountMu.Lock()
	app.accounts = accountRegistry{
		ActiveAccountID: "acc-pkg",
		Accounts:        []storedAccount{{ID: "acc-pkg", Email: "user@example.com", LastUsed: 1}},
	}
	app.syncCachedAccountLocked()
	app.accountMu.Unlock()

	pkgPath := filepath.Join(dir, "TestApp.pkg")
	fake := &pkgDownloadTestStore{destinationPath: pkgPath}
	app.storeMu.Lock()
	app.storeSessions = map[storeSessionKey]appstore.AppStore{
		app.currentStoreSessionKey("acc-pkg", false): fake,
	}
	app.storeMu.Unlock()

	res, err := app.downloadFromStore(context.Background(), "acc-pkg", "com.example.mac", 101, "1", dir, "macos", false, nil)
	if err != nil {
		t.Fatalf("downloadFromStore(.pkg) failed: %v", err)
	}
	if !res.Success || res.Output != pkgPath || fake.replicateCalls != 0 {
		t.Fatalf("expected ReplicateSinf to be skipped for .pkg: res=%+v replicateCalls=%d", res, fake.replicateCalls)
	}

	ipaPath := filepath.Join(dir, "TestApp.ipa")
	fake.destinationPath = ipaPath
	res, err = app.downloadFromStore(context.Background(), "acc-pkg", "com.example.ios", 102, "2", dir, "iphone", false, nil)
	if err != nil {
		t.Fatalf("downloadFromStore(.ipa) failed: %v", err)
	}
	if !res.Success || res.Output != ipaPath || fake.replicateCalls != 1 {
		t.Fatalf("expected ReplicateSinf once for .ipa: res=%+v replicateCalls=%d", res, fake.replicateCalls)
	}
}

type spyCookieJar struct {
	cookies   []*gohttp.Cookie
	saveCalls int
}

func (s *spyCookieJar) SetCookies(_ *url.URL, cookies []*gohttp.Cookie) {
	s.cookies = append(s.cookies, cookies...)
}

func (s *spyCookieJar) Cookies(*url.URL) []*gohttp.Cookie {
	return s.cookies
}

func (s *spyCookieJar) Save() error {
	s.saveCalls++
	return nil
}

func TestSyncCookieJarOnlySavesWhenDirty(t *testing.T) {
	spy := &spyCookieJar{}
	jar := newSyncCookieJar(spy)
	targetURL, _ := url.Parse("https://example.com")

	if err := jar.Save(); err != nil {
		t.Fatal(err)
	}
	if spy.saveCalls != 0 {
		t.Fatalf("Save on clean jar called inner Save %d times", spy.saveCalls)
	}

	jar.SetCookies(targetURL, nil)
	if err := jar.Save(); err != nil {
		t.Fatal(err)
	}
	if spy.saveCalls != 0 {
		t.Fatalf("empty SetCookies must not mark jar dirty, saveCalls=%d", spy.saveCalls)
	}

	jar.SetCookies(targetURL, []*gohttp.Cookie{{Name: "sid", Value: "abc"}})
	if err := jar.Save(); err != nil {
		t.Fatal(err)
	}
	if spy.saveCalls != 1 {
		t.Fatalf("expected 1 Save after SetCookies, got %d", spy.saveCalls)
	}

	if err := jar.Save(); err != nil {
		t.Fatal(err)
	}
	if spy.saveCalls != 1 {
		t.Fatalf("subsequent Save without mutations should skip inner Save, got %d", spy.saveCalls)
	}
}
