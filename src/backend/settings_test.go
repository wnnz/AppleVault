package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
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
