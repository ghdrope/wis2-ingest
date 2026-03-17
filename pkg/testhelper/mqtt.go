package testhelper

import (
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MockMQTTClient is a test helepr that implements the mqtt.Client interface.
// It allows simulating connections, subscriptions, and tokens without a real broker.
type MockMQTTClient struct {
	ConnectFunc   func() mqtt.Token // custom Connect function for test behavior
	ConnectCalled bool              // flag indicating whether Connect() was called
}

// Connect simulates connecting to an MQTT broker.
// It sets connectCalled to true and returns either the custom token or a default mockToken.
func (m *MockMQTTClient) Connect() mqtt.Token {
	m.ConnectCalled = true
	if m.ConnectFunc != nil {
		return m.ConnectFunc()
	}
	return &MockToken{}
}

// Disconnect simulates disconnecting from the MQTT broker.
// This mock implementation does nothing.
func (m *MockMQTTClient) Disconnect(quiesce uint) {}

// Subscribe simulates subscribing to a topic with a callback.
// Always returns a mockToken in tests.
func (m *MockMQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &MockToken{}
}

// MockToken implements mqtt.Token interface for testing purposes.
// It simulates an immediately completed token.
type MockToken struct{}

// Wait blocks until the token is marked done.
// In this mock, it returns immediately with true.
func (t *MockToken) Wait() bool { return true }

// WaitTimeout blocks until the token is done or timeout occurs.
// In this mock, it returns immediately with true.
func (t *MockToken) WaitTimeout(time.Duration) bool { return true }

// Error returns any error associated with the token.
// This mock always returns nil.
func (t *MockToken) Error() error { return nil }

// Done returns a channel that is closed when the token completes.
// This mock returns a channel closed immediately to simulate a completed token.
func (t *MockToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

// MqttMessage is a test helper that implements mqtt.Message interface.
// It allows sending mock MQTT messages to handlers in tests.
type MqttMessage struct {
	TopicField   string // Topic of the message
	PayloadField []byte // Payload of the message
}

// Duplicate indicates whether the message is a duplicate. Always false in mock.
func (m MqttMessage) Duplicate() bool { return false }

// Qos returns the Quality of Service level of the message. Always 1 in mock.
func (m MqttMessage) Qos() byte { return 1 }

// Retained indicates if the message is retained. Always false in mock.
func (m MqttMessage) Retained() bool { return false }

// Topic returns the topic of the message.
func (m MqttMessage) Topic() string { return m.TopicField }

// MessageID returns the message ID. Always 0 in mock.
func (m MqttMessage) MessageID() uint16 { return 0 }

// Payload returns the message payload.
func (m MqttMessage) Payload() []byte { return m.PayloadField }
