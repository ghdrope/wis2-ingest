package mqtt

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wis2-ingest/internal/config"
	"wis2-ingest/internal/validate"

	"go.uber.org/zap"
)

// Processor handles MQTT message processing pipeline:
// parse → download → validate → store.
type Processor struct {
	client *Client
}

// NewProcessor creates a message processor.
func NewProcessor(c *Client) *Processor {
	return &Processor{client: c}
}

// Process executes full message processing pipeline.
func (p *Processor) Process(msg Msg) {

	start := time.Now()
	topic := msg.Topic()
	ctx := context.Background()

	var payload Payload

	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		p.client.logger.Error("invalid JSON payload", zap.Error(err), zap.String("topic", topic))
		p.client.metrics.IncFailure(ctx, topic)
		return
	}

	pendingDir := filepath.Join(p.client.cfg.OutputDir, "pending_validation")
	outputDir := p.client.cfg.OutputDir

	if err := os.MkdirAll(pendingDir, 0755); err != nil {
		p.client.logger.Error("failed to create pending directory", zap.Error(err), zap.String("dir", pendingDir))
		p.client.metrics.IncFailure(ctx, topic)
		return
	}

	authPolicy := p.client.findAuthPolicy(topic)

	for _, link := range payload.Links {
		if !isValidLink(link) {
			continue
		}

		p.processLink(ctx, topic, link, pendingDir, outputDir, authPolicy, start)
	}
}

// processLink handles download → validate → store for a single link.
func (p *Processor) processLink(
	ctx context.Context,
	topic string,
	link Link,
	pendingDir string,
	outputDir string,
	authPolicy *config.AuthPolicy,
	start time.Time,
) {

	// Idempotency key
	key := fmt.Sprintf("%s|%s|%s", topic, link.Href, link.Type)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))

	finalPath := filepath.Join(outputDir, hash+".bufr")
	tmpPath := filepath.Join(pendingDir, hash+".tmp")
	lockPath := finalPath + ".lock"

	// Idempotency check
	if _, err := os.Stat(finalPath); err == nil {
		return
	}

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return // already being processed
	}
	defer func() {
		_ = lockFile.Close()
		_ = os.Remove(lockPath)
	}()

	req, err := http.NewRequest(http.MethodGet, link.Href, nil)
	if err != nil {
		p.client.logger.Error("failed to create request", zap.Error(err))
		return
	}

	if authPolicy != nil {

		password, err := authPolicy.Password.ResolveValue()
		if err != nil {
			p.client.logger.Error(
				"failed to resolve authentication secret",
				zap.Error(err),
				zap.String("policy", authPolicy.Name),
			)

			p.client.metrics.IncFailure(ctx, topic)
			return
		}

		req.SetBasicAuth(
			authPolicy.Username,
			password,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		p.client.logger.Error("download failed", zap.Error(err), zap.String("href", link.Href))
		p.client.metrics.IncFailure(ctx, topic)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		p.client.logger.Error("authentication failed", zap.Error(err), zap.String("href", link.Href))
		p.client.metrics.IncFailure(ctx, topic)
		return
	}

	if resp.StatusCode != http.StatusOK {
		return
	}

	if err := p.client.saveToFile(tmpPath, resp.Body); err != nil {
		p.client.logger.Error("failed to store file", zap.Error(err), zap.String("file", tmpPath))
		p.client.metrics.IncFailure(ctx, topic)
		_ = os.Remove(tmpPath)
		return
	}

	if _, err := validate.IsBUFR(tmpPath); err != nil {

		p.client.logger.Error("invalid BUFR file", zap.Error(err), zap.String("file", tmpPath))

		_ = os.Remove(tmpPath)

		p.client.metrics.IncFailure(ctx, topic)
		return
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {

		p.client.logger.Error("failed to move file",
			zap.Error(err),
			zap.String("source", tmpPath),
			zap.String("dest", finalPath),
		)

		p.client.metrics.IncFailure(ctx, topic)
		return
	}

	p.client.metrics.IncSuccess(ctx, topic)
	p.client.metrics.ObserveLatency(ctx, topic, start)

	p.client.logger.Info("stored file",
		zap.String("file", finalPath),
		zap.String("topic", topic),
	)
}

// isValidLink reports whether a link is eligible for download.
//
// A valid link must:
//   - have Rel set to "canonical"
//   - contain a non-empty Href
//   - have a MIME type containing "application/octet-stream"
func isValidLink(link Link) bool {
	return link.Rel == "canonical" &&
		link.Href != "" &&
		strings.Contains(link.Type, "application/octet-stream")
}
