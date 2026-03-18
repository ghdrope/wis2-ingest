package filesystem

// Test Includes:
// - TestEnsureDirs_SingleDir       : Verifies a single directory is created.
// - TestEnsureDirs_MultipleDirs    : Verifies multiple directories are created at once.
// - TestEnsureDirs_ExistingDir     : Verifies function does not fail if directory already exists.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnsureDirs_SingleDir verifies that EnsureDirs creates a single directory.
func TestEnsureDirs_SingleDir(t *testing.T) {
	temp := t.TempDir() // Go provides a temporary test directory
	dirPath := filepath.Join(temp, "single")

	err := EnsureDirs(dirPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", dirPath)
	}
}

// TestEnsureDirs_MultipleDirs verifies that EnsureDirs creates multiple directories.
func TestEnsureDirs_MultipleDirs(t *testing.T) {
	temp := t.TempDir()
	dir1 := filepath.Join(temp, "dir1")
	dir2 := filepath.Join(temp, "dir2")
	dir3 := filepath.Join(temp, "dir3")

	err := EnsureDirs(dir1, dir2, dir3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, d := range []string{dir1, dir2, dir3} {
		info, err := os.Stat(d)
		if err != nil {
			t.Fatalf("directory %s was not created: %v", d, err)
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}
}

// TestEnsureDirs_ExistingDir verifies that EnsureDirs does not fail if the directory already exists.
func TestEnsureDirs_ExistingDir(t *testing.T) {
	temp := t.TempDir()
	dir := filepath.Join(temp, "existing")

	// create dir first
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// call EnsureDirs on the same directory
	err = EnsureDirs(dir)
	if err != nil {
		t.Fatalf("expected no error when directory already exists, got %v", err)
	}
}
