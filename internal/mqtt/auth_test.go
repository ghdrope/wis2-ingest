package mqtt

import (
	"testing"
	"wis2-ingest/internal/config"
)

// TestFindAuthPolicyMatch verifies that a matching topic
// returns the correct authentication policy.
func TestFindAuthPolicyMatch(t *testing.T) {

	c := &Client{
		cfg: &config.IngestOptions{
			AuthPolicies: []config.AuthPolicy{
				{
					Topic:    "test/topic",
					Username: "user",
					Password: config.SecretRef{
						Value: "pass",
					},
				},
			},
		},
	}

	policy := c.findAuthPolicy("test/topic")

	if policy == nil {
		t.Fatal("expected policy, got nil")
		return
	}

	if policy.Username != "user" {
		t.Errorf("expected username user, got %s", policy.Username)
	}
}

// TestFindAuthPolicyNoMatch verifies that no matching topic
// returns nil.
func TestFindAuthPolicyNoMatch(t *testing.T) {

	c := &Client{
		cfg: &config.IngestOptions{
			AuthPolicies: []config.AuthPolicy{},
		},
	}

	policy := c.findAuthPolicy("missing/topic")

	if policy != nil {
		t.Fatal("expected nil policy")
	}
}
