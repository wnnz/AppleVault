package store

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransferExecutesDirectRangeAndStripsCrossHostSensitiveHeaders(t *testing.T) {
	payload := []byte("0123456789abcdef")

	var targetHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "a.cdn.test":
			w.Header().Set("Location", "http://b.cdn.test/blob")
			w.WriteHeader(http.StatusFound)
		case "b.cdn.test":
			targetHeaders = r.Header.Clone()
			w.Header().Set("Content-Range", "bytes 4-7/16")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(payload[4:8])
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	base.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", server.Listener.Addr().String())
	}

	// Pass nil macAddress: if transfer wrongly routed through the Unicorn VM execute
	// path, b.hardware() would immediately fail instead of completing the request.
	backend := New(nil, nil, context.Background(), base)
	exec := backend.Client("transfer")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://a.cdn.test/start", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", "bytes=4-7")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("X-Token", "secret-token")
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Cookie", "session=abc")

	resp, err := exec.Do(req)
	if err != nil {
		t.Fatalf("transfer Do failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "4567" {
		t.Fatalf("body = %q, want 4567", body)
	}

	if targetHeaders.Get("Range") != "bytes=4-7" {
		t.Fatalf("redirected Range header = %q", targetHeaders.Get("Range"))
	}
	for _, stripped := range []string{"Authorization", "Cookie", "X-Token", "X-Dsid", "Icloud-Dsid"} {
		if got := targetHeaders.Get(stripped); got != "" {
			t.Fatalf("expected cross-host sensitive header %s to be stripped, got %q", stripped, got)
		}
	}
}
