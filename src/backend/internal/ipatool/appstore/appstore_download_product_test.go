package appstore

import (
	"testing"

	"AppleVault/src/backend/internal/ipatool/http"
)

func TestDownloadProductRequestUsesEndpointVersionKey(t *testing.T) {
	store := &appstore{}
	account := Account{DirectoryServicesID: "dsid", PasswordToken: "token", StoreFront: "143465-2,34"}
	app := App{ID: 414478124}

	tests := []struct {
		name     string
		endpoint downloadProductEndpoint
		wantKey  string
	}{
		{name: "volume store", endpoint: volumeStoreEndpoint(account), wantKey: "externalVersionId"},
		{name: "redownload", endpoint: downloadProductEndpoint{URL: "https://downloaddispatch.itunes.apple.com/r/redownload", ExternalVersionIDKey: "appExtVrsId"}, wantKey: "appExtVrsId"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := store.downloadProductRequest(account, app, "AABBCC", "890654806", tt.endpoint)
			payload, ok := req.Payload.(*http.XMLPayload)
			if !ok {
				t.Fatalf("unexpected payload type %T", req.Payload)
			}
			if got := payload.Content[tt.wantKey]; got != "890654806" {
				t.Fatalf("%s = %v, want 890654806", tt.wantKey, got)
			}
		})
	}
}

func TestDownloadDispatchEndpoints(t *testing.T) {
	redownload, err := newRedownloadEndpoint("https://downloaddispatch.itunes.apple.com/r/redownload")
	if err != nil || redownload.ExternalVersionIDKey != "appExtVrsId" {
		t.Fatalf("redownload endpoint: endpoint=%+v err=%v", redownload, err)
	}

	update, err := newUpdateEndpoint("https://downloaddispatch.itunes.apple.com/up/updateProduct")
	if err != nil || update.ExternalVersionIDKey != "appExtVrsId" {
		t.Fatalf("update endpoint: endpoint=%+v err=%v", update, err)
	}

	if _, err := newRedownloadEndpoint("https://example.com/r/redownload"); err == nil {
		t.Fatal("expected an untrusted redownload endpoint to be rejected")
	}
}
