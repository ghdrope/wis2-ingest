package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureDirs verifies single, multiple, nested and existing directories
// are all created without error.
func TestEnsureDirs(t *testing.T) {
	base := t.TempDir()
	existing := filepath.Join(base, "existing")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	dirs := []string{
		filepath.Join(base, "dir1"),
		filepath.Join(base, "dir2"),
		filepath.Join(base, "nested", "a", "b"),
		existing,
	}

	if err := EnsureDirs(dirs...); err != nil {
		t.Fatalf("EnsureDirs() error: %v", err)
	}
	for _, d := range dirs {
		if info, err := os.Stat(d); err != nil || !info.IsDir() {
			t.Errorf("%s was not created as a directory (err=%v)", d, err)
		}
	}
}

// TestEnsureDirsNoArgs verifies calling without paths is a no-op.
func TestEnsureDirsNoArgs(t *testing.T) {
	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() error: %v", err)
	}
}

// TestEnsureDirsError verifies the failing path is reported and later
// paths are not created.
func TestEnsureDirsError(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(file, "sub")
	after := filepath.Join(base, "after")

	err := EnsureDirs(bad, after)
	if err == nil || !strings.Contains(err.Error(), "failed to create directory "+bad) {
		t.Fatalf("EnsureDirs() error = %v; want failure for %s", err, bad)
	}
	if _, err := os.Stat(after); !os.IsNotExist(err) {
		t.Errorf("%s must not be created after a failure", after)
	}
}

// TestCheckDirWritable verifies each outcome of the writability check.
func TestCheckDirWritable(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "file.txt")
	if err := os.WriteFile(file, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "writable directory", path: base},
		{name: "does not exist", path: filepath.Join(base, "missing"), wantErr: "directory does not exist"},
		{name: "is a file", path: file, wantErr: "path exists but is not a directory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDirWritable(tt.path)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckDirWritable() error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("CheckDirWritable() error = %v; want containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestCheckDirWritableLeavesNoFiles verifies the probe file is removed.
func TestCheckDirWritableLeavesNoFiles(t *testing.T) {
	dir := t.TempDir()
	if err := CheckDirWritable(dir); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("probe file left behind: %v", entries)
	}
}

// TestCheckDirWritableReadOnly verifies a read-only directory is rejected.
func TestCheckDirWritableReadOnly(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write to read-only directories")
	}
	dir := filepath.Join(t.TempDir(), "ro")
	if err := os.Mkdir(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := CheckDirWritable(dir)
	if err == nil || !strings.Contains(err.Error(), "directory is not writable") {
		t.Fatalf("CheckDirWritable() error = %v; want not writable", err)
	}
}
