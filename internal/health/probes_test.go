package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeConn is a ConnectionChecker with a fixed state.
type fakeConn bool

func (f fakeConn) Connected() bool { return bool(f) }

func probe(t *testing.T, mux *http.ServeMux, path string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code, rec.Body.String()
}

// fakeClock replaces the probe clock for the duration of the test and
// returns a function that moves it forward.
func fakeClock(t *testing.T) func(time.Duration) {
	t.Helper()
	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	orig := now
	now = func() time.Time { return current }
	t.Cleanup(func() { now = orig })
	return func(d time.Duration) { current = current.Add(d) }
}

// TestProbes verifies /ready reflects the MQTT connection, while /live
// tolerates a disconnection within the grace period (and always succeeds in
// debug mode).
func TestProbes(t *testing.T) {
	tests := []struct {
		name      string
		debug     string
		connected bool
		ready     int
		readyBody string
		live      int
		liveBody  string
	}{
		{
			name: "connected", connected: true,
			ready: http.StatusOK, readyBody: "ok",
			live: http.StatusOK, liveBody: "alive",
		},
		{
			name: "disconnected within the grace period", connected: false,
			ready: http.StatusServiceUnavailable, readyBody: "not ready",
			live: http.StatusOK, liveBody: "alive",
		},
		{
			name: "disconnected in debug mode", debug: "true", connected: false,
			ready: http.StatusServiceUnavailable, readyBody: "not ready",
			live: http.StatusOK, liveBody: "alive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DEBUG", tt.debug)
			_ = fakeClock(t)

			mux := http.NewServeMux()
			RegisterProbeHandlers(mux, fakeConn(tt.connected))

			if code, body := probe(t, mux, "/ready"); code != tt.ready || body != tt.readyBody {
				t.Errorf("/ready = %d %q; want %d %q", code, body, tt.ready, tt.readyBody)
			}
			if code, body := probe(t, mux, "/live"); code != tt.live || body != tt.liveBody {
				t.Errorf("/live = %d %q; want %d %q", code, body, tt.live, tt.liveBody)
			}
		})
	}
}

// probeState is a ConnectionChecker whose state can change after registration.
type probeState struct{ connected bool }

func (p *probeState) Connected() bool { return p.connected }

// TestProbesFollowConnectionChanges verifies the probes are evaluated per
// request, not once at registration time.
func TestProbesFollowConnectionChanges(t *testing.T) {
	t.Setenv("DEBUG", "")
	_ = fakeClock(t)
	state := &probeState{}
	mux := http.NewServeMux()
	RegisterProbeHandlers(mux, state)

	if code, _ := probe(t, mux, "/ready"); code != http.StatusServiceUnavailable {
		t.Fatalf("/ready before connect = %d; want 503", code)
	}
	state.connected = true
	if code, _ := probe(t, mux, "/ready"); code != http.StatusOK {
		t.Fatalf("/ready after connect = %d; want 200", code)
	}
	if code, _ := probe(t, mux, "/live"); code != http.StatusOK {
		t.Fatalf("/live after connect = %d; want 200", code)
	}
}

// TestLivenessGracePeriod verifies /live only fails after the client has
// been disconnected for longer than the grace period, counted from the last
// time it was seen connected (or from startup), and recovers on reconnect.
func TestLivenessGracePeriod(t *testing.T) {
	t.Setenv("DEBUG", "")
	advance := fakeClock(t)
	state := &probeState{}
	mux := http.NewServeMux()
	RegisterProbeHandlers(mux, state)

	advance(livenessGracePeriod - time.Second)
	if code, _ := probe(t, mux, "/live"); code != http.StatusOK {
		t.Fatalf("/live just before the grace period since startup = %d; want 200", code)
	}

	advance(2 * time.Second)
	if code, body := probe(t, mux, "/live"); code != http.StatusServiceUnavailable || body != "not alive" {
		t.Fatalf("/live after the grace period since startup = %d %q; want 503 \"not alive\"", code, body)
	}

	state.connected = true
	if code, _ := probe(t, mux, "/live"); code != http.StatusOK {
		t.Fatalf("/live after reconnecting = %d; want 200", code)
	}

	state.connected = false
	advance(livenessGracePeriod - time.Second)
	if code, _ := probe(t, mux, "/live"); code != http.StatusOK {
		t.Fatalf("/live within the grace period since the last connection = %d; want 200", code)
	}
	advance(2 * time.Second)
	if code, _ := probe(t, mux, "/live"); code != http.StatusServiceUnavailable {
		t.Fatalf("/live after the grace period since the last connection = %d; want 503", code)
	}
}
