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

**Prerequisites:**

Ensure all Go services are running and connected to the `opendif-network`:
- Orchestration Engine (port 4000)
- Consent Engine (port 8081)
- Policy Decision Point (port 8082)
- Portal Backend (port 3000)
- Audit Service (port 3001)

---

## Metrics Overview

### HTTP Request Metrics

| Metric                          | Type      | Labels                                    | Purpose                    |
| ------------------------------- | --------- | ----------------------------------------- | -------------------------- |
| `http_requests_total`           | Counter   | `method`, `route`, `status_code` | Request volume by endpoint |
| `http_request_duration_seconds` | Histogram | `method`, `route`               | API latency percentiles    |

**Label Definitions:**
- `method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `route`: Normalized route path (e.g., `/consents`, `/policies`)
- `status_code`: HTTP response status code (200, 404, 500, etc.)

### External Call Metrics

| Metric                            | Type      | Labels                                    | Purpose                    |
| --------------------------------- | --------- | ----------------------------------------- | -------------------------- |
| `external_calls_total`            | Counter   | `external_target`, `external_operation`, `external_success` | External call volume       |
| `external_call_duration_seconds`   | Histogram | `external_target`, `external_operation`    | External call latency      |
| `external_call_errors_total`       | Counter   | `external_target`, `external_operation`    | Failed external calls      |

**Label Definitions:**
- `external_target`: Target service or system (e.g., `postgres`, `redis`, `external-api`)
- `external_operation`: Operation type (e.g., `query`, `insert`, `get`, `set`)
- `external_success`: Success status (`true` or `false`)

### Database Metrics

| Metric                    | Type      | Labels                    | Purpose               |
| ------------------------- | --------- | ------------------------- | --------------------- |
| `db_latency_seconds`      | Histogram | `db_name`, `db_operation` | Database query timing |

**Label Definitions:**
- `db_name`: Database name or identifier
- `db_operation`: Database operation type (e.g., `select`, `insert`, `update`, `delete`)

### Business Event Metrics

| Metric                    | Type    | Labels                          | Purpose                |
| ------------------------- | ------- | ------------------------------- | ---------------------- |
| `business_events_total`   | Counter | `business_action`, `business_outcome` | Business KPI tracking |

**Label Definitions:**
- `business_action`: Business action type (e.g., `consent_created`, `policy_evaluated`)
- `business_outcome`: Outcome of the action (e.g., `success`, `failure`, `pending`)

### Workflow Metrics

| Metric                          | Type            | Labels          | Purpose               |
| ------------------------------- | --------------- | --------------- | --------------------- |
| `workflow_duration_seconds`     | Histogram       | `workflow_name` | End-to-end workflow timing |
| `workflow_inflight`             | UpDownCounter   | `workflow_name` | Active workflow count |

**Label Definitions:**
- `workflow_name`: Name of the workflow (e.g., `data_exchange`, `consent_flow`)

### Cache Metrics

| Metric                | Type    | Labels                    | Purpose          |
| --------------------- | ------- | ------------------------- | ----------------- |
| `cache_events_total`  | Counter | `cache_name`, `cache_result` | Cache hit/miss tracking |

**Label Definitions:**
- `cache_name`: Cache identifier or name
- `cache_result`: Cache operation result (`hit` or `miss`)

### Policy Decision Metrics (PDP)

| Metric                        | Type      | Labels            | Purpose                |
| ----------------------------- | --------- | ----------------- | ---------------------- |
| `decision_latency_seconds`    | Histogram | `decision_type`   | Policy evaluation time |
| `decision_failures_total`     | Counter   | `failure_reason`  | Policy decision errors |

**Label Definitions:**
- `decision_type`: Type of policy decision (e.g., `allow`, `deny`, `conditional`)
- `failure_reason`: Reason for decision failure (e.g., `policy_not_found`, `evaluation_error`)

### Go Runtime Metrics (Automatic)

The monitoring package automatically instruments Go runtime metrics:
- `process_cpu_seconds_total` - CPU usage
- `go_memstats_*` - Memory statistics (alloc, sys, heap, etc.)
- `go_goroutines` - Goroutine count
- `go_gc_duration_seconds` - Garbage collection pause times

---

## Useful Prometheus Queries

**Request Rate by Endpoint:**
```promql
sum by (route, method) (rate(http_requests_total[5m]))
```

**95th Percentile Latency by Endpoint:**
```promql
histogram_quantile(0.95, sum by (route, le) (rate(http_request_duration_seconds_bucket[5m])))
```

**Error Rate by Endpoint:**
```promql
sum by (route) (rate(http_requests_total{status_code=~"5.."}[5m]))
```

**Top 10 Slowest Endpoints:**
```promql
topk(10, histogram_quantile(0.95, sum by (route, le) (rate(http_request_duration_seconds_bucket[5m]))))
```

**External Call Error Rate:**
```promql
sum by (external_target, external_operation) (rate(external_call_errors_total[5m]))
```

**95th Percentile Database Latency:**
```promql
histogram_quantile(0.95, sum by (db_name, db_operation, le) (rate(db_latency_seconds_bucket[5m])))
```

**Service Availability:**
```promql
up{job=~"orchestration-engine|consent-engine|policy-decision-point|portal-backend|audit-service"}
```

**Current Metric Values (All):**
```promql
{__name__=~"http_.*|external_.*|db_.*|business_.*|workflow_.*|cache_.*|decision_.*"}
```

---

## Grafana Dashboard

Pre-configured dashboard: **Go Services Metrics**

**URL:** http://localhost:3002/d/go-services/go-services-metrics

**Panels:**

- HTTP Traffic (req/s)
- HTTP Latency (P95)
- Service Health (1=up, 0=down)
- External Calls per Second
- External Call Error %
- Business Events

---

## Stop Services

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

## Production Deployment

This setup is for **local development only**. For production:

1. **Use Managed Service**: Grafana Cloud (free tier), Datadog, New Relic
2. **Or Self-Host**: Deploy Prometheus HA, Thanos/Mimir for long-term storage
3. **Security Hardening**: Change Grafana admin password, enable OAuth/SSO, use reverse proxy
4. **Storage & Retention**: Adjust `--storage.tsdb.retention.time` based on storage capacity
5. **Alerting**: Configure Alertmanager for production alerts

---

## Architecture

```
Go Services (Port 3000, 4000, 8081, 8082, 3001)
    ↓ /metrics endpoint (Prometheus format)
