package version

import (
	"strings"
	"testing"
)

// Test Includes:
// - TestGet                     : Verifies Get() returns a non-empty Info struct with correct defaults.
// - TestInfoString               : Verifies Info.String() outputs all fields and includes expected labels.
// - TestInfoShort                : Verifies Info.Short() returns only the Version field.
// - TestStringHelper             : Verifies String() helper returns the same output as Info.String().
// - TestShortHelper              : Verifies Short() helper returns the same output as Info.Short().
// - TestStringReflectAllFields   : Verifies that the reflect-based String() includes all struct fields.

// TestGet verifies Get() returns a non-empty Info struct with correct defaults.
func TestGet(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Errorf("Get() returned empty Version")
	}
	if info.GitCommit == "" {
		t.Errorf("Get() returned empty GitCommit")
	}
	if info.BuildDate == "" {
		t.Errorf("Get() returned empty BuildDate")
	}
	if info.GoVersion == "" {
		t.Errorf("Get() returned empty GoVersion")
	}
	if info.Platform == "" {
		t.Errorf("Get() returned empty Platform")
	}
}

// TestInfoString verifies Info.String() outputs all fields and includesexpected labels.
func TestInfoString(t *testing.T) {
	info := Get()
	out := info.String()

	fields := []string{"Version", "GitCommit", "BuildDate", "GoVersion", "Platform"}
	for _, f := range fields {
		if !strings.Contains(out, f) {
			t.Errorf("Info.String() output missing field %q", f)
		}
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			t.Errorf("Info.String() line %q missing colon", line)
		}
	}
}

// TestInfoShort verifies Info.Short() returns only the Version field.
func TestInfoShort(t *testing.T) {
	info := Get()
	short := info.Short()

	if !strings.Contains(short, info.Version) {
		t.Errorf("Info.Short() output %q does not contain Version %q", short, info.Version)
	}
}

// TestStringHelper verifies String() helper returns the same output as Info.String().
func TestStringHelper(t *testing.T) {
	info := Get()
	if got, want := String(), info.String(); got != want {
		t.Errorf("String() helper = %q; want %q", got, want)
	}
}

// TestShortHelper verifies Short() helper returns the same output as Info.Short().
func TestShortHelper(t *testing.T) {
	info := Get()
	if got, want := Short(), info.Short(); got != want {
		t.Errorf("Short() helper = %q; want %q", got, want)
	}
}

// TestStringReflectAllFields verifies that the reflect-based String() includes
// all struct fields.
func TestStringReflectAllFields(t *testing.T) {
	info := Info{
		Version:   "v1.2.3",
		GitCommit: "abc123",
		BuildDate: "2026-03-19T10:00:00Z",
		GoVersion: "go1.26",
		Platform:  "linux/amd64",
	}

	out := info.String()

	values := []string{"v1.2.3", "abc123", "2026-03-19T10:00:00Z", "go1.26", "linux/amd64"}
	for _, val := range values {
		if !strings.Contains(out, val) {
			t.Errorf("Info.String() missing value %q", val)
		}
	}
}
