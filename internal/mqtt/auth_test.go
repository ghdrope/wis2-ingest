package mqtt

import (
	"testing"
	"wis2-ingest/internal/config"
)

// TestMatchTopic verifies MQTT wildcard semantics ("+" single level,
// "#" multi level, including the parent level).
func TestMatchTopic(t *testing.T) {
	tests := []struct {
		pattern, topic string
		want           bool
	}{
		{"a/b/c", "a/b/c", true},
		{"a/b/c", "a/b/d", false},
		{"a/b", "a/b/c", false},
		{"a/b/c", "a/b", false},
		{"a/+/c", "a/x/c", true},
		{"a/+/c", "a/x/y/c", false},
		{"a/+", "a", false},
		{"+/+", "a/b", true},
		{"a/#", "a/b/c/d", true},
		{"a/#", "a", true}, // "#" also matches the parent level (MQTT spec)
		{"#", "anything/at/all", true},
		{"origin/a/wis2/+/data/#", "origin/a/wis2/fr-meteofrance/data/core/weather", true},
		{"origin/a/wis2/+/data/#", "origin/a/wis2/fr-meteofrance/metadata", false},
		{"a/b", "a/B", false}, // case-sensitive
	}
	for _, tt := range tests {
		if got := matchTopic(tt.pattern, tt.topic); got != tt.want {
			t.Errorf("matchTopic(%q, %q) = %v; want %v", tt.pattern, tt.topic, got, tt.want)
		}
	}
}

// TestFindAuthPolicy verifies the first matching policy wins, in
// configuration order.
func TestFindAuthPolicy(t *testing.T) {
	c := &Client{cfg: &config.IngestOptions{AuthPolicies: []config.AuthPolicy{
		{Name: "specific", Topic: "origin/a/wis2/fr/#"},
		{Name: "generic", Topic: "origin/a/wis2/#"},
	}}}

	tests := []struct {
		topic string
		want  string // policy name, "" for none
	}{
		{"origin/a/wis2/fr/data", "specific"},
		{"origin/a/wis2/de/data", "generic"},
		{"origin/b/wis2/fr/data", ""},
	}
	for _, tt := range tests {
		got := c.findAuthPolicy(tt.topic)
		switch {
		case tt.want == "" && got != nil:
			t.Errorf("findAuthPolicy(%q) = %q; want nil", tt.topic, got.Name)
		case tt.want != "" && (got == nil || got.Name != tt.want):
			t.Errorf("findAuthPolicy(%q) = %v; want %q", tt.topic, got, tt.want)
		}
	}
}

// TestFindAuthPolicyNoPolicies verifies nil is returned without policies.
func TestFindAuthPolicyNoPolicies(t *testing.T) {
	c := &Client{cfg: &config.IngestOptions{}}
	if got := c.findAuthPolicy("any/topic"); got != nil {
		t.Errorf("findAuthPolicy() = %+v; want nil", got)
	}
}

// TestFindAuthPolicyReturnsCopy verifies callers cannot mutate the
// configured policy through the returned pointer.
func TestFindAuthPolicyReturnsCopy(t *testing.T) {
	c := &Client{cfg: &config.IngestOptions{AuthPolicies: []config.AuthPolicy{{Topic: "#", Username: "orig"}}}}

	c.findAuthPolicy("x").Username = "changed"

	if c.cfg.AuthPolicies[0].Username != "orig" {
		t.Error("configured policy was mutated through findAuthPolicy result")
	}
}
