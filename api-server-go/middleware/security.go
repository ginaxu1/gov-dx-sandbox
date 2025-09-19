package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Remove server information
		w.Header().Set("Server", "")

		next.ServeHTTP(w, r)
	})
}

// InputValidation validates common input patterns
func InputValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for suspicious patterns in URL
		url := r.URL.Path
		if containsSuspiciousPatterns(url) {
			slog.Warn("Suspicious URL pattern detected", "url", url, "ip", getClientIP(r))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Check Content-Type for POST/PUT requests
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				slog.Warn("Invalid Content-Type", "contentType", contentType, "ip", getClientIP(r))
				http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityLogging logs security-relevant events
func SecurityLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := getClientIP(r)
		userAgent := r.Header.Get("User-Agent")

		// Log the request
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"ip", clientIP,
			"userAgent", userAgent,
			"timestamp", start,
		)

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Log the response
		duration := time.Since(start)
		slog.Info("HTTP Response",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"ip", clientIP,
		)

		// Log security events
		if wrapped.statusCode >= 400 {
			slog.Warn("HTTP Error Response",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"ip", clientIP,
				"userAgent", userAgent,
			)
		}
	})
}

// containsSuspiciousPatterns checks for common attack patterns
func containsSuspiciousPatterns(url string) bool {
	suspicious := []string{
		"../", "..\\", "..%2f", "..%5c", // Path traversal
		"<script", "javascript:", "vbscript:", // XSS
		"union select", "drop table", "insert into", // SQL injection
		"exec(", "eval(", "system(", // Code injection
		"../../", "..\\..", // More path traversal
	}

	urlLower := strings.ToLower(url)
	for _, pattern := range suspicious {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}

	return false
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
