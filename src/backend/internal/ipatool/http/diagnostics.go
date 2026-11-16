package http

import "net/url"

// DiagnosticURL deliberately omits credentials, query strings and fragments.
func DiagnosticURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "[invalid URL]"
	}
	return parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
}