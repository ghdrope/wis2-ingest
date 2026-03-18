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
