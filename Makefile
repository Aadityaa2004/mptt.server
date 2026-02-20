# MPTT Server - build, test, and lint targets
# See CONTRIBUTING.md for full development setup.

.PHONY: build test test-unit test-integration lint lint-go lint-frontend clean help

# Go build: all production services
build:
	go build ./...

# Build individual services (for local run)
build-api:
	go build -o bin/api-service ./src/production/MQT.ApiService
build-ingestor:
	go build -o bin/ingestor ./src/production/MQT.IngestorService

# Unit tests (no integration tag)
test: test-unit
test-unit:
	go test ./src/test/unit/...

# Integration tests (require Postgres; use docker-compose.test.yml or TEST_DATABASE_URL)
test-integration:
	go test -tags=integration ./src/test/integration/...

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
	go clean -cache -testcache

# Help
help:
	@echo "Targets:"
	@echo "  build          - Build all Go packages"
	@echo "  test / test-unit - Run Go unit tests"
	@echo "  test-integration - Run Go integration tests (needs DB)"
	@echo "  test-frontend  - Run frontend tests"
	@echo "  lint           - Run Go and frontend linters"
	@echo "  lint-go        - Go vet + golangci-lint"
	@echo "  lint-frontend  - ESLint in frontend"
	@echo "  clean          - Remove bin/ and go cache"
