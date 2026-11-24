package validate

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsBUFR verifies files are accepted only when they contain the BUFR
// start marker followed by the 7777 end marker. The start marker may come
// after a GTS bulletin header.
func TestIsBUFR(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "BUFR message", content: "BUFR\x00\x00\x00\x04body7777", want: true},
		{name: "marker after a bulletin header", content: "ISMD01 LFPW 010000\r\r\nBUFR...7777", want: true},
		{name: "html error page", content: "<html><body>404 not found</body></html>"},
		{name: "html page mentioning BUFR", content: "<html><body>BUFR file not found</body></html>"},
		{name: "end marker only before BUFR", content: "7777 BUFR truncated"},
		{name: "lowercase marker", content: "bufr"},
		{name: "empty file", content: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := IsBUFR(path)
			if err != nil {
				t.Fatalf("IsBUFR() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsBUFR(%q) = %v; want %v", tt.content, got, tt.want)
			}
		})
	}
}

// TestIsBUFRErrors verifies filesystem errors are returned with false.
func TestIsBUFRErrors(t *testing.T) {
	for name, path := range map[string]string{
		"missing file": filepath.Join(t.TempDir(), "missing"),
		"directory":    t.TempDir(),
	} {
		t.Run(name, func(t *testing.T) {
			ok, err := IsBUFR(path)
			if err == nil || ok {
				t.Errorf("IsBUFR(%s) = %v, %v; want false, error", path, ok, err)
			}
		})
	}
}
