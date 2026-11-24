package validate

import (
	"bytes"
	"io"
	"os"
)

// IsBUFR performs a BUFR validation.
//
// The validation checks whether the downloaded content contains the BUFR
// start marker followed by the "7777" end marker. This is intended to
// distinguish BUFR messages from HTML error pages and other unexpected
// content, including pages that only mention "BUFR". The start marker may
// come after a GTS bulletin header.
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

	start := bytes.Index(data, []byte("BUFR"))
	if start < 0 {
		return false, nil
	}

	return bytes.Contains(data[start+len("BUFR"):], []byte("7777")), nil
}
