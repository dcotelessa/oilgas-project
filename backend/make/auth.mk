# Authentication and User Management
.PHONY: create-admin create-tenant list-tenants list-users validate-rls

create-admin: ## Create admin user
	@echo "$(GREEN)Creating admin user...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set. Run 'make debug-env' to check environment.$(RESET)"; \
		exit 1; \
	fi
	@read -p "Email: " email; \
	read -s -p "Password: " password; echo; \
	read -p "Company (optional): " company; \
	go run scripts/utilities/create_user.go --email=$$email --password=$$password --role=admin --company="$$company" --tenant=default

create-tenant: ## Create new tenant
	@echo "$(GREEN)Creating new tenant...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set. Run 'make debug-env' to check environment.$(RESET)"; \
		exit 1; \
	fi
	@read -p "Tenant name: " name; \
	read -p "Tenant slug (optional): " slug; \
	if [ -z "$$slug" ]; then \
		go run scripts/utilities/create_tenant.go --name="$$name"; \
	else \
		go run scripts/utilities/create_tenant.go --name="$$name" --slug=$$slug; \
	fi

list-tenants: ## List all tenants
	@echo "$(GREEN)Active tenants:$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@psql "$(DATABASE_URL)" -c "SELECT name, slug, database_type, is_active, created_at FROM store.tenants ORDER BY created_at;" 2>/dev/null || echo "Database not accessible"

list-users: ## List all users
	@echo "$(GREEN)System users:$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@psql "$(DATABASE_URL)" -c "SELECT u.email, u.role, t.slug as tenant, u.last_login, u.created_at FROM auth.users u JOIN store.tenants t ON u.tenant_id = t.id WHERE u.deleted_at IS NULL ORDER BY u.created_at;" 2>/dev/null || echo "Database not accessible"

validate-rls: ## Validate Row-Level Security
	@echo "$(GREEN)Validating Row-Level Security...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	go run scripts/utilities/validate_rls.go
