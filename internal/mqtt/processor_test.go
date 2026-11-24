package mqtt

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"wis2-ingest/internal/config"
	"wis2-ingest/pkg/testhelper"

	"go.uber.org/zap"
)

const validBUFR = "BUFR\x00\x00\x00\x04fake-bufr-body7777"

// fileServer serves body with status and counts requests.
func fileServer(t *testing.T, status int, body string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// listFiles returns the names of the regular files in dir (non-recursive).
func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// assertNoLeftovers checks no tmp or lock files remain after processing.
func assertNoLeftovers(t *testing.T, outputDir string) {
	t.Helper()
	if pending := listFiles(t, filepath.Join(outputDir, "pending_validation")); len(pending) != 0 {
		t.Errorf("pending_validation not empty: %v", pending)
	}
	for _, name := range listFiles(t, outputDir) {
		if filepath.Ext(name) == ".lock" {
			t.Errorf("lock file left behind: %s", name)
		}
	}
}

// TestProcessStoresValidBUFR verifies the happy path: the file is downloaded,
// validated and moved to the output directory, and success is counted.
func TestProcessStoresValidBUFR(t *testing.T) {
	topic := uniqueTopic("test/process/success")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	link := bufrLink(srv.URL + "/file.bufr")

	NewProcessor(c).Process(message(t, topic, link))

	finalPath := expectedFinalPath(c.cfg.OutputDir, topic, link)
	data, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("expected stored file %s: %v", finalPath, err)
	}
	if string(data) != validBUFR {
		t.Errorf("stored content = %q; want %q", data, validBUFR)
	}
	if hits.Load() != 1 {
		t.Errorf("server hits = %d; want 1", hits.Load())
	}
	assertNoLeftovers(t, c.cfg.OutputDir)
	if got := successCount(t, topic); got != 1 {
		t.Errorf("success count = %v; want 1", got)
	}
	if got := failureCount(t, topic); got != 0 {
		t.Errorf("failure count = %v; want 0", got)
	}
}

// TestProcessIsIdempotent verifies an already stored file is not downloaded
// again when the same notification is received twice.
func TestProcessIsIdempotent(t *testing.T) {
	topic := uniqueTopic("test/process/idempotent")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	msg := message(t, topic, bufrLink(srv.URL+"/file.bufr"))

	NewProcessor(c).Process(msg)
	NewProcessor(c).Process(msg)

	if hits.Load() != 1 {
		t.Errorf("server hits = %d; want 1 (second message must be skipped)", hits.Load())
	}
	if got := successCount(t, topic); got != 1 {
		t.Errorf("success count = %v; want 1", got)
	}
}

// TestProcessSkipsLockedLink verifies a link already being processed
// (lock file present) is skipped.
func TestProcessSkipsLockedLink(t *testing.T) {
	topic := uniqueTopic("test/process/locked")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	link := bufrLink(srv.URL + "/file.bufr")

	lock := expectedFinalPath(c.cfg.OutputDir, topic, link) + ".lock"
	if err := os.WriteFile(lock, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	NewProcessor(c).Process(message(t, topic, link))

	if hits.Load() != 0 {
		t.Errorf("server hits = %d; want 0", hits.Load())
	}
	if _, err := os.Stat(lock); err != nil {
		t.Errorf("lock owned by another worker must not be removed: %v", err)
	}
}

// TestProcessStoresEachLink verifies every valid link in a notification is
// processed, and the same URL on different topics is stored separately.
func TestProcessStoresEachLink(t *testing.T) {
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	a, b := bufrLink(srv.URL+"/a.bufr"), bufrLink(srv.URL+"/b.bufr")

	NewProcessor(c).Process(message(t, "test/process/multi/1", a, b))
	NewProcessor(c).Process(message(t, "test/process/multi/2", a))

	for _, p := range []string{
		expectedFinalPath(c.cfg.OutputDir, "test/process/multi/1", a),
		expectedFinalPath(c.cfg.OutputDir, "test/process/multi/1", b),
		expectedFinalPath(c.cfg.OutputDir, "test/process/multi/2", a),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected stored file %s: %v", p, err)
		}
	}
	if hits.Load() != 3 {
		t.Errorf("server hits = %d; want 3", hits.Load())
	}
}

