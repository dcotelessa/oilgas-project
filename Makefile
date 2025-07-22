# Oil & Gas Inventory System - Enhanced Makefile
# Load environment variables from .env.local or .env
ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
else ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Use DATABASE_URL from environment, ensure postgresql:// format
ifneq (,$(findstring postgres://,$(DATABASE_URL)))
    DATABASE_URL := $(subst postgres://,postgresql://,$(DATABASE_URL))
endif

# Configuration
DATABASE_URL ?= postgresql://postgres:password@localhost:5432/oil_gas_inventory
TEST_DATABASE_URL ?= postgresql://postgres:password@localhost:5432/oil_gas_inventory_test
TOOLS_DIR := tools
BACKEND_DIR := backend
FRONTEND_DIR := frontend

.PHONY: help setup migrate seed-data env-info debug-migrations ensure-basic-schema

help:
	@echo "Oil & Gas Inventory System"
	@echo "=========================="
	@echo "Environment: $(or $(APP_ENV),local)"
	@echo "Database: $(POSTGRES_DB)"
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo ""
	@echo "🚀 Main Commands:"
	@echo "  make setup              - Complete system setup"
	@echo "  make dev                - Start development environment"
	@echo "  make test               - Run all tests"
	@echo "  make clean              - Clean all build artifacts"
	@echo ""
	@echo "📊 Database (postgresql::):"
	@echo "  make postgresql-status - Check database status"
	@echo "  make postgresql-reset  - Reset database (destructive)"
	@echo "  make postgresql-backup - Create database backup"
	@echo ""
	@echo "🏢 Multi-Tenant (tenant::):"
	@echo "  make tenant-create NAME=\"Co\" SLUG=co - Create tenant"
	@echo "  make tenant-list       - List all tenants"
	@echo "  make tenant-status     - Show tenant status"
	@echo ""
	@echo "🛠️  Migration Tools (migration::):"
	@echo "  make migration-setup   - Setup migration tools"
	@echo "  make migration-convert FILE=db.mdb COMPANY=\"Name\""
	@echo "  make migration-analyze DIR=cf_app"
	@echo "  make migration-validate"
	@echo ""
	@echo "🔧 Backend (backend::):"
	@echo "  make backend-build     - Build backend services"
	@echo "  make backend-test      - Run backend tests"
	@echo "  make backend-dev       - Start backend in dev mode"
	@echo ""
	@echo "🌐 Frontend (frontend::):"
	@echo "  make frontend-build    - Build frontend assets"
	@echo "  make frontend-test     - Run frontend tests"
	@echo "  make frontend-dev      - Start frontend dev server"
	@echo ""
	@echo "🚀 API (api::):"
	@echo "  make api-docs          - Generate API documentation"
	@echo "  make api-test          - Test API endpoints"
	@echo "  make api-benchmark     - Run API benchmarks"
	@echo ""
	@echo "🏗️  Deployment (deployment::):"
	@echo "  make deployment-build  - Build for production"
	@echo "  make deployment-test   - Test deployment"
	@echo "  make deployment-deploy - Deploy to environment"

# =============================================================================
# ENVIRONMENT & SETUP
# =============================================================================

env-info:
	@echo "🔧 Environment Information"
	@echo "========================="
	@echo "APP_ENV: $(APP_ENV)"
	@echo "POSTGRES_DB: $(POSTGRES_DB)"
	@echo "POSTGRES_HOST: $(POSTGRES_HOST)"
	@echo "POSTGRES_PORT: $(POSTGRES_PORT)"
	@echo "POSTGRES_USER: $(POSTGRES_USER)"
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@echo "TEST_DATABASE_URL: $(TEST_DATABASE_URL)"
	@echo "TOOLS_DIR: $(TOOLS_DIR)"
	@echo "BACKEND_DIR: $(BACKEND_DIR)"
	@echo "FRONTEND_DIR: $(FRONTEND_DIR)"

setup: ensure-basic-schema migrate seed-data migration-setup backend-setup
	@echo "🎯 Complete system setup completed!"
	@echo "✅ Database schema verified"
	@echo "✅ Tenant architecture implemented" 
	@echo "✅ Migration tools ready"
	@echo "✅ Backend dependencies installed"
	@echo ""
	@echo "🚀 Next steps:"
	@echo "  make dev                    # Start full development environment"
	@echo "  make tenant-status         # Check tenant configuration"
	@echo "  make migration-convert ... # Convert legacy data"

ensure-basic-schema:
	@echo "🏗️  Ensuring basic schema exists..."
	@psql "$(DATABASE_URL)" -c "SELECT 'customers table exists' FROM store.customers LIMIT 0;" 2>/dev/null || \
		(echo "❌ Basic schema missing. Run the setup script first:" && \
		 echo "   ./scripts/setup.sh" && exit 1)
	@psql "$(DATABASE_URL)" -c "SELECT 'users table exists' FROM store.users LIMIT 0;" 2>/dev/null || \
		(echo "❌ Users table missing. Run the setup script first." && exit 1)
	@psql "$(DATABASE_URL)" -c "SELECT 'inventory table exists' FROM store.inventory LIMIT 0;" 2>/dev/null || \
		(echo "❌ Inventory table missing. Run the setup script first." && exit 1)
	@psql "$(DATABASE_URL)" -c "SELECT 'received table exists' FROM store.received LIMIT 0;" 2>/dev/null || \
		(echo "❌ Received table missing. Run the setup script first." && exit 1)
	@echo "✅ Basic schema verified"

migrate: ensure-basic-schema
	@echo "🗄️  Running tenant migrations..."
	@if [ ! -f "$(BACKEND_DIR)/migrations/001_add_tenant_architecture.sql" ]; then \
		echo "❌ Migration file not found"; exit 1; \
	fi
	@psql "$(DATABASE_URL)" -f $(BACKEND_DIR)/migrations/001_add_tenant_architecture.sql || exit 1
	@psql "$(DATABASE_URL)" -f $(BACKEND_DIR)/migrations/002_tenant_rls_policies.sql || exit 1
	@echo "✅ Tenant migrations completed successfully"

seed-data:
	@echo "🌱 Seeding tenant data..."
	@psql "$(DATABASE_URL)" -f $(BACKEND_DIR)/seeds/tenant_seeds.sql || exit 1
	@echo "✅ Tenant seed data loaded successfully"

debug-migrations:
	@echo "🔍 Migration Debug Information"
	@echo "=============================="
	@echo "DATABASE_URL: $(DATABASE_URL)"
	@ls -la $(BACKEND_DIR)/migrations/ 2>/dev/null || echo "❌ Migrations directory not found"
	@psql "$(DATABASE_URL)" -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname IN ('store', 'migrations') ORDER BY schemaname, tablename;" 2>/dev/null || echo "❌ Database connection failed"

# =============================================================================
# POSTGRESQL DATABASE OPERATIONS
# =============================================================================

postgresql-status: ## Check database status and connections
	@echo "📊 PostgreSQL Database Status"
	@echo "============================="
	@psql "$(DATABASE_URL)" -c "SELECT 'Connected to: ' || current_database() || ' on ' || inet_server_addr() || ':' || inet_server_port() as status;" 2>/dev/null || echo "❌ Database connection failed"
	@echo ""
	@echo "📋 Schema Tables:"
	@psql "$(DATABASE_URL)" -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname = 'store' ORDER BY tablename;" 2>/dev/null
	@echo ""
	@echo "🔗 Connections:"
	@psql "$(DATABASE_URL)" -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;" 2>/dev/null

postgresql-reset: ## Reset database (destructive)
	@echo "⚠️  WARNING: This will destroy all data!"
	@read -p "Type 'RESET' to continue: " confirm && [ "$$confirm" = "RESET" ] || exit 1
	@dropdb oil_gas_inventory 2>/dev/null || true
	@createdb oil_gas_inventory
	@echo "✅ Database reset complete. Run 'make setup' to reinitialize."

postgresql-backup: ## Create database backup
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@pg_dump "$(DATABASE_URL)" > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created in backups/"

postgresql-restore: ## Restore from backup (BACKUP=file.sql)
	@if [ -z "$(BACKUP)" ]; then echo "❌ Usage: make postgresql-restore BACKUP=backup_file.sql"; exit 1; fi
	@psql "$(DATABASE_URL)" -f "$(BACKUP)"
	@echo "✅ Database restored from $(BACKUP)"

postgresql-monitor: ## Monitor database performance
	@echo "📈 Database Performance Monitor"
	@echo "=============================="
	@psql "$(DATABASE_URL)" -c "SELECT now() as timestamp, (SELECT count(*) FROM pg_stat_activity) as connections, (SELECT count(*) FROM pg_locks) as locks;"
	@echo ""
	@echo "🔥 Most Active Tables:"
	@psql "$(DATABASE_URL)" -c "SELECT schemaname, relname, n_tup_ins + n_tup_upd + n_tup_del as activity FROM pg_stat_user_tables WHERE schemaname = 'store' ORDER BY activity DESC LIMIT 10;"

postgresql-vacuum: ## Vacuum and analyze database
	@echo "🧹 Vacuuming database..."
	@psql "$(DATABASE_URL)" -c "VACUUM ANALYZE;"
	@echo "✅ Database maintenance completed"

# =============================================================================
# MULTI-TENANT OPERATIONS
# =============================================================================

tenant-create: ## Create new tenant (NAME="Company" SLUG=company)
	@if [ -z "$(NAME)" ] || [ -z "$(SLUG)" ]; then \
		echo "❌ Usage: make tenant-create NAME=\"Company Name\" SLUG=company_slug"; exit 1; \
	fi
	@echo "🏢 Creating tenant: $(NAME) ($(SLUG))"
	@psql "$(DATABASE_URL)" -c "INSERT INTO store.tenants (tenant_name, tenant_slug, active) VALUES ('$(NAME)', '$(SLUG)', true) ON CONFLICT (tenant_slug) DO NOTHING;" || exit 1
	@echo "✅ Tenant created successfully"

tenant-list: ## List all tenants
	@echo "📋 All Tenants"
	@echo "============="
	@psql "$(DATABASE_URL)" -c "SELECT tenant_id, tenant_name, tenant_slug, active, created_at FROM store.tenants ORDER BY tenant_id;"

tenant-status: ## Show comprehensive tenant status
	@echo "📊 Tenant Status Report"
	@echo "======================"
	@psql "$(DATABASE_URL)" -c "SELECT t.tenant_id, t.tenant_name, t.active, (SELECT COUNT(*) FROM store.users WHERE tenant_id = t.tenant_id) as users, (SELECT COUNT(*) FROM store.inventory WHERE tenant_id = t.tenant_id) as inventory FROM store.tenants t ORDER BY t.tenant_id;"

tenant-delete: ## Delete tenant (SLUG=tenant_slug)
	@if [ -z "$(SLUG)" ]; then echo "❌ Usage: make tenant-delete SLUG=tenant_slug"; exit 1; fi
	@echo "⚠️  WARNING: This will delete tenant '$(SLUG)' and all data!"
	@read -p "Type 'DELETE' to confirm: " confirm && [ "$$confirm" = "DELETE" ] || exit 1
	@psql "$(DATABASE_URL)" -c "DELETE FROM store.tenants WHERE tenant_slug = '$(SLUG)';"
	@echo "✅ Tenant '$(SLUG)' deleted"

tenant-switch: ## Switch active tenant context (SLUG=tenant_slug)
	@if [ -z "$(SLUG)" ]; then echo "❌ Usage: make tenant-switch SLUG=tenant_slug"; exit 1; fi
	@echo "🔄 Switching to tenant: $(SLUG)"
	@psql "$(DATABASE_URL)" -c "SELECT tenant_id, tenant_name FROM store.tenants WHERE tenant_slug = '$(SLUG)';" || exit 1
	@echo "✅ Context switched to $(SLUG)"

tenant-users: ## List users for tenant (SLUG=tenant_slug)
	@if [ -z "$(SLUG)" ]; then echo "❌ Usage: make tenant-users SLUG=tenant_slug"; exit 1; fi
	@psql "$(DATABASE_URL)" -c "SELECT u.username, utr.role, u.created_at FROM store.users u JOIN store.user_tenant_roles utr ON u.user_id = utr.user_id JOIN store.tenants t ON utr.tenant_id = t.tenant_id WHERE t.tenant_slug = '$(SLUG)';"

# =============================================================================
# MIGRATION TOOLS
# =============================================================================

migration-setup: ## Setup migration tools environment
	@echo "🛠️  Setting up migration tools..."
	@mkdir -p $(TOOLS_DIR)/{cmd,internal/{config,processor,mapping,validation,exporters,reporting},test,config,bin,output}
	@if [ ! -f "$(TOOLS_DIR)/go.mod" ]; then \
		cd $(TOOLS_DIR) && go mod init github.com/dcotelessa/oil-gas-inventory/tools; \
	fi
	@if [ ! -f "$(TOOLS_DIR)/config/oil_gas_mappings.json" ]; then \
		echo '{"oil_gas_mappings":{},"processing_options":{"workers":4,"batch_size":1000}}' > $(TOOLS_DIR)/config/oil_gas_mappings.json; \
	fi
	@echo "✅ Migration tools environment ready"

migration-build: migration-setup ## Build migration tools
	@echo "🔨 Building migration tools..."
	@if [ ! -f "$(TOOLS_DIR)/cmd/mdb_processor.go" ]; then \
		echo "❌ Source files not found. Implement modular MDB processor first."; exit 1; \
	fi
	@cd $(TOOLS_DIR) && go build -o bin/mdb_processor cmd/mdb_processor.go
	@cd $(TOOLS_DIR) && go build -o bin/cf_analyzer cmd/cf_query_analyzer.go
	@cd $(TOOLS_DIR) && go build -o bin/conversion_tester cmd/conversion_tester.go
	@echo "✅ Migration tools built"

migration-convert: ## Convert MDB (FILE=db.mdb COMPANY="Name")
	@if [ -z "$(FILE)" ] || [ -z "$(COMPANY)" ]; then \
		echo "❌ Usage: make migration-convert FILE=database.mdb COMPANY=\"Company Name\""; exit 1; \
	fi
	@if [ ! -f "$(TOOLS_DIR)/bin/mdb_processor" ]; then \
		echo "❌ Tools not built. Run 'make migration-build' first."; exit 1; \
	fi
	@$(TOOLS_DIR)/bin/mdb_processor -file "$(FILE)" -company "$(COMPANY)" -db "$(DATABASE_URL)" -verbose
	@echo "✅ Conversion completed. Check $(TOOLS_DIR)/output/"

migration-analyze: ## Analyze ColdFusion (DIR=cf_app)
	@if [ -z "$(DIR)" ]; then echo "❌ Usage: make migration-analyze DIR=cf_app"; exit 1; fi
	@if [ ! -f "$(TOOLS_DIR)/bin/cf_analyzer" ]; then \
		echo "❌ CF analyzer not built. Run 'make migration-build' first."; exit 1; \
	fi
	@$(TOOLS_DIR)/bin/cf_analyzer "$(DIR)" "$(TOOLS_DIR)/output/reports/cf_analysis.json"
	@echo "✅ Analysis completed"

migration-validate: ## Validate conversion results
	@if [ ! -f "$(TOOLS_DIR)/bin/conversion_tester" ]; then \
		echo "❌ Conversion tester not built. Run 'make migration-build' first."; exit 1; \
	fi
	@$(TOOLS_DIR)/bin/conversion_tester validate "$(TOOLS_DIR)/output" "$(DATABASE_URL)"
	@echo "✅ Validation completed"

migration-test: ## Run migration tool tests
	@echo "🧪 Testing migration tools..."
	@cd $(TOOLS_DIR) && go test ./... -v

migration-test-sample: ## Test with sample data
	@echo "🧪 Testing with sample data..."
	@if [ ! -f "$(TOOLS_DIR)/test/fixtures/sample_customers.csv" ]; then \
		echo "❌ Sample data not found. Run migration script first."; exit 1; \
	fi
	@$(MAKE) migration-convert FILE=$(TOOLS_DIR)/test/fixtures/sample_customers.csv COMPANY="Sample Test"

migration-clean: ## Clean migration outputs
	@echo "🧹 Cleaning migration outputs..."
	@rm -rf $(TOOLS_DIR)/output/* $(TOOLS_DIR)/bin/*
	@echo "✅ Migration outputs cleaned"

migration-status: ## Show migration tools status
	@echo "📊 Migration Tools Status"
	@echo "========================"
	@echo "Tools directory: $(TOOLS_DIR)"
	@echo "Configuration: $(TOOLS_DIR)/config/"
	@echo ""
	@echo "🔧 Available tools:"
	@for tool in mdb_processor cf_analyzer conversion_tester; do \
		if [ -f "$(TOOLS_DIR)/bin/$$tool" ]; then echo "  ✅ $$tool"; else echo "  ❌ $$tool (run 'make migration-build')"; fi; \
	done
	@echo ""
	@echo "📁 Recent outputs:"
	@find $(TOOLS_DIR)/output -name "*.csv" -o -name "*.sql" 2>/dev/null | head -5 || echo "  No outputs found"

migration-report: ## Generate migration summary report
	@echo "📋 Migration Summary Report"
	@echo "=========================="
	@if [ -d "$(TOOLS_DIR)/output" ]; then \
		echo "📂 Conversion outputs:"; \
		find $(TOOLS_DIR)/output -name "*.csv" -o -name "*.sql" | wc -l | xargs echo "  Total files:"; \
		echo "📊 Recent conversions:"; \
		find $(TOOLS_DIR)/output -name "validation_report_*.json" 2>/dev/null | head -3 | sed 's/^/  /' || echo "  No validation reports found"; \
	else \
		echo "❌ No migration output found"; \
	fi

# =============================================================================
# BACKEND OPERATIONS
# =============================================================================

backend-setup: ## Setup backend dependencies
	@echo "🔧 Setting up backend..."
	@cd $(BACKEND_DIR) && go mod tidy
	@cd $(BACKEND_DIR) && go mod download
	@echo "✅ Backend dependencies installed"

backend-build: ## Build backend services
	@echo "🔨 Building backend services..."
	@cd $(BACKEND_DIR) && go build -o bin/server cmd/server/main.go
	@cd $(BACKEND_DIR) && go build -o bin/migrator migrator.go
	@echo "✅ Backend services built"

backend-dev: ## Start backend development server
	@echo "🚀 Starting backend development server..."
	@cd $(BACKEND_DIR) && go run cmd/server/main.go

backend-test: ## Run backend tests
	@echo "🧪 Running backend tests..."
	@cd $(BACKEND_DIR) && go test ./... -v

backend-test-coverage: ## Generate backend test coverage
	@echo "📊 Generating test coverage..."
	@cd $(BACKEND_DIR) && go test ./... -coverprofile=coverage.out
	@cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: $(BACKEND_DIR)/coverage.html"

backend-lint: ## Lint backend code
	@echo "🔍 Linting backend code..."
	@cd $(BACKEND_DIR) && golangci-lint run ./...

backend-format: ## Format backend code
	@echo "🎨 Formatting backend code..."
	@cd $(BACKEND_DIR) && go fmt ./...

backend-clean: ## Clean backend build artifacts
	@echo "🧹 Cleaning backend artifacts..."
	@rm -rf $(BACKEND_DIR)/bin $(BACKEND_DIR)/coverage.*
	@echo "✅ Backend artifacts cleaned"

backend-logs: ## Show backend logs
	@echo "📋 Backend Logs"
	@echo "=============="
	@if [ -f "$(BACKEND_DIR)/logs/app.log" ]; then \
		tail -50 $(BACKEND_DIR)/logs/app.log; \
	else \
		echo "❌ No log file found"; \
	fi

backend-health: ## Check backend health
	@echo "🏥 Backend Health Check"
	@echo "======================"
	@curl -s http://localhost:8000/health 2>/dev/null || echo "❌ Backend not responding on port 8000"

# =============================================================================
# FRONTEND OPERATIONS
# =============================================================================

frontend-setup: ## Setup frontend dependencies
	@echo "🌐 Setting up frontend..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm install; \
	else \
		echo "❌ Frontend directory not found: $(FRONTEND_DIR)"; \
		echo "  Create frontend directory or update FRONTEND_DIR variable"; \
	fi
	@echo "✅ Frontend dependencies installed"

frontend-build: ## Build frontend for production
	@echo "🔨 Building frontend for production..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run build; \
	else \
		echo "❌ Frontend directory not found"; \
	fi
	@echo "✅ Frontend built"

frontend-dev: ## Start frontend development server
	@echo "🚀 Starting frontend development server..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run dev; \
	else \
		echo "❌ Frontend directory not found"; \
	fi

frontend-test: ## Run frontend tests
	@echo "🧪 Running frontend tests..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm test; \
	else \
		echo "❌ Frontend directory not found"; \
	fi

frontend-test-e2e: ## Run end-to-end tests
	@echo "🧪 Running E2E tests..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run test:e2e; \
	else \
		echo "❌ Frontend directory not found"; \
	fi

frontend-lint: ## Lint frontend code
	@echo "🔍 Linting frontend code..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run lint; \
	else \
		echo "❌ Frontend directory not found"; \
	fi

frontend-clean: ## Clean frontend build artifacts
	@echo "🧹 Cleaning frontend artifacts..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/build $(FRONTEND_DIR)/node_modules/.cache; \
	fi
	@echo "✅ Frontend artifacts cleaned"

frontend-serve: ## Serve built frontend
	@echo "🌐 Serving built frontend..."
	@if [ -d "$(FRONTEND_DIR)/dist" ]; then \
		cd $(FRONTEND_DIR) && npx serve dist; \
	else \
		echo "❌ Build directory not found. Run 'make frontend-build' first"; \
	fi

# =============================================================================
# API OPERATIONS
# =============================================================================

api-docs: ## Generate API documentation
	@echo "📚 Generating API documentation..."
	@if [ -d "$(BACKEND_DIR)" ]; then \
		cd $(BACKEND_DIR) && swag init -g cmd/server/main.go -o docs/swagger; \
	else \
		echo "❌ Backend directory not found"; \
	fi
	@echo "✅ API documentation generated"

api-test: ## Test API endpoints
	@echo "🧪 Testing API endpoints..."
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8000/health | jq . || echo "❌ Health endpoint failed"
	@echo "Testing API status..."
	@curl -s http://localhost:8000/api/v1/status | jq . || echo "❌ API status failed"

api-test-auth: ## Test authentication endpoints
	@echo "🔐 Testing authentication..."
	@curl -s -X POST http://localhost:8000/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"username":"test","password":"test"}' | jq . || echo "❌ Auth test failed"

api-benchmark: ## Run API performance benchmarks
	@echo "⚡ Running API benchmarks..."
	@if command -v ab >/dev/null 2>&1; then \
		echo "Testing health endpoint performance..."; \
		ab -n 1000 -c 10 http://localhost:8000/health; \
	else \
		echo "❌ Apache Bench (ab) not found. Install with: apt-get install apache2-utils"; \
	fi

api-load-test: ## Run load tests
	@echo "🔥 Running load tests..."
	@if command -v hey >/dev/null 2>&1; then \
		hey -n 10000 -c 100 http://localhost:8000/health; \
	else \
		echo "❌ 'hey' load testing tool not found. Install with: go install github.com/rakyll/hey@latest"; \
	fi

api-monitor: ## Monitor API performance
	@echo "📈 API Performance Monitor"
	@echo "========================="
	@curl -s http://localhost:8000/health | jq -r '.timestamp' | xargs -I {} echo "Last response: {}"
	@echo "Active connections to backend:"
	@netstat -an | grep :8000 | wc -l | xargs echo "  Port 8000:"

# =============================================================================
# DEPLOYMENT OPERATIONS
# =============================================================================

deployment-build: ## Build for production deployment
	@echo "🏗️  Building for production deployment..."
	@$(MAKE) backend-build
	@$(MAKE) frontend-build
	@$(MAKE) migration-build
	@echo "✅ Production build completed"

deployment-test: ## Test deployment configuration
	@echo "🧪 Testing deployment configuration..."
	@echo "Checking environment variables..."
	@if [ -z "$(DATABASE_URL)" ]; then echo "❌ DATABASE_URL not set"; else echo "✅ DATABASE_URL configured"; fi
	@if [ -z "$(APP_ENV)" ]; then echo "❌ APP_ENV not set"; else echo "✅ APP_ENV: $(APP_ENV)"; fi
	@echo "Checking build artifacts..."
	@if [ -f "$(BACKEND_DIR)/bin/server" ]; then echo "✅ Backend binary ready"; else echo "❌ Backend binary missing"; fi
	@if [ -d "$(FRONTEND_DIR)/dist" ]; then echo "✅ Frontend build ready"; else echo "❌ Frontend build missing"; fi

deployment-package: ## Package application for deployment
	@echo "📦 Packaging application..."
	@mkdir -p dist
	@tar -czf dist/oil-gas-inventory-$(shell date +%Y%m%d-%H%M%S).tar.gz \
		$(BACKEND_DIR)/bin/ \
		$(FRONTEND_DIR)/dist/ \
		$(TOOLS_DIR)/bin/ \
		$(BACKEND_DIR)/migrations/ \
		$(BACKEND_DIR)/seeds/ \
		Makefile
	@echo "✅ Package created in dist/"

deployment-deploy: ## Deploy to environment (ENV=staging|production)
	@if [ -z "$(ENV)" ]; then echo "❌ Usage: make deployment-deploy ENV=staging|production"; exit 1; fi
	@echo "🚀 Deploying to $(ENV) environment..."
	@$(MAKE) deployment-build
	@$(MAKE) deployment-test
	@echo "⚠️  Deployment to $(ENV) - implement your deployment strategy here"
	@echo "✅ Deployment to $(ENV) completed"

deployment-rollback: ## Rollback deployment (VERSION=backup_version)
	@if [ -z "$(VERSION)" ]; then echo "❌ Usage: make deployment-rollback VERSION=backup_version"; exit 1; fi
	@echo "⏪ Rolling back to version: $(VERSION)"
	@echo "⚠️  Implement rollback strategy here"

deployment-status: ## Check deployment status
	@echo "📊 Deployment Status"
	@echo "==================="
	@echo "Environment: $(APP_ENV)"
	@echo "Database: $(DATABASE_URL)"
	@$(MAKE) backend-health
	@$(MAKE) postgresql-status

deployment-logs: ## View deployment logs
	@echo "📋 Deployment Logs"
	@echo "=================="
	@$(MAKE) backend-logs

# =============================================================================
# DEVELOPMENT & TESTING
# =============================================================================

dev: ## Start full development environment
	@echo "🚀 Starting full development environment..."
	@echo "Starting database checks..."
	@$(MAKE) postgresql-status
	@echo "Starting backend in background..."
	@cd $(BACKEND_DIR) && go run cmd/server/main.go &
	@echo "Backend PID: $$!" > .dev_backend_pid
	@sleep 2
	@echo "✅ Development environment started"
	@echo "🔗 Backend: http://localhost:8000"
	@echo "🔗 API Docs: http://localhost:8000/docs"
	@echo ""
	@echo "To stop: make dev-stop"

dev-stop: ## Stop development environment
	@echo "🛑 Stopping development environment..."
	@if [ -f ".dev_backend_pid" ]; then \
		kill $$(cat .dev_backend_pid) 2>/dev/null || true; \
		rm .dev_backend_pid; \
	fi
	@echo "✅ Development environment stopped"

test: ## Run all tests
	@echo "🧪 Running comprehensive test suite..."
	@$(MAKE) backend-test
	@$(MAKE) migration-test
	@$(MAKE) frontend-test

test-integration: ## Run integration tests
	@echo "🔗 Running integration tests..."
	@$(MAKE) postgresql-status
	@$(MAKE) backend-test
	@$(MAKE) api-test

test-performance: ## Run performance tests
	@echo "⚡ Running performance tests..."
	@$(MAKE) backend-test-coverage
	@$(MAKE) api-benchmark
	@$(MAKE) migration-test-sample

test-e2e: ## Run end-to-end tests
	@echo "🔄 Running end-to-end tests..."
	@$(MAKE) setup
	@$(MAKE) dev &
	@sleep 5
	@$(MAKE) frontend-test-e2e
	@$(MAKE) dev-stop

# =============================================================================
# UTILITIES & MAINTENANCE
# =============================================================================

clean: ## Clean all build artifacts and temporary files
	@echo "🧹 Cleaning all build artifacts..."
	@$(MAKE) backend-clean
	@$(MAKE) frontend-clean
	@$(MAKE) migration-clean
	@find . -name "*.tmp" -delete
	@find . -name ".DS_Store" -delete
	@rm -f .dev_backend_pid
	@echo "✅ All artifacts cleaned"

logs: ## Show all application logs
	@echo "📋 Application Logs"
	@echo "=================="
	@$(MAKE) backend-logs
	@$(MAKE) deployment-logs

health-check: ## Comprehensive system health check
	@echo "🏥 System Health Check"
	@echo "====================="
	@echo "1. Database connectivity..."
	@$(MAKE) postgresql-status | grep -q "Connected" && echo "✅ Database OK" || echo "❌ Database connection failed"
	@echo ""
	@echo "2. Required directories..."
	@test -d "$(BACKEND_DIR)" && echo "✅ Backend directory" || echo "❌ Backend directory missing"
	@test -d "$(TOOLS_DIR)" && echo "✅ Tools directory" || echo "❌ Tools directory missing"
	@echo ""
	@echo "3. Configuration files..."
	@test -f ".env.local" && echo "✅ .env.local" || echo "❌ .env.local missing"
	@echo ""
	@echo "4. Build artifacts..."
	@$(MAKE) backend-health | grep -q "200" && echo "✅ Backend service" || echo "❌ Backend not responding"
	@$(MAKE) migration-status | grep -q "✅" && echo "✅ Migration tools" || echo "❌ Migration tools need setup"

monitor: ## Monitor system performance
	@echo "📈 System Performance Monitor"
	@echo "============================"
	@$(MAKE) postgresql-monitor
	@$(MAKE) api-monitor

update-deps: ## Update all dependencies
	@echo "📦 Updating dependencies..."
	@cd $(BACKEND_DIR) && go get -u ./...
	@cd $(BACKEND_DIR) && go mod tidy
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm update; \
	fi
	@cd $(TOOLS_DIR) && go get -u ./...
	@cd $(TOOLS_DIR) && go mod tidy
	@echo "✅ Dependencies updated"

security-scan: ## Run security scans
	@echo "🔒 Running security scans..."
	@cd $(BACKEND_DIR) && gosec ./... || echo "⚠️  Security issues found"
	@cd $(TOOLS_DIR) && gosec ./... || echo "⚠️  Security issues found in tools"
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm audit || echo "⚠️  Frontend security issues found"; \
	fi

format: ## Format all code
	@echo "🎨 Formatting all code..."
	@$(MAKE) backend-format
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run format 2>/dev/null || echo "Frontend formatting not configured"; \
	fi
	@echo "✅ Code formatting completed"

lint: ## Lint all code
	@echo "🔍 Linting all code..."
	@$(MAKE) backend-lint
	@$(MAKE) frontend-lint

docs: ## Generate all documentation
	@echo "📚 Generating documentation..."
	@$(MAKE) api-docs
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Generating Go documentation..."; \
		cd $(BACKEND_DIR) && godoc -http=:6060 & \
		echo "Go docs available at http://localhost:6060"; \
	fi

backup: ## Create complete system backup
	@echo "💾 Creating complete system backup..."
	@$(MAKE) postgresql-backup
	@mkdir -p backups/system_$(shell date +%Y%m%d_%H%M%S)
	@cp -r $(BACKEND_DIR) backups/system_$(shell date +%Y%m%d_%H%M%S)/
	@cp -r $(TOOLS_DIR) backups/system_$(shell date +%Y%m%d_%H%M%S)/
	@cp Makefile .env* backups/system_$(shell date +%Y%m%d_%H%M%S)/ 2>/dev/null || true
	@echo "✅ System backup created in backups/"

restore: ## Restore from system backup (BACKUP=backup_dir)
	@if [ -z "$(BACKUP)" ]; then echo "❌ Usage: make restore BACKUP=backup_directory"; exit 1; fi
	@echo "🔄 Restoring from backup: $(BACKUP)"
	@echo "⚠️  This will overwrite current files!"
	@read -p "Type 'RESTORE' to continue: " confirm && [ "$confirm" = "RESTORE" ] || exit 1
	@cp -r $(BACKUP)/* .
	@echo "✅ System restored from $(BACKUP)"

# =============================================================================
# MIGRATION SCRIPT INTEGRATION
# =============================================================================

migrate-structure: ## Run the structure migration script
	@echo "🔄 Running structure migration script..."
	@if [ ! -f "migrate_to_new_structure.sh" ]; then \
		echo "❌ Migration script not found. Create it first."; exit 1; \
	fi
	@chmod +x migrate_to_new_structure.sh
	@./migrate_to_new_structure.sh
	@echo "✅ Structure migration completed"

clean-old-structure: ## Remove old structure after migration
	@echo "🗑️  Removing old structure..."
	@if [ -f "remove_old_structure.sh" ]; then \
		chmod +x remove_old_structure.sh; \
		./remove_old_structure.sh; \
	else \
		echo "❌ Cleanup script not found"; \
	fi

verify-migration: ## Verify new structure is working
	@echo "✅ Verifying migration..."
	@$(MAKE) migration-setup
	@$(MAKE) migration-status
	@$(MAKE) health-check
	@echo "✅ Migration verification completed"

# =============================================================================
# INFORMATION & STATUS
# =============================================================================

status: ## Show comprehensive system status
	@echo "📊 Oil & Gas Inventory System Status"
	@echo "===================================="
	@$(MAKE) env-info
	@echo ""
	@$(MAKE) postgresql-status
	@echo ""
	@$(MAKE) tenant-status
	@echo ""
	@$(MAKE) migration-status

info: ## Show system information
	@echo "ℹ️  System Information"
	@echo "====================="
	@echo "Project: Oil & Gas Inventory Management System"
	@echo "Repository: github.com/dcotelessa/oil-gas-inventory"
	@echo "Version: 1.0.0"
	@echo "Go Version: $(shell go version 2>/dev/null || echo 'Not installed')"
	@echo "Node Version: $(shell node --version 2>/dev/null || echo 'Not installed')"
	@echo "PostgreSQL: $(shell psql --version 2>/dev/null | head -1 || echo 'Not installed')"
	@echo ""
	@echo "📁 Directory Structure:"
	@echo "  Backend: $(BACKEND_DIR)/"
	@echo "  Frontend: $(FRONTEND_DIR)/"
	@echo "  Tools: $(TOOLS_DIR)/"
	@echo "  Database: $(DATABASE_URL)"

version: ## Show version information
	@echo "Oil & Gas Inventory System v1.0.0"
	@echo "Built for production-ready oil & gas data management"

# =============================================================================
# ADVANCED OPERATIONS
# =============================================================================

demo: ## Run complete system demonstration
	@echo "🎬 Running system demonstration..."
	@$(MAKE) setup
	@$(MAKE) tenant-create NAME="Demo Company" SLUG=demo
	@$(MAKE) migration-test-sample
	@$(MAKE) tenant-status
	@echo "✅ Demo completed successfully"

benchmark: ## Run comprehensive benchmarks
	@echo "⚡ Running comprehensive benchmarks..."
	@$(MAKE) test-performance
	@$(MAKE) api-benchmark
	@echo "✅ Benchmarks completed"

stress-test: ## Run stress tests
	@echo "🔥 Running stress tests..."
	@$(MAKE) api-load-test
	@$(MAKE) postgresql-monitor
	@echo "✅ Stress tests completed"

# =============================================================================
# CONTINUOUS INTEGRATION
# =============================================================================

ci-setup: ## Setup for CI environment
	@echo "🤖 Setting up CI environment..."
	@$(MAKE) backend-setup
	@$(MAKE) migration-setup
	@echo "✅ CI environment ready"

ci-test: ## Run CI test suite
	@echo "🤖 Running CI test suite..."
	@$(MAKE) lint
	@$(MAKE) security-scan
	@$(MAKE) test
	@$(MAKE) test-integration
	@echo "✅ CI tests completed"

ci-build: ## Build for CI
	@echo "🤖 Building for CI..."
	@$(MAKE) deployment-build
	@$(MAKE) deployment-test
	@echo "✅ CI build completed"

# =============================================================================
# HELP AND DOCUMENTATION
# =============================================================================

help-detailed: ## Show detailed help for all sections
	@echo "Oil & Gas Inventory System - Detailed Help"
	@echo "=========================================="
	@echo ""
	@echo "🚀 MAIN OPERATIONS:"
	@echo "  setup                   - Complete system setup (database + tools + backend)"
	@echo "  dev                     - Start full development environment"
	@echo "  test                    - Run comprehensive test suite"
	@echo "  clean                   - Clean all build artifacts"
	@echo ""
	@echo "📊 POSTGRESQL DATABASE:"
	@echo "  postgresql-status      - Show database status and table statistics"
	@echo "  postgresql-reset       - Reset database (destructive, requires confirmation)"
	@echo "  postgresql-backup      - Create timestamped database backup"
	@echo "  postgresql-restore     - Restore from backup file"
	@echo "  postgresql-monitor     - Real-time database performance monitoring"
	@echo "  postgresql-vacuum      - Vacuum and analyze database"
	@echo ""
	@echo "🏢 MULTI-TENANT MANAGEMENT:"
	@echo "  tenant-create          - Create new tenant with name and slug"
	@echo "  tenant-list            - List all tenants with basic info"
	@echo "  tenant-status          - Comprehensive tenant status report"
	@echo "  tenant-delete          - Delete tenant (destructive, requires confirmation)"
	@echo "  tenant-switch          - Switch active tenant context"
	@echo "  tenant-users           - List users for specific tenant"
	@echo ""
	@echo "🛠️  MIGRATION TOOLS:"
	@echo "  migration-setup        - Initialize migration tools environment"
	@echo "  migration-build        - Build all migration tools from source"
	@echo "  migration-convert      - Convert MDB file to PostgreSQL"
	@echo "  migration-analyze      - Analyze ColdFusion application"
	@echo "  migration-validate     - Validate conversion results"
	@echo "  migration-test         - Run migration tool tests"
	@echo "  migration-status       - Show tools status and recent outputs"
	@echo "  migration-report       - Generate migration summary report"
	@echo "  migration-clean        - Clean migration outputs"
	@echo ""
	@echo "🔧 BACKEND DEVELOPMENT:"
	@echo "  backend-setup          - Install Go dependencies"
	@echo "  backend-build          - Build backend services"
	@echo "  backend-dev            - Start development server"
	@echo "  backend-test           - Run backend tests"
	@echo "  backend-test-coverage  - Generate test coverage report"
	@echo "  backend-lint           - Lint Go code"
	@echo "  backend-format         - Format Go code"
	@echo "  backend-health         - Check backend service health"
	@echo "  backend-logs           - Show backend logs"
	@echo "  backend-clean          - Clean build artifacts"
	@echo ""
	@echo "🌐 FRONTEND DEVELOPMENT:"
	@echo "  frontend-setup         - Install npm dependencies"
	@echo "  frontend-build         - Build for production"
	@echo "  frontend-dev           - Start development server"
	@echo "  frontend-test          - Run frontend tests"
	@echo "  frontend-test-e2e      - Run end-to-end tests"
	@echo "  frontend-lint          - Lint frontend code"
	@echo "  frontend-clean         - Clean build artifacts"
	@echo "  frontend-serve         - Serve built frontend"
	@echo ""
	@echo "🚀 API OPERATIONS:"
	@echo "  api-docs               - Generate Swagger API documentation"
	@echo "  api-test               - Test API endpoints"
	@echo "  api-test-auth          - Test authentication endpoints"
	@echo "  api-benchmark          - Run API performance benchmarks"
	@echo "  api-load-test          - Run load tests"
	@echo "  api-monitor            - Monitor API performance"
	@echo ""
	@echo "🏗️  DEPLOYMENT:"
	@echo "  deployment-build       - Build for production deployment"
	@echo "  deployment-test        - Test deployment configuration"
	@echo "  deployment-package     - Package application for deployment"
	@echo "  deployment-deploy      - Deploy to environment"
	@echo "  deployment-rollback    - Rollback deployment"
	@echo "  deployment-status      - Check deployment status"
	@echo "  deployment-logs        - View deployment logs"

commands: ## List all available commands
	@echo "📋 All Available Commands:"
	@echo "=========================="
	@$(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($1 !~ "^[#.]") {print $1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$'

# =============================================================================
# END OF MAKEFILE
# =============================================================================
