package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// IngestOptions defines the configuration required for the WIS2 ingest process.
//
// Optional fields have default values. Mandatory fields are enforced dynamically
// via the `mandatory:"true"` struct tag.
type IngestOptions struct {
	Host     string `env:"WIS2_MQTT_HOST" mandatory:"false"`                      // MQTT broker host
	Username string `env:"WIS2_MQTT_USERNAME" mandatory:"false"`                  // MQTT username
	Password string `env:"WIS2_MQTT_PASSWORD" mandatory:"false" sensitive:"true"` // MQTT password

	Topics    []string `env:"WIS2_MQTT_TOPIC" mandatory:"true"`       // MQTT topics (mandatory)
	OutputDir string   `env:"WIS2_OUTPUT_DIRECTORY" mandatory:"true"` // Output directory (mandatory)
}

// NewIngestConfig returns a new IngestOptions struct with default values.
func NewIngestOptions() *IngestOptions {
	return &IngestOptions{
		Host:     "globalbroker.meteo.fr",
		Username: "everyone",
		Password: "everyone",
	}
}

// Load reads environment variables, overrides defaults, and validates mandatory fields.
func (o *IngestOptions) Load() error {
	v := reflect.ValueOf(o).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		envName := field.Tag.Get("env")
		if envName == "" {
			envName = field.Name
		}

		val := os.Getenv(envName)
		if val == "" {
			continue // keep default
		}

		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(val)
		case reflect.Slice:
			v.Field(i).Set(reflect.ValueOf(splitAndTrim(val, ";")))
		}
	}

	// Validate mandatory fields
	if err := o.validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// validate checks that all mandatory fields are defined.
func (o *IngestOptions) validate() error {
	v := reflect.ValueOf(o).Elem()
	t := v.Type()
	var missing []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		mandatory := field.Tag.Get("mandatory")
		if mandatory != "true" {
			continue
		}

		envName := field.Tag.Get("env")
		if envName == "" {
			envName = field.Name
		}

		value := v.Field(i).Interface()
		if v.Field(i).Kind() == reflect.String && strings.TrimSpace(value.(string)) == "" {
			missing = append(missing, envName)
		}
		if v.Field(i).Kind() == reflect.Slice && len(value.([]string)) == 0 {
			missing = append(missing, envName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing mandatory environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// PrintVerbose prints all configuration fields and their current values.
// The env tag is used for display.
func (o *IngestOptions) PrintVerbose() string {
	v := reflect.ValueOf(o).Elem()
	t := v.Type()

	var sb strings.Builder
	sb.WriteString("Configuration values:\n")
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip sensitive fields
		if field.Tag.Get("sensitive") == "true" {
			continue
		}

		envName := field.Tag.Get("env")
		if envName == "" {
			envName = field.Name
		}
		value := v.Field(i).Interface()
		fmt.Fprintf(&sb, "  %-20s : %v\n", envName, value)
	}
	return sb.String()
}

// splitAndTrim splits a string by sep and trims whitespace.
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
