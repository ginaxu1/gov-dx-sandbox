package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

var (
	// Metrics instruments
	httpRequestsCounter metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	metricsHandler      http.Handler
	initialized         bool
	initOnce            sync.Once
	initErr             error
)

// Config holds the configuration for OpenTelemetry metrics
type Config struct {
	// ExporterType can be "prometheus", "otlp", or "none" (disabled)
	ExporterType string
	// ServiceName is the name of the service
	ServiceName string
	// OTLPEndpoint is the OTLP endpoint URL (for Datadog, New Relic, etc.)
	OTLPEndpoint string
	// OTLPHeaders are additional headers for OTLP exporter (e.g., API keys)
	OTLPHeaders map[string]string
}

// DefaultConfig returns a default configuration
func DefaultConfig(serviceName string) Config {
	return Config{
		ExporterType: getEnvOrDefault("OTEL_METRICS_EXPORTER", "prometheus"),
		ServiceName:  serviceName,
		OTLPEndpoint: getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTLPHeaders:  parseHeaders(getEnvOrDefault("OTEL_EXPORTER_OTLP_HEADERS", "")),
	}
}

// Initialize sets up OpenTelemetry metrics with the given configuration
func Initialize(config Config) error {
	if initialized {
		return nil // Already initialized
	}

	ctx := context.Background()

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion("dev"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create meter provider based on exporter type
	var reader sdkmetric.Reader
	var handler http.Handler

	switch config.ExporterType {
	case "prometheus", "":
		// Use Prometheus exporter (default for local dev)
		// Create a Prometheus registry for the exporter
		reg := prometheus.NewRegistry()
		exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
		if err != nil {
			return fmt.Errorf("failed to create Prometheus exporter: %w", err)
		}
		reader = exporter
		// Use promhttp.HandlerFor with the custom registry
		handler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
		metricsHandler = handler
		slog.Info("Initialized OpenTelemetry metrics with Prometheus exporter",
			"service", config.ServiceName)

	case "otlp":
		// Use OTLP exporter (for Datadog, New Relic, etc.)
		if config.OTLPEndpoint == "" {
			return fmt.Errorf("OTLP endpoint is required when using OTLP exporter")
		}

		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(config.OTLPEndpoint),
			otlpmetrichttp.WithInsecure(), // Use WithTLSClientConfig for production
		}

		// Add headers if provided
		if len(config.OTLPHeaders) > 0 {
			opts = append(opts, otlpmetrichttp.WithHeaders(config.OTLPHeaders))
		}

		exporter, err := otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		reader = sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(15*time.Second))
		metricsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("# Metrics exported via OTLP\n"))
		})
		slog.Info("Initialized OpenTelemetry metrics with OTLP exporter",
			"service", config.ServiceName,
			"endpoint", config.OTLPEndpoint)

	case "none":
		// Disabled - use no-op reader
		reader = sdkmetric.NewManualReader()
		metricsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("# Metrics disabled\n"))
		})
		slog.Info("OpenTelemetry metrics disabled",
			"service", config.ServiceName)

	default:
		return fmt.Errorf("unknown exporter type: %s (supported: prometheus, otlp, none)", config.ExporterType)
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := otel.Meter("portal-backend")

	// Create instruments
	httpRequestsCounter, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_requests_total counter: %w", err)
	}

	httpRequestDuration, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_request_duration_seconds histogram: %w", err)
	}

	initialized = true
	return nil
}

// ensureInitialized ensures OpenTelemetry is initialized with default config
func ensureInitialized() {
	initOnce.Do(func() {
		serviceName := getEnvOrDefault("SERVICE_NAME", "portal-backend")
		config := DefaultConfig(serviceName)
		initErr = Initialize(config)
		if initErr != nil {
			slog.Warn("Failed to initialize OpenTelemetry metrics, metrics will be disabled",
				"error", initErr)
		}
	})
}

// MetricsHandler returns the metrics HTTP handler
func MetricsHandler() http.Handler {
	ensureInitialized()
	if metricsHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("# Metrics not initialized\n"))
		})
	}
	return metricsHandler
}

// MetricsMiddleware records metrics for each request using OpenTelemetry
func MetricsMiddleware(next http.Handler) http.Handler {
	ensureInitialized()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !initialized {
			// If metrics not initialized, just pass through
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		// Use the URL path as the label
		// Note: In a real app with path parameters (e.g. /users/123),
		// you'd want to use the route pattern (e.g. /users/{id}) to avoid high cardinality.
		// For now, we'll use the raw path as requested.
		path := r.URL.Path

		// Record metrics with attributes
		httpRequestsCounter.Add(context.Background(), 1,
			metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", path),
				attribute.Int("http.status_code", rw.statusCode),
			),
		)
		httpRequestDuration.Record(context.Background(), duration,
			metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", path),
			),
		)
	})
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	// Parse format: "key1=value1,key2=value2"
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}
