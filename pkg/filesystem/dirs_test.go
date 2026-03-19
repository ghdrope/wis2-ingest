package filesystem

// Test Includes:
// - TestEnsureDirs_SingleDir           : Verifies a single directory is created.
// - TestEnsureDirs_MultipleDirs        : Verifies multiple directories are created at once.
// - TestEnsureDirs_ExistingDir         : Verifies function does not fail if directory already exists.
// - TestCheckDirWritable_Success       : Verifies a valid writable directory passes the check.
// - TestCheckDirWritable_NotExist      : Verifies error when directory does not exist.
// - TestCheckDirWritable_NotDirectory  : Verifies error when path is a file.

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEnsureDirs_SingleDir verifies a single directory is created.
func TestEnsureDirs_SingleDir(t *testing.T) {
	temp := t.TempDir()
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

// TestEnsureDirs_MultipleDirs verifies multiple directories are created at once.
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

// TestEnsureDirs_ExistingDir verifies function does not fail if directory already exists.
func TestEnsureDirs_ExistingDir(t *testing.T) {
	temp := t.TempDir()
	dir := filepath.Join(temp, "existing")

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = EnsureDirs(dir)
	if err != nil {
		t.Fatalf("expected no error when directory already exists, got %v", err)
	}
}

// TestCheckDirWritable_Success verifies a valid writable directory passes the check.
func TestCheckDirWritable_Success(t *testing.T) {
	temp := t.TempDir()

	err := CheckDirWritable(temp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestCheckDirWritable_NotExist verifies error when directory does not exist.
func TestCheckDirWritable_NotExist(t *testing.T) {
	path := filepath.Join(os.TempDir(), "non-existent-dir-xyz")

	err := CheckDirWritable(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestCheckDirWritable_NotDirectory verifies error when path is a file.
func TestCheckDirWritable_NotDirectory(t *testing.T) {
	temp := t.TempDir()
	file := filepath.Join(temp, "file.txt")

	if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err := CheckDirWritable(file)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
