# Integration Test Makefile

.PHONY: test-integration test-integration-verbose test-integration-coverage test-integration-race test-integration-all

# Basic integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./test/integration/... -tags=integration

# Verbose integration tests with detailed output
test-integration-verbose:
	@echo "Running verbose integration tests..."
	go test -v ./test/integration/... -tags=integration -test.v

# Integration tests with coverage
test-integration-coverage:
	@echo "Running integration tests with coverage..."
	go test -v ./test/integration/... -tags=integration -coverprofile=integration-coverage.out
	go tool cover -html=integration-coverage.out -o integration-coverage.html
	@echo "Coverage report generated: integration-coverage.html"

# Integration tests with race detection
test-integration-race:
	@echo "Running integration tests with race detection..."
	go test -v ./test/integration/... -tags=integration -race

# Run all integration test variations
test-integration-all: test-integration test-integration-race test-integration-coverage

# Setup test database
test-db-setup:
	@echo "Setting up test database..."
	docker-compose -f docker-compose.test.yml up -d postgres-test
	sleep 5
	go run migrator.go migrate test

# Teardown test database  
test-db-teardown:
	@echo "Tearing down test database..."
	docker-compose -f docker-compose.test.yml down -v

# Full integration test cycle
test-integration-full: test-db-setup test-integration-all test-db-teardown

# Continuous integration testing
test-ci:
	@echo "Running CI integration tests..."
	go test -v ./test/integration/... -tags=integration -short -race -coverprofile=ci-coverage.out

