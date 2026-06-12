package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	prom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Metrics provides OpenTelemetry metrics tracking for a service.
type Metrics struct {
	filesSuccess metric.Int64Counter
	filesFailed  metric.Int64Counter
	latency      metric.Float64Histogram
}

// New initializes OpenTelemetry metrics for the specified service and
// returns a Metrics instance together with an Prometheus http handler
// that exposes the metrics endpoint.
func New(service string) (*Metrics, http.Handler, error) {

	exporter, err := prom.New()
	if err != nil {
		return nil, nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)

	otel.SetMeterProvider(provider)

	meter := otel.Meter(service)

	filesSuccess, err := meter.Int64Counter("wis2_ingest_files_success_total")
	if err != nil {
		return nil, nil, err
	}

	filesFailed, err := meter.Int64Counter("wis2_ingest_files_failed_total")
	if err != nil {
		return nil, nil, err
	}

	latency, err := meter.Float64Histogram("wis2_ingest_latency_seconds")
	if err != nil {
		return nil, nil, err
	}

	return &Metrics{
		filesSuccess: filesSuccess,
		filesFailed:  filesFailed,
		latency:      latency,
	}, promhttp.Handler(), nil
}
