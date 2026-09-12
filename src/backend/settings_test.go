package backend

import (
	"encoding/json"
	"os"
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
