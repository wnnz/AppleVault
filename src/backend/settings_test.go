package backend

import (
	"encoding/json"
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
	}
	for _, tc := range tests {
		if got := isTemporaryExecutableDir(tc.dir); got != tc.want {
			t.Errorf("isTemporaryExecutableDir(%q) = %v, want %v", tc.dir, got, tc.want)
		}
	}
}
