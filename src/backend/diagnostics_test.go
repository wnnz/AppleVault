package backend

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDiagnosticLogRedactsSecretsAndRemainsValidJSON(t *testing.T) {
	a := &App{dataDirOverride: t.TempDir()}
	password := `secret&password"with\escapes`
	a.addDiagnosticSecrets(password, "123456", "session-token", "test@example.com")
	a.writeDiagnostic("test.failure", map[string]interface{}{
		"error":  "login for test@example.com: " + password,
		"nested": map[string]interface{}{"html": html.EscapeString(password), "query": url.QueryEscape(password)},
		"values": []string{"session-token", "123456"}, "status": 404,
		"download_error": `Get "https://cdn.apple.com/a.ipa?token=unregistered-signed-token": denied`,
	})
	data, err := os.ReadFile(a.diagnosticLogPath())
	if err != nil {
		t.Fatal(err)
	}
	var row map[string]interface{}
	if err := json.Unmarshal(data, &row); err != nil {
		t.Fatalf("redaction corrupted JSON: %v", err)
	}
	for _, secret := range []string{"secret", "session-token", "123456", "test@example.com", url.QueryEscape(password), "unregistered-signed-token"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("diagnostic log exposed %q", secret)
		}
	}
	if !strings.Contains(string(data), "[redacted]") || !strings.Contains(string(data), "404") {
		t.Fatal("redaction must preserve useful status information")
	}
}

func TestConcurrentDiagnosticWritesProduceCompleteLines(t *testing.T) {
	a := &App{dataDirOverride: t.TempDir()}
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			a.writeDiagnostic("test.concurrent", map[string]interface{}{"sequence": i})
		}(i)
	}
	group.Wait()
	data, err := os.ReadFile(a.diagnosticLogPath())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 20 {
		t.Fatalf("lost log events: got %d", len(lines))
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatal("concurrent writes interleaved JSON records")
		}
	}
}

func TestCleanupDiagnosticLogs(t *testing.T) {
	a := &App{dataDirOverride: t.TempDir()}
	logsDir := filepath.Join(a.getDataDir(), "logs")
	if err := os.MkdirAll(logsDir, 0700); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	oldDate := now.AddDate(0, 0, -30).Format("2006-01-02")
	todayDate := now.Format("2006-01-02")

	oldFile := filepath.Join(logsDir, "appstore-"+oldDate+".jsonl")
	oldRotated := filepath.Join(logsDir, "appstore-"+oldDate+".jsonl.1")
	todayFile := filepath.Join(logsDir, "appstore-"+todayDate+".jsonl")
	for _, p := range []string{oldFile, oldRotated, todayFile} {
		if err := os.WriteFile(p, []byte("{}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	a.cleanupDiagnosticLogs(14*24*time.Hour, now)

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("expected old log file to be removed, err=%v", err)
	}
	if _, err := os.Stat(oldRotated); !os.IsNotExist(err) {
		t.Fatalf("expected old rotated log file to be removed, err=%v", err)
	}
	if _, err := os.Stat(todayFile); err != nil {
		t.Fatalf("expected today's log file to be kept, err=%v", err)
	}
}

func TestDiagnosticSecretsCapped(t *testing.T) {
	a := &App{dataDirOverride: t.TempDir()}
	for i := 0; i < 600; i++ {
		a.addDiagnosticSecrets(fmt.Sprintf("secret-%d", i))
	}

	a.diagnosticMu.Lock()
	defer a.diagnosticMu.Unlock()
	if len(a.diagnosticSecrets) > maxDiagnosticSecrets {
		t.Fatalf("expected at most %d diagnostic secrets, got %d", maxDiagnosticSecrets, len(a.diagnosticSecrets))
	}
	foundLast := false
	for _, s := range a.diagnosticSecrets {
		if s == "secret-599" {
			foundLast = true
			break
		}
	}
	if !foundLast {
		t.Fatal("expected latest secret-599 to be retained")
	}
}
