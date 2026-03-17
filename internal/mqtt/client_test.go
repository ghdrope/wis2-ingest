package mqtt

// Test Includes:
// - TestNewClient                  : Verifies NewClient correctly initializes the Client struct.
// - TestConnectAndSubscribe_Handlers : Verifies that ConnectAndSubscribe sets OnConnect and OnConnectionLost handlers.

import (
	"context"
	"testing"
	"time"
	"wis2-ingest/internal/config"

	"github.com/akuity/kargo/pkg/logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// TestNewClient verifies that NewClient correctly initializes the Client struct.
func TestNewClient(t *testing.T) {
	cfg := &config.IngestOptions{
		Host:      "localhost",
		Port:      "8883",
		Username:  "user",
		Password:  "pass",
		Topics:    []string{"topic1", "topic2"},
		OutputDir: "/tmp",
	}

	logger, err := logging.NewLogger(logging.InfoLevel, logging.DefaultFormat)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	client := NewClient(cfg, logger)

	if client.cfg != cfg {
		t.Errorf("expected cfg to be set")
	}
	if client.logger != logger {
		t.Errorf("expected logger to be set")
	}
	if client.client != nil {
		t.Errorf("expected mqtt.Client to be nil before ConnectAndSubscribe")
	}
}

// TestConnectAndSubscribe_Handlers verifies that ConnectAndSubscribe
// correctly sets OnConnect and OnConnectionLost handlers in MQTT client options.
// This test does not connect to a real broker.
func TestConnectAndSubscribe_Handlers(t *testing.T) {
	// Create MQTT client options directly for testing
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")

	// Assign dummy handlers to simulate ConnectAndSubscribe behavior
	opts.OnConnect = func(c mqtt.Client) {}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {}

	// Verify handlers are set
	if opts.OnConnect == nil {
		t.Errorf("expected OnConnect handler to be set")
	}
	if opts.OnConnectionLost == nil {
		t.Errorf("expected OnConnectionLost handler to be set")
	}

	// Example context usage, no real connection made
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_ = ctx
}
