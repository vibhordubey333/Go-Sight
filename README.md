# Go-Sight

Go-Sight is a small Go API instrumented with Prometheus metrics and visualized in Grafana. It includes Docker Compose setup plus optional Kubernetes notes.

## What it does

- Exposes basic API routes for health checks and sample responses
- Exposes `/metrics` with OpenTelemetry runtime and host metrics
- Scrapes metrics with Prometheus and visualizes them in Grafana

## What is included

- Go API with `/health`, `/metrics`, and example routes
- Prometheus scraping the API metrics
- OpenTelemetry runtime + host metrics exposed at `/metrics`
- Grafana with provisioned Prometheus data source
- Dashboard auto-provisioning from JSON files
- Local persistence for Grafana and Prometheus

## Requirements

- Go 1.22+ (for local run)
- Docker + Docker Compose (for full stack)

## Quick start (Docker Compose)

```bash
make compose-up
```

Open:
- API: `http://localhost:8000/health`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (user: `admin`, pass: `password`)

## Run locally (API only)

```bash
make run
```

## API endpoints

- `GET /health` -> service health
- `GET /metrics` -> Prometheus metrics
- `GET /v1/users` -> sample response
- `GET /v1/compute?n=10000` -> sample compute

### Curl examples

```bash
curl http://localhost:8000/health
curl http://localhost:8000/v1/users
curl "http://localhost:8000/v1/compute?n=10000"
curl http://localhost:8000/metrics | head -n 20
```

## Grafana dashboards

Dashboards are loaded from files at startup:
- Provisioning config: `docker/grafana-dashboards.yml`
- Dashboard JSON: `docker/dashboards/*.json`

To add a dashboard:
1. Export JSON in Grafana UI
2. Save it to `docker/dashboards/`
3. Restart Grafana: `make compose-down` then `make compose-up`

Notes:
- Datasource is provisioned as `prometheus` (uid `prometheus`).
- For provisioned dashboards, remove any `__inputs` blocks.
- Use `{job=~"$job"}` in queries if you add a `job` variable.

## Troubleshooting "No data"

1. Check Prometheus targets: `http://localhost:9090/targets`
2. Verify a metric exists in Prometheus, e.g. `system_cpu_time_seconds_total`
3. In Grafana, ensure datasource is `prometheus` and `job` is `go-api`

## Data persistence

- Grafana data is stored in Postgres volume `postgres_data`
- Prometheus data is stored in volume `prometheus_data`

Use:
- `make compose-down` to stop containers (keeps data)
- `make compose-clean` to stop containers and remove volumes

## Kubernetes notes

See `Prometheus-Grafana-Readme.md` for the Helm/minikube setup steps.

## Make targets

- `make run` - run the API locally
- `make compose-up` - build and start the stack
- `make compose-down` - stop the stack
- `make compose-clean` - stop and remove volumes
