# Testing Plan for CI Success

This document outlines how to ensure all tests pass and GitHub CI shows green (✓) on push.

## Pre-Push Checklist

Before pushing, run these commands locally from the **project root**:

### 1. Go (Backend)
```bash
go build ./...
go vet ./...
go test ./src/test/unit/... ./src/test/testutil/... ./src/production/... -count=1 -v
golangci-lint run
```

### 2. Frontend (Next.js)
```bash
cd src/production/mqt.frontend
npm ci
npm run lint
npm run test
```

### 3. Python Bridge
```bash
cd src/production/MQT.Bridge
pip install -r requirements.txt
pip install --target ./vendor paho-mqtt
cd $OLDPWD
PYTHONPATH=./src/production/MQT.Bridge/vendor:$PYTHONPATH python -m unittest src.test.unit.bridge.test_mqtt_bridge -v
```

## CI Pipeline Jobs

The CI workflow (`.github/workflows/ci.yml`) runs these jobs on every push to `main` and `dev-prod_ready`:

| Job | What it runs |
|-----|--------------|
| **go** | Build, vet, unit tests, golangci-lint, govulncheck |
| **frontend** | `npm ci`, `npm run lint`, `npm run test`, npm audit |
| **bridge** | Python bridge unit tests |
| **trivy** | Security scan (continue-on-error: true) |

## Common CI Failure Causes

### Lint Errors
- **ESLint**: Run `cd src/production/mqt.frontend && npm run lint` to catch frontend issues before push
- **golangci-lint**: Run `golangci-lint run` (uses `.golangci.yml` auto-discovered in repo root)

### Test Failures
- **Go tests**: Ensure all packages in `./src/test/unit/...`, `./src/test/testutil/...`, `./src/production/...` pass
- **Frontend tests**: Run `npm run test` in `src/production/mqt.frontend`. Fix mock setup for:
  - `window.matchMedia` (layout tests) – add to test setup
  - Missing mocks for `sensorService.getDevice` (DeviceCard)
  - `act()` warnings – wrap async updates in `act()`
- **Bridge tests**: Verify MQTT bridge unit tests pass with correct PYTHONPATH

### golangci-lint-action
- **v6** does not support a `config` input – golangci-lint auto-discovers `.golangci.yml` in the repo root. Do not pass `config:` in the workflow.

## Quick Commands

```bash
# One-liner to run all local checks (from repo root)
(go build ./... && go vet ./... && go test ./src/test/unit/... ./src/test/testutil/... ./src/production/... -count=1) && \
(cd src/production/mqt.frontend && npm run lint && npm run test -- --run) && \
echo "All checks passed"
```

**Note:** If frontend tests hit "JavaScript heap out of memory" or "Worker exited unexpectedly", the `npm run test` script sets `NODE_OPTIONS='--max-old-space-size=8192'`. On resource-constrained machines, use `npm run test:ci` which runs with `--no-file-parallelism`.

## Fixes Applied (Recent)

- **golangci-lint-action**: Removed invalid `config` input
- **ThemeToggle**: Replaced `useEffect` + `setState` with `useSyncExternalStore` (avoids react-hooks/set-state-in-effect)
- **Navbar**: Moved `NavLinks` outside component to satisfy react-hooks/static-components
- **ReadingsChart**: Moved `ReadingsChartTooltip` outside component; fixed parsing error (extra brace)
- **WeatherForecast**: Moved `TempTooltip` and `PrecipitationTooltip` outside; added proper TooltipProps types
- **CurrentWeather.test**: Replaced `as any` with typed mock; fixed text matcher for "Toronto"
- **usePiPreferences**: Replaced `any` with typed object; removed unused `hexToTailwind` import
