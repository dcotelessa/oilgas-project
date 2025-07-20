# Testing and Quality Assurance
.PHONY: test test-unit test-integration test-coverage test-auth

test: ## Run all tests
	@echo "$(GREEN)Running all tests...$(RESET)"
	go test -v ./...

test-unit: ## Run unit tests only
	@echo "$(GREEN)Running unit tests...$(RESET)"
	go test -v -short ./internal/...

test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(RESET)"
	@echo "$(YELLOW)Setting up test database...$(RESET)"
	$(MAKE) test-db-setup
	go test -v -run Integration ./test/...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(RESET)"

test-auth: ## Test authentication system
	@echo "$(GREEN)Testing authentication system...$(RESET)"
	go test -v ./internal/auth/...

test-db-setup: ## Setup test database
	@echo "$(GREEN)Setting up test database...$(RESET)"
	@if [ -z "$TEST_DATABASE_URL" ]; then \
		echo "$(YELLOW)TEST_DATABASE_URL not set, using default...$(RESET)"; \
		export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/oil_gas_test?sslmode=disable"; \
	fi
	@echo "$(GREEN)✅ Test database ready$(RESET)"