// TestProcessIgnoresInvalidLinks verifies links that are not canonical
// octet-stream downloads are never requested.
func TestProcessIgnoresInvalidLinks(t *testing.T) {
	topic := uniqueTopic("test/process/invalid-links")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)

	NewProcessor(c).Process(message(t, topic,
		Link{Href: srv.URL, Rel: "via", Type: "application/octet-stream"},
		Link{Href: srv.URL, Rel: "canonical", Type: "text/html"},
		Link{Href: "", Rel: "canonical", Type: "application/octet-stream"},
	))

	if hits.Load() != 0 {
		t.Errorf("server hits = %d; want 0", hits.Load())
	}
	if files := listFiles(t, c.cfg.OutputDir); len(files) != 0 {
		t.Errorf("unexpected files: %v", files)
	}
}

// TestProcessFailures verifies each failure mode stores nothing, leaves no
// temporary files behind and is counted as a failure.
func TestProcessFailures(t *testing.T) {
	notBUFR, _ := fileServer(t, http.StatusOK, "<html>error page</html>")
	unauthorized, _ := fileServer(t, http.StatusUnauthorized, "")
	closed := httptest.NewServer(http.NotFoundHandler())
	closedURL := closed.URL
	closed.Close()

	tests := []struct {
		name  string
		topic string
		href  string
	}{
		{name: "content is not BUFR", topic: uniqueTopic("test/process/fail/not-bufr"), href: notBUFR.URL},
		{name: "401 unauthorized", topic: uniqueTopic("test/process/fail/401"), href: unauthorized.URL},
		{name: "server unreachable", topic: uniqueTopic("test/process/fail/unreachable"), href: closedURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t)
			NewProcessor(c).Process(message(t, tt.topic, bufrLink(tt.href)))

			if files := listFiles(t, c.cfg.OutputDir); len(files) != 0 {
				t.Errorf("unexpected files stored: %v", files)
			}
			assertNoLeftovers(t, c.cfg.OutputDir)
			if got := failureCount(t, tt.topic); got != 1 {
				t.Errorf("failure count = %v; want 1", got)
			}
		})
	}
}

// TestProcessNonOKStatus verifies non-200 responses (e.g. 403, 404, 500)
// store nothing and are counted as failures, so they show up in the
// dashboards.
func TestProcessNonOKStatus(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			topic := uniqueTopic("test/process/status/" + http.StatusText(status))
			srv, _ := fileServer(t, status, validBUFR)
			c := newTestClient(t)

			NewProcessor(c).Process(message(t, topic, bufrLink(srv.URL)))

			if files := listFiles(t, c.cfg.OutputDir); len(files) != 0 {
				t.Errorf("unexpected files stored: %v", files)
			}
			assertNoLeftovers(t, c.cfg.OutputDir)
			if got := successCount(t, topic); got != 0 {
				t.Errorf("success count = %v; want 0", got)
			}
			if got := failureCount(t, topic); got != 1 {
				t.Errorf("failure count = %v; want 1", got)
			}
		})
	}
}

// TestProcessInvalidJSON verifies malformed payloads are counted as failures.
func TestProcessInvalidJSON(t *testing.T) {
	topic := uniqueTopic("test/process/invalid-json")
	c := newTestClient(t)

	NewProcessor(c).Process(testhelper.MqttMessage{TopicField: topic, PayloadField: []byte("{not json")})

	if got := failureCount(t, topic); got != 1 {
		t.Errorf("failure count = %v; want 1", got)
	}
}

// TestProcessOutputDirNotCreatable verifies a failure is counted when the
// pending directory cannot be created.
func TestProcessOutputDirNotCreatable(t *testing.T) {
	topic := uniqueTopic("test/process/bad-output-dir")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)

	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	c.cfg.OutputDir = file

	NewProcessor(c).Process(message(t, topic, bufrLink(srv.URL)))

	if hits.Load() != 0 {
		t.Errorf("server hits = %d; want 0", hits.Load())
	}
	if got := failureCount(t, topic); got != 1 {
		t.Errorf("failure count = %v; want 1", got)
	}
}

// TestProcessInvalidURL verifies an unparsable href is skipped without
// leaving a lock file behind.
func TestProcessInvalidURL(t *testing.T) {
	topic := uniqueTopic("test/process/invalid-url")
	c := newTestClient(t)

	NewProcessor(c).Process(message(t, topic, bufrLink("http://[::1")))

	if files := listFiles(t, c.cfg.OutputDir); len(files) != 0 {
		t.Errorf("unexpected files: %v", files)
	}
}

