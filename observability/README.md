# Observability Stack for OpenDIF MVP

Local development stack: **Go Services** → **Prometheus** → **Grafana**

Collects real-time metrics from all Go services for debugging performance and errors.

---

## Quick Start

```bash
cd observability
docker compose up -d
```

**Services:**

- **Prometheus**: http://localhost:9090 (raw metrics & queries)
- **Grafana**: http://localhost:3002 (dashboards, login: `admin` / `admin`)

**Service Metrics Endpoints:**

- **Orchestration Engine**: http://localhost:4000/metrics
- **Consent Engine**: http://localhost:8081/metrics
- **Policy Decision Point**: http://localhost:8082/metrics
- **API Server Go**: http://localhost:3000/metrics
- **Audit Service**: http://localhost:3001/metrics (when instrumented)

---

## Metrics Overview

All services use the shared `exchange/pkg/monitoring` package which exposes the following metrics:

### HTTP Request Metrics

| Metric                          | Type      | Labels                                    | Purpose                    |
| ------------------------------- | --------- | ----------------------------------------- | -------------------------- |
| `http_requests_total`           | Counter   | `http_method`, `http_route`, `http_status_code` | Request volume by endpoint |
| `http_request_duration_seconds` | Histogram | `http_method`, `http_route`               | API latency percentiles    |

### External Call Metrics (DB, Providers, etc.)

| Metric                            | Type      | Labels                                    | Purpose                    |
| --------------------------------- | --------- | ----------------------------------------- | -------------------------- |
| `external_calls_total`            | Counter   | `external_target`, `external_operation`, `external_success` | External call volume       |
| `external_call_duration_seconds`   | Histogram | `external_target`, `external_operation`    | External call latency      |
| `external_call_errors_total`       | Counter   | `external_target`, `external_operation`    | Failed external calls      |

### Database Metrics

| Metric                    | Type      | Labels                    | Purpose               |
| ------------------------- | --------- | ------------------------- | --------------------- |
| `db_latency_seconds`      | Histogram | `db.name`, `db.operation` | Database query timing |

### Business Event Metrics

| Metric                    | Type    | Labels                          | Purpose                |
| ------------------------- | ------- | ------------------------------- | ---------------------- |
| `business_events_total`   | Counter | `business.action`, `business.outcome` | Business KPI tracking |

### Workflow Metrics

| Metric                          | Type            | Labels          | Purpose               |
| ------------------------------- | --------------- | --------------- | --------------------- |
| `workflow_duration_seconds`     | Histogram       | `workflow.name` | End-to-end workflow timing |
| `workflow_inflight`             | UpDownCounter   | `workflow.name` | Active workflow count |

### Cache Metrics

| Metric                | Type    | Labels                    | Purpose          |
| --------------------- | ------- | ------------------------- | ----------------- |
| `cache_events_total`  | Counter | `cache.name`, `cache.result` | Cache hit/miss tracking |

### Policy Decision Metrics (PDP)

| Metric                        | Type      | Labels            | Purpose                |
| ----------------------------- | --------- | ----------------- | ---------------------- |
| `decision_latency_seconds`    | Histogram | `decision.type`   | Policy evaluation time |
| `decision_failures_total`     | Counter   | `failure.reason`  | Policy decision errors |

### Go Runtime Metrics (Automatic)

The monitoring package automatically instruments Go runtime metrics:
- `process_cpu_seconds_total` - CPU usage
- `go_memstats_*` - Memory statistics
- `go_goroutines` - Goroutine count
- `go_gc_duration_seconds` - GC pause times

---

## Useful Prometheus Queries

### HTTP Request Analysis

**Request Rate by Endpoint:**
```promql
sum by (http_route, http_method) (rate(http_requests_total[5m]))
```

**95th Percentile Latency by Endpoint:**
```promql
histogram_quantile(0.95, sum by (http_route, le) (rate(http_request_duration_seconds_bucket[5m])))
```

**Error Rate by Endpoint:**
```promql
sum by (http_route) (rate(http_requests_total{http_status_code=~"5.."}[5m]))
```

**Top 10 Slowest Endpoints:**
```promql
topk(10, histogram_quantile(0.95, sum by (http_route, le) (rate(http_request_duration_seconds_bucket[5m]))))
```

### External Call Analysis

**External Call Error Rate:**
```promql
sum by (external_target, external_operation) (rate(external_call_errors_total[5m]))
```

**95th Percentile External Call Latency:**
```promql
histogram_quantile(0.95, sum by (external_target, le) (rate(external_call_duration_seconds_bucket[5m])))
```

**Failed External Calls:**
```promql
sum by (external_target) (external_call_errors_total)
```

### Database Performance

**95th Percentile Database Latency:**
```promql
histogram_quantile(0.95, sum by (db_name, db_operation, le) (rate(db_latency_seconds_bucket[5m])))
```

**Slowest Database Operations:**
```promql
topk(5, histogram_quantile(0.95, sum by (db_name, db_operation, le) (rate(db_latency_seconds_bucket[5m]))))
```

### Workflow Metrics

**Active Workflows:**
```promql
sum by (workflow_name) (workflow_inflight)
```

**Workflow Duration (95th percentile):**
```promql
histogram_quantile(0.95, sum by (workflow_name, le) (rate(workflow_duration_seconds_bucket[5m])))
```

### Policy Decision Point

**Decision Latency (95th percentile):**
```promql
histogram_quantile(0.95, sum by (decision_type, le) (rate(decision_latency_seconds_bucket[5m])))
```

**Decision Failures by Reason:**
```promql
sum by (failure_reason) (decision_failures_total)
```

### Cache Performance

**Cache Hit Rate:**
```promql
sum(rate(cache_events_total{cache_result="hit"}[5m])) / sum(rate(cache_events_total[5m]))
```

