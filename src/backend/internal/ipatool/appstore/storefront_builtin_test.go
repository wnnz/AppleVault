package appstore

import (
	"errors"
	"io"
	"net/http"
	"testing"

	apphttp "AppleVault/src/backend/internal/ipatool/http"
)

// storefrontStubClient is a minimal http.Client[[]byte] stub; only Send is
// used by storefront discovery.
type storefrontStubClient struct {
	result apphttp.Result[[]byte]
	err    error
}

func (c storefrontStubClient) Send(apphttp.Request) (apphttp.Result[[]byte], error) {
	return c.result, c.err
}

func (storefrontStubClient) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("unexpected direct request")
}

func (storefrontStubClient) NewRequest(string, string, io.Reader) (*http.Request, error) {
	return nil, errors.New("unexpected request construction")
}

func TestStorefrontFallsBackToBuiltinMap(t *testing.T) {
	store := &appstore{
		storefrontClient: storefrontStubClient{result: apphttp.Result[[]byte]{StatusCode: http.StatusNotFound}},
	}

	countryCode, err := store.countryCodeFromStoreFront("143465-19,29")
	if err != nil {
		t.Fatal(err)
	}

	if countryCode != "CN" {
		t.Fatalf("expected CN, got %s", countryCode)
	}

	countryCode, err = store.countryCodeFromStoreFront("143441-1,29")
	if err != nil {
		t.Fatal(err)
	}

	if countryCode != "US" {
		t.Fatalf("expected US, got %s", countryCode)
	}
}
