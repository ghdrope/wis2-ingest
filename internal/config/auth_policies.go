package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v2"
)

// SecretRef represents a secret that can be either:
//   - inline value (dev/local)
//   - K8s mounted file secret (prod)
type SecretRef struct {
	Value      string `yaml:"value"`
	SecretName string `yaml:"secretName"`
	SecretKey  string `yaml:"secretKey"`
}

// Resolve returns the actual secret value.
func (s SecretRef) ResolveValue() (string, error) {

	if s.SecretName == "" {
		return s.Value, nil
	}

	if s.SecretKey == "" {
		return "", fmt.Errorf("secretKey is required when secretName is set")
	}

	path := filepath.Join(DefaultSecretBasePath, s.SecretName, s.SecretKey)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file %s: %w", path, err)
	}

	return string(data), nil
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
func (o *IngestOptions) LoadAuthPolicies(path string) error {

	data, err := os.ReadFile(path)
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
