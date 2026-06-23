package config

import "testing"

// TestNewIngestOptions verifies that default values
// are correctly initialized.
func TestNewIngestOptions(t *testing.T) {

	opts := NewIngestOptions()

	if opts.Host != "globalbroker.meteo.fr" {
		t.Errorf("unexpected host: %s", opts.Host)
	}

	if opts.Username != "everyone" {
		t.Errorf("unexpected username: %s", opts.Username)
	}

	if opts.Password != "everyone" {
		t.Errorf("unexpected password")
	}
}
