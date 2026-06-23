package validate

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsBUFR_ValidFile verifies that a file containing the BUFR
// marker is correctly identified as a valid BUFR file.
func TestIsBUFR_ValidFile(t *testing.T) {

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "valid.bufr")

	content := []byte("some binary header ... BUFR ... trailing data")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ok, err := IsBUFR(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ok {
		t.Error("expected file to be valid BUFR")
	}
}

// TestIsBUFR_InvalidFile verifies that a file without the BUFR
// marker is correctly rejected.
func TestIsBUFR_InvalidFile(t *testing.T) {

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.bufr")

	content := []byte("this is just an html error page")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ok, err := IsBUFR(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ok {
		t.Error("expected file to be invalid BUFR")
	}
}

// TestIsBUFR_MissingFile verifies that a missing file
// returns an error from the filesystem.
func TestIsBUFR_MissingFile(t *testing.T) {

	path := filepath.Join(t.TempDir(), "does-not-exist.bufr")

	ok, err := IsBUFR(path)
	if err == nil {
		t.Fatal("expected error for missing file")
	}

	if ok {
		t.Error("expected false result for missing file")
	}
}
