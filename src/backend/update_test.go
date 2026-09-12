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
