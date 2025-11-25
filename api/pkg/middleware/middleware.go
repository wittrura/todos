package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type requestIDKeyType struct{}

var requestIDKey = requestIDKeyType{}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestID string
		if header := r.Header.Get("X-Request-ID"); header != "" {
			requestID = header
		} else {
			requestID = RequestIDFromContext(r.Context())
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

type Logger interface {
	Info(msg string, fields map[string]any)
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	Status int
}

// WriteHeader captures the status code before calling the underlying WriteHeader.
func (rec *statusRecorder) WriteHeader(code int) {
	rec.Status = code
	rec.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         http.StatusOK, // Default status if WriteHeader isn't called
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		logger.Info("Completed HTTP request", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"request_id":  RequestIDFromContext(r.Context()),
			"status":      recorder.Status,
			"duration_ms": duration.Milliseconds(),
		})

	})
}

func RequestIDFromContext(ctx context.Context) string {
	if id := ctx.Value(requestIDKey); id != nil {
		return id.(string)
	}
	return uuid.New().String()
}
