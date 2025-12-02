# Observability Stack

Local development observability stack for monitoring Go services using Prometheus and Grafana.

## How to Start the Stack

1. Ensure all Go services are running and connected to the `opendif-network`:
   - Orchestration Engine (port 4000)
   - Consent Engine (port 8081)
   - Policy Decision Point (port 8082)
   - Portal Backend (port 3000)
   - Audit Service (port 3001)

2. Start the observability stack:
```bash
cd observability
docker compose up -d
```

3. Wait 10-15 seconds for services to initialize.

## How to Access Grafana

- **URL**: http://localhost:3002
- **Login**: `admin` / `admin` (default password)
- **Dashboard**: http://localhost:3002/d/go-services/go-services-metrics

## Metrics Being Collected

The observability stack collects the following high-level metrics from all Go services:

- **HTTP Request Metrics**: Request volume, latency, and error rates by endpoint
- **External Call Metrics**: Database queries, API calls, and external service interactions
- **Business Event Metrics**: Key business actions and outcomes
- **Service Health Metrics**: Service availability and uptime
- **Go Runtime Metrics**: CPU, memory, goroutines, and garbage collection

For detailed metric definitions, label descriptions, and PromQL queries, see the [Metrics Reference](docs/metrics-reference.md).

## Detailed Documentation

1. **[Metrics Reference](docs/metrics-reference.md)**: Complete metric tables, label definitions, and PromQL query examples
2. **[Architecture & Internals](docs/architecture.md)**: System architecture, detailed PromQL queries, troubleshooting guide, and production deployment notes