Prometheus (localhost:9090, 30d retention)
    ↓ PromQL queries
Grafana (localhost:3002, dashboards)
```

**Stack Components:**

- **prometheus**: `prom/prometheus:v2.55.1`
- **grafana**: `grafana/grafana:11.2.0`

**Network:**

All services run on a shared Docker network (`opendif-network`) to enable service discovery. Services are referenced by their Docker Compose service names in Prometheus configuration (e.g., `orchestration-engine:4000`).

**Volumes:**

- `prometheus-data`: Metric storage (30 day retention)
- `grafana-data`: Dashboard configs & user data

---

## Troubleshooting

### Prometheus Can't Scrape Services

**Issue**: Targets show as DOWN in Prometheus (http://localhost:9090/targets)

**Solutions:**

1. Verify services are running and exposing metrics:
   ```bash
   curl http://localhost:4000/metrics
   curl http://localhost:8081/metrics
   curl http://localhost:8082/metrics
   ```

2. Check Prometheus logs:
   ```bash
   docker compose logs prometheus
   ```

3. Verify network connectivity:
   - Ensure all services are on the same `opendif-network`
   - Check service names match Prometheus configuration

### Grafana Can't Connect to Prometheus

**Issue**: "Data source is not working" in Grafana

**Solutions:**

1. Verify Prometheus is running: `curl http://localhost:9090/-/healthy`
2. Check datasource URL in `grafana/provisioning/datasources/datasource.yml` (should be `http://prometheus:9090`)
3. Ensure both containers are on the same Docker network (`opendif-network`)

### Network Issues

**Issue**: Services can't communicate with each other

**Solutions:**

1. Verify network exists: `docker network ls | grep opendif-network`
2. Check service is on network: `docker network inspect opendif-network`
3. Recreate network if needed:
   ```bash
   docker compose down
   docker network rm opendif-network
   docker compose up -d
   ```

---

## How to Add Metrics to New Go Services

1. **Import the monitoring package:**
   ```go
   import "github.com/gov-dx-sandbox/exchange/shared/monitoring"
   ```

2. **Expose metrics endpoint in main.go:**
   ```go
   mux.Handle("/metrics", monitoring.Handler())
   ```

3. **Wrap HTTP handlers with metrics middleware:**
   ```go
   handler := monitoring.HTTPMetricsMiddleware(mux)
   ```

4. **Add service to Prometheus configuration:**
   Edit `prometheus/prometheus.yml`:
   ```yaml
   - job_name: your-service
     metrics_path: /metrics
     static_configs:
       - targets:
           - your-service:PORT
         labels:
           service: 'your-service'
           port: 'PORT'
   ```

5. **Ensure service is on `opendif-network`:**
   In your service's `docker-compose.yml`:
   ```yaml
   services:
     your-service:
       networks:
         - opendif-network
   
   networks:
     opendif-network:
       name: opendif-network
       external: true
   ```

6. **Restart Prometheus:**
   ```bash
   docker compose restart prometheus
   ```

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
