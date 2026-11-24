package utils

import (
	"os"
	"testing"
)

// TestIsDebug verifies only DEBUG=true enables debug mode.
func TestIsDebug(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"", false},
		{"yes", false},
		{"TRUE", false}, // exact match only
		{"1", false},
	}
	for _, tt := range tests {
		t.Run("DEBUG="+tt.value, func(t *testing.T) {
			t.Setenv("DEBUG", tt.value)
			if got := IsDebug(); got != tt.want {
				t.Errorf("IsDebug() = %v; want %v (DEBUG=%q)", got, tt.want, tt.value)
			}
		})
	}
}

// TestIsDebugUnset verifies debug mode is off when DEBUG is not set at all.
func TestIsDebugUnset(t *testing.T) {
	t.Setenv("DEBUG", "") // registers restore of the original value
	if err := os.Unsetenv("DEBUG"); err != nil {
		t.Fatal(err)
	}
	if IsDebug() {
		t.Error("IsDebug() = true; want false when DEBUG is unset")
	}
}
