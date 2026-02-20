# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Production-ready tooling: Makefile, CONTRIBUTING.md, SECURITY.md, LICENSE, CHANGELOG.md.
- Docs: architecture diagram and runbook (docs/architecture.md, docs/runbook.md).
- Test layout: src/test/unit, src/test/integration, src/test/acceptance.
- CI/CD: GitHub Actions (ci.yml, build-push.yml), .golangci.yml.
- Observability: Prometheus and Grafana configs, runbook.
- Security: govulncheck, npm audit, optional Trivy in CI.

## [1.0.0] - (release date TBD)

- Initial production deployment (RPi, Cloudflare Tunnel).
- Microservices: API, MQTT Ingestor, MQTT Bridge, Mosquitto, Email, Frontend.
- RBAC, JWT auth, health checks, Docker Compose and RPi deployment docs.
