package health

import (
	"net/http"
	"wis2-ingest/internal/mqtt"
	"wis2-ingest/pkg/utils"
)

// ReadyFunc returns true if service is ready.
type ReadyFunc func() bool

// Register adds health endpoints into the existing HTTP multiplexer, mux.
func Register(mux *http.ServeMux, mqttClient *mqtt.Client) {

	// Readiness
	var ready ReadyFunc = func() bool { return true }

	// Liveness
	var live ReadyFunc = func() bool { return true }

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
