package backend

import (
	"context"
	"errors"
	"sync"
	"testing"

	"AppleVault/src/backend/internal/ipatool/appstore"
)

type spyKeychain struct {
	mu          sync.Mutex
	data        map[string][]byte
	getCalls    int
	setCalls    int
	removeCalls int
}

func newSpyKeychain() *spyKeychain {
	return &spyKeychain{data: make(map[string][]byte)}
}

func (s *spyKeychain) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getCalls++
	val, ok := s.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return append([]byte(nil), val...), nil
}

func (s *spyKeychain) Set(key string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setCalls++
	s.data[key] = append([]byte(nil), data...)
	return nil
}

func (s *spyKeychain) Remove(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeCalls++
	delete(s.data, key)
	return nil
}

func TestCachingKeychainReadsOnceAndDeduplicatesWrites(t *testing.T) {
	spy := newSpyKeychain()
	spy.data["account"] = []byte("secret-v1")
	kc := newCachingKeychain(spy)

	v1, err := kc.Get("account")
	if err != nil || string(v1) != "secret-v1" {
		t.Fatalf("first Get = %q, %v", v1, err)
	}
	v2, err := kc.Get("account")
	if err != nil || string(v2) != "secret-v1" {
		t.Fatalf("second Get = %q, %v", v2, err)
	}
	if spy.getCalls != 1 {
		t.Fatalf("expected 1 underlying Get call, got %d", spy.getCalls)
	}

	if err := kc.Set("account", []byte("secret-v1")); err != nil {
		t.Fatal(err)
	}
	if spy.setCalls != 0 {
		t.Fatalf("identical Set must not write through, got %d Set calls", spy.setCalls)
	}

	if err := kc.Set("account", []byte("secret-v2")); err != nil {
		t.Fatal(err)
	}
	if spy.setCalls != 1 {
		t.Fatalf("changed Set must write through once, got %d", spy.setCalls)
	}

	v3, err := kc.Get("account")
	if err != nil || string(v3) != "secret-v2" {
		t.Fatalf("Get after Set = %q, %v", v3, err)
	}
	if spy.getCalls != 1 {
		t.Fatalf("Get after Set must remain cached, got %d inner Get calls", spy.getCalls)
	}

	if err := kc.Remove("account"); err != nil {
		t.Fatal(err)
	}
	if spy.removeCalls != 1 {
		t.Fatalf("expected 1 Remove call, got %d", spy.removeCalls)
	}
	if _, err := kc.Get("account"); err == nil {
		t.Fatal("expected removed key lookup to fail")
	}
}

type lookupSpyStore struct {
	appstore.AppStore
	lookupCalls int
	lastInput   appstore.LookupInput
	app         appstore.App
}

func (l *lookupSpyStore) Lookup(input appstore.LookupInput) (appstore.LookupOutput, error) {
	l.lookupCalls++
	l.lastInput = input
	return appstore.LookupOutput{App: l.app}, nil
}

func TestResolveStoreAppCachedSkipsAndCachesLookups(t *testing.T) {
	app := NewApp()
	acc := appstore.Account{StoreFront: "143441-1,29"}
	spy := &lookupSpyStore{
		app: appstore.App{ID: 414478124, BundleID: "com.tencent.xin", Name: "WeChat", Price: 0},
	}

	// When appID is already known and full metadata is not needed, Lookup must be skipped.
	resolved, err := app.resolveStoreAppCached(context.Background(), spy, acc, "com.tencent.xin", 414478124, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if spy.lookupCalls != 0 || resolved.ID != 414478124 || resolved.BundleID != "com.tencent.xin" {
		t.Fatalf("unexpected direct resolution: calls=%d app=%+v", spy.lookupCalls, resolved)
	}

	// First Lookup when appID == 0 must hit store.Lookup and populate lookupCache.
	ctx := context.Background()
	resolved, err = app.resolveStoreAppCached(ctx, spy, acc, "com.tencent.xin", 0, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if spy.lookupCalls != 1 || resolved.Name != "WeChat" || spy.lastInput.Context != ctx {
		t.Fatalf("unexpected initial Lookup result: calls=%d app=%+v", spy.lookupCalls, resolved)
	}

	// Subsequent calls (even with requireFullMetadata == true) should hit the cache.
	resolved2, err := app.resolveStoreAppCached(ctx, spy, acc, "com.tencent.xin", 414478124, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if spy.lookupCalls != 1 || resolved2.Name != "WeChat" {
		t.Fatalf("expected second Lookup to reuse cache, calls=%d app=%+v", spy.lookupCalls, resolved2)
	}
}

func TestVersionMetadataCachePersistsAcrossAppInstances(t *testing.T) {
	dir := t.TempDir()
	app1 := NewApp()
	app1.dataDirOverride = dir

	want := VersionMetadataResult{
		ExternalVersionID: "871234567",
		DisplayVersion:    "8.0.55",
		ReleaseDate:       "2025-12-01",
		FileSize:          2048,
		DisplayFileSize:   formatBytes(2048),
		Success:           true,
	}
	app1.putCachedVersionMetadata("com.example.App", 987654, want)

	// Empty version ID represents dynamic "latest" lookups and must never be cached.
	app1.putCachedVersionMetadata("com.example.App", 987654, VersionMetadataResult{
		ExternalVersionID: "",
		DisplayVersion:    "9.9.9",
		Success:           true,
	})
	if _, ok := app1.getCachedVersionMetadata("com.example.App", 987654, ""); ok {
		t.Fatal("empty versionID must not be stored or returned from cache")
	}

	app2 := NewApp()
	app2.dataDirOverride = dir

	gotByID, ok := app2.getCachedVersionMetadata("", 987654, "871234567")
	if !ok || gotByID != want {
		t.Fatalf("cache miss or mismatch by appID: ok=%v got=%+v want=%+v", ok, gotByID, want)
	}

	gotByBundle, ok := app2.getCachedVersionMetadata("COM.EXAMPLE.APP", 0, "871234567")
	if !ok || gotByBundle != want {
		t.Fatalf("cache miss or mismatch by bundleID: ok=%v got=%+v want=%+v", ok, gotByBundle, want)
	}

	batch := app2.getCachedVersionMetadataBatch("com.example.App", 987654, []string{"871234567", "999999999"})
	if len(batch) != 1 || batch["871234567"] != want {
		t.Fatalf("unexpected batch lookup result: %+v", batch)
	}
}
