package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"
	"wis2-ingest/internal/config"

	"github.com/akuity/kargo/pkg/logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client wraps an MQTT client and associated configuration and logger.
type Client struct {
	cfg       *config.IngestOptions
	logger    *logging.Logger
	client    mqtt.Client
	connected bool
}

// NewClient creates a new MQTT client instance with the given configuration and logger.
func NewClient(cfg *config.IngestOptions, logger *logging.Logger) *Client {
	return &Client{
		cfg:    cfg,
		logger: logger,
	}
}

// ConnectAndSubscribe connects to the MQTT broker and subscribes to configured topics.
// It handles automatic reconnects and logs all connection events.
func (c *Client) ConnectAndSubscribe(ctx context.Context) error {
	broker := fmt.Sprintf("tls://%s:8883", c.cfg.Host)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(fmt.Sprintf("wis2-ingest-%d", time.Now().Unix()))
	opts.SetTLSConfig(&tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: c.cfg.Host,
	})
	opts.SetUsername(c.cfg.Username)
	opts.SetPassword(c.cfg.Password)

	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	// OnConnect subscribes to all configured topics
	opts.OnConnect = func(client mqtt.Client) {
		c.logger.Info("Connected to MQTT broker")
		c.connected = true
		for _, t := range c.cfg.Topics {
			handler := func(cl mqtt.Client, msg mqtt.Message) {
				c.messageHandler(cl, msg)
			}
			token := client.Subscribe(t, 1, handler)
			token.Wait()
			if token.Error() != nil {
				c.logger.Error(token.Error(), "Subscribe error", "topic", t)
			} else {
				c.logger.Info(fmt.Sprintf("Subscribed to topic: %s", t))
			}
		}
	}

	// OnConnectionLost logs the error
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		c.logger.Error(err, "Connection lost")
		c.connected = false
	}

	c.client = mqtt.NewClient(opts)

	// Attempt connection
	token := c.client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to connect to broker: %w", err)
	}

	c.logger.Info("MQTT client initialized successfully")
	return nil
}

// Connected returns true if the MQTT client is currently connected.
func (c *Client) Connected() bool {
	return c.connected
}
