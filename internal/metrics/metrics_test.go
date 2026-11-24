package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"wis2-ingest/pkg/testhelper"
)

// New registers an exporter in the global Prometheus registry, and calling
// it twice in one process makes the /metrics endpoint fail with duplicate
// target_info samples. Production calls it once per process, so tests share
// a single instance too.
var (
	sharedOnce    sync.Once
	sharedMetrics *Metrics
	sharedHandler http.Handler
	sharedErr     error
)

// topicSeq makes uniqueTopic return a new topic on every call.
var topicSeq atomic.Int64

// uniqueTopic returns base with a unique suffix, so counters start from zero
// even when the tests run more than once (-count > 1).
func uniqueTopic(base string) string {
	return fmt.Sprintf("%s/%d", base, topicSeq.Add(1))
}

func shared(t *testing.T) (*Metrics, http.Handler) {
	t.Helper()
	sharedOnce.Do(func() {
		sharedMetrics, sharedHandler, sharedErr = New("wis2-ingest-test")
	})
	if sharedErr != nil {
		t.Fatalf("New() error: %v", sharedErr)
	}
	return sharedMetrics, sharedHandler
}

// TestNew verifies New returns usable metrics and a handler.
func TestNew(t *testing.T) {
	m, h := shared(t)
	if m == nil || h == nil {
		t.Fatalf("New() = %v, %v; want non-nil metrics and handler", m, h)
	}
}

// TestCounters verifies success/failure counters are exported under the
// names used by the Grafana dashboards, labelled by topic.
func TestCounters(t *testing.T) {
	m, h := shared(t)
	ctx := context.Background()
	topic := uniqueTopic("test/counters")

	m.IncSuccess(ctx, topic)
	m.IncSuccess(ctx, topic)
	m.IncFailure(ctx, topic)

	if got := testhelper.MetricValue(t, h, "wis2_ingest_files_success_total", topic); got != 2 {
		t.Errorf("wis2_ingest_files_success_total = %v; want 2", got)
	}
	if got := testhelper.MetricValue(t, h, "wis2_ingest_files_failed_total", topic); got != 1 {
		t.Errorf("wis2_ingest_files_failed_total = %v; want 1", got)
	}
	if got := testhelper.MetricValue(t, h, "wis2_ingest_files_success_total", "other/topic"); got != 0 {
		t.Errorf("counter leaked to another topic: %v", got)
	}
}

// TestObserveLatency verifies latency is recorded in seconds.
func TestObserveLatency(t *testing.T) {
	m, h := shared(t)
	topic := uniqueTopic("test/latency")

	m.ObserveLatency(context.Background(), topic, time.Now().Add(-2*time.Second))

	if got := testhelper.MetricValue(t, h, "wis2_ingest_latency_seconds_count", topic); got != 1 {
		t.Errorf("wis2_ingest_latency_seconds_count = %v; want 1", got)
	}
	if got := testhelper.MetricValue(t, h, "wis2_ingest_latency_seconds_sum", topic); got < 2 || got > 10 {
		t.Errorf("wis2_ingest_latency_seconds_sum = %v; want ~2 (seconds)", got)
	}
}

// TestNilMetricsIsSafe verifies a nil *Metrics is a no-op, since the MQTT
// client allows running without metrics.
func TestNilMetricsIsSafe(t *testing.T) {
	var m *Metrics
	ctx := context.Background()
	m.IncSuccess(ctx, "t")
	m.IncFailure(ctx, "t")
	m.ObserveLatency(ctx, "t", time.Now())
}
