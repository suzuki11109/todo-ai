package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// TimeoutMiddleware adds a timeout to requests
type TimeoutMiddleware struct {
	timeout time.Duration
}

// NewTimeoutMiddleware creates a new timeout middleware
func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		timeout: timeout,
	}
}

// Middleware wraps the handler with a timeout
func (t *TimeoutMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), t.timeout)
		defer cancel()

		// Create a custom response writer to capture status code
		rw := &timeoutResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute handler with timeout context
		done := make(chan struct{})
		go func() {
			next.ServeHTTP(rw, r.WithContext(ctx))
			close(done)
		}()

		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestTimeout)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "request timeout",
					"message": "The request took too long to complete",
				})
				return
			}
		}

		// Ensure we write the actual status code
		if rw.statusCode != http.StatusOK {
			w.WriteHeader(rw.statusCode)
		}
	})
}

// timeoutResponseWriter is a wrapper to capture the status code
type timeoutResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *timeoutResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}