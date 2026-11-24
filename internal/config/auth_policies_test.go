package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestLoadAuthPolicies verifies that all policy fields, including the
// K8s secret reference, are parsed exactly as rendered by the Helm chart.
func TestLoadAuthPolicies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-policies.yaml")
	content := `authPolicies:
  - name: inline
    topic: origin/a/wis2/#
    username: user
    password:
      value: pass
  - name: from-secret
    topic: origin/b/wis2/+/data
    username: other
    password:
      secretName: wis2-ingest-auth-policies-secrets
      secretKey: other-password
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := NewIngestOptions()
	if err := opts.LoadAuthPolicies(path); err != nil {
		t.Fatalf("LoadAuthPolicies() error: %v", err)
	}

	want := []AuthPolicy{
		{
			Name:     "inline",
			Topic:    "origin/a/wis2/#",
			Username: "user",
			Password: SecretRef{Value: "pass"},
		},
		{
			Name:     "from-secret",
			Topic:    "origin/b/wis2/+/data",
			Username: "other",
			Password: SecretRef{
				SecretName: "wis2-ingest-auth-policies-secrets",
				SecretKey:  "other-password",
			},
		},
	}
	if !reflect.DeepEqual(opts.AuthPolicies, want) {
		t.Errorf("AuthPolicies =\n%+v\nwant\n%+v", opts.AuthPolicies, want)
	}
}

// TestLoadAuthPoliciesEmptyList verifies the chart's fallback "[]" renders
// to no policies without error.
func TestLoadAuthPoliciesEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-policies.yaml")
	if err := os.WriteFile(path, []byte("authPolicies:\n  []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := NewIngestOptions()
	if err := opts.LoadAuthPolicies(path); err != nil {
		t.Fatalf("LoadAuthPolicies() error: %v", err)
	}
	if len(opts.AuthPolicies) != 0 {
		t.Errorf("AuthPolicies = %+v; want empty", opts.AuthPolicies)
	}
}

// TestLoadAuthPoliciesMissingFile verifies that a missing file is ignored,
// so deployments without auth policies keep working.
func TestLoadAuthPoliciesMissingFile(t *testing.T) {
	opts := NewIngestOptions()
	if err := opts.LoadAuthPolicies(filepath.Join(t.TempDir(), "missing.yaml")); err != nil {
		t.Fatalf("LoadAuthPolicies() error = %v; want nil", err)
	}
	if opts.AuthPolicies != nil {
		t.Errorf("AuthPolicies = %+v; want nil", opts.AuthPolicies)
	}
}

// TestLoadAuthPoliciesErrors verifies read and parse errors are reported.
func TestLoadAuthPoliciesErrors(t *testing.T) {
	dir := t.TempDir()
	invalid := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(invalid, []byte("invalid: ["), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "invalid yaml", path: invalid, wantErr: "failed to parse auth policies yaml"},
		{name: "path is a directory", path: dir, wantErr: "failed to read auth policies file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewIngestOptions().LoadAuthPolicies(tt.path)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("LoadAuthPolicies() error = %v; want containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestSecretRefResolveValue verifies inline values and K8s mounted secret
// files are resolved correctly.
func TestSecretRefResolveValue(t *testing.T) {
	base := t.TempDir()
	orig := secretsBasePath
	secretsBasePath = base
	t.Cleanup(func() { secretsBasePath = orig })

	if err := os.MkdirAll(filepath.Join(base, "my-secret"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "my-secret", "password"), []byte("s3cr3t"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "my-secret", "with-newline"), []byte("s3cr3t\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		ref     SecretRef
		want    string
		wantErr string
	}{
		{name: "inline value", ref: SecretRef{Value: "inline"}, want: "inline"},
		{name: "empty ref", ref: SecretRef{}, want: ""},
		{
			name: "mounted secret file",
			ref:  SecretRef{SecretName: "my-secret", SecretKey: "password"},
			want: "s3cr3t",
		},
		{
			name: "trailing newline is not part of the secret",
			ref:  SecretRef{SecretName: "my-secret", SecretKey: "with-newline"},
			want: "s3cr3t",
		},
		{
			name: "secret file takes precedence over inline value",
			ref:  SecretRef{Value: "ignored", SecretName: "my-secret", SecretKey: "password"},
			want: "s3cr3t",
		},
		{
			name:    "missing secret key",
			ref:     SecretRef{SecretName: "my-secret"},
			wantErr: "secretKey is required when secretName is set",
		},
		{
			name:    "missing secret file",
			ref:     SecretRef{SecretName: "my-secret", SecretKey: "nope"},
			wantErr: "failed to read secret file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ref.ResolveValue()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ResolveValue() error = %v; want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveValue() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveValue() = %q; want %q", got, tt.want)
			}
		})
	}
}
