package utils

import (
	"os"
	"testing"
)

// TestIsDebug validates the behavior of IsDebug under different
// environment variable configurations.
func TestIsDebug(t *testing.T) {

	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "debug enabled when DEBUG=true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "debug disabled when DEBUG=false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "debug disabled when DEBUG is empty",
			envValue: "",
			expected: false,
		},
		{
			name:     "debug disabled when DEUBG has invalid value",
			envValue: "yes",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			original := os.Getenv("DEBUG")

			t.Cleanup(func() {
				_ = os.Setenv("DEBUG", original)
			})

			// Set test-specific environment value.
			_ = os.Setenv("DEBUG", tt.envValue)

			got := IsDebug()

			if got != tt.expected {
				t.Errorf(
					"IsDebug() = %v, want %v (DEBUG=%q)",
					got,
					tt.expected,
					tt.envValue,
				)
			}
		})
	}
}
