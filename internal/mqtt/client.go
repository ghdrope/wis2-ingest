package mqtt

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
	"wis2-ingest/internal/config"
	"wis2-ingest/internal/metrics"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// Client wraps an MQTT client and its configuration, logger and metrics.
type Client struct {
	cfg       *config.IngestOptions
	logger    *zap.Logger
	client    mqtt.Client
	connected atomic.Bool

	metrics *metrics.Metrics
}

// NewClient creates a new MQTT client instance.
func NewClient(cfg *config.IngestOptions, logger *zap.Logger, m *metrics.Metrics) *Client {
	if m == nil {
		logger.Info("metrics is nil - some observability will be disabled.")
	}

	return &Client{
		cfg:     cfg,
		logger:  logger,
		metrics: m,
	}
}

// clientID returns a unique MQTT client ID. Brokers disconnect a client when
// another one connects with the same ID, so replicas started in the same
// second must not share it. It stays within the 23 characters guaranteed
// by MQTT 3.1.1.
func clientID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b) // never returns an error
	return "wis2-ingest-" + hex.EncodeToString(b)
}

// Start runs the MQTT connection loop asynchronously. The client
// disconnects from the broker when ctx is cancelled.
func (c *Client) Start(ctx context.Context) {
	go c.connectLoop(ctx)
}

// clientOptions builds the paho options: TLS connection to the configured
// broker, automatic reconnection and the connection callbacks.
func (c *Client) clientOptions() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tls://%s:8883", c.cfg.Host)).
		SetClientID(clientID()).
		SetTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: c.cfg.Host,
		}).
		SetUsername(c.cfg.Username).
		SetPassword(c.cfg.Password)

	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	opts.OnConnect = c.onConnect
	opts.OnConnectionLost = c.onConnectionLost

	return opts
}

// onConnect subscribes to the configured topics. It runs on every
// (re)connection, since subscriptions don't survive a clean session.
//
// The client only reports Connected once every subscription succeeded: a
// client that misses topics is not ready, and if it stays that way the
// liveness probe eventually restarts it to retry.
func (c *Client) onConnect(client mqtt.Client) {

	c.logger.Info("Connected to MQTT broker")

	handler := func(_ mqtt.Client, msg mqtt.Message) {
		c.messageHandler(msg)
	}

	subscribed := true
	for _, topic := range c.cfg.Topics {

		token := client.Subscribe(topic, 1, handler)
		token.Wait()

		if token.Error() != nil {
			c.logger.Error("subscribe error", zap.Error(token.Error()), zap.String("topic", topic))
			subscribed = false
			continue
		}

		c.logger.Info("subscribed to topic", zap.String("topic", topic))
	}

	c.connected.Store(subscribed)
}

// onConnectionLost marks the client as disconnected. paho reconnects on its
// own and calls onConnect again.
func (c *Client) onConnectionLost(_ mqtt.Client, err error) {
	c.logger.Error("MQTT connection lost", zap.Error(err))
	c.connected.Store(false)
}

// connectLoop manages the connection lifecycle: it connects, and
// disconnects when ctx is cancelled.
func (c *Client) connectLoop(ctx context.Context) {

	c.client = mqtt.NewClient(c.clientOptions())

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping MQTT connect loop")
			return
		default:
		}

		// With ConnectRetry the token only completes once connected, so
		// also watch ctx to be able to stop while the broker is unreachable.
		token := c.client.Connect()
		select {
		case <-token.Done():
		case <-ctx.Done():
			c.client.Disconnect(0)
			c.logger.Info("Stopping MQTT connect loop")
			return
		}

		if err := token.Error(); err != nil {
			c.logger.Error("MQTT connect failed", zap.Error(err))
			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
			}
			continue
		}

		c.logger.Info("MQTT client initialized successfully")

		<-ctx.Done()
		c.logger.Info("Disconnecting from MQTT broker")
		c.client.Disconnect(250)
		c.connected.Store(false)
		return
	}
}

// Connected returns true if the MQTT client is currently connected.
func (c *Client) Connected() bool {
	return c.connected.Load()
}

// saveToFile persists the HTTP response body to the given file path.
// An existing file is overwritten: it can only be a temporary file left
// behind by a crashed process, since the caller holds the lock for it.
func (c *Client) saveToFile(path string, body io.ReadCloser) error {

	defer func() {
		_ = body.Close()
	}()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	_, err = io.Copy(f, body)
	return err
}
