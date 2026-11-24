package testhelper

// MqttMessage is a test helper that implements the mqtt.Message interface
// (and the smaller mqtt.Msg interface used by the processor).
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

// Ack acknowledges the message. No-op in mock.
func (m MqttMessage) Ack() {}
