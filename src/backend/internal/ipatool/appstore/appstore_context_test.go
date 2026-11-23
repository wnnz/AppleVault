package appstore

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	gohttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apphttp "AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/util/operatingsystem"
)

type stubDownloadClient struct {
	captured []apphttp.Request
	sendFunc func(req apphttp.Request) (apphttp.Result[downloadResult], error)
}

func (s *stubDownloadClient) Send(req apphttp.Request) (apphttp.Result[downloadResult], error) {
	s.captured = append(s.captured, req)
	if s.sendFunc != nil {
		return s.sendFunc(req)
	}
	return apphttp.Result[downloadResult]{StatusCode: gohttp.StatusOK}, nil
}

func (s *stubDownloadClient) Do(*gohttp.Request) (*gohttp.Response, error) {
	return nil, errors.New("unexpected Do call")
}

func (s *stubDownloadClient) NewRequest(string, string, io.Reader) (*gohttp.Request, error) {
	return nil, errors.New("unexpected NewRequest call")
}

func createMinimalIPAZip(t *testing.T, dir, name string) string {
	t.Helper()
	ipaPath := filepath.Join(dir, name)
	f, err := os.Create(ipaPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	infoEntry, err := zw.Create("Payload/Example.app/Info.plist")
	if err != nil {
		t.Fatal(err)
	}
	const infoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleExecutable</key><string>Example</string>
<key>CFBundleSupportedPlatforms</key><array><string>iPhoneOS</string></array>
</dict></plist>`
	if _, err := infoEntry.Write([]byte(infoPlist)); err != nil {
		t.Fatal(err)
	}
	binEntry, err := zw.Create("Payload/Example.app/Example")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := binEntry.Write([]byte("binary")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return ipaPath
}

func TestSendCommerceDownloadRequestPropagatesContext(t *testing.T) {
	stub := &stubDownloadClient{
		sendFunc: func(req apphttp.Request) (apphttp.Result[downloadResult], error) {
			if req.Context != nil {
				if err := req.Context.Err(); err != nil {
					return apphttp.Result[downloadResult]{}, err
				}
			}
			return apphttp.Result[downloadResult]{StatusCode: gohttp.StatusOK}, nil
		},
	}
	store := &appstore{downloadClient: stub}

	ctx, cancel := context.WithCancel(context.Background())
	req := apphttp.Request{URL: "https://example.invalid/download", Method: apphttp.MethodPOST}

	if _, err := store.sendCommerceDownloadRequest(ctx, req, "001122334455"); err != nil {
		t.Fatalf("sendCommerceDownloadRequest() error = %v", err)
	}
	if len(stub.captured) != 1 || stub.captured[0].Context != ctx {
		t.Fatalf("captured request context = %v, want %v", stub.captured[0].Context, ctx)
	}

	cancel()
	_, err := store.sendCommerceDownloadRequest(ctx, req, "001122334455")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled after cancel, got %v", err)
	}
}

func TestReplicateSinfTruncatesStaleTmpAndRespectsCancel(t *testing.T) {
	dir := t.TempDir()
	ipaPath := createMinimalIPAZip(t, dir, "Example.ipa")
	tmpPath := ipaPath + ".tmp"
	if err := os.WriteFile(tmpPath, bytes.Repeat([]byte("X"), 64*1024), 0644); err != nil {
		t.Fatal(err)
	}

	store := &appstore{os: operatingsystem.New()}
	wantSinf := []byte("replicated-sinf-payload")
	err := store.ReplicateSinf(ReplicateSinfInput{
		Context:     context.Background(),
		PackagePath: ipaPath,
		Sinfs:       []Sinf{{ID: 1, Data: wantSinf}},
	})
	if err != nil {
		t.Fatalf("ReplicateSinf() error = %v", err)
	}
	if _, statErr := os.Stat(tmpPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected %s to be removed, stat err = %v", tmpPath, statErr)
	}

	zr, err := zip.OpenReader(ipaPath)
	if err != nil {
		t.Fatalf("failed to reopen replicated IPA as zip: %v", err)
	}
	foundSinf := false
	for _, f := range zr.File {
		if f.Name == "Payload/Example.app/SC_Info/Example.sinf" {
			rc, openErr := f.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			got, readErr := io.ReadAll(rc)
			_ = rc.Close()
			if readErr != nil || !bytes.Equal(got, wantSinf) {
				t.Fatalf("unexpected replicated sinf data (%q, %v)", got, readErr)
			}
			foundSinf = true
		}
	}
	_ = zr.Close()
	if !foundSinf {
		t.Fatal("replicated IPA did not contain expected SC_Info/Example.sinf")
	}

	ipaPath2 := createMinimalIPAZip(t, dir, "Canceled.ipa")
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = store.ReplicateSinf(ReplicateSinfInput{
		Context:     canceledCtx,
		PackagePath: ipaPath2,
		Sinfs:       []Sinf{{ID: 1, Data: wantSinf}},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled for canceled ReplicateSinf, got %v", err)
	}
	if _, statErr := os.Stat(ipaPath2 + ".tmp"); !os.IsNotExist(statErr) {
		t.Fatalf("expected canceled ReplicateSinf not to leave .tmp, stat err = %v", statErr)
	}
}

func TestPackageFileNameSanitizesAndAppendsExternalVersionID(t *testing.T) {
	app := App{ID: 414478124, BundleID: "com.example/../../evil:app"}
	got := packageFileNameWithVersionID(app, "1.0/../../bad:ver", "871234567", PlatformIPhone)
	if strings.ContainsAny(got, `/\:`) || strings.Contains(got, "..") {
		t.Fatalf("package file name %q contains unsafe path characters", got)
	}
	if !strings.Contains(got, "871234567") || !strings.HasSuffix(got, ".ipa") {
		t.Fatalf("unexpected IPA file name %q", got)
	}

	noVerID := packageFileNameWithVersionID(App{ID: 42, BundleID: "com.example.mac"}, "2.0", "", PlatformMacOS)
	if noVerID != "com.example.mac_42_2.0.pkg" {
		t.Fatalf("macOS package file name = %q, want com.example.mac_42_2.0.pkg", noVerID)
	}

	sanitized := sanitizeFileComponent("  ..1.0/../../bad:ver  ")
	if strings.ContainsAny(sanitized, `/\: `) || strings.Contains(sanitized, "..") {
		t.Fatalf("sanitizeFileComponent returned unsafe component %q", sanitized)
	}
}

func TestValidateFileMD5AndIPAZipContextCancel(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "blob.bin")
	payload := []byte("hello-md5-context")
	if err := os.WriteFile(dataPath, payload, 0600); err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum(payload)
	hexSum := hex.EncodeToString(sum[:])

	ipaPath := createMinimalIPAZip(t, dir, "CheckCancel.ipa")

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := validateFileMD5Context(canceledCtx, dataPath, hexSum); !errors.Is(err, context.Canceled) {
		t.Fatalf("validateFileMD5Context() with canceled ctx = %v, want context.Canceled", err)
	}
	if err := validateIPAZipContext(canceledCtx, ipaPath); !errors.Is(err, context.Canceled) {
		t.Fatalf("validateIPAZipContext() with canceled ctx = %v, want context.Canceled", err)
	}
}
