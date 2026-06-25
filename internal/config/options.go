package config

const (
	DefaultConfigBasePath        = "/etc/wis2-ingest"
	DefaultAuthPoliciesConfigMap = "auth-policies-configmap.yaml"
)

// IngestOptions contains all runtime configuration for wis2-ingest.
//
// Values are loaded from environment variables and optional ConfigMaps.
type IngestOptions struct {
	Host     string `env:"WIS2_MQTT_HOST" mandatory:"false"`                      // MQTT broker host
	Username string `env:"WIS2_MQTT_USERNAME" mandatory:"false"`                  // MQTT username
	Password string `env:"WIS2_MQTT_PASSWORD" mandatory:"false" sensitive:"true"` // MQTT password

	Topics    []string `env:"WIS2_MQTT_TOPIC" mandatory:"true"`       // MQTT topics (mandatory)
	OutputDir string   `env:"WIS2_OUTPUT_DIRECTORY" mandatory:"true"` // Output directory (mandatory)

	AuthPoliciesConfigMapPath string       `env:"WIS2_AUTH_POLICIES_CONFIGMAP_PATH" mandatory:"false" print:"false"` // Override auth-policies ConfigMap path
	AuthPolicies              []AuthPolicy `mandatory:"false" print:"false"`                                         // Custom authentication policies
}

// NewIngestOptions creates a configuration instance placed
// with the default broker credentials.
func NewIngestOptions() *IngestOptions {
	return &IngestOptions{
		Host:     "globalbroker.meteo.fr",
		Username: "everyone",
		Password: "everyone",
	}
}
