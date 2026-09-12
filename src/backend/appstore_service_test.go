package backend

import (
	"testing"
	"time"

	"AppleVault/src/backend/internal/ipatool/appstore"
)

func TestMapStoreApp(t *testing.T) {
	purchasedAt := time.Date(2026, time.September, 12, 8, 30, 0, 0, time.UTC)
	result := mapStoreApp(appstore.App{
		ID: 123, BundleID: "com.example.app", Name: "Example", Version: "2.0",
		Price: 1.5, PurchaseDate: purchasedAt,
	})

	if result.ID != 123 || result.BundleID != "com.example.app" || result.DisplayPrice != "¥1.50" {
		t.Fatalf("unexpected mapped app: %#v", result)
	}
	if result.PurchaseDate != purchasedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected purchase date: %q", result.PurchaseDate)
	}
}

func TestParseStorePlatformAliases(t *testing.T) {
	platform, err := parseStorePlatform(" iOS ")
	if err != nil {
		t.Fatal(err)
	}
	if platform != appstore.PlatformIPhone {
		t.Fatalf("expected iphone platform, got %q", platform)
	}

	if _, err := parseStorePlatform("android"); err == nil {
		t.Fatal("expected unsupported platform to fail")
	}
}

func TestIsRetryableAppleResponse(t *testing.T) {
	for _, message := range []string{
		"unexpected response from Apple (HTTP 500)",
		"empty or non-plist body",
		"invalid response",
	} {
		if !isRetryableAppleResponse(assertionError(message)) {
			t.Fatalf("expected %q to be retryable", message)
		}
	}
	if isRetryableAppleResponse(assertionError("invalid credentials")) {
		t.Fatal("credentials error must not be retried as a transient response")
	}
}

type assertionError string

func (e assertionError) Error() string { return string(e) }
