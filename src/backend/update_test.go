package backend

import "testing"

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"1.3.1", "1.3.0", true},
		{"v2.0.0", "1.9.9", true},
		{"1.3.0", "1.3.0", false},
		{"1.2.9", "1.3.0", false},
	}
	for _, test := range tests {
		if got := isVersionNewer(test.candidate, test.current); got != test.want {
			t.Errorf("isVersionNewer(%q, %q) = %v", test.candidate, test.current, got)
		}
	}
}

func TestDisplayVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.5.0", "1.5"},
		{"1.10.0", "1.10"},
		{"1.5.2", "1.5.2"},
	}
	for _, tc := range tests {
		if got := displayVersion(tc.input); got != tc.want {
			t.Errorf("displayVersion(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
