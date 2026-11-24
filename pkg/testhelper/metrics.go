package testhelper

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// MetricValue scrapes a Prometheus handler and returns the value of the
// sample named name whose labels include topic="<topic>". It returns 0 when
// the sample is not present (e.g. a counter that was never incremented).
func MetricValue(t *testing.T, h http.Handler, name, topic string) float64 {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics endpoint returned %d: %s", rec.Code, rec.Body.String())
	}

	label := `topic="` + topic + `"`
	sc := bufio.NewScanner(rec.Body)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, name+"{") || !strings.Contains(line, label) {
			continue
		}
		fields := strings.Fields(line)
		v, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil {
			t.Fatalf("cannot parse metric line %q: %v", line, err)
		}
		return v
	}
	return 0
}
