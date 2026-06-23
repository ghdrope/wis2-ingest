package filesystem

import (
	"fmt"
	"os"
)

// EnsureDirs ensures that all provided directory paths exist,
// creating them when necessary.
func EnsureDirs(paths ...string) error {
	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", p, err)
		}
	}
	return nil
}

// CheckDirWritable verifies that a directory exists and that
// files can be created within it.
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
