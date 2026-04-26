package middleware

import (
	"encoding/json"
	"net/http"
	"runtime"

	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and logs the error
type RecoveryMiddleware struct {
	logger *zap.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{logger: logger}
}

// Middleware recovers from panics and returns a 500 error
func (r *RecoveryMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				r.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", req.URL.Path),
					zap.String("method", req.Method),
					zap.String("stack", string(stack[:length])),
				)

				// Return generic error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "internal server error",
				})
			}
		}()

		next.ServeHTTP(w, req)
	})
}