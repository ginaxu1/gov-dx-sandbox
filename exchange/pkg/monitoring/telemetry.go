package monitoring

import (
	"context"
	"net/http"
	"sync"
	"time"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	exporter               *prometheus.Exporter
	meterProvider          *sdkmetric.MeterProvider
	meterName              string
	requestCounter         metric.Int64Counter
	latencyHist            metric.Float64Histogram
	externalCallCounter    metric.Int64Counter
	externalCallLatency    metric.Float64Histogram
	externalCallErrCounter metric.Int64Counter
	businessEventCounter   metric.Int64Counter
	workflowDurationHist   metric.Float64Histogram
	workflowInFlight       metric.Int64UpDownCounter
	dbLatencyHist          metric.Float64Histogram
	cacheEventCounter      metric.Int64Counter
	decisionLatencyHist    metric.Float64Histogram
	decisionFailureCounter metric.Int64Counter
	initOnce               sync.Once
	httpHandler            http.Handler
)

// Config captures the minimal setup parameters shared across services.
type Config struct {
	ServiceName   string
	ResourceAttrs map[string]string
}

// Setup configures OpenTelemetry metrics with a Prometheus exporter and runtime instrumentation.
func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "unknown-service"
	}

	var attrs []attribute.KeyValue
	attrs = append(attrs, semconv.ServiceName(cfg.ServiceName))
	for k, v := range cfg.ResourceAttrs {
		attrs = append(attrs, attribute.String(k, v))
	}

	var initErr error

	initOnce.Do(func() {
		exp, err := prometheus.New(prometheus.WithoutUnits())
		if err != nil {
			initErr = err
			return
		}

		res, err := resource.Merge(
			resource.Default(),
			resource.NewSchemaless(attrs...),
		)
		if err != nil {
			initErr = err
			return
		}

		meterName = cfg.ServiceName
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(exp),
			sdkmetric.WithResource(res),
		)

		otel.SetMeterProvider(meterProvider)
		exporter = exp
		httpHandler = promhttp.Handler()

		meter := meterProvider.Meter(meterName)
		requestCounter, err = meter.Int64Counter(
			"http_requests_total",
			metric.WithDescription("Total number of HTTP requests processed"),
		)
		if err != nil {
			initErr = err
			return
		}

		latencyHist, err = meter.Float64Histogram(
			"http_request_duration_seconds",
			metric.WithDescription("HTTP request duration in seconds"),
		)
		if err != nil {
			initErr = err
			return
		}

		externalCallCounter, err = meter.Int64Counter(
			"external_calls_total",
			metric.WithDescription("Total number of external calls (DB, providers, etc.)"),
		)
		if err != nil {
			initErr = err
			return
		}

		externalCallLatency, err = meter.Float64Histogram(
			"external_call_duration_seconds",
			metric.WithDescription("Duration of external calls in seconds"),
		)
		if err != nil {
			initErr = err
			return
		}

		externalCallErrCounter, err = meter.Int64Counter(
			"external_call_errors_total",
			metric.WithDescription("Number of failed external calls"),
		)
		if err != nil {
			initErr = err
			return
		}

		businessEventCounter, err = meter.Int64Counter(
			"business_events_total",
			metric.WithDescription("Business event counts by action and outcome"),
		)
		if err != nil {
			initErr = err
			return
		}

		workflowDurationHist, err = meter.Float64Histogram(
			"workflow_duration_seconds",
			metric.WithDescription("End-to-end workflow durations"),
		)
		if err != nil {
			initErr = err
			return
		}

		workflowInFlight, err = meter.Int64UpDownCounter(
			"workflow_inflight",
			metric.WithDescription("Number of workflows currently processing"),
		)
		if err != nil {
			initErr = err
			return
		}

		dbLatencyHist, err = meter.Float64Histogram(
			"db_latency_seconds",
			metric.WithDescription("Database latency segmented by datastore and operation"),
		)
		if err != nil {
			initErr = err
			return
		}

		cacheEventCounter, err = meter.Int64Counter(
			"cache_events_total",
			metric.WithDescription("Cache hit/miss counts"),
		)
		if err != nil {
			initErr = err
			return
		}

		decisionLatencyHist, err = meter.Float64Histogram(
			"decision_latency_seconds",
			metric.WithDescription("Policy decision latency"),
		)
		if err != nil {
			initErr = err
			return
		}

		decisionFailureCounter, err = meter.Int64Counter(
			"decision_failures_total",
			metric.WithDescription("Policy decision failures grouped by reason"),
		)
		if err != nil {
			initErr = err
			return
		}

		// Start Go runtime metrics (goroutines, GC, etc.)
		_ = runtime.Start(
			runtime.WithMinimumReadMemStatsInterval(10*time.Second),
			runtime.WithMeterProvider(meterProvider),
		)
	})

	if initErr != nil {
		return nil, initErr
	}

	return func(ctx context.Context) error {
		if meterProvider != nil {
			return meterProvider.Shutdown(ctx)
		}
		return nil
	}, nil
}

