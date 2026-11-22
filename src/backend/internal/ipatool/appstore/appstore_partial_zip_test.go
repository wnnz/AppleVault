package appstore

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	apphttp "AppleVault/src/backend/internal/ipatool/http"
	"howett.net/plist"
)

func buildLargeTestIPA(t *testing.T, infoMethod uint16, infoContent []byte, infoModTime time.Time) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	infoHeader := &zip.FileHeader{
		Name:   "Payload/Example.app/Info.plist",
		Method: infoMethod,
	}
	if !infoModTime.IsZero() {
		infoHeader.Modified = infoModTime
	}

	w, err := zw.CreateHeader(infoHeader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(infoContent); err != nil {
		t.Fatal(err)
	}

	// Add a large incompressible-ish gap before the Central Directory so the entry's
	// local header falls well outside the initial 65 KB ZIP trailer fetch.
	padWriter, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "Payload/Example.app/Padding.bin",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatal(err)
	}
	padding := make([]byte, 128<<10)
	for i := range padding {
		padding[i] = byte(i * 31)
	}
	if _, err := padWriter.Write(padding); err != nil {
		t.Fatal(err)
	}

	// Populate > 2500 entries so the Central Directory exceeds 150 KB and would
	// otherwise force dozens of 4 KB sequential reads under archive/zip's bufio.
	for i := 0; i < 2500; i++ {
		ew, err := zw.CreateHeader(&zip.FileHeader{
			Name:   fmt.Sprintf("Payload/Example.app/Resources/asset-%04d.dat", i),
			Method: zip.Store,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ew.Write([]byte{1}); err != nil {
			t.Fatal(err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestReadVersionMetadataFromIPACoalescesRangeRequests(t *testing.T) {
	plistData, err := plist.Marshal(map[string]any{
		"CFBundleShortVersionString": "8.0.50",
		"releaseDate":                "2025-11-12T08:09:10Z",
	}, plist.XMLFormat)
	if err != nil {
		t.Fatal(err)
	}

	ipaBytes := buildLargeTestIPA(t, zip.Deflate, plistData, time.Time{})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.ServeContent(w, r, "app.ipa", time.Time{}, bytes.NewReader(ipaBytes))
	}))
	defer server.Close()

	store := &appstore{httpClient: passthroughClient{}}
	meta, size, err := store.readVersionMetadataFromIPA(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("readVersionMetadataFromIPA failed: %v", err)
	}

	if size != int64(len(ipaBytes)) {
		t.Fatalf("size = %d, want %d", size, len(ipaBytes))
	}
	if meta.DisplayVersion != "8.0.50" {
		t.Fatalf("DisplayVersion = %q, want 8.0.50", meta.DisplayVersion)
	}
	wantDate := time.Date(2025, 11, 12, 8, 9, 10, 0, time.UTC)
	if !meta.ReleaseDate.Equal(wantDate) {
		t.Fatalf("ReleaseDate = %v, want %v", meta.ReleaseDate, wantDate)
	}

	if count := requests.Load(); count > 3 {
		t.Fatalf("expected coalesced metadata reads to use <= 3 HTTP requests, got %d", count)
	}
}

func TestReadVersionMetadataFromIPAFallsBackWhenSuffixRangeRejected(t *testing.T) {
	plistData, err := plist.Marshal(map[string]any{
		"bundleShortVersionString": "3.2.1",
	}, plist.BinaryFormat)
	if err != nil {
		t.Fatal(err)
	}

	modTime := time.Date(2024, 6, 15, 10, 20, 30, 0, time.UTC)
	ipaBytes := buildLargeTestIPA(t, zip.Store, plistData, modTime)

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		rng := r.Header.Get("Range")
		if strings.HasPrefix(rng, "bytes=-") {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", len(ipaBytes)))
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if !strings.HasPrefix(rng, "bytes=") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(ipaBytes)
			return
		}

		spec := strings.TrimPrefix(rng, "bytes=")
		parts := strings.SplitN(spec, "-", 2)
		start, err1 := strconv.ParseInt(parts[0], 10, 64)
		end, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil || start < 0 || end < start || end >= int64(len(ipaBytes)) {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(ipaBytes)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(ipaBytes[start : end+1])
	}))
	defer server.Close()

	store := &appstore{httpClient: passthroughClient{}}
	meta, size, err := store.readVersionMetadataFromIPA(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("fallback readVersionMetadataFromIPA failed: %v", err)
	}

	if size != int64(len(ipaBytes)) || meta.DisplayVersion != "3.2.1" {
		t.Fatalf("unexpected metadata: meta=%+v size=%d", meta, size)
	}
	if !meta.ReleaseDate.Equal(modTime) {
		t.Fatalf("ReleaseDate = %v, want fallback zip Modified %v", meta.ReleaseDate, modTime)
	}

	if count := requests.Load(); count > 5 {
		t.Fatalf("expected fallback reads to stay bounded (<= 5 HTTP requests), got %d", count)
	}
}

func TestReadVersionMetadataFromIPAHonorsCanceledContext(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	store := &appstore{httpClient: passthroughClient{}}
	_, _, err := store.readVersionMetadataFromIPA(ctx, server.URL)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if requests.Load() != 0 {
		t.Fatalf("canceled context must abort before issuing network requests, got %d", requests.Load())
	}
}

type fakeBagClient struct {
	calls int
}

func (f *fakeBagClient) Send(apphttp.Request) (apphttp.Result[bagResult], error) {
	f.calls++
	return apphttp.Result[bagResult]{
		StatusCode: http.StatusOK,
		Data: bagResult{
			URLBag: urlBag{
				EntDownloadEndpoint: "https://downloaddispatch.itunes.apple.com/WebObjects/DownloadDispatch.woa/wa/ent/download",
				RedownloadEndpoint:  "https://downloaddispatch.itunes.apple.com/r/redownload",
			},
		},
	}, nil
}

func (*fakeBagClient) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("unexpected Do")
}

func (*fakeBagClient) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func TestFetchURLBagUsesInMemoryTTLCache(t *testing.T) {
	client := &fakeBagClient{}
	store := &appstore{bagClient: client}

	bag1, err := store.fetchURLBag(context.Background(), "AABBCCDDEEFF")
	if err != nil {
		t.Fatal(err)
	}
	bag2, err := store.fetchURLBag(context.Background(), "AABBCCDDEEFF")
	if err != nil {
		t.Fatal(err)
	}

	if client.calls != 1 {
		t.Fatalf("expected cached urlBag to trigger 1 request, got %d", client.calls)
	}
	if bag1 != bag2 || bag1.EntDownloadEndpoint == "" {
		t.Fatalf("unexpected cached bags: %+v vs %+v", bag1, bag2)
	}
}
