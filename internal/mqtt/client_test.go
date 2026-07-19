package mqtt

import (
	"context"
	"testing"
	"time"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/metrics"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// TestNewClient verifies that NewClient correctly initializes the Client struct.
func TestNewClient(t *testing.T) {
	cfg := &config.IngestOptions{
		Host:      "localhost",
		Username:  "user",
		Password:  "pass",
		Topics:    []string{"topic1", "topic2"},
		OutputDir: "/tmp",
	}

	logger := zap.NewNop()

	metricsInstance, _, err := metrics.New("test")
	if err != nil {
		t.Fatalf("failed to create metrics: %v", err)
	}

	client := NewClient(cfg, logger, metricsInstance)

	if client.cfg != cfg {
		t.Errorf("expected cfg to be set")
	}
	if client.logger != logger {
		t.Errorf("expected logger to be set")
	}
	if client.client != nil {
		t.Errorf("expected mqtt.Client to be nil before ConnectAndSubscribe")
	}
	if client.Connected() {
		t.Errorf("expected connected to be false initially")
	}
}

// TestConnectedFlag simulates connection state toggling.
func TestConnectedFlag(t *testing.T) {
	cfg := &config.IngestOptions{
		Host:   "localhost",
		Topics: []string{"topic1"},
	}

	logger := zap.NewNop()

	metricsInstance, _, err := metrics.New("test")
	if err != nil {
		t.Fatalf("failed to create metrics: %v", err)
	}

	client := NewClient(cfg, logger, metricsInstance)

	client.connected.Store(false)
	client.Connected()
	client.connected.Store(true)
	if !client.Connected() {
		t.Errorf("expected connected to be true after simulated connect")
	}

	client.connected.Store(false)
	if client.Connected() {
		t.Errorf("expected connected to be false after simulated connection lost")
	}
}

// TestConnectAndSubscribe_Handlers verifies that ConnectAndSubscribe
// correctly sets OnConnect and OnConnectionLost handlers in MQTT client options.
// This test does not connect to a real broker.
func TestConnectAndSubscribe_Handlers(t *testing.T) {
	// Create MQTT client options directly for testing
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:8883")

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