// Handler returns the Prometheus /metrics handler.
func Handler() http.Handler {
	if httpHandler != nil {
		return httpHandler
	}
	return http.NotFoundHandler()
}

// HTTPMetricsMiddleware records request counts and latency.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestCounter == nil || latencyHist == nil {
			next.ServeHTTP(w, r)
			return
		}

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(recorder, r)

		attrs := attributeSet(r.Method, r.URL.Path, recorder.status)
		requestCounter.Add(r.Context(), 1, metric.WithAttributes(attrs...))
		latencyHist.Record(r.Context(), time.Since(start).Seconds(), metric.WithAttributes(attrs...))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(statusCode int) {
	s.status = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

func attributeSet(method, route string, status int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.status_code", status),
	}
}

// RecordExternalCall tracks latency and errors for downstream dependencies (DB, providers, etc.).
func RecordExternalCall(ctx context.Context, target, operation string, duration time.Duration, err error) {
	if externalCallCounter == nil || externalCallLatency == nil {
		return
	}

	success := err == nil
	attrs := []attribute.KeyValue{
		attribute.String("external.target", target),
		attribute.String("external.operation", operation),
		attribute.Bool("external.success", success),
	}

	externalCallCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	externalCallLatency.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))

	if err != nil && externalCallErrCounter != nil {
		externalCallErrCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordBusinessEvent records custom business KPIs like processed jobs or schema changes.
func RecordBusinessEvent(ctx context.Context, action string, success bool) {
	if businessEventCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("business.action", action),
		attribute.String("business.outcome", outcomeLabel(success)),
	}

	businessEventCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func outcomeLabel(success bool) string {
	if success {
		return "success"
	}
	return "failure"
}

// RecordWorkflowDuration logs how long a named workflow took.
func RecordWorkflowDuration(ctx context.Context, workflow string, duration time.Duration) {
	if workflowDurationHist == nil {
		return
	}

	workflowDurationHist.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("workflow.name", workflow),
	))
}

// WorkflowInFlightAdd adjusts the in-flight workflow counter (use delta +1 / -1).
func WorkflowInFlightAdd(ctx context.Context, workflow string, delta int64) {
	if workflowInFlight == nil {
		return
	}

	workflowInFlight.Add(ctx, delta, metric.WithAttributes(
		attribute.String("workflow.name", workflow),
	))
}

// RecordDBLatency records datastore read/write duration.
func RecordDBLatency(ctx context.Context, datastore, operation string, duration time.Duration) {
	if dbLatencyHist == nil {
		return
	}

	dbLatencyHist.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("db.name", datastore),
		attribute.String("db.operation", operation),
	))
}

// RecordCacheEvent increments cache hit/miss counters.
func RecordCacheEvent(ctx context.Context, cacheName string, hit bool) {
	if cacheEventCounter == nil {
		return
	}

	result := "miss"
	if hit {
		result = "hit"
	}

	cacheEventCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("cache.name", cacheName),
		attribute.String("cache.result", result),
	))
}

// RecordDecisionLatency tracks how long it took to evaluate a policy decision.
func RecordDecisionLatency(ctx context.Context, decisionType string, duration time.Duration) {
	if decisionLatencyHist == nil {
		return
	}

	decisionLatencyHist.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("decision.type", decisionType),
	))
}

// RecordDecisionFailure increments failure counter with a reason label.
func RecordDecisionFailure(ctx context.Context, reason string) {
	if decisionFailureCounter == nil {
		return
	}

	decisionFailureCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("failure.reason", reason),
	))
}
