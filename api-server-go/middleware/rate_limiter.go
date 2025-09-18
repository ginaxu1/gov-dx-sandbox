package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	maxReqs  int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		maxReqs:  maxRequests,
		window:   window,
	}
}

// IsAllowed checks if a request from the given IP is allowed
func (rl *RateLimiter) IsAllowed(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Get existing requests for this IP
	requests := rl.requests[clientIP]

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if now.Sub(reqTime) < rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.maxReqs {
		slog.Warn("Rate limit exceeded", "ip", clientIP, "requests", len(validRequests), "limit", rl.maxReqs)
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[clientIP] = validRequests

	return true
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(maxRequests, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !limiter.IsAllowed(clientIP) {
				slog.Warn("Rate limit exceeded", "ip", clientIP, "path", r.URL.Path)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
