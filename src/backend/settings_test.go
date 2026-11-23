package backend

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewAppHasNoDefaultKeychainPassphrase(t *testing.T) {
	app := NewApp()
	if app.settings.KeychainPassphrase != "" {
		t.Fatalf("new app should not provide a default keychain passphrase, got %q", app.settings.KeychainPassphrase)
	}
}

func TestSavedKeychainPassphraseIsPreserved(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()
	saved := Settings{KeychainPassphrase: "123456", DefaultPlatform: "iphone"}
	data, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(app.getSettingsFilePath(), data, 0600); err != nil {
		t.Fatal(err)
	}

	app.loadSettings()
	if app.GetSettings().KeychainPassphrase != "123456" {
		t.Fatal("existing default passphrase must remain available for stored credentials")
	}
	if app.GetThemeHintShown() {
		t.Fatal("legacy settings without a theme hint field should show the hint")
	}
	if err = app.MarkThemeHintShown(); err != nil {
		t.Fatal(err)
	}
	if !app.GetThemeHintShown() {
		t.Fatal("theme hint state was not saved")
	}
	reloaded := NewApp()
	reloaded.dataDirOverride = app.dataDirOverride
	reloaded.loadSettings()
	if !reloaded.GetThemeHintShown() || reloaded.GetSettings().KeychainPassphrase != "123456" {
		t.Fatal("theme hint state and existing passphrase should persist together")
	}
	if err = app.SaveSettings(Settings{KeychainPassphrase: "stale-value", DefaultPlatform: "ipad"}); err != nil {
		t.Fatal(err)
	}
	if app.GetSettings().KeychainPassphrase != "123456" {
		t.Fatal("saving unrelated settings must not overwrite the keychain passphrase")
	}
	if !app.GetThemeHintShown() {
		t.Fatal("saving unrelated settings must not reset the theme hint state")
	}
	if err = app.SetKeychainPassphrase("new-secret"); err != nil {
		t.Fatal(err)
	}
	if app.GetSettings().KeychainPassphrase != "new-secret" {
		t.Fatal("explicit passphrase update was not saved")
	}
}

func TestLoadSettingsBacksUpCorruptFile(t *testing.T) {
	app := NewApp()
	app.dataDirOverride = t.TempDir()
	settingsPath := app.getSettingsFilePath()
	if err := os.WriteFile(settingsPath, []byte("{bad"), 0600); err != nil {
		t.Fatal(err)
	}

	app.loadSettings()

	matches, err := filepath.Glob(settingsPath + ".corrupt-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 corrupt backup file, got %d (%v)", len(matches), matches)
	}
	if _, statErr := os.Stat(settingsPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected original corrupt settings.json to be moved away, stat err=%v", statErr)
	}
}

func TestWriteFileAtomicReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")

	if err := writeFileAtomic(target, []byte(`{"v":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(target, []byte(`{"v":2}`), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"v":2}` {
		t.Fatalf("expected second write contents, got %q", string(got))
	}

	leftovers, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("expected no leftover temp files, got %v", leftovers)
	}
}

func TestKeychainPassphraseEncryptedAtRest(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("DPAPI encryption test only runs on Windows")
	}
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	const secret = "super-secret-pass-123"
	if err := app.SetKeychainPassphrase(secret); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(app.getSettingsFilePath())
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if strings.Contains(content, secret) {
		t.Fatal("settings.json should not contain plaintext passphrase")
	}
	if !strings.Contains(content, "protectedKeychainPassphrase") {
		t.Fatal("settings.json should contain protectedKeychainPassphrase")
	}

	reloaded := NewApp()
	reloaded.dataDirOverride = app.dataDirOverride
	reloaded.loadSettings()
	if got := reloaded.GetSettings().KeychainPassphrase; got != secret {
		t.Fatalf("expected decrypted passphrase %q, got %q", secret, got)
	}
}

func TestPlaintextPassphraseMigrated(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("DPAPI encryption test only runs on Windows")
	}
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	const secret = "legacy-plain-secret-456"
	if err := os.WriteFile(app.getSettingsFilePath(), []byte(`{"keychainPassphrase":"`+secret+`"}`), 0600); err != nil {
		t.Fatal(err)
	}

	app.loadSettings()
	if got := app.GetSettings().KeychainPassphrase; got != secret {
		t.Fatalf("expected passphrase %q after migration, got %q", secret, got)
	}

	raw, err := os.ReadFile(app.getSettingsFilePath())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatal("migrated settings.json should no longer contain plaintext passphrase")
	}
	if !strings.Contains(string(raw), "protectedKeychainPassphrase") {
		t.Fatal("migrated settings.json should contain protectedKeychainPassphrase")
	}
}

func TestUndecryptablePassphraseBlobPreserved(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("DPAPI encryption test only runs on Windows")
	}
	app := NewApp()
	app.dataDirOverride = t.TempDir()

	const blob = "bm90LWEtZHBhcGktYmxvYg=="
	if err := os.WriteFile(app.getSettingsFilePath(), []byte(`{"protectedKeychainPassphrase":"`+blob+`"}`), 0600); err != nil {
		t.Fatal(err)
	}

	app.loadSettings()
	if got := app.GetSettings().KeychainPassphrase; got != "" {
		t.Fatalf("expected empty passphrase when blob cannot be decrypted, got %q", got)
	}

	if err := app.MarkThemeHintShown(); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(app.getSettingsFilePath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), blob) {
		t.Fatalf("expected undecryptable blob to be preserved in settings.json, got %s", string(raw))
	}
}

