package health

import (
	"fmt"
	"net/http"
	"wis2-ingest/internal/mqtt"
	"wis2-ingest/pkg/utils"
)

// ReadyFunc returns true if service is ready.
type ReadyFunc func() bool

// Start starts K8s probes.
// Implementation prepared for safe debug mode usage.
func Start(mqttClient *mqtt.Client) {
	if utils.IsDebug() {
		initializeProbes(
			func() bool { return true },
			func() bool { return true },
		)
	} else {
		initializeProbes(
			func() bool { return true },
			func() bool { return mqttClient.Connected() },
		)
	}
}

// initializeProbes iniciates an HTTP server for readiness and liveness.
// ready assess whether service is ready.
// live is optional; if nil, always assumes true.
// The server runs in a goroutine and doesn't block execution.
func initializeProbes(ready ReadyFunc, live ReadyFunc) {
	writeResp := func(w http.ResponseWriter, code int, msg string) {
		w.WriteHeader(code)
		if _, err := w.Write([]byte(msg)); err != nil {
			fmt.Printf("failed to write response: %v\n", err)
		}
	}

	go func() {
		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			if ready() {
				writeResp(w, http.StatusOK, "ok")
			} else {
				writeResp(w, http.StatusServiceUnavailable, "not ready")
			}
		})

		http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
			if live == nil || live() {
				writeResp(w, http.StatusOK, "alive")
			} else {
				writeResp(w, http.StatusServiceUnavailable, "not alive")
			}
		})

		addr := ":8080"
		fmt.Printf("Health HTTP server listening on %s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Health HTTP server failed: %v\n", err)
		}
	}()
}
