package config

import (
	"testing"
)

// TestLoadFromEnvOverridesDefaults verifies that
// environment variables override default values.
func TestLoadFromEnvOverridesDefaults(t *testing.T) {

	t.Setenv("WIS2_MQTT_HOST", "localhost")
	t.Setenv("WIS2_MQTT_TOPIC", "topic1;topic2")
	t.Setenv("WIS2_OUTPUT_DIRECTORY", "/tmp")

	opts := NewIngestOptions()

	if err := opts.loadFromEnv(); err != nil {
		t.Fatal(err)
	}

	if opts.Host != "localhost" {
		t.Fatalf("expected localhost got %s", opts.Host)
	}

	if len(opts.Topics) != 2 {
		t.Fatalf("expected 2 topics got %d", len(opts.Topics))
	}
}

// TestValidateSuccess verifies that validation succeeds
// when all mandatory values are present.
func TestValidateSuccess(t *testing.T) {

	opts := &IngestOptions{
		Topics:    []string{"topic"},
		OutputDir: "/tmp",
	}

	if err := opts.validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidateMissingMandatory verifies that validation
// fails when mandatory values are missing.
func TestValidateMissingMandatory(t *testing.T) {

	opts := &IngestOptions{}

	if err := opts.validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

// TestLoadAppliesEnvironmentVariables verifies that
// Load executes the complete loading workflow.
func TestLoadAppliesEnvironmentVariables(t *testing.T) {

	t.Setenv("WIS2_MQTT_TOPIC", "topic")
	t.Setenv("WIS2_OUTPUT_DIRECTORY", "/tmp")

	opts := NewIngestOptions()

	oldPath := DefaultAuthPoliciesConfigMapPath

	defer func() {
		_ = oldPath
	}()

	if err := opts.Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidate verifies that IngestOptions validation correctly
// enforces mandatory fields such as Topics and OutputDir.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    IngestOptions
		wantErr bool
	}{
		{
			name: "valid",
			opts: IngestOptions{
				Topics:    []string{"topic"},
				OutputDir: "/tmp",
			},
		},
		{
			name: "missing topics",
			opts: IngestOptions{
				OutputDir: "/tmp",
			},
			wantErr: true,
		},
		{
			name: "missing output dir",
			opts: IngestOptions{
				Topics: []string{"topic"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
		})
	}
}
