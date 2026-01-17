package version

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want bool
	}{
		{"0.1.0", "0.1.1", true},
		{"0.1.1", "0.1.0", false},
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "1.0.0", true},
		{"1.0.0", "0.9.9", false},
		{"0.1", "0.1.1", true},
		{"0.1.1", "0.1", false},
		{"0.1.0", "0.1.0.1", true},
	}

	for _, tt := range tests {
		if got := isNewer(tt.v1, tt.v2); got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}
