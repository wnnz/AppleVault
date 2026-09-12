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
	if err = app.SaveSettings(Settings{KeychainPassphrase: "stale-value", DefaultPlatform: "ipad"}); err != nil {
		t.Fatal(err)
	}
	if app.GetSettings().KeychainPassphrase != "123456" {
		t.Fatal("saving unrelated settings must not overwrite the keychain passphrase")
	}
	if err = app.SetKeychainPassphrase("new-secret"); err != nil {
		t.Fatal(err)
	}
	if app.GetSettings().KeychainPassphrase != "new-secret" {
		t.Fatal("explicit passphrase update was not saved")
	}
}
