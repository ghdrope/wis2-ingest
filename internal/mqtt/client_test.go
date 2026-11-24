package mqtt

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wis2-ingest/internal/config"
	"wis2-ingest/pkg/testhelper"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// TestNewClient verifies the client is created disconnected and without an
// underlying MQTT client until Start is called. Nil metrics are allowed.
func TestNewClient(t *testing.T) {
	cfg := &config.IngestOptions{Host: "localhost"}
	logger := zap.NewNop()

	c := NewClient(cfg, logger, nil)

	if c.cfg != cfg || c.logger != logger || c.metrics != nil {
		t.Errorf("NewClient() did not keep the given dependencies")
	}
	if c.client != nil {
		t.Error("underlying MQTT client must not be created before Start")
	}
	if c.Connected() {
		t.Error("Connected() = true; want false before connecting")
	}
}

// TestConnectLoopStopsOnCancelledContext verifies the connect loop exits
// without dialing when the context is already cancelled.
func TestConnectLoopStopsOnCancelledContext(t *testing.T) {
	c := NewClient(&config.IngestOptions{Host: "127.0.0.1"}, zap.NewNop(), nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		c.connectLoop(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("connectLoop did not return after context cancellation")
	}
	if c.client == nil {
		t.Error("connectLoop must build the MQTT client")
	}
	if c.Connected() {
		t.Error("Connected() = true; want false")
	}
}

// TestSaveToFile verifies the body is written and closed, and that an
// existing file (a temporary file left behind by a crash) is overwritten.
func TestSaveToFile(t *testing.T) {
	c := &Client{}
	path := filepath.Join(t.TempDir(), "out.tmp")

	body := &trackingCloser{Reader: strings.NewReader("payload")}
	if err := c.saveToFile(path, body); err != nil {
		t.Fatalf("saveToFile() error: %v", err)
	}
	if !body.closed {
		t.Error("saveToFile() must close the body")
	}
	if data, _ := os.ReadFile(path); string(data) != "payload" {
		t.Errorf("file content = %q; want %q", data, "payload")
	}

	second := &trackingCloser{Reader: strings.NewReader("new")}
	if err := c.saveToFile(path, second); err != nil {
		t.Fatalf("saveToFile() over an existing file error: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "new" {
		t.Errorf("file content = %q; want %q (old content must be truncated)", data, "new")
	}

	third := &trackingCloser{Reader: strings.NewReader("x")}
	if err := c.saveToFile(filepath.Join(t.TempDir(), "missing", "out.tmp"), third); err == nil {
		t.Error("saveToFile() must fail when the directory does not exist")
	}
	if !third.closed {
		t.Error("saveToFile() must close the body on error")
	}
}

type trackingCloser struct {
	io.Reader
	closed bool
}

func (t *trackingCloser) Close() error {
	t.closed = true
	return nil
}

// TestConnectLoopStopsWhileConnecting verifies the connect loop returns when
// the context is cancelled while the broker is unreachable (paho keeps
// retrying the connection, so the token never completes on its own).
func TestConnectLoopStopsWhileConnecting(t *testing.T) {
	c := NewClient(&config.IngestOptions{Host: "127.0.0.1"}, zap.NewNop(), nil)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		c.connectLoop(ctx)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("connectLoop did not return after cancelling while connecting")
	}
	if c.Connected() {
		t.Error("Connected() = true; want false")
	}
}

// TestClientID verifies client IDs are unique and fit the 23 characters
// guaranteed by MQTT 3.1.1.
func TestClientID(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := clientID()
		if len(id) > 23 {
			t.Fatalf("clientID() = %q (%d chars); want at most 23", id, len(id))
		}
		if !strings.HasPrefix(id, "wis2-ingest-") {
			t.Fatalf("clientID() = %q; want prefix wis2-ingest-", id)
		}
		if seen[id] {
			t.Fatalf("clientID() returned %q twice", id)
		}
		seen[id] = true
	}
}

// fakeToken is a completed paho token with an optional error.
type fakeToken struct{ err error }

