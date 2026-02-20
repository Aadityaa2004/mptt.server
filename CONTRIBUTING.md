# Contributing to MPT.MQTT_Server

Thank you for your interest in contributing. This document explains how to set up the project, run tests, and submit changes.

## Prerequisites

- **Go 1.25+** – [golang.org/dl](https://go.dev/dl/)
- **Node.js LTS** – for the frontend ([nodejs.org](https://nodejs.org/))
- **Docker & Docker Compose** – for running the full stack or test dependencies (Postgres)

## Getting Started

1. Clone the repository.
2. Copy env examples and set required variables (see [README.md](README.md) and [env.production.example](env.production.example)). Do **not** commit `.env` or `.env.production`.
3. Install Go dependencies: `go mod download`
4. Install frontend dependencies: `cd src/production/mqt.frontend && npm ci`

## Building

- **All Go code:** `make build`
- **API service binary:** `make build-api` (output in `bin/api-service`)
- **Ingestor binary:** `make build-ingestor` (output in `bin/ingestor`)

## Running Tests

### Unit tests (Go)

```bash
make test-unit
# or
go test ./src/test/unit/... ./src/production/...
```

No database or external services required.

### Integration tests (Go)

Integration tests use the `integration` build tag and require a running Postgres and optionally the API service.

**Option A – Docker Compose (recommended)**

```bash
docker compose -f docker-compose.test.yml up -d
export TEST_DATABASE_URL="postgres://iot_user:iot_password@localhost:5432/iot?sslmode=disable"
go test -tags=integration ./src/test/integration/...
docker compose -f docker-compose.test.yml down
```

**Option B – Existing stack**

If you already have Postgres and the API running (e.g. via `docker-compose.yml`), set `TEST_DATABASE_URL` (and any `TEST_*` vars the tests expect) and run:

```bash
go test -tags=integration ./src/test/integration/...
```

### Frontend tests

```bash
make test-frontend
# or
cd src/production/mqt.frontend && npm run test
```

### Acceptance / E2E

See `src/test/acceptance/` for high-level flow tests. These typically require the full stack to be up; run manually or in CI as a separate step.

## Linting

- **Go:** `make lint-go` (runs `go vet` and `golangci-lint` if installed).
- **Frontend:** `make lint-frontend` (runs ESLint in the frontend).

Install golangci-lint for Go:  
`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Submitting Changes

1. Create a branch from `main` (or the default branch).
2. Make your changes. Keep the scope focused.
3. Run `make build`, `make test-unit`, and `make lint`. Fix any failures.
4. If you touch integration paths, run integration tests as above.
5. Open a pull request with a clear description. Reference any related issues.

## Code and Repo Guidelines

- **Do not modify** code under `src/production/` unless explicitly part of an approved change; the plan is to add tooling and tests around production without changing it unnecessarily.
- New dependencies (Go or npm) should be compatible with the project license (see [LICENSE](LICENSE)).
- Never commit secrets or `.env.production`. CI uses GitHub (or GitLab) secrets for Docker Hub and deployment.

## Makefile Reference

Run `make help` for a list of targets. Main ones: `build`, `test`, `test-unit`, `test-integration`, `test-frontend`, `lint`, `lint-go`, `lint-frontend`, `clean`.
