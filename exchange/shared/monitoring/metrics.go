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
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
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
		method := r.Method
		status := strconv.Itoa(rw.statusCode)

		// Normalize route, but use "unknown" for 404s to prevent cardinality explosion
		route := normalizeRoute(r.URL.Path)
		if rw.statusCode == http.StatusNotFound {
			route = "unknown"
		}

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
// Falls back to "unknown" for unrecognized patterns to prevent cardinality explosion
// Examples:
//   - /consents/123 -> /consents/:id
//   - /api/v1/policy/metadata -> /api/v1/policy/metadata (static path preserved)
//   - /data-owner/user@example.com -> /data-owner/:id
//   - /random/unknown/path -> unknown (fallback to prevent cardinality explosion)
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

	// Known static routes that should be preserved
	knownRoutes := map[string]bool{
		"/health":                         true,
		"/metrics":                        true,
		"/debug":                          true,
		"/api/v1/policy/metadata":         true,
		"/api/v1/policy/decide":           true,
		"/api/v1/policy/update-allowlist": true,
	}

	// Check if this is a known route
	fullPath := "/" + strings.Join(parts, "/")
	if knownRoutes[fullPath] {
		return fullPath
	}

	// For paths with 2 segments, check if second looks like an ID
	if len(parts) == 2 {
		if looksLikeID(parts[1]) {
			return "/" + parts[0] + "/:id"
		}
		// Unknown 2-segment path - fallback to prevent cardinality explosion
		return "unknown"
	}

	// For paths with 3+ segments, check if last segment looks like an ID
	if len(parts) >= 3 {
		lastPart := parts[len(parts)-1]
		if looksLikeID(lastPart) {
			// Replace last segment with :id
			normalized := parts[:len(parts)-1]
			normalizedPath := "/" + strings.Join(normalized, "/") + "/:id"
			// Only return if it's a reasonable pattern (max 4 segments)
			if len(normalized) <= 3 {
				return normalizedPath
			}
		}
		// Unknown long path - fallback to prevent cardinality explosion
		return "unknown"
	}

	// Single segment path - only allow known single segments
	if len(parts) == 1 {
		singlePath := "/" + parts[0]
		if knownRoutes[singlePath] {
			return singlePath
		}
		// Unknown single segment - fallback
		return "unknown"
	}

	// Fallback for any unrecognized pattern
	return "unknown"
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
