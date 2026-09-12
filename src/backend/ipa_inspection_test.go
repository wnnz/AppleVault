package backend

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"howett.net/plist"
)

func createTestIPA(t *testing.T, info ipaBundleInfo, signed bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "Example.ipa")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	infoEntry, _ := archive.Create("Payload/Example.app/Info.plist")
	data, _ := plist.Marshal(info, plist.XMLFormat)
	_, _ = infoEntry.Write(data)
	if signed {
		signature, _ := archive.Create("Payload/Example.app/_CodeSignature/CodeResources")
		_, _ = signature.Write([]byte("signature"))
	}
	if err = archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInspectIPAAndDeviceCompatibility(t *testing.T) {
	app := NewApp()
	path := createTestIPA(t, ipaBundleInfo{
		DisplayName: "Example", BundleID: "com.example.app", ShortVersion: "2.0", BuildVersion: "200",
		MinimumOSVersion: "18.0", SupportedPlatforms: []string{"iPhoneOS"}, DeviceFamily: []int{1, 2},
	}, true)
	result, err := app.InspectIPA(path, "iPhone17,1", "17.6")
	if err != nil {
		t.Fatal(err)
	}
	if result.AppName != "Example" || result.BundleID != "com.example.app" || !result.Signed {
		t.Fatalf("unexpected IPA metadata: %#v", result)
	}
	if result.Compatible || !strings.Contains(result.CompatibilityMessage, "低于") {
		t.Fatalf("expected minimum OS incompatibility: %#v", result)
	}
}

func TestInspectIPARejectsMissingSignature(t *testing.T) {
	app := NewApp()
	path := createTestIPA(t, ipaBundleInfo{BundleID: "com.example.unsigned"}, false)
	result, err := app.InspectIPA(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Compatible || result.Signed {
		t.Fatalf("unsigned IPA should not be compatible: %#v", result)
	}
}

func TestAppleVersionComparison(t *testing.T) {
	if compareAppleVersions("17.6.1", "17.6") <= 0 || compareAppleVersions("17.5", "18.0") >= 0 {
		t.Fatal("unexpected Apple version comparison")
	}
	if isAppleVersion("已锁定") || !isAppleVersion("17.6.1") {
		t.Fatal("unexpected Apple version validation")
	}
}
