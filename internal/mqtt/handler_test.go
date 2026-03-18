package mqtt

// Test Includes:
// - TestMessageHandler_StoresBufrFiles : Verifies that .bufr payloads are downloaded and stored correctly.
// - TestMessageHandler_SkipsNonBufr    : Verifies non-.bufr files are ignored.
// - TestMessageHandler_InvalidJSON      : Verifies invalid JSON payloads are handled gracefully.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"wis2-ingest/internal/config"

	"wis2-ingest/pkg/testhelper"

	"github.com/akuity/kargo/pkg/logging"
)

// MqttMessage implements the minimal interface Msg for testing.
type MqttMessage struct {
	TopicField   string
	PayloadField []byte
}

// TestMessageHandler_StoresBufrFiles ensures that canonical .bufr files
// are downloaded and stored in the configured output directory.
func TestMessageHandler_StoresBufrFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Start mock HTTP server to serve .bufr file
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("bufr content")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	cfg := &config.IngestOptions{OutputDir: tempDir}
	logger, err := logging.NewLogger(logging.InfoLevel, logging.DefaultFormat)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	client := &Client{cfg: cfg, logger: logger}

	// Prepare payload with one .bufr file and one non-.bufr file
	payload := Payload{
		Links: []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		}{
			{Href: ts.URL + "/file.bufr", Rel: "canonical"},
			{Href: ts.URL + "/file.txt", Rel: "canonical"},
		},
	}
	data, _ := json.Marshal(payload)

	// Use testhelper MqttMessage (implements minimal interface for handler)
	msg := testhelper.MqttMessage{TopicField: "test", PayloadField: data}

	client.messageHandler(nil, msg)

	// Verify .bufr file exists in correct directory
	dir := tempDir
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read output directory: %v", err)
	}

	foundBufr := false
	foundTxt := false
	for _, f := range files {
		switch filepath.Ext(f.Name()) {
		case ".bufr":
			foundBufr = true
		case ".txt":
			foundTxt = true
		}
	}

	if !foundBufr {
		t.Errorf("expected .bufr file to be stored")
	}
	if foundTxt {
		t.Errorf("non-.bufr file should not be stored")
	}
}

// TestMessageHandler_InvalidJSON ensures that invalid JSON payloads are handled gracefully
// and do not panic.
func TestMessageHandler_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.IngestOptions{OutputDir: tempDir}
	logger, err := logging.NewLogger(logging.InfoLevel, logging.DefaultFormat)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	client := &Client{cfg: cfg, logger: logger}

	// Invalid JSON payload
	msg := testhelper.MqttMessage{TopicField: "test", PayloadField: []byte("{invalid json}")}

	// Should not panic
	client.messageHandler(nil, msg)
}
