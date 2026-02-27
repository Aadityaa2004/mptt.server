# MPTT Server - build, test, and lint targets
# See CONTRIBUTING.md for full development setup.

.PHONY: build test test-unit test-integration test-acceptance test-bridge test-all lint lint-go lint-frontend clean help install-hooks pre-push

# Go build: all production services
build:
	go build ./...

# Build individual services (for local run)
build-api:
	go build -o bin/api-service ./src/production/MQT.ApiService
build-ingestor:
	go build -o bin/ingestor ./src/production/MQT.IngestorService
build-email:
	go build -o bin/email-service ./src/production/MQT.EmailService

# Unit tests (excludes integration/acceptance; runs all *_test.go without build tags)
test: test-unit
test-unit:
	go test ./src/test/unit/... ./src/test/testutil/... ./src/production/... -count=1

# Coverage report
coverage:
	go test ./src/test/unit/... ./src/production/... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out
	@echo ""
	@go tool cover -func=coverage.out | tail -1

# Integration tests (requires docker-compose -f docker-compose.test.yml up)
# Export TEST_API_URL=http://localhost:9002 (optional)
test-integration:
	go test -tags=integration ./src/test/integration/... -count=1 -v

# Acceptance tests (E2E; requires full stack)
test-acceptance:
	go test -tags=acceptance ./src/test/acceptance/... -count=1 -v

# Bridge (Python) unit tests
test-bridge:
	(pip3 install -q -r src/production/MQT.Bridge/requirements.txt 2>/dev/null || true) && \
	PYTHONPATH=./src/production/MQT.Bridge/vendor:$${PYTHONPATH} python3 -m unittest src.test.unit.bridge.test_mqtt_bridge -v

# Run all tests (unit + integration if stack up + bridge)
test-all: test-unit test-bridge
	@echo "Run 'make test-integration' with docker-compose.test.yml up for integration tests"
	@echo "Run 'make test-acceptance' for E2E acceptance tests"

# Lint
lint: lint-go lint-frontend
lint-go:
	go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
lint-frontend:
	cd src/production/mqt.frontend && npm run lint

# Frontend tests
test-frontend:
	cd src/production/mqt.frontend && npm run test

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf src/production/MQT.Bridge/vendor/
	go clean -cache -testcache

# Install git hooks (pre-push: build, test, lint, security, secret scan)
install-hooks:
	chmod +x .githooks/pre-push
	git config core.hooksPath .githooks
	@echo "Pre-push hooks installed. Push will run: build, tests, vet, golangci-lint, govulncheck, secret scan."

# Run pre-push checks manually (same as hook)
pre-push: build test-unit lint-go
	@command -v govulncheck >/dev/null 2>&1 && govulncheck ./... || true
	@echo "Pre-push checks done. Install hooks with: make install-hooks"

# Help
help:
	@echo "Targets:"
	@echo "  build            - Build all Go packages"
	@echo "  test / test-unit - Run Go unit tests"
	@echo "  coverage         - Run tests and show coverage report"
	@echo "  test-integration - Run integration tests (needs docker-compose.test.yml)"
	@echo "  test-acceptance  - Run E2E acceptance tests"
	@echo "  test-bridge      - Run Bridge Python unit tests"
	@echo "  test-all         - Run unit + bridge tests"
	@echo "  test-frontend    - Run frontend tests"
	@echo "  lint             - Run Go and frontend linters"
	@echo "  lint-go          - Go vet + golangci-lint"
	@echo "  lint-frontend    - ESLint in frontend"
	@echo "  install-hooks    - Install pre-push hook (build, test, lint, security, secrets)"
	@echo "  pre-push         - Run pre-push checks manually"
	@echo "  clean            - Remove bin/, vendor/, go cache"
