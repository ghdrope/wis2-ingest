package utils

import "os"

// IsDebug detects debug mode.
func IsDebug() bool {
	return os.Getenv("DEBUG") == "true"
}
