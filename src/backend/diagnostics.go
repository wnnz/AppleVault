package backend

import (
	apihttp "AppleVault/src/backend/internal/ipatool/http"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	maxDiagnosticLogBytes = 10 << 20
	maxDiagnosticSecrets  = 512
)

var diagnosticURLPattern = regexp.MustCompile(`https?://[^\s"'<>]+`)

func (a *App) addDiagnosticSecrets(values ...string) {
	a.diagnosticMu.Lock()
	defer a.diagnosticMu.Unlock()
	for _, value := range values {
		if value == "" {
			continue
		}
		variants := []string{value, html.EscapeString(value), url.QueryEscape(value)}
		if encoded, err := json.Marshal(value); err == nil {
			variants = append(variants, string(encoded[1:len(encoded)-1]))
		}
		for _, variant := range variants {
			if variant == "" {
				continue
			}
			found := false
			for _, existing := range a.diagnosticSecrets {
				if existing == variant {
					found = true
					break
				}
			}
			if !found {
				a.diagnosticSecrets = append(a.diagnosticSecrets, variant)
			}
		}
	}
	if len(a.diagnosticSecrets) > maxDiagnosticSecrets {
		trimmed := make([]string, maxDiagnosticSecrets)
		copy(trimmed, a.diagnosticSecrets[len(a.diagnosticSecrets)-maxDiagnosticSecrets:])
		a.diagnosticSecrets = trimmed
	}
}

func (a *App) cleanupDiagnosticLogs(maxAge time.Duration, now time.Time) {
	if maxAge <= 0 {
		return
	}
	cutoff := now.Add(-maxAge)
	matches, err := filepath.Glob(filepath.Join(a.getDataDir(), "logs", "appstore-*.jsonl*"))
	if err != nil {
		return
	}
	for _, path := range matches {
		base := filepath.Base(path)
		var logTime time.Time
		trimmed := strings.TrimPrefix(base, "appstore-")
		if len(trimmed) >= 10 {
			if parsed, parseErr := time.ParseInLocation("2006-01-02", trimmed[:10], now.Location()); parseErr == nil {
				logTime = parsed
			}
		}
		if logTime.IsZero() {
			if info, statErr := os.Stat(path); statErr == nil {
				logTime = info.ModTime()
			}
		}
		if !logTime.IsZero() && logTime.Before(cutoff) {
			_ = os.Remove(path)
		}
	}
}

func (a *App) diagnosticLogPath() string {
	return filepath.Join(a.getDataDir(), "logs", "appstore-"+time.Now().Format("2006-01-02")+".jsonl")
}

func (a *App) writeDiagnostic(event string, fields map[string]interface{}) {
	a.diagnosticMu.Lock()
	defer a.diagnosticMu.Unlock()
	row := map[string]interface{}{"time": time.Now().Format(time.RFC3339Nano), "pid": os.Getpid(), "event": event, "fields": redactDiagnosticValue(fields, a.diagnosticSecrets)}
	data, err := json.Marshal(row)
	if err != nil {
		log.Printf("encode diagnostic log: %v", err)
		return
	}
	path := a.diagnosticLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		log.Printf("create diagnostic directory: %v", err)
		return
	}
	if info, err := os.Stat(path); err == nil && info.Size() >= maxDiagnosticLogBytes {
		_ = os.Remove(path + ".1")
		if err := os.Rename(path, path+".1"); err != nil {
			log.Printf("rotate diagnostic log: %v", err)
			return
		}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("open diagnostic log: %v", err)
		return
	}
	_, err = fmt.Fprintln(file, string(data))
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		log.Printf("write diagnostic log: %v", err)
	}
}

func redactDiagnosticValue(value interface{}, secrets []string) interface{} {
	switch value := value.(type) {
	case string:
		value = diagnosticURLPattern.ReplaceAllStringFunc(value, apihttp.DiagnosticURL)
		for _, secret := range secrets {
			value = strings.ReplaceAll(value, secret, "[redacted]")
		}
		return value
	case map[string]interface{}:
		copy := make(map[string]interface{}, len(value))
		for key, item := range value {
			copy[key] = redactDiagnosticValue(item, secrets)
		}
		return copy
	case []interface{}:
		copy := make([]interface{}, len(value))
		for i, item := range value {
			copy[i] = redactDiagnosticValue(item, secrets)
		}
		return copy
	case []string:
		copy := make([]interface{}, len(value))
		for i, item := range value {
			copy[i] = redactDiagnosticValue(item, secrets)
		}
		return copy
	default:
		return value
	}
}
