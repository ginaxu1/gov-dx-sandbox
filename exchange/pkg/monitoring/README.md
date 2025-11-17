## Monitoring Package Usage

This package centralizes OpenTelemetry setup and metric helpers that all Go services can share. Below is a quick guide for wiring a service to the shared instrumentation.

### 1. Add the dependency

```bash
# In the service directory
go get github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring@v0.0.0
```

In `go.mod` keep (or add) the replace so the module resolves locally:

```
replace github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring => ../exchange/pkg/monitoring
```

### 2. Initialize OTEL

In `main.go` (or equivalent bootstrap file):

```go
shutdown, err := monitoring.Setup(context.Background(), monitoring.Config{
    ServiceName: "consent-engine", // service-specific
    ResourceAttrs: map[string]string{
        "environment": cfg.Environment,
        "version":     Version,
    },
})
if err != nil { log.Fatal(err) }
defer shutdown(context.Background())
```

### 3. HTTP exposure

```go
mux.Handle("/metrics", monitoring.Handler())
server := &http.Server{
    Addr:    addr,
    Handler: monitoring.HTTPMetricsMiddleware(mux),
}
```

This automatically emits request counters, latency histograms, and status-class labels.

### 4. Common helpers

| Helper | When to use |
| ------ | ----------- |
| `RecordExternalCall(ctx, target, operation, duration, err)` | Wrap outbound HTTP/DB calls (e.g., PDP, providers). |
| `RecordDBLatency(ctx, datastore, operation, duration)` | SQL exec/query wrappers in `consent-engine`. |
| `RecordCacheEvent(ctx, cacheName, hit)` | Cache hit/miss tracking (works with sync.Map caches). |
| `WorkflowInFlightAdd` / `RecordWorkflowDuration` | Track queue depth + total processing time for workflows (e.g., GraphQL or consent flows). |
| `RecordDecisionLatency` / `RecordDecisionFailure` | Policy decision timing & failure labels (used by `policy-decision-point`). |

### 5. Service-specific guidance

- **api-server-go:** Use HTTP middleware + business event counters (`RecordBusinessEvent`) to capture edge success/failure metrics. Register `/metrics` before any reverse proxy.

- **consent-engine:** Call `RecordDBLatency` around `Exec/Query`, use `RecordCacheEvent` for pending consent caches, and keep `RecordExternalCall` for outbound calls if/when they are added.

- **policy-decision-point:** Wrap `GetPolicyDecision` logic with `RecordDecisionLatency` and `RecordDecisionFailure` (already wired) and register `/metrics`. If the service continues to use GORM, wrap DB calls with `RecordExternalCall`.

### 6. Prometheus scraping

Every service now exposes `/metrics`; add the target to Prometheus (or use the shared `exchange/monitoring` Docker Compose stack).

### 7. Reusing updates

Future changes to metric names, OTEL exporters, or middleware only need to be made inside this package. All services import the same code, so upgrades remain uniform.

