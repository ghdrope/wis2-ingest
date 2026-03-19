package config

// Test Includes:
// - TestIngestOptions_Load_Success           : Verifies valid environment variables populate IngestOptions correctly.
// - TestIngestOptions_Load_Defaults          : Ensures defaults are applied when optional env vars are missing.
// - TestIngestOptions_Load_MissingTopic      : Checks error is returned when WIS2_MQTT_TOPIC is not defined.
// - TestIngestOptions_Load_MissingOutputDir  : Checks error is returned when WIS2_OUTPUT_DIRECTORY is not defined.

import (
	"reflect"
	"strings"
	"testing"
	"wis2-ingest/pkg/testhelper"
)

// TestIngestOptions_Load_Success verifies that valid environment variables
// are correctly parsed into the IngestOptions struct.
func TestIngestOptions_Load_Success(t *testing.T) {
	cleanup := []func(){
		testhelper.SetEnv("WIS2_MQTT_HOST", "test-host"),
		testhelper.SetEnv("WIS2_MQTT_PORT", "1234"),
		testhelper.SetEnv("WIS2_MQTT_USERNAME", "user"),
		testhelper.SetEnv("WIS2_MQTT_PASSWORD", "pass"),
		testhelper.SetEnv("WIS2_MQTT_TOPIC", "topic1; topic2 ;topic3"),
		testhelper.SetEnv("WIS2_OUTPUT_DIRECTORY", "/tmp/output"),
	}
	defer func() {
		for _, fn := range cleanup {
			fn()
		}
	}()

	opts := NewIngestOptions()
	err := opts.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedTopics := []string{"topic1", "topic2", "topic3"}

	if opts.Host != "test-host" {
		t.Errorf("expected Host %q, got %q", "test-host", opts.Host)
	}
	if opts.Port != "1234" {
		t.Errorf("expected Port %q, got %q", "1234", opts.Port)
	}
	if opts.Username != "user" {
		t.Errorf("expected Username %q, got %q", "user", opts.Username)
	}
	if opts.Password != "pass" {
		t.Errorf("expected Password %q, got %q", "pass", opts.Password)
	}
	if !reflect.DeepEqual(opts.Topics, expectedTopics) {
		t.Errorf("expected Topics %v, got %v", expectedTopics, opts.Topics)
	}
	if opts.OutputDir != "/tmp/output" {
		t.Errorf("expected OutputDir %q, got %q", "/tmp/output", opts.OutputDir)
	}
}

// TestIngestOptions_Load_Defaults verifies that defaults are applied when optional
// environment variables are not set.
func TestIngestOptions_Load_Defaults(t *testing.T) {
	cleanup := []func(){
		testhelper.UnsetEnv("WIS2_MQTT_HOST"),
		testhelper.UnsetEnv("WIS2_MQTT_PORT"),
		testhelper.UnsetEnv("WIS2_MQTT_USERNAME"),
		testhelper.UnsetEnv("WIS2_MQTT_PASSWORD"),
		testhelper.SetEnv("WIS2_MQTT_TOPIC", "topic"),
		testhelper.SetEnv("WIS2_OUTPUT_DIRECTORY", "/tmp/output"),
	}
	defer func() {
		for _, fn := range cleanup {
			fn()
		}
	}()

	opts := NewIngestOptions()
	err := opts.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if opts.Host != "globalbroker.meteo.fr" {
		t.Errorf("expected default Host, got %q", opts.Host)
	}
	if opts.Port != "8883" {
		t.Errorf("expected default Port, got %q", opts.Port)
	}
	if opts.Username != "everyone" {
		t.Errorf("expected default Username, got %q", opts.Username)
	}
	if opts.Password != "everyone" {
		t.Errorf("expected default Password, got %q", opts.Password)
	}
}

// TestIngestOptions_Load_MissingTopic verifies that an error is returned
// when WIS2_MQTT_TOPIC is not defined.
func TestIngestOptions_Load_MissingTopic(t *testing.T) {
	cleanup := []func(){
		testhelper.UnsetEnv("WIS2_MQTT_TOPIC"),
		testhelper.SetEnv("WIS2_OUTPUT_DIRECTORY", "/tmp/output"),
	}
	defer func() {
		for _, fn := range cleanup {
			fn()
		}
	}()

	opts := NewIngestOptions()
	err := opts.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "WIS2_MQTT_TOPIC") {
		t.Errorf("expected error mentioning WIS2_MQTT_TOPIC, got %q", err.Error())
	}
}

// TestIngestOptions_Load_MissingOutputDir verifies that an error is returned
// when WIS2_OUTPUT_DIRECTORY is not defined.
func TestIngestOptions_Load_MissingOutputDir(t *testing.T) {
	cleanup := []func(){
		testhelper.SetEnv("WIS2_MQTT_TOPIC", "topic"),
		testhelper.UnsetEnv("WIS2_OUTPUT_DIRECTORY"),
	}
	defer func() {
		for _, fn := range cleanup {
			fn()
		}
	}()

	opts := NewIngestOptions()
	err := opts.Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "WIS2_OUTPUT_DIRECTORY") {
		t.Errorf("expected error mentioning WIS2_OUTPUT_DIRECTORY, got %q", err.Error())
	}
}
