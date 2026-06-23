package mqtt

import (
	"strings"
	"wis2-ingest/internal/config"
)

// findAuthPolicy returns the authentication policy associated with
// the exact MQTT topic being processed, or nil if no match exists.
func (c *Client) findAuthPolicy(topic string) *config.AuthPolicy {

	for i := range c.cfg.AuthPolicies {

		policy := c.cfg.AuthPolicies[i]

		if matchTopic(policy.Topic, topic) {
			return &policy
		}
	}

	return nil
}

// matchTopic checks MQTT-style topic pattern matching.
// Supports:
//   - single-level wildcard
//     # multi-level wildcard (end only)
func matchTopic(pattern, topic string) bool {

	pLevels := strings.Split(pattern, "/")
	tLevels := strings.Split(topic, "/")

	for i := 0; i < len(pLevels); i++ {

		p := pLevels[i]

		// multi-level wildcard
		if p == "#" {
			return true
		}

		// topic ended early
		if i >= len(tLevels) {
			return false
		}

		t := tLevels[i]

		// single-level wildcard
		if p == "+" {
			continue
		}

		// exact match
		if p != t {
			return false
		}
	}

	// topic must not have extra levels unless pattern ends with #
	return len(pLevels) == len(tLevels)
}
