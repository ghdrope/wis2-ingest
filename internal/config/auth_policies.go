package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// SecretRef represents a secret that can be either:
//   - inline value (dev/local)
//   - K8s mounted file secret (prod)
type SecretRef struct {
	Value      string `yaml:"value"`
	SecretName string `yaml:"secretName"`
	SecretKey  string `yaml:"secretKey"`
}

// secretsBasePath is the directory where K8s secrets are mounted.
// It is a variable only so tests can point it to a temporary directory.
var secretsBasePath = DefaultConfigBasePath

// ResolveValue returns the actual secret value.
func (s SecretRef) ResolveValue() (string, error) {

	if s.SecretName == "" {
		return s.Value, nil
	}

	if s.SecretKey == "" {
		return "", fmt.Errorf("secretKey is required when secretName is set")
	}

	filePath := filepath.Join(secretsBasePath, s.SecretName, s.SecretKey)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file %s: %w", filePath, err)
	}

	// Secret files often end with a newline (e.g. created with
	// "kubectl create secret --from-file"), which is not part of the value.
	return strings.TrimRight(string(data), "\r\n"), nil
}

// AuthPolicy defines download credentials associated with MQTT topic patterns.
//
// When a notification is received on a matching topic, these credentials
// are used to authenticate the payload download request.
type AuthPolicy struct {
	Name     string    `yaml:"name"`
	Topic    string    `yaml:"topic"`
	Username string    `yaml:"username"`
	Password SecretRef `yaml:"password" sensitive:"true"`
}

// AuthPoliciesConfig represents the authentication policy configuration
// loaded from auth-policies.yaml.
type AuthPoliciesConfig struct {
	AuthPolicies []AuthPolicy `yaml:"authPolicies"`
}

// LoadAuthPolicies loads download authentication policies from
// the configured auth-policies YAML file.
//
// Missing files are ignored to allow deployments without
// additional authentication requirements.
func (o *IngestOptions) LoadAuthPolicies(filePath string) error {

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read auth policies file: %w", err)
	}

	var cfg AuthPoliciesConfig

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse auth policies yaml: %w", err)
	}

	o.AuthPolicies = cfg.AuthPolicies

	return nil
}
