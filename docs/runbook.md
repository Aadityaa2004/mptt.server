# Runbook

Operational guide for alerts and common fixes.

## Alerts (Prometheus)

| Alert | Meaning | Action |
|-------|---------|--------|
| **ApiServiceDown** | API service (`api-service:9002`) is unreachable for >1 minute. | 1. Check API container: `docker ps` / `docker compose ps`. 2. Check logs: `docker compose logs api-service`. 3. Restart if needed: `docker compose restart api-service`. 4. If DB-related, check Postgres and connectivity. |
| **MqttIngestorDown** | MQTT Ingestor (`mqtt-ingestor:9003`) is unreachable for >1 minute. | 1. Check Ingestor container and logs: `docker compose logs mqtt-ingestor`. 2. Ensure Mosquitto and API are up (Ingestor depends on them). 3. Restart: `docker compose restart mqtt-ingestor`. |

## Health endpoints

- **API:** `GET /health/live` (liveness), `GET /health/ready` (readiness). Returns 200 when OK.
- **MQTT Ingestor:** `GET /health`. Returns 200 when OK.

Use these for load balancers, Kubernetes probes, or manual checks.

## Common issues

- **API won’t start:** Check Postgres is healthy and env vars (e.g. `POSTGRES_*`, `JWT_SECRET_KEY`) are set. See [env.production.example](../env.production.example) and [RPI_DEPLOYMENT_STEPS.md](../RPI_DEPLOYMENT_STEPS.md).
- **Ingestor not receiving data:** Verify MQTT Bridge and Mosquitto are running; check topic filter (`MQTT_TOPIC`) and broker credentials.
- **High memory/CPU on RPi:** Limit container resources in `docker-compose` and monitor with `docker stats`.

## Testing layout

- **Go tests in production:** Go unit tests remain co-located with their packages under `src/production/...` so they can exercise unexported internals.
- **Central test tree:** Cross-cutting Go tests and higher-level flows live under `src/test`:
  - `src/test/unit/...` for Go unit tests that sit outside production packages (e.g. JWT service), grouped by topic (`jwt`, `bridge`, etc.).
  - `src/test/integration/...` for Go integration tests with the `integration` build tag, grouped by topic (`auth`, `email`, `internal`, `health`, etc.).
  - `src/test/acceptance/...` for Go acceptance/E2E tests with the `acceptance` build tag, grouped by flow (`e2e`, etc.).
- **Bridge (Python) tests:** Python unit tests for the MQTT Bridge live under `src/test/unit/bridge` and are run via `make test-bridge` or the CI bridge job.

## Escalation

For security or data issues, follow [SECURITY.md](../SECURITY.md). For deployment rollback, see [RPI_DEPLOYMENT_STEPS.md](../RPI_DEPLOYMENT_STEPS.md).
