package health

import (
	"net/http"
	"sync"
	"time"
	"wis2-ingest/pkg/utils"
)

// ProbeFunc reports whether a health probe succeeds.
type ProbeFunc func() bool

// ConnectionChecker reports whether the MQTT client is connected.
// It is satisfied by *mqtt.Client.
type ConnectionChecker interface {
	Connected() bool
}

// livenessGracePeriod is how long the MQTT client may stay disconnected
// before the liveness probe fails. The client reconnects on its own, so
// short broker outages must not restart the pod; a restart is only the last
// resort when the client never comes back.
var livenessGracePeriod = 5 * time.Minute

// now is the clock used by the liveness probe. Tests replace it.
var now = time.Now

// RegisterProbeHandlers registers HTTP liveness and readiness
// endpoints on the provided ServeMux.
//
//   - /ready succeeds while the MQTT client is connected and subscribed.
//   - /live succeeds unless the client has been disconnected for longer
//     than livenessGracePeriod. In debug mode it always succeeds, to keep
//     the pod running while debugging without a broker.
func RegisterProbeHandlers(mux *http.ServeMux, mqttClient ConnectionChecker) {

	// Readiness
	ready := func() bool {
		return mqttClient.Connected()
	}

	// Liveness
	var mu sync.Mutex
	lastConnected := now() // the grace period also covers the first connection

	live := func() bool {
		if utils.IsDebug() {
			return true
		}

		mu.Lock()
		defer mu.Unlock()

		if mqttClient.Connected() {
			lastConnected = now()
			return true
		}

		return now().Sub(lastConnected) < livenessGracePeriod
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
