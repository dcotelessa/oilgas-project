# Authentication and User Management (Single-Tenant)
.PHONY: create-admin list-users validate-db validate-rls health-check

# User Management (Placeholders for future implementation)
create-admin: ## Create admin user (placeholder)
	@echo "$(GREEN)Creating admin user...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set. Run 'make debug-env' to check environment.$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)📋 Admin user creation will be implemented with authentication system$(RESET)"
	@echo "$(BLUE)🔧 Current focus: Repository layer and API endpoints$(RESET)"
	@cd scripts/utilities && go run *.go create-user

list-users: ## List system users (placeholder)
	@echo "$(GREEN)System users:$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)📋 User listing will be implemented with authentication system$(RESET)"
	@echo "$(BLUE)🔧 For now, checking customer records:$(RESET)"
	@psql "$(DATABASE_URL)" -c "SELECT customer_id, customer, contact, email, created_at FROM store.customers WHERE deleted = false ORDER BY created_at;" 2>/dev/null || echo "Database not accessible"

# Database Validation
validate-db: ## Validate database schema and connectivity
	@echo "$(GREEN)Validating database schema...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@cd scripts/utilities && go run *.go validate-db

validate-rls: ## Validate Row-Level Security (future multi-tenant)
	@echo "$(GREEN)Validating Row-Level Security...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@cd scripts/utilities && go run *.go validate-rls

health-check: ## Comprehensive system health check
	@echo "$(GREEN)Running system health check...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@cd scripts/utilities && go run *.go health-check

# Multi-tenant commands (commented out for Step 10 implementation)
# create-tenant: ## Create new tenant (Step 10)
# list-tenants: ## List all tenants (Step 10)

# Future authentication features (Step 10)
.PHONY: help-auth
help-auth: ## Show authentication help
	@echo "$(BLUE)Authentication Features (Implementation Status):$(RESET)"
	@echo "  ✅ Database validation       - $(GREEN)Implemented$(RESET)"
	@echo "  ✅ Health checks            - $(GREEN)Implemented$(RESET)"
	@echo "  🔄 User management          - $(YELLOW)Planned for Phase 3$(RESET)"
	@echo "  🔄 Session management       - $(YELLOW)Planned for Phase 3$(RESET)"
	@echo "  🔮 Multi-tenant support     - $(CYAN)Planned for Step 10$(RESET)"
	@echo "  🔮 Row-Level Security       - $(CYAN)Planned for Step 10$(RESET)"
	@echo ""
	@echo "$(BLUE)Current Focus:$(RESET)"
	@echo "  - Repository layer implementation"
	@echo "  - API endpoint functionality"
	@echo "  - Single-tenant operations"
