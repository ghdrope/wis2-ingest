package health

import (
	"net/http"
	"wis2-ingest/internal/mqtt"
	"wis2-ingest/pkg/utils"
)

// ProbeFunc reports whether a health probe succeeds.
type ProbeFunc func() bool

// RegisterProbeHandlers registers HTTP liveness and readiness
// endpoints on the provided ServeMux.
func RegisterProbeHandlers(mux *http.ServeMux, mqttClient *mqtt.Client) {

	// Readiness
	ready := func() bool {
		return mqttClient.Connected()
	}

	// Liveness
	live := func() bool {
		return true
	}

	if !utils.IsDebug() {
		live = func() bool {
			return mqttClient.Connected()
		}
	}

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if ready() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
	})

	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		if live() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("alive"))
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not alive"))
	})
}
