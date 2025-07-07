# Integration Test Makefile
# Run from: backend/test/integration/

.PHONY: test-integration test-integration-verbose test-integration-coverage test-integration-race test-integration-all

# Basic integration tests (no migrator needed!)
test-integration: test-db-setup
	@echo "Running integration tests..."
	go test -v ./... -tags=integration

# Verbose integration tests with detailed output
test-integration-verbose: test-db-setup
	@echo "Running verbose integration tests..."
	go test -v ./... -tags=integration -test.v

# Integration tests with coverage
test-integration-coverage: test-db-setup
	@echo "Running integration tests with coverage..."
	go test -v ./... -tags=integration -coverprofile=integration-coverage.out
	go tool cover -html=integration-coverage.out -o integration-coverage.html
	@echo "Coverage report generated: integration-coverage.html"

# Integration tests with race detection
test-integration-race: test-db-setup
	@echo "Running integration tests with race detection..."
	go test -v ./... -tags=integration -race

# Run all integration test variations
test-integration-all: test-integration test-integration-race test-integration-coverage

# Setup integration test environment (no migrator - testutil handles schema!)
setup-integration:
	@echo "🔧 Setting up integration test environment..."
	@if [ ! -f .env.test ]; then \
		echo "💡 Creating .env.test with database connection..."; \
		echo "DATABASE_URL=postgres://postgres:test123@localhost:5433/oilgas_inventory_test" > .env.test; \
		echo "TEST_DATABASE_URL=postgres://postgres:test123@localhost:5433/oilgas_inventory_test" >> .env.test; \
		echo "APP_ENV=test" >> .env.test; \
	fi

# Setup test database (just start container - no migrations!)
test-db-setup: setup-integration
	@echo "🐘 Setting up test database..."
	docker-compose -f docker-compose.test.yml up -d postgres-test
	@echo "⏳ Waiting for database to be ready..."
	@until docker exec oilgas-postgres-test pg_isready -U postgres >/dev/null 2>&1; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "✅ Test database ready (schema will be created by testutil)"

# Teardown test database  
test-db-teardown:
	@echo "Tearing down test database..."
	docker-compose -f docker-compose.test.yml down -v

# Full integration test cycle
test-integration-full: test-db-setup test-integration-all test-db-teardown

# Continuous integration testing
test-ci: test-db-setup
	@echo "Running CI integration tests..."
	go test -v ./... -tags=integration -short -race -coverprofile=ci-coverage.out
	$(MAKE) test-db-teardown

# Debug helpers
debug-env:
	@echo "🔍 Debug environment from $(PWD)"
	@echo "Current directory: $(shell pwd)"
	@echo ".env.test exists: $(shell test -f .env.test && echo "✅ YES" || echo "❌ NO")"
	@echo "Docker container running: $(shell docker ps -q -f name=oilgas-postgres-test | wc -l | tr -d ' ') containers"
	@if [ -f .env.test ]; then \
		echo "DATABASE_URL: $(shell grep DATABASE_URL .env.test)"; \
		echo "TEST_DATABASE_URL: $(shell grep TEST_DATABASE_URL .env.test)"; \
	fi

debug-db:
	@echo "🔍 Debug test database..."
	@docker-compose -f docker-compose.test.yml ps
	@echo "\nDatabase connection test:"
	@docker exec oilgas-postgres-test psql -U postgres -d oilgas_inventory_test -c "SELECT 'Connection successful!' as status;" 2>/dev/null || echo "❌ Connection failed"

test-connection:
	@echo "🔌 Testing database connection..."
	@docker exec oilgas-postgres-test psql -U postgres -d oilgas_inventory_test -c "SELECT version();"

clean:
	@echo "🧹 Cleaning up..."
	rm -f integration-coverage.out integration-coverage.html ci-coverage.out
	docker-compose -f docker-compose.test.yml down -v
