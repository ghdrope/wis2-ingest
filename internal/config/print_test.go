package config

import (
	"strings"
	"testing"
)

// TestPrintVerboseIncludesNonSensitiveFields verifies that
// non-sensitive configuration values are rendered.
func TestPrintVerboseIncludesNonSensitiveFields(t *testing.T) {

	opts := &IngestOptions{
		Host:      "localhost",
		Username:  "user",
		Password:  "secret",
		Topics:    []string{"topic"},
		OutputDir: "/tmp",
	}

	output := opts.PrintVerbose()

	if !strings.Contains(output, "localhost") {
		t.Fatal("expected host in output")
	}

	if !strings.Contains(output, "/tmp") {
		t.Fatal("expected output dir in output")
	}
}

// TestPrintVerboseOmitsSensitiveFields verifies that
// sensitive values are excluded from verbose output.
func TestPrintVerboseOmitsSensitiveFields(t *testing.T) {

	opts := &IngestOptions{
		Password: "super-secret",
	}

	output := opts.PrintVerbose()

	if strings.Contains(output, "super-secret") {
		t.Fatal("password should not be printed")
	}
}
