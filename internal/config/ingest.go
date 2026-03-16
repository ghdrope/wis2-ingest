package config

import (
	"fmt"
	"os"
	"strings"

	kargo_os "github.com/akuity/kargo/pkg/os"
)

// IngestOptions defines the configuration required for the WIS2 ingest process.
//
// The configuration is sourced from environment variables. Defaults are provided
// for some values.
type IngestOptions struct {
	Port     string
	Username string
	Password string
	Host     string

	Topics    []string
	OutputDir string
}

// NewIngestConfig initializes the ingest configuration.
//
// It loads configuration from environment variables and validates required values.
// This method should be called during command startup before any ingest processing
// begins.
func (o *IngestOptions) Load() error {
	if err := o.load(); err != nil {
		return err
	}
	return nil
}

// newIngestConfig reads environment variables and populates the IngestOptions struct.
//
// The following environment variables are used:
//
//	WIS2_MQTT_HOST
//	WIS2_MQTT_PORT
//	WIS2_MQTT_USERNAME
//	WIS2_MQTT_PASSWORD
//	WIS2_MQTT_TOPIC
//	WIS2_OUTPUT_DIRECTORY
//
// REMOTE_MQTT_TOPIC supports multiple topics separated by ';'.
func (o *IngestOptions) load() error {

	// MQTT connection parameters
	o.Port = kargo_os.GetEnv("WIS2_MQTT_PORT", "8883")
	o.Username = kargo_os.GetEnv("WIS2_MQTT_USERNAME", "everyone")
	o.Password = kargo_os.GetEnv("WIS2_MQTT_PASSWORD", "everyone")
	o.Host = kargo_os.GetEnv("WIS2_MQTT_HOST", "globalbroker.meteo.fr")

	// Topics are mandatory
	topicStr := os.Getenv("WIS2_MQTT_TOPIC")
	if topicStr == "" {
		return fmt.Errorf("WIS2_MQTT_TOPIC must be defined")
	}
	o.Topics = strings.Split(topicStr, ";")
	for i := range o.Topics {
		o.Topics[i] = strings.TrimSpace(o.Topics[i])
	}

	// OutputDir is mandatory
	if o.OutputDir = os.Getenv("WIS2_OUTPUT_DIRECTORY"); o.OutputDir == "" {
		return fmt.Errorf("WIS2_OUTPUT_DIRECTORY must be defined")
	}

	return nil
}