func TestIsTemporaryExecutableDir(t *testing.T) {
	tests := []struct {
		dir  string
		want bool
	}{
		{`C:\Users\x\AppData\Local\Temp\go-build123\b001\exe`, true},
		{`D:\Templates\AppleVault`, false},
		{`D:\Apps\AppleVault`, false},
		{`/tmp/abc`, true},
		{`/private/var/folders/xx/T/AppTranslocation/ABC/d/AppleVault.app/Contents/MacOS`, true},
	}
	for _, tc := range tests {
		if got := isTemporaryExecutableDir(tc.dir); got != tc.want {
			t.Errorf("isTemporaryExecutableDir(%q) = %v, want %v", tc.dir, got, tc.want)
		}
	}
}

func TestFormatBytesBounds(t *testing.T) {
	if got := formatBytes(0); got != "-" {
		t.Fatalf("formatBytes(0) = %q, want '-'", got)
	}
	if got := formatBytes(512); got != "512 B" {
		t.Fatalf("formatBytes(512) = %q, want '512 B'", got)
	}
	if got := formatBytes(1024); got != "1.00 KB" {
		t.Fatalf("formatBytes(1024) = %q, want '1.00 KB'", got)
	}
	large := formatBytes(1 << 50)
	if !strings.Contains(large, "TB") {
		t.Fatalf("formatBytes(1<<50) = %q, expected TB unit", large)
	}
}

func TestSecretAESGCMRoundTripAndTamper(t *testing.T) {
	key1 := bytes.Repeat([]byte{0x1a}, 32)
	key2 := bytes.Repeat([]byte{0x2b}, 32)

	if sealed, err := sealSecretAESGCM(key1, ""); err != nil || sealed != "" {
		t.Fatalf("empty seal got (%q, %v), want (\"\", nil)", sealed, err)
	}
	if plain, err := openSecretAESGCM(key1, ""); err != nil || plain != "" {
		t.Fatalf("empty open got (%q, %v), want (\"\", nil)", plain, err)
	}

	const secret = "passphrase-!@#$-中文"
	sealed, err := sealSecretAESGCM(key1, secret)
	if err != nil {
		t.Fatalf("sealSecretAESGCM failed: %v", err)
	}
	if !strings.HasPrefix(sealed, aesGCMTokenPrefix) {
		t.Fatalf("sealed secret %q missing prefix %q", sealed, aesGCMTokenPrefix)
	}

	plain, err := openSecretAESGCM(key1, sealed)
	if err != nil || plain != secret {
		t.Fatalf("openSecretAESGCM = (%q, %v), want %q", plain, err, secret)
	}

	if _, err := sealSecretAESGCM(key1[:16], secret); err == nil {
		t.Fatal("expected error for invalid 16-byte seal key")
	}
	if _, err := openSecretAESGCM(key1[:16], sealed); err == nil {
		t.Fatal("expected error for invalid 16-byte open key")
	}
	if _, err := openSecretAESGCM(key2, sealed); err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
	if _, err := openSecretAESGCM(key1, "plain:"+secret); err == nil {
		t.Fatal("expected error for unsupported prefix")
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(sealed, aesGCMTokenPrefix))
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0xff
	tampered := aesGCMTokenPrefix + base64.StdEncoding.EncodeToString(raw)
	if _, err := openSecretAESGCM(key1, tampered); err == nil {
		t.Fatal("expected authentication error for tampered ciphertext")
	}
}

func TestParseAndValidateProxyURL(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"127.0.0.1:10808",
		"ftp://127.0.0.1:21",
		"http://",
		"://bad-url",
	}
	for _, raw := range invalid {
		if parsed, err := parseAndValidateProxyURL(raw); err == nil {
			t.Errorf("parseAndValidateProxyURL(%q) unexpectedly succeeded: %v", raw, parsed)
		}
	}

	valid := []string{
		"http://127.0.0.1:10808",
		"https://proxy.example.com:8443",
		"socks5://127.0.0.1:1080",
		"socks5h://user:pass@127.0.0.1:1080",
	}
	for _, raw := range valid {
		if parsed, err := parseAndValidateProxyURL(raw); err != nil || parsed == nil {
			t.Errorf("parseAndValidateProxyURL(%q) failed: %v", raw, err)
		}
	}
}

func TestNewAppDefaultsProxyDisabled(t *testing.T) {
	app := NewApp()
	if app.GetSettings().EnableProxy {
		t.Fatal("expected new app to default EnableProxy to false")
	}
}

func TestGetRecentLogsBuffersStartupLogs(t *testing.T) {
	app := NewApp()
	for i := 0; i < 120; i++ {
		app.emitLog(fmt.Sprintf("startup-log-%03d", i))
	}

	logs := app.GetRecentLogs()
	if len(logs) != maxRecentLogs {
		t.Fatalf("GetRecentLogs() returned %d items, want %d", len(logs), maxRecentLogs)
	}
	if !strings.Contains(logs[0], "startup-log-020") {
		t.Fatalf("oldest retained log = %q, want startup-log-020", logs[0])
	}
	if !strings.Contains(logs[len(logs)-1], "startup-log-119") {
		t.Fatalf("newest retained log = %q, want startup-log-119", logs[len(logs)-1])
	}

	logs[0] = "mutated-by-caller"
	again := app.GetRecentLogs()
	if again[0] == "mutated-by-caller" {
		t.Fatal("GetRecentLogs must return a defensive copy")
	}
}
