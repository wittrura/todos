package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "example.com/todos/pkg/handlers"
)

// TestMetricsHandler_ExposesPrometheusMetrics verifies that the metrics handler
// responds with a successful status code and a body that looks like Prometheus
// text exposition format, including some of the default Go runtime metrics.
func TestMetricsHandler_ExposesPrometheusMetrics(t *testing.T) {
	h := NewMetricsHandler()
	if h == nil {
		t.Fatal("metricsHandler returned nil handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 from /metrics, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Fatalf("expected Content-Type header to be set on /metrics")
	}
	// promhttp.Handler typically returns:
	// "text/plain; version=0.0.4; charset=utf-8"
	if !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("expected Content-Type to start with %q, got %q", "text/plain", ct)
	}

	bodyBytes := rr.Body.Bytes()
	body := string(bodyBytes)

	if !strings.Contains(body, "# HELP") {
		t.Fatalf("expected /metrics body to contain a HELP comment, got:\n%s", body[:min(200, len(body))])
	}

	// "go_gc_duration_seconds" is a standard metric from the Go client.
	if !strings.Contains(body, "go_gc_duration_seconds") {
		t.Fatalf("expected /metrics body to contain Go runtime metric %q, got:\n%s", "go_gc_duration_seconds", body[:min(2000, len(body))])
	}
}

// TestMetricsHandler_IsSafeToCallMultipleTimes ensures that the handler can be
// invoked multiple times without panicking and always returns 200. This guards
// against any accidental stateful behavior in your metrics setup.
func TestMetricsHandler_IsSafeToCallMultipleTimes(t *testing.T) {
	h := NewMetricsHandler()
	if h == nil {
		t.Fatal("metricsHandler returned nil handler")
	}

	for i := range 3 {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("call %d: expected status 200 from /metrics, got %d", i+1, rr.Result().StatusCode)
		}
	}
}

// min is a small helper for truncating long bodies in error messages.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