// TestProcessUsesAuthPolicy verifies the credentials of the matching auth
// policy are sent as HTTP basic auth.
func TestProcessUsesAuthPolicy(t *testing.T) {
	topic := uniqueTopic("origin/a/wis2/test/process/auth")
	var gotUser, gotPass string
	var gotAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, gotAuth = r.BasicAuth()
		_, _ = w.Write([]byte(validBUFR))
	}))
	t.Cleanup(srv.Close)

	c := newTestClient(t,
		config.AuthPolicy{Name: "other", Topic: "origin/b/#", Username: "wrong", Password: config.SecretRef{Value: "wrong"}},
		config.AuthPolicy{Name: "match", Topic: "origin/a/wis2/#", Username: "user", Password: config.SecretRef{Value: "pass"}},
	)

	NewProcessor(c).Process(message(t, topic, bufrLink(srv.URL)))

	if !gotAuth || gotUser != "user" || gotPass != "pass" {
		t.Errorf("basic auth = (%q, %q, %v); want (user, pass, true)", gotUser, gotPass, gotAuth)
	}
	if got := successCount(t, topic); got != 1 {
		t.Errorf("success count = %v; want 1", got)
	}
}

// TestProcessWithoutAuthPolicy verifies no credentials are sent when no
// policy matches the topic.
func TestProcessWithoutAuthPolicy(t *testing.T) {
	var gotAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, gotAuth = r.BasicAuth()
		_, _ = w.Write([]byte(validBUFR))
	}))
	t.Cleanup(srv.Close)

	c := newTestClient(t, config.AuthPolicy{Topic: "origin/b/#", Username: "u", Password: config.SecretRef{Value: "p"}})
	NewProcessor(c).Process(message(t, "origin/a/test/process/no-auth", bufrLink(srv.URL)))

	if gotAuth {
		t.Error("basic auth sent although no policy matches the topic")
	}
}

// TestProcessSecretResolutionFailure verifies the download is not attempted
// (and a failure is counted) when the policy secret cannot be resolved.
func TestProcessSecretResolutionFailure(t *testing.T) {
	topic := uniqueTopic("test/process/secret-failure")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t, config.AuthPolicy{
		Name:     "broken",
		Topic:    "#",
		Username: "u",
		Password: config.SecretRef{SecretName: "missing-key"}, // secretKey missing
	})

	NewProcessor(c).Process(message(t, topic, bufrLink(srv.URL)))

	if hits.Load() != 0 {
		t.Errorf("server hits = %d; want 0", hits.Load())
	}
	if got := failureCount(t, topic); got != 1 {
		t.Errorf("failure count = %v; want 1", got)
	}
}

// TestProcessWithoutMetrics verifies processing works when the client was
// created without metrics (NewClient accepts nil).
func TestProcessWithoutMetrics(t *testing.T) {
	srv, _ := fileServer(t, http.StatusOK, validBUFR)
	c := NewClient(&config.IngestOptions{OutputDir: t.TempDir()}, zap.NewNop(), nil)
	link := bufrLink(srv.URL)

	NewProcessor(c).Process(message(t, "test/process/no-metrics", link))
	NewProcessor(c).Process(testhelper.MqttMessage{TopicField: "t", PayloadField: []byte("bad")})

	if _, err := os.Stat(expectedFinalPath(c.cfg.OutputDir, "test/process/no-metrics", link)); err != nil {
		t.Errorf("expected stored file: %v", err)
	}
}

// TestMessageHandlerDelegatesToProcessor verifies the MQTT entrypoint runs
// the processing pipeline.
func TestMessageHandlerDelegatesToProcessor(t *testing.T) {
	topic := uniqueTopic("test/handler")
	srv, _ := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	link := bufrLink(srv.URL)

	c.messageHandler(message(t, topic, link))

	if _, err := os.Stat(expectedFinalPath(c.cfg.OutputDir, topic, link)); err != nil {
		t.Errorf("expected stored file: %v", err)
	}
}

