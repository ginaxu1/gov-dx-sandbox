package middleware

import (
	"net/http"
	"os"
	"strconv"
)

// CORSMiddleware creates a CORS middleware
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", getCORSMaxAge())

			// Handle preflight (OPTIONS) requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware() func(http.Handler) http.Handler {
	return CORSMiddleware()
}

// getCORSMaxAge gets the CORS max age from environment variable or returns default
func getCORSMaxAge() string {
	if value := os.Getenv("CORS_MAX_AGE"); value != "" {
		// Validate that it's a valid number
		if _, err := strconv.Atoi(value); err == nil {
			return value
		}
	}
	return "86400" // Default: 24 hours
}