**Cache Hit/Miss Count:**
```promql
sum by (cache_name, cache_result) (cache_events_total)
```

### Business Events

**Business Event Success Rate:**
```promql
sum(rate(business_events_total{business_outcome="success"}[5m])) / sum(rate(business_events_total[5m]))
```

**Business Events by Action:**
```promql
sum by (business_action, business_outcome) (business_events_total)
```

### Service Health

**All Metrics for a Service:**
```promql
{service="orchestration-engine"}
```

**Request Rate by Service:**
```promql
sum by (service) (rate(http_requests_total[5m]))
```

**Error Rate by Service:**
```promql
sum by (service) (rate(http_requests_total{http_status_code=~"5.."}[5m]))
```

---

## 📈 Grafana Dashboard

Pre-configured dashboard: **Go Services Metrics**

**URL:** http://localhost:3002/d/go-services/go-services-metrics

**Panels:**

- HTTP Request Total Count (by endpoint)
- HTTP Latency (95th percentile)
- HTTP Error Rate
- External Call Latency
- External Call Errors
- Database Latency
- Workflow Duration
- Policy Decision Latency
- Cache Hit Rate
- Business Events
- Go Runtime Metrics (CPU, Memory, Goroutines)

---

## 🛑 Stop Services

```bash
docker compose down
```

**Keep data (volumes persist):**
```bash
docker compose stop
```

**Remove everything (data + volumes):**
```bash
docker compose down -v
```

---

## Architecture

```
Go Services (Port 3000, 4000, 8081, 8082)
    ↓ /metrics endpoint (Prometheus format)
Prometheus (localhost:9090, 30d retention)
    ↓ PromQL queries
Grafana (localhost:3002, dashboards)
```

**Stack Components:**

- **prometheus**: `prom/prometheus:v2.55.1`
- **grafana**: `grafana/grafana:11.2.0`

**Volumes:**

- `prometheus-data`: Metric storage (30 day retention)
- `grafana-data`: Dashboard configs & user data

**Data Persistence:**

- Prometheus data persists in `prometheus-data` volume
- Grafana dashboards and configs persist in `grafana-data` volume
- Data survives container restarts

---

## Configuration

### Prometheus Configuration

Edit `prometheus/prometheus.yml` to:
- Add new service targets
- Adjust scrape intervals
- Configure alerting rules

### Grafana Configuration

- **Datasource**: Auto-provisioned from `grafana/provisioning/datasources/datasource.yml`
- **Dashboards**: Auto-loaded from `grafana/dashboards/`
- **Login**: `admin` / `admin` (change in production!)

### Adding a New Service

1. Ensure the service exposes `/metrics` endpoint using `monitoring.Handler()`
   - For services in `exchange/`: Use `exchange/pkg/monitoring`
   - For root-level services: Import `github.com/gov-dx-sandbox/exchange/pkg/monitoring` with appropriate replace directive
2. Add the service to `prometheus/prometheus.yml`:
   ```yaml
   - job_name: your-service
     metrics_path: /metrics
     static_configs:
       - targets:
           - host.docker.internal:PORT
         labels:
           service: 'your-service'
           port: 'PORT'
   ```
3. Restart Prometheus: `docker compose restart prometheus`

---

## Production Deployment

This setup is for **local development only**. For production:

1. **Use Managed Service**: Grafana Cloud (free tier), Datadog, New Relic
2. **Or Self-Host**: Deploy Prometheus HA, Thanos/Mimir for long-term storage
3. **Enable Authentication**: Change Grafana admin password, enable OAuth
4. **Network Security**: Restrict Prometheus/Grafana access, use reverse proxy
5. **Retention**: Adjust `--storage.tsdb.retention.time` based on storage capacity
6. **Alerting**: Configure alertmanager for production alerts

---

## Service Instrumentation

All services use the shared `exchange/pkg/monitoring` package. See `exchange/pkg/monitoring/README.md` for integration details.

**Note**: Currently, `api-server-go` and `audit-service` are in the Prometheus config but may not be fully instrumented yet. They will appear as DOWN in Prometheus until metrics endpoints are added.

**Quick Integration:**

```go
// In main.go
shutdown, err := monitoring.Setup(context.Background(), monitoring.Config{
    ServiceName: "your-service",
    ResourceAttrs: map[string]string{
        "environment": "local",
        "version":     "1.0.0",
    },
})
if err != nil { log.Fatal(err) }
defer shutdown(context.Background())

// In HTTP server setup
mux.Handle("/metrics", monitoring.Handler())
server := &http.Server{
    Addr:    addr,
    Handler: monitoring.HTTPMetricsMiddleware(mux),
}
```

---

## Troubleshooting

### Prometheus can't scrape services

**Issue**: Targets show as DOWN in Prometheus

**Solutions**:
1. Verify services are running: `curl http://localhost:PORT/metrics`
2. Check Prometheus config: `docker compose logs prometheus`
3. For Docker services, ensure `host.docker.internal` resolves correctly
4. Check firewall/network settings

### Grafana can't connect to Prometheus

**Issue**: "Data source is not working" in Grafana

**Solutions**:
1. Verify Prometheus is running: `curl http://localhost:9090/-/healthy`
2. Check datasource URL in `grafana/provisioning/datasources/datasource.yml`
3. Ensure both containers are on the same Docker network
4. Check Grafana logs: `docker compose logs grafana`

### No metrics appearing

**Issue**: Services expose metrics but nothing shows in Prometheus

**Solutions**:
1. Verify service is in Prometheus config
2. Check service labels match Prometheus job name
3. Wait 15-30 seconds for scrape interval
4. Query directly: `http_requests_total` in Prometheus UI

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- Service-specific observability docs:
  - `exchange/pkg/monitoring/README.md`

