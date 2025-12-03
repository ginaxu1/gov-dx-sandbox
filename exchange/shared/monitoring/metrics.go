package monitoring

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request counter
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	// HTTP request duration histogram
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	// External Call Metrics
	externalCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_calls_total",
			Help: "Total number of external service calls",
		},
		[]string{"external_target", "external_operation"},
	)

	externalCallErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_call_errors_total",
			Help: "Total number of failed external service calls",
		},
		[]string{"external_target", "external_operation"},
	)

	// External call duration histogram
	externalCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_call_duration_seconds",
			Help:    "External service call duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"external_target", "external_operation"},
	)

	// Business Event Metrics
	businessEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_events_total",
			Help: "Total number of business events",
		},
		[]string{"business_action", "business_outcome"},
	)
)

// Handler returns the Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// HTTPMetricsMiddleware wraps an HTTP handler to record metrics
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		route := normalizeRoute(r.URL.Path)
		method := r.Method
		status := strconv.Itoa(rw.statusCode)

		httpRequestsTotal.WithLabelValues(method, route, status).Inc()
		httpRequestDuration.WithLabelValues(method, route).Observe(duration)
	})
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

// normalizeRoute simplifies route paths for metrics
// Preserves static paths while normalizing dynamic IDs
// Examples:
//   - /consents/123 -> /consents/:id
//   - /api/v1/policy/metadata -> /api/v1/policy/metadata (static path preserved)
//   - /data-owner/user@example.com -> /data-owner/:id
func normalizeRoute(path string) string {
	if path == "" || path == "/" {
		return "/"
	}

	// Remove query string
	if qIdx := strings.Index(path, "?"); qIdx >= 0 {
		path = path[:qIdx]
	}

	parts := strings.Split(path, "/")
	// Remove empty first element from split
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return "/"
	}

	// For paths with 2 segments, check if second looks like an ID
	if len(parts) == 2 {
		if looksLikeID(parts[1]) {
			return "/" + parts[0] + "/:id"
		}
		// Both segments are static, return full path
		return "/" + strings.Join(parts, "/")
	}

	// For paths with 3+ segments, check if last segment looks like an ID
	if len(parts) >= 3 {
		lastPart := parts[len(parts)-1]
		if looksLikeID(lastPart) {
			// Replace last segment with :id
			normalized := parts[:len(parts)-1]
			return "/" + strings.Join(normalized, "/") + "/:id"
		}
		// Last segment is static, keep full path
		return "/" + strings.Join(parts, "/")
	}

	// Single segment path
	return "/" + parts[0]
}

// looksLikeID checks if a string looks like a dynamic ID (UUID, numeric, email, or long alphanumeric)
func looksLikeID(s string) bool {
	if s == "" {
		return false
	}

	// Check for UUID pattern (e.g., "123e4567-e89b-12d3-a456-426614174000" or "consent_abc123")
	if strings.Contains(s, "_") || strings.Contains(s, "-") {
		// Likely an ID with prefix or UUID
		return true
	}

	// Check if it's all numeric (e.g., "123")
	allNumeric := true
	for _, r := range s {
		if r < '0' || r > '9' {
			allNumeric = false
			break
		}
	}
	if allNumeric && len(s) > 0 {
		return true
	}

	// Check if it looks like an email (contains @)
	if strings.Contains(s, "@") {
		return true
	}

	// Check if it's a long alphanumeric string (likely an ID)
	if len(s) > 10 {
		alphanumeric := true
		for _, r := range s {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
				alphanumeric = false
				break
			}
		}
		if alphanumeric {
			return true
		}
	}

	return false
}

// RecordExternalCall records an external service call
func RecordExternalCall(target, operation string, duration time.Duration, err error) {
	externalCallsTotal.WithLabelValues(target, operation).Inc()
	externalCallDuration.WithLabelValues(target, operation).Observe(duration.Seconds())
	if err != nil {
		externalCallErrors.WithLabelValues(target, operation).Inc()
	}
}

// RecordBusinessEvent records a business event
func RecordBusinessEvent(action, outcome string) {
	businessEventsTotal.WithLabelValues(action, outcome).Inc()
}
