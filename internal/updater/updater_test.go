package updater

import (
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		{"0.2.0", "0.1.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.1.1", "0.1.0", true},
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.2.0", false},
		{"0.1.0", "0.1.1", false},
		{"", "0.1.0", false},
		{"0.1.0", "", false},
	}

	for _, tc := range tests {
		result := isNewer(tc.latest, tc.current)
		if result != tc.expected {
			t.Errorf("isNewer(%q, %q) = %v, expected %v", tc.latest, tc.current, result, tc.expected)
		}
	}
}
