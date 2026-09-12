package appstore

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	apphttp "AppleVault/src/backend/internal/ipatool/http"
	"AppleVault/src/backend/internal/ipatool/util/operatingsystem"
)

func downloadTestStore() *appstore {
	return &appstore{
		httpClient: apphttp.NewClient[interface{}](apphttp.Args{Context: context.Background()}),
		os:         operatingsystem.New(),
	}
}

func TestDownloadFileRestartsWhenRangeIsIgnored(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "bytes=3-" {
			t.Errorf("unexpected range: %q", r.Header.Get("Range"))
		}
		_, _ = io.WriteString(w, "fresh")
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "download.tmp")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := downloadTestStore().downloadFile(server.URL, path, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "fresh" {
		t.Fatalf("expected restarted download, got %q", data)
	}
}

func TestDownloadFileAppendsValidPartialResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 3-5/6")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "def")
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "download.tmp")
	_ = os.WriteFile(path, []byte("abc"), 0600)
	if err := downloadTestStore().downloadFile(server.URL, path, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "abcdef" {
		t.Fatalf("expected resumed download, got %q", data)
	}
}

func TestDownloadFileAcceptsAlreadyCompleteRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Range", "bytes */6")
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "download.tmp")
	_ = os.WriteFile(path, []byte("abcdef"), 0600)
	if err := downloadTestStore().downloadFile(server.URL, path, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "abcdef" {
		t.Fatalf("expected complete temporary file to remain unchanged, got %q", data)
	}
}

func TestDownloadIntegrityChecks(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file")
	data := []byte("verified")
	_ = os.WriteFile(filePath, data, 0600)
	sum := md5.Sum(data)
	if err := validateFileMD5(filePath, hex.EncodeToString(sum[:])); err != nil {
		t.Fatal(err)
	}
	if err := validateFileMD5(filePath, "00000000000000000000000000000000"); err == nil {
		t.Fatal("expected checksum mismatch")
	}

	ipaPath := filepath.Join(dir, "valid.ipa")
	archive, _ := os.Create(ipaPath)
	writer := zip.NewWriter(archive)
	entry, _ := writer.Create("Payload/Example.app/Info.plist")
	_, _ = entry.Write([]byte("plist"))
	_ = writer.Close()
	_ = archive.Close()
	if err := validateIPAZip(ipaPath); err != nil {
		t.Fatal(err)
	}
}
