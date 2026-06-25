package mqtt

// messageHandler is the MQTT entrypoint.
// It delegates processing to the Processor pipeline.
func (c *Client) messageHandler(msg Msg) {
	p := NewProcessor(c)
	p.Process(msg)
}
