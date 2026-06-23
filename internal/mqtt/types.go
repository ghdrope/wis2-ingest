package mqtt

// Properties represents metadata from incoming MQTT messages.
type Properties struct {
	DataID string `json:"data_id"`
}

// Link represents a downloadable resource in an MQTT message.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
	Type string `json:"type"`
}

// Payload represents the MQTT message structure.
type Payload struct {
	Properties Properties `json:"properties"`
	Links      []Link     `json:"links"`
}

// Msg defines the minimal interface required for testing MQTT messages.
type Msg interface {
	Topic() string
	Payload() []byte
}
