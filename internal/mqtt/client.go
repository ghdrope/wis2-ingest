package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/metrics"

	"github.com/akuity/kargo/pkg/logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client wraps an MQTT client and its configuration, logger and metrics.
type Client struct {
	cfg       *config.IngestOptions
	logger    *logging.Logger
	client    mqtt.Client
	connected atomic.Bool

	metrics *metrics.Metrics
}

// NewClient creates a new MQTT client instance.
func NewClient(cfg *config.IngestOptions, logger *logging.Logger, m *metrics.Metrics) *Client {
	if m == nil {
		logger.Info("metrics is nil - some observability will be disabled.")
	}

	return &Client{
		cfg:     cfg,
		logger:  logger,
		metrics: m,
	}
}

// Start runs the MQTT connection loop asynchronously.
func (c *Client) Start(ctx context.Context) {
	go c.connectLoop(ctx)
}

// connectLoop manages connection lifecycle and topic subscriptions.
func (c *Client) connectLoop(ctx context.Context) {

	broker := fmt.Sprintf("tls://%s:8883", c.cfg.Host)

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("wis2-ingest-%d", time.Now().Unix())).
		SetTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: c.cfg.Host,
		}).
		SetUsername(c.cfg.Username).
		SetPassword(c.cfg.Password)

	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	// OnConnect subscribes to configured topics.
	opts.OnConnect = func(client mqtt.Client) {

		c.logger.Info("Connected to MQTT broker")
		c.connected.Store(true)

		for _, topic := range c.cfg.Topics {

			handler := func(cl mqtt.Client, msg mqtt.Message) {
				c.messageHandler(msg)
			}

			token := client.Subscribe(topic, 1, handler)
			token.Wait()

			if token.Error() != nil {
				c.logger.Error(token.Error(), "subscribe error", "topic", topic)
				continue
			}

			c.logger.Info(fmt.Sprintf("subscribed to topic: %s", topic))
		}
	}

	// OnConnectionLost logs connection failures.
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		c.logger.Error(err, "MQTT connection lost")
		c.connected.Store(false)
	}

	c.client = mqtt.NewClient(opts)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping MQTT connect loop")
			return
		default:
		}

		token := c.client.Connect()
		token.Wait()

		if err := token.Error(); err != nil {
			c.logger.Error(err, "MQTT connect failed")
			time.Sleep(5 * time.Second)
			continue
		}

		c.logger.Info("MQTT client initialized successfully")
		return
	}
}

// Connected returns true if the MQTT client is currently connected.
func (c *Client) Connected() bool {
	return c.connected.Load()
}

// saveToFile persists the HTTP response body to the given file path.
func (c *Client) saveToFile(path string, body io.ReadCloser) error {

	defer func() {
		_ = body.Close()
	}()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	_, err = io.Copy(f, body)
	return err
}
