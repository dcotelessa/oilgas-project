# =============================================================================
# TESTING MODULE - make/testing.mk
# =============================================================================
# Comprehensive testing framework for all components

.PHONY: test-setup test-unit test-integration test-tenant test-all

# =============================================================================
# TESTING SETUP
# =============================================================================

test-setup: ## 🛠️  Setup testing environment
	@echo "$(GREEN)🧪 Setting up testing environment...$(RESET)"
	@mkdir -p test/fixtures test/tmp logs/test
	@export DATABASE_URL="$$DATABASE_URL_TEST" && $(MAKE) db-migrate
	@echo "$(GREEN)✅ Testing environment ready$(RESET)"

test-clean-setup: test-clean test-setup ## 🛠️  Clean and setup testing

# =============================================================================
# UNIT TESTS
# =============================================================================

test-unit: ## 🧪 Run unit tests
	@echo "$(YELLOW)🧪 Running unit tests...$(RESET)"
	@go test -v -short ./internal/...
	@echo "$(GREEN)✅ Unit tests complete$(RESET)"

test-unit-coverage: ## 🧪 Run unit tests with coverage
	@echo "$(YELLOW)🧪 Running unit tests with coverage...$(RESET)"
	@go test -v -short -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(RESET)"

test-models: ## 🧪 Test data models
	@echo "$(YELLOW)🧪 Testing data models...$(RESET)"
	@go test -v ./internal/models/...

test-services: ## 🧪 Test service layer
	@echo "$(YELLOW)🧪 Testing service layer...$(RESET)"
	@go test -v ./internal/services/...

# =============================================================================
# INTEGRATION TESTS
# =============================================================================

test-integration: ## 🧪 Run integration tests
	@echo "$(YELLOW)🧪 Running integration tests...$(RESET)"
	@go test -v -tags=integration ./test/integration/...
	@echo "$(GREEN)✅ Integration tests complete$(RESET)"

test-api-integration: ## 🧪 Test API integration
	@echo "$(YELLOW)🧪 Testing API integration...$(RESET)"
	@go test -v -tags=integration ./test/api/...

test-database-integration: ## 🧪 Test database integration
	@echo "$(YELLOW)🧪 Testing database integration...$(RESET)"
	@go test -v -tags=integration ./test/database/...

# =============================================================================
# TENANT ISOLATION TESTS
# =============================================================================

test-tenant: ## 🧪 Test tenant isolation
	@echo "$(YELLOW)🧪 Testing tenant isolation...$(RESET)"
	@go test -v -tags=tenant ./test/tenant/...
	@echo "$(GREEN)✅ Tenant isolation tests complete$(RESET)"

test-tenant-security: ## 🧪 Test tenant security
	@echo "$(YELLOW)🧪 Testing tenant security...$(RESET)"
	@go test -v ./test/security/...

test-tenant-performance: ## 🧪 Test tenant performance
	@echo "$(YELLOW)🧪 Testing tenant performance...$(RESET)"
	@go test -v -bench=. ./test/performance/...

# =============================================================================
# COMPREHENSIVE TESTS
# =============================================================================

test-all: test-unit test-integration test-tenant ## 🧪 Run all tests
	@echo "$(GREEN)✅ All tests complete$(RESET)"

test-comprehensive: ## 🧪 Comprehensive test suite with coverage
	@echo "$(YELLOW)🧪 Running comprehensive test suite...$(RESET)"
	@$(MAKE) test-unit-coverage
	@$(MAKE) test-integration
	@$(MAKE) test-tenant
	@$(MAKE) test-performance
	@echo "$(GREEN)✅ Comprehensive testing complete$(RESET)"

test-performance: ## 🧪 Performance and load tests
	@echo "$(YELLOW)🧪 Running performance tests...$(RESET)"
	@go test -v -bench=. -benchmem ./test/performance/...
	@echo "$(GREEN)✅ Performance tests complete$(RESET)"

