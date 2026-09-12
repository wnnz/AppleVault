package http

import (
	"context"
	gohttp "net/http"
	"net/url"
	"testing"
)

type testCookieJar struct{}

func (testCookieJar) Cookies(*url.URL) []*gohttp.Cookie     { return nil }
func (testCookieJar) SetCookies(*url.URL, []*gohttp.Cookie) {}
func (testCookieJar) Save() error                           { return nil }

func TestNewRequestUsesConfiguredContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient[interface{}](Args{CookieJar: testCookieJar{}, Context: ctx})
	request, err := client.NewRequest(gohttp.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	if request.Context().Err() != context.Canceled {
		t.Fatalf("expected canceled request context, got %v", request.Context().Err())
	}
}
