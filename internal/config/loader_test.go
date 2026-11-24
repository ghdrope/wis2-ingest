package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// clearEnv makes sure no configuration variable leaks from the host
// environment into a test. Empty values are treated as "unset" by the loader.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"WIS2_MQTT_HOST",
		"WIS2_MQTT_USERNAME",
		"WIS2_MQTT_PASSWORD",
		"WIS2_MQTT_TOPIC",
		"WIS2_OUTPUT_DIRECTORY",
		"WIS2_AUTH_POLICIES_CONFIGMAP_PATH",
		"AuthPolicies",
	} {
		t.Setenv(name, "")
	}
}

// TestLoadFromEnv verifies environment variables override defaults,
// lists are split on ";" and unset variables keep the defaults.
func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("WIS2_MQTT_HOST", "localhost")
	t.Setenv("WIS2_MQTT_PASSWORD", "secret")
	t.Setenv("WIS2_MQTT_TOPIC", " topic1 ; topic2 ;")
	t.Setenv("WIS2_OUTPUT_DIRECTORY", "/data")
	t.Setenv("WIS2_AUTH_POLICIES_CONFIGMAP_PATH", "/cfg/policies.yaml")

	opts := NewIngestOptions()
	if err := opts.loadFromEnv(); err != nil {
		t.Fatalf("loadFromEnv() error: %v", err)
	}

	want := &IngestOptions{
		Host:                      "localhost",
		Username:                  "everyone", // unset -> default kept
		Password:                  "secret",
		Topics:                    []string{"topic1", "topic2"},
		OutputDir:                 "/data",
		AuthPoliciesConfigMapPath: "/cfg/policies.yaml",
	}
	if !reflect.DeepEqual(opts, want) {
		t.Errorf("loadFromEnv() =\n%+v\nwant\n%+v", opts, want)
	}
}

// TestLoadFromEnvIgnoresNonStringSlices is a regression test: fields without
// an env tag fall back to the field name, and setting an env var named
// "AuthPolicies" used to panic (cannot assign []string to []AuthPolicy).
func TestLoadFromEnvIgnoresNonStringSlices(t *testing.T) {
	clearEnv(t)
	t.Setenv("AuthPolicies", "anything")

	opts := NewIngestOptions()
	if err := opts.loadFromEnv(); err != nil {
		t.Fatalf("loadFromEnv() error: %v", err)
	}
	if opts.AuthPolicies != nil {
		t.Errorf("AuthPolicies = %+v; want nil (not loadable from env)", opts.AuthPolicies)
	}
}

// TestValidate verifies mandatory fields are enforced and reported by
// their environment variable names.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    IngestOptions
		wantErr string
	}{
		{
			name: "valid",
			opts: IngestOptions{Topics: []string{"topic"}, OutputDir: "/tmp"},
		},
		{
			name:    "missing topics",
			opts:    IngestOptions{OutputDir: "/tmp"},
			wantErr: "missing mandatory environment variables: WIS2_MQTT_TOPIC",
		},
		{
			name:    "missing output dir",
			opts:    IngestOptions{Topics: []string{"topic"}},
			wantErr: "missing mandatory environment variables: WIS2_OUTPUT_DIRECTORY",
		},
		{
			name:    "whitespace-only output dir",
			opts:    IngestOptions{Topics: []string{"topic"}, OutputDir: "   "},
			wantErr: "missing mandatory environment variables: WIS2_OUTPUT_DIRECTORY",
		},
		{
			name:    "all mandatory missing",
			opts:    IngestOptions{Host: "optional-only"},
			wantErr: "missing mandatory environment variables: WIS2_MQTT_TOPIC, WIS2_OUTPUT_DIRECTORY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validate() error = %v; want %q", err, tt.wantErr)
			}
		})
	}
}

// TestLoad verifies the complete workflow: env -> auth policies -> validation.
func TestLoad(t *testing.T) {
	writePolicies := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "auth-policies.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("success with auth policies", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("WIS2_MQTT_TOPIC", "origin/a/wis2/#")
		t.Setenv("WIS2_OUTPUT_DIRECTORY", "/data")
		t.Setenv("WIS2_AUTH_POLICIES_CONFIGMAP_PATH", writePolicies(t,
			"authPolicies:\n  - name: p1\n    topic: origin/a/#\n    username: u\n"))

		opts := NewIngestOptions()
		if err := opts.Load(); err != nil {
			t.Fatalf("Load() error: %v", err)
		}
		if len(opts.AuthPolicies) != 1 || opts.AuthPolicies[0].Name != "p1" {
			t.Errorf("AuthPolicies = %+v; want one policy named p1", opts.AuthPolicies)
		}
	})

	t.Run("invalid auth policies file", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("WIS2_MQTT_TOPIC", "topic")
		t.Setenv("WIS2_OUTPUT_DIRECTORY", "/data")
		t.Setenv("WIS2_AUTH_POLICIES_CONFIGMAP_PATH", writePolicies(t, "invalid: ["))

		if err := NewIngestOptions().Load(); err == nil {
			t.Fatal("Load() expected error for invalid auth policies yaml")
		}
	})

	t.Run("missing mandatory values", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("WIS2_AUTH_POLICIES_CONFIGMAP_PATH", filepath.Join(t.TempDir(), "missing.yaml"))

		if err := NewIngestOptions().Load(); err == nil {
			t.Fatal("Load() expected validation error")
		}
	})
}

// TestLoadAuthPoliciesDefaultPath verifies the default ConfigMap path is
// used when no override is set. The default path does not exist outside the
// cluster, which must not be an error.
func TestLoadAuthPoliciesDefaultPath(t *testing.T) {
	if _, err := os.Stat(filepath.Join(DefaultConfigBasePath, DefaultAuthPoliciesConfigMap)); err == nil {
		t.Skip("default auth policies file exists on this machine")
	}

	o := &IngestOptions{}
	if err := o.loadAuthPolicies(); err != nil {
		t.Fatalf("loadAuthPolicies() with the default path error: %v", err)
	}
	if len(o.AuthPolicies) != 0 {
		t.Errorf("AuthPolicies = %v; want none", o.AuthPolicies)
	}
}
