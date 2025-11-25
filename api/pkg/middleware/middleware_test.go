package middleware_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	. "example.com/todos/pkg/middleware"
)

// TestRequestIDMiddleware_GeneratesAndPropagatesID verifies that a new
// request ID is generated when the client does not provide one, and that
// it is available both in the response header and in the request context.
func TestRequestIDMiddleware_GeneratesAndPropagatesID(t *testing.T) {
	var seenCtxReqID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenCtxReqID = RequestIDFromContext(r.Context())
		if seenCtxReqID == "" {
			t.Errorf("expected non-empty request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	h := RequestIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	respReqID := rr.Header().Get("X-Request-Id")
	if respReqID == "" {
		t.Fatalf("expected X-Request-Id header to be set")
	}

	if seenCtxReqID != respReqID {
		t.Fatalf("expected context request ID %q to match response header %q", seenCtxReqID, respReqID)
	}
}

// TestRequestIDMiddleware_PreservesExistingID verifies that if the client
// sends an X-Request-Id header, the middleware reuses that value instead of
// generating a new one, and propagates it into the context.
func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	const clientReqID = "client-provided-id-123"

	var seenCtxReqID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenCtxReqID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	h := RequestIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	req.Header.Set("X-Request-Id", clientReqID)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	respReqID := rr.Header().Get("X-Request-Id")
	if respReqID != clientReqID {
		t.Fatalf("expected X-Request-Id header %q, got %q", clientReqID, respReqID)
	}

	if seenCtxReqID != clientReqID {
		t.Fatalf("expected context request ID %q, got %q", clientReqID, seenCtxReqID)
	}
}

// TestLoggingMiddleware_LogsBasicRequestDetails verifies that the logging
// middleware produces exactly one structured log entry per request, and that
// it includes method, path, status code, and request ID.
func TestLoggingMiddleware_LogsBasicRequestDetails(t *testing.T) {
	logger := newFakeLogger()

	// Final handler that sets a specific status code and also verifies
	// that a request ID is available in the context.
	var handlerCtxReqID string
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCtxReqID = RequestIDFromContext(r.Context())
		if handlerCtxReqID == "" {
			t.Errorf("expected request ID in context within final handler")
		}
		w.WriteHeader(http.StatusCreated) // 201
	})

	// Chain: request ID first, then logging middleware, then final handler.
	h := RequestIDMiddleware(LoggingMiddleware(logger, finalHandler))

	req := httptest.NewRequest(http.MethodPost, "/todos/123", nil)
	req.Header.Set("X-Request-Id", "req-abc-123")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.msg != "Completed HTTP request" {
		t.Errorf("expected log message to be 'Completed HTTP request' but got '%s'", entry.msg)
	}

	// Validate core fields.
	method, ok := entry.fields["method"].(string)
	if !ok || method != http.MethodPost {
		t.Errorf("expected method %q, got %#v", http.MethodPost, entry.fields["method"])
	}

	path, ok := entry.fields["path"].(string)
	if !ok || path != "/todos/123" {
		t.Errorf("expected path %q, got %#v", "/todos/123", entry.fields["path"])
	}

	status, ok := entry.fields["status"].(int)
	if !ok || status != http.StatusCreated {
		t.Errorf("expected status %d, got %#v", http.StatusCreated, entry.fields["status"])
	}

	reqID, ok := entry.fields["request_id"].(string)
	if !ok || reqID == "" {
		t.Fatalf("expected non-empty request_id field, got %#v", entry.fields["request_id"])
	}

	_, ok = entry.fields["duration_ms"].(int64)
	if !ok {
		t.Fatalf("expected to find duration field, but did not")
	}

	// The logged request_id should match what the handler saw in context.
	if reqID != handlerCtxReqID {
		t.Fatalf("expected logged request_id %q to match handler context %q", reqID, handlerCtxReqID)
	}
}

// TestLoggingMiddleware_CallsNextHandler ensures that the logging middleware
// does not short-circuit the request and always calls the next handler.
func TestLoggingMiddleware_CallsNextHandler(t *testing.T) {
	logger := newFakeLogger()

	called := false
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h := LoggingMiddleware(logger, finalHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200 from final handler, got %d", rr.Result().StatusCode)
	}
	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
}

// fakeLogger implements Logger and records structured log entries
// for verification in tests.
type fakeLogger struct {
	mu      sync.Mutex
	entries []logEntry
}

type logEntry struct {
	msg    string
	fields map[string]any
}

var _ Logger = (*fakeLogger)(nil)

func newFakeLogger() *fakeLogger {
	return &fakeLogger{}
}

func (l *fakeLogger) Info(msg string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Make a shallow copy so tests aren't affected by later mutation.
	cp := make(map[string]any, len(fields))
	maps.Copy(cp, fields)
	l.entries = append(l.entries, logEntry{
		msg:    msg,
		fields: cp,
	})
}

func (l *fakeLogger) Entries() []logEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]logEntry, len(l.entries))
	copy(out, l.entries)
	return out
}
