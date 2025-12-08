package middleware

import (
	"net/http"
)

// responseWriter is a wrapper around http.ResponseWriter to capture the status code
// This is used by the OpenTelemetry metrics middleware in otel_metrics.go
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware is defined in otel_metrics.go
// It uses OpenTelemetry under the hood and supports multiple exporters (Prometheus, OTLP, etc.)
