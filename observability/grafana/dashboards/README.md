# Grafana dashboards

Place JSON dashboard exports here (e.g. a dashboard that shows `up` for api-service and mqtt-ingestor jobs, or health check success).

To create a dashboard:
1. Run Grafana and add Prometheus as a data source (e.g. `http://prometheus:9090`).
2. Create panels using Prometheus metrics (e.g. `up{job="api-service"}`).
3. Export the dashboard as JSON and save it in this folder for version control.
