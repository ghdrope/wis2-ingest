package utils

// Test Includes:
// - TestGetLogVars              : Verifies GetLogVars parses env vars and applies defaults on invalid values.
// - TestGetLogVars_Defaults     : Verifies default logging values are returned when env vars are unset.

import (
	"os"
	"testing"
	"wis2-ingest/pkg/testhelper"

	"github.com/akuity/kargo/pkg/logging"
)

// TestGetLogVars verifies that GetLogVars correctly parses environment variables
// and falls back to defaults when invalid values are provided.
func TestGetLogVars(t *testing.T) {
	tests := []struct {
		name           string
		logLevelEnv    string
		logFormatEnv   string
		expectedLevel  logging.Level
		expectedFormat logging.Format
	}{
		{
			name:           "valid level and format",
			logLevelEnv:    "debug",
			logFormatEnv:   "json",
			expectedLevel:  logging.DebugLevel,
			expectedFormat: logging.JSONFormat,
		},
		{
			name:           "invalid level defaults to info",
			logLevelEnv:    "invalid",
			logFormatEnv:   "json",
			expectedLevel:  logging.InfoLevel,
			expectedFormat: logging.JSONFormat,
		},
		{
			name:           "invalid format defaults to default format",
			logLevelEnv:    "info",
			logFormatEnv:   "invalid",
			expectedLevel:  logging.InfoLevel,
			expectedFormat: logging.DefaultFormat,
		},
		{
			name:           "both invalid values fallback to defaults",
			logLevelEnv:    "invalid",
			logFormatEnv:   "invalid",
			expectedLevel:  logging.InfoLevel,
			expectedFormat: logging.DefaultFormat,
		},
		{
			name:           "empty values fallback to defaults",
			logLevelEnv:    "",
			logFormatEnv:   "",
			expectedLevel:  logging.InfoLevel,
			expectedFormat: logging.DefaultFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure tests do not run in parallel due to shared env vars
			// (environment variables are process-wide)

			cleanupLevel := testhelper.SetEnv(logging.LogLevelEnvVar, tt.logLevelEnv)
			defer cleanupLevel()

			cleanupFormat := testhelper.SetEnv(logging.LogFormatEnvVar, tt.logFormatEnv)
			defer cleanupFormat()

			level, format := GetLogVars()

			if level != tt.expectedLevel {
				t.Errorf("expected level %v, got %v", tt.expectedLevel, level)
			}

			if format != tt.expectedFormat {
				t.Errorf("expected format %v, got %v", tt.expectedFormat, format)
			}
		})
	}
}

// TestGetLogVars_Defaults verifies behavior when environment variables are unset.
func TestGetLogVars_Defaults(t *testing.T) {
	// Unset environment variables explicitly
	_ = os.Unsetenv(logging.LogLevelEnvVar)
	_ = os.Unsetenv(logging.LogFormatEnvVar)

	level, format := GetLogVars()

	if level != logging.InfoLevel {
		t.Errorf("expected default level %v, got %v", logging.InfoLevel, level)
	}

	if format != logging.DefaultFormat {
		t.Errorf("expected default format %v, got %v", logging.DefaultFormat, format)
	}
}
