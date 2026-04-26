package middleware

import (
	"net/http"
	"strings"

	"github.com/aki/todo-ai/internal/config"
	"go.uber.org/zap"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	config config.CORSConfig
	logger *zap.Logger
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(cfg config.CORSConfig, logger *zap.Logger) *CORSMiddleware {
	return &CORSMiddleware{
		config: cfg,
		logger: logger,
	}
}

// Middleware adds CORS headers to responses
func (c *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		} else if len(c.config.AllowOrigins) == 0 || c.config.AllowOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if c.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if the given origin is in the allowed list
func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	// If wildcard is set, allow all
	for _, allowed := range c.config.AllowOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// Allow subdomain wildcard (e.g., https://*.example.com)
		if strings.HasPrefix(allowed, "*.") {
			// Extract domain from allowed (remove *. prefix)
			domain := strings.TrimPrefix(allowed, "*.")
			// Check if origin ends with the domain
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}