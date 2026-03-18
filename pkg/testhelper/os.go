package testhelper

import "os"

// SetEnv is a helper that sets an environment variable and returns a cleanup
// function to restore the previous value after the test completes.
func SetEnv(key, value string) func() {
	originalValue, hadOriginal := os.LookupEnv(key)
	_ = os.Setenv(key, value)

	return func() {
		if hadOriginal {
			_ = os.Setenv(key, originalValue)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

// UnsetEnv ensures a variable is removed and restores it after the test.
func UnsetEnv(key string) func() {
	originalValue, hadOriginal := os.LookupEnv(key)
	_ = os.Unsetenv(key)

	return func() {
		if hadOriginal {
			_ = os.Setenv(key, originalValue)
		}
	}
}
