package mqtt

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Payload represents the JSON structure in incoming MQTT messages.
type Payload struct {
	Properties struct {
		DataID string `json:"data_id"`
	} `json:"properties"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
}

// messageHandler processes incoming MQTT messages.
// It downloads canonical .bufr files and stores them in a time-partitioned directory.
func (c *Client) messageHandler(client mqtt.Client, msg mqtt.Message) {
	now := time.Now().UTC()

	var payload Payload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		c.logger.Error(err, "invalid JSON payload", "topic", msg.Topic())
		return
	}

	// Build target directory by UTC timestamp: YYYY/MM/DD/HH
	dir := filepath.Join(
		c.cfg.OutputDir,
		now.Format("2006"), // year
		now.Format("01"),   // month
		now.Format("02"),   // day
		now.Format("15"),   // hour
	)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.logger.Error(err, "failed to create directory", "dir", dir)
		return
	}

	// Process each canonical link
	for _, link := range payload.Links {
		if link.Rel != "canonical" || link.Href == "" {
			continue
		}

		filename := filepath.Base(link.Href)

		// Only process .bufr files
		if filepath.Ext(filename) != ".bufr" {
			continue
		}

		outPath := filepath.Join(dir, filename)

		// Skip existing files
		if _, err := os.Stat(outPath); err == nil {
			continue
		}

		// Download the file
		resp, err := http.Get(link.Href)
		if err != nil {
			c.logger.Error(err, "failed to download href", "href", link.Href)
			continue
		}

		// Ensure file is properly closed
		f, err := os.OpenFile(outPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			c.logger.Error(err, "cannot create file", "file", outPath)
			resp.Body.Close()
			continue
		}

		_, err = io.Copy(f, resp.Body)
		resp.Body.Close()
		f.Close()
		if err != nil {
			c.logger.Error(err, "failed writing file", "file", outPath)
			continue
		}

		c.logger.Info("bufr file stored", "file", outPath, "topic", msg.Topic())
	}
}
