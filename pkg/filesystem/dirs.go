package filesystem

import (
	"fmt"
	"os"
)

// EnsureDirs creates directories if they do not already exist.
// It accepts one or more paths and ensures each directory exists.
func EnsureDirs(paths ...string) error {
	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", p, err)
		}
	}
	return nil
}

// CheckDirWritable verifies that the given path exists and is writable.
// It does NOT create the directory. Returns error if not accessible.
func CheckDirWritable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("failed to access directory %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}

	// Try creating a temp file to check writability
	f, err := os.CreateTemp(path, ".wis2-test-*")
	if err != nil {
		return fmt.Errorf("directory is not writable: %s", path)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Remove(f.Name()); err != nil {
		return fmt.Errorf("failed to remove temp file: %w", err)
	}

	return nil
}
