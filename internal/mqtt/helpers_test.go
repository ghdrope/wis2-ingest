package mqtt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"wis2-ingest/internal/config"
	"wis2-ingest/internal/metrics"
	"wis2-ingest/pkg/testhelper"

	"go.uber.org/zap"
)

// metrics.New can only be called once per process (global Prometheus
// registry), so all tests in this package share one instance and use
// unique topics to keep their counters apart.
var (
	metricsOnce    sync.Once
	sharedMetrics  *metrics.Metrics
	metricsHandler http.Handler
	metricsErr     error
)

func testMetrics(t *testing.T) (*metrics.Metrics, http.Handler) {
	t.Helper()
	metricsOnce.Do(func() {
		sharedMetrics, metricsHandler, metricsErr = metrics.New("wis2-ingest-mqtt-test")
	})
	if metricsErr != nil {
		t.Fatalf("metrics.New() error: %v", metricsErr)
	}
	return sharedMetrics, metricsHandler
}

// topicSeq makes uniqueTopic return a new topic on every call.
var topicSeq atomic.Int64

// uniqueTopic returns base with a unique suffix. Metrics are shared by the
// whole process, so each test (and each run with -count > 1) needs its own
// topic to start from zero.
func uniqueTopic(base string) string {
	return fmt.Sprintf("%s/%d", base, topicSeq.Add(1))
}

// newTestClient returns a Client writing into a fresh temporary directory.
func newTestClient(t *testing.T, policies ...config.AuthPolicy) *Client {
	t.Helper()
	m, _ := testMetrics(t)
	cfg := &config.IngestOptions{
		OutputDir:    t.TempDir(),
		AuthPolicies: policies,
	}
	return NewClient(cfg, zap.NewNop(), m)
}

func successCount(t *testing.T, topic string) float64 {
	_, h := testMetrics(t)
	return testhelper.MetricValue(t, h, "wis2_ingest_files_success_total", topic)
}

func failureCount(t *testing.T, topic string) float64 {
	_, h := testMetrics(t)
	return testhelper.MetricValue(t, h, "wis2_ingest_files_failed_total", topic)
}

// bufrLink returns a link that passes isValidLink.
func bufrLink(href string) Link {
	return Link{Href: href, Rel: "canonical", Type: "application/octet-stream"}
}

// message builds a WIS2 notification message with the given links.
func message(t *testing.T, topic string, links ...Link) testhelper.MqttMessage {
	t.Helper()
	data, err := json.Marshal(Payload{Properties: Properties{DataID: "id"}, Links: links})
	if err != nil {
		t.Fatal(err)
	}
	return testhelper.MqttMessage{TopicField: topic, PayloadField: data}
}

// expectedFinalPath mirrors the idempotency key used by processLink.
func expectedFinalPath(outputDir, topic string, link Link) string {
	key := fmt.Sprintf("%s|%s|%s", topic, link.Href, link.Type)
	return filepath.Join(outputDir, fmt.Sprintf("%x", sha256.Sum256([]byte(key)))+".bufr")
}
