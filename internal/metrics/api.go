package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// FILES
// IncSuccess increases counter by 1 when a file from a given topic is
// successfully processed.
func (m *Metrics) IncSuccess(ctx context.Context, topic string) {
	if m == nil {
		return
	}

	m.filesSuccess.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("topic", topic),
		),
	)
}

// IncFailure increases counter by 1 when a file from a given topic
// fails to be processed.
func (m *Metrics) IncFailure(ctx context.Context, topic string) {
	if m == nil {
		return
	}

	m.filesFailed.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("topic", topic),
		),
	)
}

// ObserveLatency records the elapsed processing time for a file from a given topic.
func (m *Metrics) ObserveLatency(ctx context.Context, topic string, start time.Time) {
	if m == nil {
		return
	}

	m.latency.Record(ctx,
		time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String("topic", topic),
		),
	)
}
