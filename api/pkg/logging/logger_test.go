package logging_test

import (
	"bytes"
	"strings"
	"testing"

	. "example.com/todos/pkg/logging"
)

func TestInfo(t *testing.T) {
	var buf bytes.Buffer

	l := NewLogger(&buf)

	l.Info("Completed HTTP request", map[string]any{
		"method":      "GET",
		"path":        "/todos",
		"request_id":  "123-456",
		"status":      200,
		"duration_ms": 150,
	})

	got := buf.String()
	if !strings.Contains(got, "method=GET") {
		t.Errorf("expected to find 'method=GET', but did not")
	}
	if !strings.Contains(got, "path=/todos") {
		t.Errorf("expected to find 'path=/todos', but did not")
	}
	if !strings.Contains(got, "request_id=123-456") {
		t.Errorf("expected to find 'request_id=123-456', but did not")
	}
	if !strings.Contains(got, "status=200") {
		t.Errorf("expected to find 'status=200', but did not")
	}
	if !strings.Contains(got, "duration_ms=150") {
		t.Errorf("expected to find 'duration_ms=150', but did not")
	}
}