func (t fakeToken) Wait() bool                     { return true }
func (t fakeToken) WaitTimeout(time.Duration) bool { return true }
func (t fakeToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
func (t fakeToken) Error() error { return t.err }

// fakeMQTT records subscriptions. Only Subscribe is implemented: the
// embedded nil interface panics if anything else is called.
type fakeMQTT struct {
	pahomqtt.Client
	failTopic string
	handlers  map[string]pahomqtt.MessageHandler
}

func (f *fakeMQTT) Subscribe(topic string, qos byte, h pahomqtt.MessageHandler) pahomqtt.Token {
	if topic == f.failTopic {
		return fakeToken{err: errors.New("subscribe refused")}
	}
	if f.handlers == nil {
		f.handlers = map[string]pahomqtt.MessageHandler{}
	}
	f.handlers[topic] = h
	return fakeToken{}
}

// TestOnConnect verifies every configured topic is subscribed, the client is
// marked connected, and received messages reach the processor.
func TestOnConnect(t *testing.T) {
	c := newTestClient(t)
	ok1, ok2 := uniqueTopic("test/onconnect/a"), uniqueTopic("test/onconnect/b")
	c.cfg.Topics = []string{ok1, ok2}
	fake := &fakeMQTT{}

	c.onConnect(fake)

	if !c.Connected() {
		t.Error("Connected() = false after subscribing to every topic")
	}
	if len(fake.handlers) != 2 || fake.handlers[ok1] == nil || fake.handlers[ok2] == nil {
		t.Fatalf("subscribed topics = %v; want %s and %s", fake.handlers, ok1, ok2)
	}

	fake.handlers[ok1](fake, testhelper.MqttMessage{TopicField: ok1, PayloadField: []byte("not json")})
	if got := failureCount(t, ok1); got != 1 {
		t.Errorf("failure count = %v; want 1 (message must reach the processor)", got)
	}
}

// TestOnConnectionLost verifies the client is marked disconnected.
func TestOnConnectionLost(t *testing.T) {
	c := newTestClient(t)
	c.connected.Store(true)

	c.onConnectionLost(nil, errors.New("broker went away"))

	if c.Connected() {
		t.Error("Connected() = true after onConnectionLost")
	}
}

// TestClientOptions verifies the broker address, TLS, credentials and
// reconnection settings used to connect.
func TestClientOptions(t *testing.T) {
	c := NewClient(&config.IngestOptions{
		Host:     "broker.example.org",
		Username: "user",
		Password: "pass",
	}, zap.NewNop(), nil)

	opts := c.clientOptions()

	if len(opts.Servers) != 1 || opts.Servers[0].String() != "tls://broker.example.org:8883" {
		t.Errorf("servers = %v; want [tls://broker.example.org:8883]", opts.Servers)
	}
	if opts.TLSConfig == nil || opts.TLSConfig.ServerName != "broker.example.org" || opts.TLSConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("TLS config = %+v; want server name broker.example.org and TLS >= 1.2", opts.TLSConfig)
	}
	if opts.Username != "user" || opts.Password != "pass" {
		t.Errorf("credentials = %q/%q; want user/pass", opts.Username, opts.Password)
	}
	if !opts.AutoReconnect || !opts.ConnectRetry {
		t.Error("auto reconnect and connect retry must be enabled")
	}
	if opts.OnConnect == nil || opts.OnConnectionLost == nil {
		t.Error("connection callbacks must be set")
	}
	if !strings.HasPrefix(opts.ClientID, "wis2-ingest-") {
		t.Errorf("client ID = %q; want prefix wis2-ingest-", opts.ClientID)
	}
}

// TestOnConnectSubscriptionFailure verifies a failing subscription doesn't
// stop the others, but the client is not reported as connected (not ready)
// while it misses topics. A later successful reconnection makes it ready.
func TestOnConnectSubscriptionFailure(t *testing.T) {
	c := newTestClient(t)
	ok1, ok2 := uniqueTopic("test/onconnect/a"), uniqueTopic("test/onconnect/b")
	c.cfg.Topics = []string{ok1, "test/onconnect/refused", ok2}
	fake := &fakeMQTT{failTopic: "test/onconnect/refused"}

	c.onConnect(fake)

	if c.Connected() {
		t.Error("Connected() = true with a failed subscription; want false")
	}
	if len(fake.handlers) != 2 || fake.handlers[ok1] == nil || fake.handlers[ok2] == nil {
		t.Fatalf("subscribed topics = %v; want %s and %s", fake.handlers, ok1, ok2)
	}

	fake.failTopic = ""
	c.onConnect(fake)
	if !c.Connected() {
		t.Error("Connected() = false after a reconnection with every subscription")
	}
}
