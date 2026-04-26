package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
	"go.uber.org/zap"
)

// RateLimiterMiddleware implements rate limiting using token bucket algorithm
type RateLimiterMiddleware struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
	logger   *zap.Logger
}

// NewRateLimiterMiddleware creates a new rate limiter
// rate: requests per second, burst: maximum burst size
func NewRateLimiterMiddleware(r rate.Limit, burst int, logger *zap.Logger) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rate:     r,
		burst:    burst,
		logger:   logger,
	}
}

// getLimiter returns a rate limiter for the given IP address
func (r *RateLimiterMiddleware) getLimiter(ip string) *rate.Limiter {
	limiter, exists := r.limiters.Load(ip)
	if !exists {
		limiter = rate.NewLimiter(r.rate, r.burst)
		r.limiters.Store(ip, limiter)
	}
	return limiter.(*rate.Limiter)
}

// Middleware limits the number of requests per IP
func (r *RateLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get client IP (considering proxy headers)
		ip := getClientIP(req)

		limiter := r.getLimiter(ip)

		if !limiter.Allow() {
			r.logger.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", req.URL.Path),
			)
			respondTooManyRequests(w)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// getClientIP extracts the real client IP from request
func getClientIP(req *http.Request) string {
	// Check X-Forwarded-For header
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

func respondTooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "rate limit exceeded",
		"message": "Too many requests, please try again later",
	})
}