// TestIsValidLink verifies link eligibility rules.
func TestIsValidLink(t *testing.T) {
	tests := []struct {
		name string
		link Link
		want bool
	}{
		{name: "canonical octet-stream", link: bufrLink("http://x/f"), want: true},
		{name: "octet-stream with parameters", link: Link{Href: "http://x/f", Rel: "canonical", Type: "application/octet-stream; charset=binary"}, want: true},
		{name: "not canonical", link: Link{Href: "http://x/f", Rel: "via", Type: "application/octet-stream"}},
		{name: "wrong type", link: Link{Href: "http://x/f", Rel: "canonical", Type: "application/json"}},
		{name: "empty href", link: Link{Rel: "canonical", Type: "application/octet-stream"}},
		{name: "zero value", link: Link{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLink(tt.link); got != tt.want {
				t.Errorf("isValidLink(%+v) = %v; want %v", tt.link, got, tt.want)
			}
		})
	}
}

// TestProcessTakesOverStaleLock verifies a lock left behind by a crashed
// process does not block the file forever, while a recent lock is kept.
func TestProcessTakesOverStaleLock(t *testing.T) {
	topic := uniqueTopic("test/process/stale-lock")
	srv, hits := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	link := bufrLink(srv.URL + "/file.bufr")

	finalPath := expectedFinalPath(c.cfg.OutputDir, topic, link)
	lock := finalPath + ".lock"
	if err := os.WriteFile(lock, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * staleLockAge)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}

	NewProcessor(c).Process(message(t, topic, link))

	if hits.Load() != 1 {
		t.Errorf("server hits = %d; want 1", hits.Load())
	}
	if _, err := os.Stat(finalPath); err != nil {
		t.Errorf("expected stored file after taking over the stale lock: %v", err)
	}
	assertNoLeftovers(t, c.cfg.OutputDir)
}

// TestProcessOverwritesLeftoverTmpFile verifies a temporary file left
// behind by a crash does not make the download fail forever.
func TestProcessOverwritesLeftoverTmpFile(t *testing.T) {
	topic := uniqueTopic("test/process/leftover-tmp")
	srv, _ := fileServer(t, http.StatusOK, validBUFR)
	c := newTestClient(t)
	link := bufrLink(srv.URL + "/file.bufr")

	finalPath := expectedFinalPath(c.cfg.OutputDir, topic, link)
	pending := filepath.Join(c.cfg.OutputDir, "pending_validation")
	if err := os.MkdirAll(pending, 0o755); err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(pending, strings.TrimSuffix(filepath.Base(finalPath), ".bufr")+".tmp")
	if err := os.WriteFile(tmp, []byte("partial download from a crashed process, longer than the new one"), 0o644); err != nil {
		t.Fatal(err)
	}

	NewProcessor(c).Process(message(t, topic, link))

	data, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("expected stored file: %v", err)
	}
	if string(data) != validBUFR {
		t.Errorf("stored content = %q; want %q", data, validBUFR)
	}
	if got := successCount(t, topic); got != 1 {
		t.Errorf("success count = %v; want 1", got)
	}
}

// TestProcessDownloadTimeout verifies a server that never answers fails the
// download instead of blocking the MQTT message handler forever.
func TestProcessDownloadTimeout(t *testing.T) {
	topic := uniqueTopic("test/process/timeout")
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })

	orig := httpClient
	httpClient = &http.Client{Timeout: 100 * time.Millisecond}
	t.Cleanup(func() { httpClient = orig })

	c := newTestClient(t)
	done := make(chan struct{})
	go func() {
		NewProcessor(c).Process(message(t, topic, bufrLink(srv.URL+"/slow.bufr")))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Process did not return: download has no timeout")
	}
	if got := failureCount(t, topic); got != 1 {
		t.Errorf("failure count = %v; want 1", got)
	}
	assertNoLeftovers(t, c.cfg.OutputDir)
}

// TestAcquireLock verifies the lock is exclusive, and that only a stale
// lock is taken over.
func TestAcquireLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.lock")

	f, ok := acquireLock(path)
	if !ok {
		t.Fatal("acquireLock() on a free path = false; want true")
	}
	_ = f.Close()

	if _, ok := acquireLock(path); ok {
		t.Error("acquireLock() on a held lock = true; want false")
	}

	old := time.Now().Add(-2 * staleLockAge)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	f, ok = acquireLock(path)
	if !ok {
		t.Fatal("acquireLock() on a stale lock = false; want true")
	}
	_ = f.Close()

	if _, ok := acquireLock(filepath.Join(t.TempDir(), "missing", "file.lock")); ok {
		t.Error("acquireLock() in a missing directory = true; want false")
	}
}
