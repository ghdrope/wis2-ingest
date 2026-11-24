package config

import (
	"testing"
)

// TestPrintVerbose verifies the exact output: printable fields are listed by
// env name, while sensitive and print:"false" fields are omitted.
func TestPrintVerbose(t *testing.T) {
	opts := &IngestOptions{
		Host:                      "localhost",
		Username:                  "user",
		Password:                  "super-secret",
		Topics:                    []string{"t1", "t2"},
		OutputDir:                 "/tmp",
		AuthPoliciesConfigMapPath: "/etc/hidden.yaml",
		AuthPolicies: []AuthPolicy{
			{Name: "p", Password: SecretRef{Value: "policy-secret"}},
		},
	}

	want := "Configuration values:\n" +
		"  WIS2_MQTT_HOST       : localhost\n" +
		"  WIS2_MQTT_USERNAME   : user\n" +
		"  WIS2_MQTT_TOPIC      : [t1 t2]\n" +
		"  WIS2_OUTPUT_DIRECTORY : /tmp\n"

	if got := opts.PrintVerbose(); got != want {
		t.Errorf("PrintVerbose() =\n%s\nwant\n%s", got, want)
	}
}
