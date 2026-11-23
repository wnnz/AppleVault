package backend

import (
	"archive/zip"
	"context"
	"errors"
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

func TestVerifyZIPEntriesContext(t *testing.T) {
	path := createTestIPA(t, ipaBundleInfo{BundleID: "com.example.verify"}, true)
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if err := verifyZIPEntriesContext(context.Background(), reader.File); err != nil {
		t.Fatalf("expected valid IPA to verify, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = verifyZIPEntriesContext(ctx, reader.File)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestInspectIPAiPhoneOnlyAllowedOnIPad(t *testing.T) {
	app := NewApp()

	iphoneOnlyPath := createTestIPA(t, ipaBundleInfo{
		DisplayName:        "iPhoneApp",
		BundleID:           "com.example.iphoneonly",
		ShortVersion:       "1.0",
		SupportedPlatforms: []string{"iPhoneOS"},
		DeviceFamily:       []int{1},
	}, true)
	resIPad, err := app.InspectIPA(iphoneOnlyPath, "iPad16,2", "18.0")
	if err != nil {
		t.Fatal(err)
	}
	if !resIPad.Compatible || !strings.Contains(resIPad.CompatibilityMessage, "兼容模式") {
		t.Fatalf("expected iPhone-only IPA to be allowed in compatibility mode on iPad, got %+v", resIPad)
	}

	ipadOnlyPath := createTestIPA(t, ipaBundleInfo{
		DisplayName:        "iPadApp",
		BundleID:           "com.example.ipadonly",
		ShortVersion:       "1.0",
		SupportedPlatforms: []string{"iPhoneOS"},
		DeviceFamily:       []int{2},
	}, true)
	resIPhone, err := app.InspectIPA(ipadOnlyPath, "iPhone17,1", "18.0")
	if err != nil {
		t.Fatal(err)
	}
	if resIPhone.Compatible {
		t.Fatalf("expected iPad-only IPA to be incompatible with iPhone, got %+v", resIPhone)
	}
}
