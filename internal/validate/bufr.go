package validate

import (
	"bytes"
	"io"
	"os"
)

// IsBUFR performs a BUFR validation.
//
// The validation checks whether the downloaded content contains
// the BUFR marker. This is intended to distinguish BUFR messages
// from HTML error pages and other unexpected content.
func IsBUFR(filePath string) (bool, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f.Close()
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		return false, err
	}

	return bytes.Contains(data, []byte("BUFR")), nil
}