test-load: ## 🧪 Load testing
	@echo "$(YELLOW)🧪 Running load tests...$(RESET)"
	@if command -v hey >/dev/null 2>&1; then \
		echo "Testing health endpoint..."; \
		hey -n 1000 -c 10 http://localhost:$(API_PORT)/health; \
	else \
		echo "$(YELLOW)💡 Install hey for load testing: go install github.com/rakyll/hey@latest$(RESET)"; \
		go test -v -tags=load ./test/load/...; \
	fi

# =============================================================================
# SECURITY TESTS
# =============================================================================

test-security: ## 🧪 Security tests
	@echo "$(YELLOW)🧪 Running security tests...$(RESET)"
	@go test -v ./test/security/...
	@echo "$(GREEN)✅ Security tests complete$(RESET)"

test-auth-security: ## 🧪 Authentication security tests
	@echo "$(YELLOW)🧪 Testing authentication security...$(RESET)"
	@go test -v ./test/security/auth/...

test-injection: ## 🧪 SQL injection tests
	@echo "$(YELLOW)🧪 Testing SQL injection protection...$(RESET)"
	@go test -v ./test/security/injection/...

# =============================================================================
# DATA TESTS
# =============================================================================

test-data-import: ## 🧪 Test data import functionality
	@echo "$(YELLOW)🧪 Testing data import...$(RESET)"
	@go test -v ./test/data/import/...

test-data-export: ## 🧪 Test data export functionality
	@echo "$(YELLOW)🧪 Testing data export...$(RESET)"
	@go test -v ./test/data/export/...

test-csv-processing: ## 🧪 Test CSV processing
	@echo "$(YELLOW)🧪 Testing CSV processing...$(RESET)"
	@go test -v ./test/data/csv/...

# =============================================================================
# END-TO-END TESTS
# =============================================================================

test-e2e: ## 🧪 End-to-end tests
	@echo "$(YELLOW)🧪 Running end-to-end tests...$(RESET)"
	@echo "$(BLUE)Starting test API server...$(RESET)"
	@TEST_ENV=true go run cmd/server/main.go &
	@API_PID=$!; \
	sleep 3; \
	go test -v ./test/e2e/...; \
	TEST_RESULT=$?; \
	kill $API_PID; \
	exit $TEST_RESULT

test-workflow: ## 🧪 Test complete MDB → CSV → API workflow
	@echo "$(YELLOW)🧪 Testing complete workflow...$(RESET)"
	@go test -v ./test/workflow/...

# =============================================================================
# TEST UTILITIES
# =============================================================================

test-fixtures: ## 🧪 Generate test fixtures
	@echo "$(YELLOW)🧪 Generating test fixtures...$(RESET)"
	@go run cmd/tools/generate-fixtures.go
	@echo "$(GREEN)✅ Test fixtures generated$(RESET)"

test-mock-data: ## 🧪 Generate mock data for testing
	@echo "$(YELLOW)🧪 Generating mock data...$(RESET)"
	@go run cmd/tools/generate-mock-data.go \
		--tenants=3 \
		--customers=100 \
		--inventory=500 \
		--workorders=50
	@echo "$(GREEN)✅ Mock data generated$(RESET)"

test-reset: ## 🧪 Reset test environment
	@echo "$(YELLOW)🧪 Resetting test environment...$(RESET)"
	@export DATABASE_URL="$DATABASE_URL_TEST" && $(MAKE) db-migrate-reset
	@$(MAKE) test-setup
	@echo "$(GREEN)✅ Test environment reset$(RESET)"

# =============================================================================
# TEST REPORTING
# =============================================================================

test-report: ## 🧪 Generate comprehensive test report
	@echo "$(YELLOW)🧪 Generating test report...$(RESET)"
	@mkdir -p reports
	@go test -v -json ./... > reports/test-results.json
	@go run cmd/tools/test-report.go reports/test-results.json > reports/test-report.html
	@echo "$(GREEN)✅ Test report: reports/test-report.html$(RESET)"

test-coverage-report: ## 🧪 Generate coverage report
	@echo "$(YELLOW)🧪 Generating coverage report...$(RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o reports/coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "$(GREEN)✅ Coverage report: reports/coverage.html$(RESET)"

