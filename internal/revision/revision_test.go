package revision

// Test Includes:
// - TestAppendRuntimeInfo_AppendsCorrectly     : Ensures runtime info is appended correctly after Docker image version line.
// - TestAppendRuntimeInfo_MultipleTopics      : Verifies that multiple topics are joined correctly as a comma-separated string.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wis2-ingest/internal/config"
)

// TestAppendRuntimeInfo_AppendsCorrectly ensures runtime info is appended correctly
// after Docker image version line.
func TestAppendRuntimeInfo_AppendsCorrectly(t *testing.T) {
	// Create a temporary directory for the revision file
	tempDir := t.TempDir()
	revFile := filepath.Join(tempDir, "wis2-ingest.rev")

	// Simulate Docker image version already written
	initialContent := "Docker Image Version: 1.0\n"
	if err := os.WriteFile(revFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create temp revision file: %v", err)
	}

	// Prepare a sample IngestOptions config
	cfg := &config.IngestOptions{
		Host:      "mqtt.example.com",
		Port:      "1883",
		Username:  "myuser",
		Topics:    []string{"topic1", "topic2"},
		OutputDir: "/data/wis2-ingest",
	}

	// Call the function to append runtime info
	if err := AppendRuntimeInfo(revFile, cfg); err != nil {
		t.Fatalf("AppendRuntimeInfo failed: %v", err)
	}

	// Read back the file contents
	data, err := os.ReadFile(revFile)
	if err != nil {
		t.Fatalf("failed to read revision file: %v", err)
	}
	content := string(data)

	// Verify that the original Docker image version line exists
	if !strings.Contains(content, initialContent) {
		t.Errorf("expected Docker image version line to remain, got: %s", content)
	}

	// Verify that runtime info was appended correctly
	expectedLines := []string{
		"Host: mqtt.example.com",
		"Port: 1883",
		"Username: myuser",
		"Topic: topic1,topic2",
		"Output Directory: /data/wis2-ingest",
	}

	for _, line := range expectedLines {
		if !strings.Contains(content, line) {
			t.Errorf("expected revision file to contain line '%s'", line)
		}
	}
}

// TestAppendRuntimeInfo_MultipleTopics verifies that multiple topics are joined
// correctly as a comma-separated string.
func TestAppendRuntimeInfo_MultipleTopics(t *testing.T) {
	tempDir := t.TempDir()
	revFile := filepath.Join(tempDir, "wis2-ingest.rev")

	// Initial Docker version
	if err := os.WriteFile(revFile, []byte("Docker Image Version: 2.0\n"), 0644); err != nil {
		t.Fatalf("failed to create temp revision file: %v", err)
	}

	cfg := &config.IngestOptions{
		Host:      "broker.local",
		Port:      "8883",
		Username:  "user123",
		Topics:    []string{"topicA", "topicB", "topicC"},
		OutputDir: "/tmp/output",
	}

	if err := AppendRuntimeInfo(revFile, cfg); err != nil {
		t.Fatalf("AppendRuntimeInfo failed: %v", err)
	}

	data, err := os.ReadFile(revFile)
	if err != nil {
		t.Fatalf("failed to read revision file: %v", err)
	}
	content := string(data)

	expectedTopicLine := "Topic: topicA,topicB,topicC"
	if !strings.Contains(content, expectedTopicLine) {
		t.Errorf("expected topic line '%s', got: %s", expectedTopicLine, content)
	}
}
