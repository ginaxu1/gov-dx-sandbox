## Observability Guide

This service already exposes OpenTelemetry metrics (HTTP latency, runtime stats, external dependency calls, workflow timing, etc). Follow the steps below to see everything in Grafana.

---

### 1. Prerequisites
* Docker Desktop (or another Docker runtime) is running.
* The orchestration-engine service is running locally and reachable on `http://localhost:4000`.



### 2. Start the Monitoring Stack (Prometheus + Grafana)
We ship a ready-made setup under `exchange/monitoring`. Run it from the repo root:

```bash
cd exchange/monitoring
docker compose up -d
```

What this does:
* Prometheus listens on **http://localhost:9090** and scrapes `http://host.docker.internal:4000/metrics`.
* Grafana listens on **http://localhost:3000** with user/pass `admin / admin`.
* A dashboard named **“Orchestration Engine Metrics”** is auto-loaded with the most relevant charts.

To stop the stack later: `docker compose down`.


### 3. Verify Prometheus Is Scraping
1. Open http://localhost:9090/targets  
2. You should see a job called `orchestration-engine` with status **UP**.  
3. Optional: run `curl http://localhost:4000/metrics` to confirm the service exposes metrics.

If the job shows **DOWN**, double-check:
* The service is running.
* The monitoring stack can reach it (Docker Desktop on macOS uses `host.docker.internal` automatically; change the hostname in `exchange/monitoring/prometheus/prometheus.yml` if needed).

### 4. Explore Grafana
1. Visit http://localhost:3000 and log in (`admin` / `admin`).
2. Grafana automatically finds the Prometheus datasource and loads the dashboard (`Browse → Orchestration Engine → Orchestration Engine Metrics`).
3. Panels you’ll see:
   - **HTTP Traffic**: request rate by method/route.
   - **HTTP Latency (P95)**: tail latency derived from the histogram.
   - **External Calls & Error %**: latency and error rate per downstream system (DB, PDP, etc.).
   - **Business Events**: success/failure counts for key workflows.
   - **Workflow Duration / In-Flight**: monitors workflow efficiency and queue depth.

Feel free to duplicate the dashboard and experiment with your own panels (gear icon → “Save As”).


### 5. Distributed Tracing (Optional)
If you also want traces:
1. Run Jaeger:
   ```bash
   docker run -d --name jaeger \
     -p 16686:16686 -p 4317:4317 -p 4318:4318 \
     jaegertracing/all-in-one:1.56
   ```
2. Export traces from the service by setting:
   ```
   OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
   OTEL_TRACES_EXPORTER=otlp
   ```
3. View traces at http://localhost:16686.


### 6. Basic Alerting (Optional)
1. Prometheus Alerting:
   * Create `alert.rules.yml` with the sample rules below and reference it from `prometheus.yml`.
   * Use Alertmanager to forward notifications (Slack/email/PagerDuty).
2. Grafana Alerting:
   * Open a panel → “Alert” tab → configure thresholds (e.g., P99 latency > 500 ms).

Sample PromQL rules:
```yaml
groups:
  - name: orchestration-engine
    rules:
      - alert: HighP99Latency
        expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 0.5
        for: 5m
      - alert: HighExternalErrorRate
        expr: rate(external_call_errors_total[5m]) > 5
        for: 5m
```

### 7. Operational Checklist
- [ ] `/metrics` reachable & Prometheus target shows **UP**.
- [ ] Grafana dashboard displays HTTP / external / workflow panels.
- [ ] (Optional) Jaeger shows traces for GraphQL requests.
- [ ] Alerts configured for latency/error spikes.
- [ ] Runbooks or on-call docs reference these dashboards.

Once the stack is running, you can confidently observe the service’s health without needing deep Prometheus or Grafana expertise.

### Distributed Tracing (Jaeger OTLP)
1. Run Jaeger all-in-one:
   ```bash
   docker run -d --name jaeger -p 16686:16686 -p 4317:4317 -p 4318:4318 \
     jaegertracing/all-in-one:1.56
   ```
2. Update service env vars to export traces:
   ```
   OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
   OTEL_TRACES_EXPORTER=otlp
   ```
3. View traces at http://localhost:16686.

### Grafana Dashboards
1. Deploy Grafana:
   ```bash
   docker run -d --name grafana -p 3000:3000 grafana/grafana
   ```
2. Add Prometheus datasource (`http://host.docker.internal:9090`).
3. Import dashboards for:
   - Four Golden Signals (latency, traffic, errors, saturation)
   - Custom metrics (`external_calls_total`, `business_events_total`).
4. Optionally add Jaeger/Tempo datasource for traces.

### Alerting
1. Configure Alertmanager with Prometheus or use Grafana Alerting.
2. Example PromQL rules:
   ```yaml
   groups:
     - name: orchestration-engine
       rules:
         - alert: HighP99Latency
           expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le)) > 0.5
           for: 5m
           labels:
             severity: warning
           annotations:
             summary: "P99 latency above 500ms"
         - alert: HighExternalErrorRate
           expr: rate(external_call_errors_total[5m]) > 5
           for: 5m
           labels:
             severity: critical
           annotations:
             summary: "External dependency error rate is elevated"
   ```
3. Route alerts to Slack, email, or PagerDuty as needed.

### Operational Checklist
- [ ] `/metrics` scraped by Prometheus.
- [ ] Runtime & dependency metrics visible in Grafana.
- [ ] Jaeger shows traces for GraphQL requests.
- [ ] Alerts configured for latency, dependency errors, and CPU saturation.
- [ ] Runbooks link alerts to mitigation steps.