# =============================================================================
# CONTINUOUS INTEGRATION
# =============================================================================

test-ci: ## 🧪 CI/CD test pipeline
	@echo "$(YELLOW)🧪 Running CI/CD test pipeline...$(RESET)"
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@$(MAKE) test-tenant
	@$(MAKE) test-security
	@$(MAKE) test-performance
	@echo "$(GREEN)✅ CI/CD pipeline complete$(RESET)"

test-quick: ## 🧪 Quick test cycle (unit + basic integration)
	@echo "$(YELLOW)🧪 Running quick test cycle...$(RESET)"
	@go test -v -short ./internal/... ./test/basic/...
	@echo "$(GREEN)✅ Quick tests complete$(RESET)"

# =============================================================================
# CLEANUP
# =============================================================================

test-clean: ## 🛠️  Clean test artifacts
	@echo "$(YELLOW)🧹 Cleaning test artifacts...$(RESET)"
	@rm -rf test/tmp/*
	@rm -rf logs/test/*.log
	@rm -rf coverage.out coverage.html
	@rm -rf reports/*
	@echo "$(GREEN)✅ Test cleanup complete$(RESET)"

# =============================================================================
# HELP
# =============================================================================

help-testing: ## 📖 Show testing commands help
	@echo "$(BLUE)Testing Module Commands$(RESET)"
	@echo "======================="
	@echo ""
	@echo "$(GREEN)🛠️  SETUP:$(RESET)"
	@echo "  test-setup            - Setup testing environment"
	@echo "  test-clean-setup      - Clean and setup testing"
	@echo "  test-reset            - Reset test environment"
	@echo ""
	@echo "$(YELLOW)🧪 UNIT TESTS:$(RESET)"
	@echo "  test-unit             - Run unit tests"
	@echo "  test-unit-coverage    - Unit tests with coverage"
	@echo "  test-models           - Test data models"
	@echo "  test-services         - Test service layer"
	@echo ""
	@echo "$(BLUE)🧪 INTEGRATION TESTS:$(RESET)"
	@echo "  test-integration      - Run integration tests"
	@echo "  test-api-integration  - Test API integration"
	@echo "  test-database-integration - Test DB integration"
	@echo ""
	@echo "$(RED)🧪 TENANT TESTS:$(RESET)"
	@echo "  test-tenant           - Test tenant isolation"
	@echo "  test-tenant-security  - Test tenant security"
	@echo "  test-tenant-performance - Test tenant performance"
	@echo ""
	@echo "$(GREEN)🧪 COMPREHENSIVE:$(RESET)"
	@echo "  test-all              - Run all tests"
	@echo "  test-comprehensive    - Full test suite with coverage"
	@echo "  test-performance      - Performance tests"
	@echo "  test-load             - Load testing"
	@echo ""
	@echo "$(YELLOW)🧪 SECURITY:$(RESET)"
	@echo "  test-security         - Security tests"
	@echo "  test-auth-security    - Auth security tests"
	@echo "  test-injection        - SQL injection tests"
	@echo ""
	@echo "$(BLUE)🧪 DATA:$(RESET)"
	@echo "  test-data-import      - Test data import"
	@echo "  test-data-export      - Test data export"
	@echo "  test-csv-processing   - Test CSV processing"
	@echo ""
	@echo "$(RED)🧪 END-TO-END:$(RESET)"
	@echo "  test-e2e              - End-to-end tests"
	@echo "  test-workflow         - Test MDB→CSV→API workflow"
	@echo ""
	@echo "$(GREEN)🧪 UTILITIES:$(RESET)"
	@echo "  test-fixtures         - Generate test fixtures"
	@echo "  test-mock-data        - Generate mock data"
	@echo "  test-report           - Generate test report"
	@echo "  test-coverage-report  - Generate coverage report"
	@echo ""
	@echo "$(YELLOW)🧪 CI/CD:$(RESET)"
	@echo "  test-ci               - CI/CD test pipeline"
	@echo "  test-quick            - Quick test cycle"
