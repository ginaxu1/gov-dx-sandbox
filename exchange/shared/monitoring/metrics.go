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
		[]string{"method", "route", "status_code"},
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
		statusCode := strconv.Itoa(rw.statusCode)

		httpRequestsTotal.WithLabelValues(method, route, statusCode).Inc()
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
// Extracts the base route pattern, e.g., /consents/123 -> /consents
func normalizeRoute(path string) string {
	if path == "" || path == "/" {
		return "/"
	}

	// Remove query string
	if qIdx := strings.Index(path, "?"); qIdx >= 0 {
		path = path[:qIdx]
	}

	// For routes with IDs, normalize to base path
	// e.g., /consents/abc123 -> /consents, /data-owner/user@example.com -> /data-owner
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		// Return base path (first two parts: "" and "route")
		return "/" + parts[1]
	}

	return path
}

