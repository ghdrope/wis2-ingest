package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadAuthPolicies verifies that valid authentication
// policies are loaded from a YAML file.
func TestLoadAuthPolicies(t *testing.T) {

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "auth-policies.yaml")

	content := `authPolicies:
  - name: test
    topic: origin/a/wis2/#
    username: user
    password:
      value: pass
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test yaml: %v", err)
	}

	opts := NewIngestOptions()

	if err := opts.LoadAuthPolicies(path); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if len(opts.AuthPolicies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(opts.AuthPolicies))
	}

	policy := opts.AuthPolicies[0]

	if policy.Name != "test" {
		t.Errorf("unexpected name: %s", policy.Name)
	}

	if policy.Username != "user" {
		t.Errorf("unexpected username: %s", policy.Username)
	}
}

// TestLoadAuthPoliciesMissingFile verifies that a missing
// auth policy file is ignored.
func TestLoadAuthPoliciesMissingFile(t *testing.T) {

	opts := NewIngestOptions()

	err := opts.LoadAuthPolicies("/does/not/exists.yaml")

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// TestLoadAuthPoliciesInvalidYAML verifies that invalid YAML
// content returns an error.
func TestLoadAuthPoliciesInvalidYAML(t *testing.T) {

	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "auth-policies.yaml")

	if err := os.WriteFile(path, []byte("invalid: ["), 0644); err != nil {
		t.Fatal(err)
	}

	opts := NewIngestOptions()

	if err := opts.LoadAuthPolicies(path); err == nil {
		t.Fatal("expected error")
	}
}
