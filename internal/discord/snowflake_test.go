package discord

import "testing"

func TestIsSnowflake(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"123456789012345678", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"123abc", false},
		{"01234", false},                      // leading zero
		{"-123", false},                       // sign
		{"99999999999999999999999999", false}, // overflows uint64
		{"18446744073709551615", true},        // max uint64
		{"18446744073709551616", false},       // max uint64 + 1
	}
	for _, tt := range tests {
		if got := IsSnowflake(tt.in); got != tt.want {
			t.Errorf("IsSnowflake(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
