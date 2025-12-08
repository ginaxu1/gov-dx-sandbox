package monitoring

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

var (
	// Metrics instruments
	httpRequestsCounter   metric.Int64Counter
	httpRequestDuration   metric.Float64Histogram
	externalCallsCounter  metric.Int64Counter
	externalCallErrors    metric.Int64Counter
	externalCallDuration  metric.Float64Histogram
	businessEventsCounter metric.Int64Counter
	metricsHandler        http.Handler
	initialized           bool
)

// Config holds the configuration for OpenTelemetry metrics
type Config struct {
	// ExporterType can be "prometheus", "otlp", or "none" (disabled)
	ExporterType string
	// ServiceName is the name of the service (e.g., "portal-backend", "orchestration-engine")
	ServiceName string
	// OTLPEndpoint is the OTLP endpoint URL (for Datadog, New Relic, etc.)
	// Example: "https://api.datadoghq.com/api/v2/otlp"
	OTLPEndpoint string
	// OTLPHeaders are additional headers for OTLP exporter (e.g., API keys)
	OTLPHeaders map[string]string
	// PrometheusPort is the port for Prometheus exporter (default: 8888)
	PrometheusPort int
}

// DefaultConfig returns a default configuration
func DefaultConfig(serviceName string) Config {
	return Config{
		ExporterType:   getEnvOrDefault("OTEL_METRICS_EXPORTER", "prometheus"),
		ServiceName:    serviceName,
		OTLPEndpoint:   getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		PrometheusPort: 8888,
		OTLPHeaders:    parseHeaders(getEnvOrDefault("OTEL_EXPORTER_OTLP_HEADERS", "")),
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
		exporter, err := prometheus.New()
		if err != nil {
			return fmt.Errorf("failed to create Prometheus exporter: %w", err)
		}
		reader = exporter
		handler = exporter
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
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{Name: "http_request_duration_seconds"},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
				},
			},
		)),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := otel.Meter("opendif")

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

	externalCallsCounter, err = meter.Int64Counter(
		"external_calls_total",
		metric.WithDescription("Total number of external service calls"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create external_calls_total counter: %w", err)
	}

	externalCallErrors, err = meter.Int64Counter(
		"external_call_errors_total",
		metric.WithDescription("Total number of failed external service calls"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create external_call_errors_total counter: %w", err)
	}

	externalCallDuration, err = meter.Float64Histogram(
		"external_call_duration_seconds",
		metric.WithDescription("External service call duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create external_call_duration_seconds histogram: %w", err)
	}

	businessEventsCounter, err = meter.Int64Counter(
		"business_events_total",
		metric.WithDescription("Total number of business events"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create business_events_total counter: %w", err)
	}

	initialized = true
	return nil
}

// otelHandler returns the metrics HTTP handler
// For Prometheus exporter, this returns the Prometheus metrics endpoint
// For OTLP exporter, this returns a simple status endpoint
func otelHandler() http.Handler {
	if metricsHandler == nil {
		// Fallback if not initialized
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("# Metrics not initialized\n"))
		})
	}
	return metricsHandler
}

// otelHTTPMetricsMiddleware wraps an HTTP handler to record metrics using OpenTelemetry
func otelHTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !initialized {
			// If metrics not initialized, just pass through
			next.ServeHTTP(w, r)
			return
		}

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

		// Record metrics with attributes
		httpRequestsCounter.Add(context.Background(), 1,
			metric.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.route", route),
				attribute.Int("http.status_code", rw.statusCode),
			),
		)
		httpRequestDuration.Record(context.Background(), duration,
			metric.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.route", route),
			),
		)
	})
}

// otelRecordExternalCall records an external service call using OpenTelemetry
func otelRecordExternalCall(target, operation string, duration time.Duration, err error) {
	if !initialized {
		return
	}

	ctx := context.Background()
	attrs := []metric.AddOption{
		metric.WithAttributes(
			attribute.String("external.target", target),
			attribute.String("external.operation", operation),
		),
	}

	externalCallsCounter.Add(ctx, 1, attrs...)
	externalCallDuration.Record(ctx, duration.Seconds(), attrs...)
	if err != nil {
		externalCallErrors.Add(ctx, 1, attrs...)
	}
}

// otelRecordBusinessEvent records a business event using OpenTelemetry
func otelRecordBusinessEvent(action, outcome string) {
	if !initialized {
		return
	}

	businessEventsCounter.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("business.action", action),
			attribute.String("business.outcome", outcome),
		),
	)
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
