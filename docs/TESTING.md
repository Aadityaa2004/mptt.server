# Testing guide

This project has multiple layers of tests:

- Go unit tests co-located with their packages under `src/production/...`
- Go tests in a central `src/test` tree (unit, integration, acceptance)
- Python unit tests for the MQTT Bridge
- Frontend tests for the React UI

The sections below describe where each set of tests lives and how to run them.

## Go tests

### Co-located Go unit tests (src/production)

- Most Go unit tests live next to the code they exercise under `src/production/...`.
- This layout allows tests to use the same package as the code and access unexported helpers when needed.
- These are run by default when you run Go tests for the module:

```bash
go test ./src/production/... -count=1
```

or via the Makefile (which also includes `src/test` unit packages):

```bash
make test          # alias for make test-unit
make test-unit
```

### Central Go test tree (src/test)

Higher-level and cross-cutting Go tests live under `src/test`:

- `src/test/unit/...`
  - Go unit tests that sit outside production packages (e.g. JWT service helpers).
  - Grouped by topic:
    - `src/test/unit/jwt/...`
    - `src/test/unit/bridge/...` (for any future Go bridge tests)
    - Additional topic folders can be added as needed.

- `src/test/integration/...`
  - Go tests with the `integration` build tag.
  - Hit running services over HTTP and other real dependencies.
  - Grouped by topic, for example:
    - `src/test/integration/auth/...`
    - `src/test/integration/email/...`
    - `src/test/integration/internal/...`
    - `src/test/integration/health/...`

- `src/test/acceptance/...`
  - Go tests with the `acceptance` build tag.
  - End-to-end / high-level flows, e.g. `src/test/acceptance/e2e/...`.

#### Running Go tests from src/test

From the repo root:

- Unit tests (no build tags):

```bash
make test-unit
# or directly
go test ./src/test/unit/... ./src/test/testutil/... ./src/production/... -count=1
```

- Integration tests (require the stack to be up; see deployment docs):

```bash
make test-integration
# or directly
go test -tags=integration ./src/test/integration/... -count=1 -v
```

- Acceptance (E2E) tests:

```bash
make test-acceptance
# or directly
go test -tags=acceptance ./src/test/acceptance/... -count=1 -v
```

## Python tests (MQTT Bridge)

Python unit tests for the MQTT Bridge live under the central test tree:

- Code under test: `src/production/MQT.Bridge/mqtt_bridge.py`
- Tests: `src/test/unit/bridge/test_mqtt_bridge.py`

The test module adjusts `sys.path` to import `mqtt_bridge` from
`src/production/MQT.Bridge` and its `vendor` directory (for `paho-mqtt`).

### Running bridge tests

From the repo root:

```bash
make test-bridge
```

This will:

- Install bridge dependencies from `src/production/MQT.Bridge/requirements.txt`
- Ensure `src/production/MQT.Bridge/vendor` is on `PYTHONPATH`
- Run:

```bash
python3 -m unittest src.test.unit.bridge.test_mqtt_bridge -v
```

You can also run the unittest command directly if you have dependencies installed.

## Frontend tests

Frontend code and tests live under:

- `src/production/mqt.frontend`

To run frontend tests:

```bash
make test-frontend
# or directly
cd src/production/mqt.frontend && npm test
```

Linting is available via:

```bash
make lint-frontend
```

## CI overview

The GitHub Actions workflow (`.github/workflows/ci.yml`) runs:

- Go build, `go vet`, and unit tests:
  - `go test ./src/test/unit/... ./src/test/testutil/... ./src/production/... -count=1 -v`
- GolangCI-Lint and `govulncheck`
- Frontend lint and tests via the Node job
- Bridge Python unit tests:
  - Install requirements and `paho-mqtt` into `src/production/MQT.Bridge/vendor`
  - Run `python -m unittest src.test.unit.bridge.test_mqtt_bridge -v` with `PYTHONPATH` pointing at `vendor`
- Trivy configuration scan

## Dependency injection and Wire

For dependency injection in Go services:

- Services and handlers are constructed via explicit `NewXxx(...)` constructors and small interfaces.
- [Google Wire](https://github.com/google/wire) is used at the service boundaries to generate the wiring code.

### Setup

- Install Wire locally:

```bash
go install github.com/google/wire/cmd/wire@latest
```

- Wire is tracked as a tool dependency in `tools.go` under the `tools` build tag.

### Running Wire

From the repo root (after the injectors are in place):

- API service:

```bash
cd src/production/MQT.ApiService
wire
```

- Email service:

```bash
cd src/production/MQT.EmailService
wire
```

- Ingestor service:

```bash
cd src/production/MQT.IngestorService
wire
```

Each directory will have a `wire.go` (hand-written) and a generated `wire_gen.go` that is committed to the repo.

## Summary of test locations

- **Production tree (`src/production/...`)**
  - Go production code and its co-located Go unit tests
  - Frontend app and its tests under `src/production/mqt.frontend`
  - No Python tests or other non-frontend tests

- **Test tree (`src/test/...`)**
  - Go unit/integration/acceptance tests grouped by topic
  - Python unit tests for the MQTT Bridge under `src/test/unit/bridge`

