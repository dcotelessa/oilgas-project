# =============================================================================
# DATABASE MODULE - make/database.mk
# =============================================================================
# Database operations, migrations, and tenant management

.PHONY: db-setup db-start db-stop db-migrate db-seed db-reset db-status

# =============================================================================
# DATABASE LIFECYCLE
# =============================================================================

db-setup: db-health db-migrate ## 🛠️  Complete database setup
	@echo "$(GREEN)✅ Database setup complete$(RESET)"

db-start: ## 🛠️  Start database (if using Docker)
	@echo "$(YELLOW)🚀 Starting database...$(RESET)"
	@if [ -f docker-compose.yml ]; then \
		docker-compose up -d postgres; \
		echo "$(GREEN)✅ Database started$(RESET)"; \
	else \
		echo "$(BLUE)💡 Using external database connection$(RESET)"; \
	fi

db-stop: ## 🛠️  Stop database (if using Docker)
	@echo "$(YELLOW)⏹️  Stopping database...$(RESET)"
	@if [ -f docker-compose.yml ]; then \
		docker-compose stop postgres; \
		echo "$(GREEN)✅ Database stopped$(RESET)"; \
	else \
		echo "$(BLUE)💡 External database - no action needed$(RESET)"; \
	fi

# =============================================================================
# MIGRATIONS
# =============================================================================

db-migrate: db-health ## 🛠️  Run database migrations
	@echo "$(YELLOW)📦 Running database migrations...$(RESET)"
	@go run cmd/migrator/main.go migrate $(ENV)
	@echo "$(GREEN)✅ Migrations complete$(RESET)"

db-migrate-status: ## 🛠️  Check migration status
	@echo "$(BLUE)📋 Migration Status$(RESET)"
	@go run cmd/migrator/main.go status $(ENV)

db-migrate-reset: ## ⚠️  Reset database (DANGEROUS)
	@echo "$(RED)⚠️  WARNING: This will delete ALL data!$(RESET)"
	@read -p "Are you sure? Type 'DELETE ALL DATA' to confirm: " confirm && \
	[ "$$confirm" = "DELETE ALL DATA" ] && \
	go run cmd/migrator/main.go reset $(ENV) && \
	echo "$(GREEN)✅ Database reset complete$(RESET)"
	#

# =============================================================================
# MIGRATION GENERATION AND EXECUTION
# =============================================================================

db-generate-migrations: ## 🔄 Generate migration SQL files
	@echo "$(YELLOW)🔄 Generating migration files...$(RESET)"
	@go run cmd/migrator/main.go generate $(ENV)
	@echo "$(GREEN)✅ Migration files generated in migrations/$(RESET)"
	@echo "$(BLUE)📁 Generated files:$(RESET)"
	@ls -la migrations/*.sql 2>/dev/null || echo "No files generated"

db-show-migrations: ## 📄 Show generated migration content
	@echo "$(BLUE)📄 Generated Migration Files$(RESET)"
	@echo "============================"
	@for file in migrations/*.sql; do \
		if [ -f "$file" ]; then \
			echo "$(YELLOW)📄 $file:$(RESET)"; \
			echo "---"; \
			head -20 "$file"; \
			echo "..."; \
			echo ""; \
		fi; \
	done

db-migrate-from-files: ## 📦 Execute migrations from generated SQL files
	@echo "$(YELLOW)📦 Executing migrations from files...$(RESET)"
	@if [ ! -f "migrations/001_store_schema.sql" ]; then \
		echo "$(RED)❌ Migration files not found. Run 'make db-generate-migrations' first$(RESET)"; \
		exit 1; \
	fi
	@for file in migrations/001_store_schema.sql migrations/002_auth_schema.sql migrations/003_seed_data.sql; do \
		if [ -f "$file" ]; then \
			echo "$(YELLOW)📦 Executing: $file$(RESET)"; \
			psql "$DATABASE_URL" -f "$file" && echo "$(GREEN)✅ Completed: $file$(RESET)"; \
		else \
			echo "$(YELLOW)⚠️  Skipping missing file: $file$(RESET)"; \
		fi; \
	done
	@echo "$(GREEN)✅ All migrations executed$(RESET)"

db-migrate-enhanced: db-generate-migrations db-migrate-from-files ## 🚀 Complete migration (generate + execute)
	@echo "$(GREEN)🚀 Enhanced migration complete!$(RESET)"
	@echo "$(BLUE)Migration files are available for review in migrations/$(RESET)"

# =============================================================================
# MIGRATION VALIDATION
# =============================================================================

db-validate-migrations: ## 🔍 Validate generated migration SQL
	@echo "$(YELLOW)🔍 Validating migration SQL syntax...$(RESET)"
	@for file in migrations/*.sql; do \
		if [ -f "$file" ]; then \
			echo "$(BLUE)Checking: $file$(RESET)"; \
			psql "$DATABASE_URL" --dry-run -f "$file" 2>/dev/null && \
				echo "$(GREEN)✅ Valid: $file$(RESET)" || \
				echo "$(RED)❌ Invalid: $file$(RESET)"; \
		fi; \
	done

db-preview-changes: ## 👀 Preview what migrations will do
	@echo "$(BLUE)👀 Migration Preview$(RESET)"
	@echo "===================="
	@echo "$(YELLOW)The following changes will be applied:$(RESET)"
	@echo ""
	@echo "$(BLUE)1. Store Schema (001_store_schema.sql):$(RESET)"
	@echo "   • Create store and migrations schemas"
	@echo "   • Create customers, grade, sizes, inventory, received tables"
	@echo "   • Add indexes for performance"
	@echo ""
	@echo "$(BLUE)2. Auth Schema (002_auth_schema.sql):$(RESET)"
	@echo "   • Create auth schema"
	@echo "   • Create tenants, users, sessions tables"
	@echo "   • Add authentication indexes and constraints"
	@echo "   • Insert default system tenant"
	@echo ""
	@echo "$(BLUE)3. Seed Data (003_seed_data.sql):$(RESET)"
	@echo "   • Insert standard oil & gas grades (J55, K55, N80, L80, P110, etc.)"
	@echo "   • Insert standard pipe sizes (4 1/2\", 5 1/2\", 7\", 9 5/8\", etc.)"
	@echo ""
	@echo "$(GREEN)Run 'make db-migrate-enhanced' to apply these changes$(RESET)"

# =============================================================================
# MIGRATION ROLLBACK (FUTURE)
# =============================================================================

db-generate-rollback: ## 🔄 Generate rollback SQL files (future)
	@echo "$(YELLOW)🔄 Generating rollback files...$(RESET)"
	@mkdir -p migrations/rollback
	@echo "-- Rollback for 003_seed_data.sql" > migrations/rollback/003_rollback.sql
	@echo "DELETE FROM store.sizes WHERE size IN ('4 1/2\"', '5\"', '5 1/2\"', '7\"', '8 5/8\"', '9 5/8\"', '10 3/4\"', '13 3/8\"');" >> migrations/rollback/003_rollback.sql
	@echo "DELETE FROM store.grade WHERE grade IN ('J55', 'K55', 'N80', 'L80', 'P110', 'C90', 'T95');" >> migrations/rollback/003_rollback.sql
	@echo ""
	@echo "-- Rollback for 002_auth_schema.sql" > migrations/rollback/002_rollback.sql  
	@echo "DROP SCHEMA IF EXISTS auth CASCADE;" >> migrations/rollback/002_rollback.sql
	@echo ""
	@echo "-- Rollback for 001_store_schema.sql" > migrations/rollback/001_rollback.sql
	@echo "DROP SCHEMA IF EXISTS store CASCADE;" >> migrations/rollback/001_rollback.sql
	@echo "DROP SCHEMA IF EXISTS migrations CASCADE;" >> migrations/rollback/001_rollback.sql
	@echo "$(GREEN)✅ Rollback files generated in migrations/rollback/$(RESET)"

# =============================================================================
# HELP ADDITIONS
# =============================================================================

help-database-migrations: ## 📖 Show migration commands help
	@echo "$(BLUE)Database Migration Commands$(RESET)"
	@echo "==========================="
	@echo ""
	@echo "$(GREEN)🔄 GENERATION:$(RESET)"
	@echo "  db-generate-migrations  - Generate migration SQL files"
	@echo "  db-show-migrations      - Show generated migration content"
	@echo "  db-generate-rollback    - Generate rollback SQL files"
	@echo ""
	@echo "$(YELLOW)📦 EXECUTION:$(RESET)"
	@echo "  db-migrate-from-files   - Execute migrations from SQL files"
	@echo "  db-migrate-enhanced     - Complete migration (generate + execute)"
	@echo ""
	@echo "$(BLUE)🔍 VALIDATION:$(RESET)"
	@echo "  db-validate-migrations  - Validate migration SQL syntax"
	@echo "  db-preview-changes      - Preview what migrations will do"
	@echo ""
	@echo "$(GREEN)Workflow:$(RESET)"
	@echo "  1. make db-preview-changes     # See what will happen"
	@echo "  2. make db-migrate-enhanced    # Generate + execute"
	@echo "  3. make db-show-migrations     # Review generated files"

# =============================================================================
# TENANT MANAGEMENT
# =============================================================================

db-tenant-create: ## 🏢 Create new tenant database
	@read -p "Tenant ID (lowercase, no spaces): " tenant && \
	echo "$(YELLOW)🏢 Creating tenant database: $$tenant$(RESET)" && \
	go run cmd/migrator/main.go tenant-create $$tenant && \
	echo "$(GREEN)✅ Tenant $$tenant created$(RESET)"

db-tenant-list: ## 🏢 List all tenant databases
	@echo "$(BLUE)🏢 Active Tenant Databases$(RESET)"
	@echo "============================="
	@psql "$$DATABASE_URL" -c "\l" | grep oilgas_ | \
		awk '{print $$1}' | sed 's/oilgas_//' | \
		while read tenant; do \
			echo "  📊 $$tenant"; \
		done || echo "$(YELLOW)No tenants found$(RESET)"

db-tenant-backup: ## 🏢 Backup specific tenant
	@read -p "Tenant ID to backup: " tenant && \
	echo "$(YELLOW)💾 Backing up tenant: $$tenant$(RESET)" && \
	mkdir -p backups && \
	pg_dump "$$DATABASE_URL" --schema-only oilgas_$$tenant > \
		"backups/tenant_$$tenant_schema_$$(date +%Y%m%d_%H%M%S).sql" && \
	pg_dump "$$DATABASE_URL" --data-only oilgas_$$tenant > \
		"backups/tenant_$$tenant_data_$$(date +%Y%m%d_%H%M%S).sql" && \
	echo "$(GREEN)✅ Tenant $$tenant backed up to backups/$(RESET)"

db-tenant-restore: ## 🏢 Restore specific tenant
	@echo "$(YELLOW)📥 Available backups:$(RESET)"
	@ls -la backups/ | grep tenant_ | tail -5
	@read -p "Backup file to restore: " backup && \
	read -p "Tenant ID: " tenant && \
	echo "$(YELLOW)📥 Restoring $$backup to tenant $$tenant$(RESET)" && \
	psql "$$DATABASE_URL/oilgas_$$tenant" < "backups/$$backup" && \
	echo "$(GREEN)✅ Tenant $$tenant restored$(RESET)"

# =============================================================================
# DATA OPERATIONS
# =============================================================================

db-seed: ## 🛠️  Seed database with sample data
	@echo "$(YELLOW)🌱 Seeding database...$(RESET)"
	@go run cmd/migrator/main.go seed $(ENV)
	@echo "$(GREEN)✅ Database seeded$(RESET)"

db-import-csv: ## 📥 Import CSV data for tenant
	@read -p "CSV file path: " file && \
	read -p "Tenant ID: " tenant && \
	read -p "Data type (customers/inventory/workorders): " datatype && \
	echo "$(YELLOW)📥 Importing $$file to $$tenant as $$datatype$(RESET)" && \
	go run cmd/tools/csv-importer.go \
		--file="$$file" \
		--tenant="$$tenant" \
		--type="$$datatype" && \
	echo "$(GREEN)✅ Import complete$(RESET)"

db-export-csv: ## 📤 Export tenant data to CSV
	@read -p "Tenant ID: " tenant && \
	read -p "Data type (customers/inventory/workorders): " datatype && \
	echo "$(YELLOW)📤 Exporting $$datatype from $$tenant$(RESET)" && \
	mkdir -p exports && \
	go run cmd/tools/csv-exporter.go \
		--tenant="$$tenant" \
		--type="$$datatype" \
		--output="exports/$$tenant_$$datatype_$$(date +%Y%m%d_%H%M%S).csv" && \
	echo "$(GREEN)✅ Export complete$(RESET)"

# =============================================================================
# MAINTENANCE
# =============================================================================

db-vacuum: ## 🛠️  Vacuum and analyze database
	@echo "$(YELLOW)🧹 Running database maintenance...$(RESET)"
	@psql "$$DATABASE_URL" -c "VACUUM ANALYZE;" && \
	echo "$(GREEN)✅ Database maintenance complete$(RESET)"

db-stats: ## 📊 Show database statistics
	@echo "$(BLUE)📊 Database Statistics$(RESET)"
	@echo "======================"
	@psql "$$DATABASE_URL" -c "\l+" | grep oilgas
	@echo ""
	@echo "Connection stats:"
	@psql "$$DATABASE_URL" -c "SELECT count(*) as connections FROM pg_stat_activity WHERE datname LIKE 'oilgas_%';"

db-connections: ## 📊 Show active connections
	@echo "$(BLUE)📊 Active Database Connections$(RESET)"
	@echo "================================"
	@psql "$$DATABASE_URL" -c "\
		SELECT datname, usename, client_addr, state, query_start \
		FROM pg_stat_activity \
		WHERE datname LIKE 'oilgas_%' \
		ORDER BY query_start DESC;"

db-health: ## 🛠️  Validate database connection
	@echo "$(YELLOW)🔍 Validating database connection...$(RESET)"
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		echo "$(YELLOW)💡 Set in .env.local or export DATABASE_URL=...$(RESET)"; \
		exit 1; \
	fi
	@psql "$$DATABASE_URL" -c "SELECT version();" > /dev/null && \
		echo "$(GREEN)✅ Database connection successful$(RESET)" || \
		(echo "$(RED)❌ Database connection failed$(RESET)" && exit 1)

# =============================================================================
# CLEANUP
# =============================================================================

db-clean: ## 🛠️  Clean database artifacts
	@echo "$(YELLOW)🧹 Cleaning database artifacts...$(RESET)"
	@rm -rf logs/db-*.log
	@echo "$(GREEN)✅ Database cleanup complete$(RESET)"

# =============================================================================
# HELP
# =============================================================================

help-database: ## 📖 Show database commands help
	@echo "$(BLUE)Database Module Commands$(RESET)"
	@echo "========================="
	@echo ""
	@echo "$(GREEN)🛠️  LIFECYCLE:$(RESET)"
	@echo "  db-setup           - Complete database setup"
	@echo "  db-start           - Start database (Docker)"
	@echo "  db-stop            - Stop database (Docker)"
	@echo ""
	@echo "$(YELLOW)📦 MIGRATIONS:$(RESET)"
	@echo "  db-migrate         - Run migrations"
	@echo "  db-migrate-status  - Check migration status"
	@echo "  db-migrate-reset   - Reset database (DANGEROUS)"
	@echo ""
	@echo "$(RED)🏢 TENANT MANAGEMENT:$(RESET)"
	@echo "  db-tenant-create   - Create new tenant"
	@echo "  db-tenant-list     - List all tenants"
	@echo "  db-tenant-backup   - Backup specific tenant"
	@echo "  db-tenant-restore  - Restore specific tenant"
	@echo ""
	@echo "$(BLUE)📊 DATA OPERATIONS:$(RESET)"
	@echo "  db-seed            - Seed with sample data"
	@echo "  db-import-csv      - Import CSV data"
	@echo "  db-export-csv      - Export to CSV"
	@echo ""
	@echo "$(GREEN)🧹 MAINTENANCE:$(RESET)"
	@echo "  db-vacuum          - Database maintenance"
	@echo "  db-stats           - Database statistics"
	@echo "  db-connections     - Active connections"
	@echo "  db-health          - Validate database connection"

# =============================================================================
# DATABASE VALIDATION & TESTING
# =============================================================================

db-test-local: ## 🧪 Test local database setup
	@echo "$(BLUE)🧪 Testing Local Database Setup$(RESET)"
	@echo "================================="
	@echo ""
	@echo "$(YELLOW)1. Testing basic connectivity...$(RESET)"
	@$(MAKE) db-health
	@echo ""
	@echo "$(YELLOW)2. Checking database schema...$(RESET)"
	@psql "$$DATABASE_URL" -c "\dt store.*" 2>/dev/null || echo "$(BLUE)💡 No store schema found yet$(RESET)"
	@echo ""
	@echo "$(YELLOW)3. Checking main tables...$(RESET)"
	@psql "$$DATABASE_URL" -c "\
		SELECT schemaname, tablename, tableowner \
		FROM pg_tables \
		WHERE schemaname = 'store' \
		ORDER BY tablename;" 2>/dev/null || echo "$(BLUE)💡 Store schema not found$(RESET)"

db-test-tables: ## 🧪 Test table record counts
	@echo "$(BLUE)🧪 Table Record Counts$(RESET)"
	@echo "======================"
	@for table in customers inventory received grades sizes; do \
		count=$$(psql "$$DATABASE_URL" -t -c "SELECT COUNT(*) FROM store.$$table;" 2>/dev/null | tr -d ' ' || echo "0"); \
		echo "  📊 store.$$table: $$count records"; \
	done

db-test-tenant: ## 🧪 Test tenant database creation
	@echo "$(BLUE)🧪 Testing Tenant Creation$(RESET)"
	@echo "=========================="
	@read -p "Create test tenant 'testlocal'? (y/N): " confirm && \
	if [ "$$confirm" = "y" ]; then \
		echo "$(YELLOW)Creating test tenant...$(RESET)"; \
		go run cmd/migrator/main.go tenant-create testlocal; \
		echo "$(GREEN)✅ Test tenant created$(RESET)"; \
	else \
		echo "$(BLUE)💡 Skipped tenant creation$(RESET)"; \
	fi

db-test-connections: ## 🧪 Test connection pooling  
	@echo "$(BLUE)🧪 Active Database Connections$(RESET)"
	@psql "$$DATABASE_URL" -c "\
		SELECT \
			datname, \
			count(*) as connections, \
			state, \
			application_name \
		FROM pg_stat_activity \
		WHERE datname LIKE 'oilgas%' OR datname = current_database() \
		GROUP BY datname, state, application_name \
		ORDER BY datname, connections DESC;"

db-connection-limits: ## 📊 Show connection limits and usage
	@echo "$(BLUE)📊 Connection Limits & Usage$(RESET)"
	@echo "=============================="
	@echo "$(YELLOW)Current connections per database:$(RESET)"
	@psql "$$DATABASE_URL" -c "\
		SELECT \
			datname, \
			numbackends as active_connections \
		FROM pg_stat_database \
		WHERE datname LIKE 'oilgas%' OR datname = current_database() \
		ORDER BY numbackends DESC;"
	@echo ""
	@echo "$(YELLOW)Server connection settings:$(RESET)"
	@psql "$$DATABASE_URL" -c "\
		SELECT \
			name, \
			setting, \
			unit, \
			short_desc \
		FROM pg_settings \
		WHERE name IN ('max_connections', 'superuser_reserved_connections', 'shared_preload_libraries') \
		ORDER BY name;"

db-test-sample-data: ## 🧪 Test sample data operations
	@echo "$(BLUE)🧪 Testing Sample Data Operations$(RESET)"
	@echo "=================================="
	@echo "$(YELLOW)Inserting test customer...$(RESET)"
	@psql "$$DATABASE_URL" -c "\
		INSERT INTO store.customers (customer, contact, email, created_at) \
		VALUES ('Test Local Company', 'John Doe', 'test@local.com', NOW()) \
		ON CONFLICT DO NOTHING;" 2>/dev/null && \
		echo "$(GREEN)✅ Test customer inserted$(RESET)" || \
		echo "$(RED)❌ Could not insert test customer$(RESET)"
	@echo ""
	@echo "$(YELLOW)Cleaning up test data...$(RESET)"
	@psql "$$DATABASE_URL" -c "\
		DELETE FROM store.customers \
		WHERE customer = 'Test Local Company' AND email = 'test@local.com';" 2>/dev/null && \
		echo "$(GREEN)✅ Test data cleaned$(RESET)" || \
		echo "$(BLUE)💡 No test data to clean$(RESET)"

db-test-comprehensive: ## 🧪 Run comprehensive database tests
	@echo "$(GREEN)🧪 Comprehensive Database Test Suite$(RESET)"
	@echo "========================================="
	@$(MAKE) db-test-local
	@echo ""
	@$(MAKE) db-test-tables
	@echo ""
	@$(MAKE) db-test-connections
	@echo ""
	@$(MAKE) db-test-sample-data
	@echo ""
	@echo "$(GREEN)🎯 Database Test Summary$(RESET)"
	@echo "========================"
	@connectivity=$$(psql "$$DATABASE_URL" -c "SELECT 'OK';" -t 2>/dev/null | tr -d ' ' || echo 'FAILED'); \
	schema_count=$$(psql "$$DATABASE_URL" -c "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name='store';" -t 2>/dev/null | tr -d ' ' || echo '0'); \
	table_count=$$(psql "$$DATABASE_URL" -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='store';" -t 2>/dev/null | tr -d ' ' || echo '0'); \
	env_status=$$([[ -n "$$DATABASE_URL" ]] && echo 'set' || echo 'missing'); \
	echo "  ✅ Connectivity: $$connectivity"; \
	echo "  ✅ Store schemas: $$schema_count"; \
	echo "  ✅ Store tables: $$table_count"; \
	echo "  ✅ Environment: DATABASE_URL $$env_status"

db-test-clean: ## 🧪 Clean test database
	@echo "$(YELLOW)🧹 Cleaning test database...$(RESET)"
	@export DATABASE_URL="$$DATABASE_URL_TEST" && \
	psql "$$DATABASE_URL_TEST" -c "DROP SCHEMA IF EXISTS store CASCADE;" && \
	echo "$(GREEN)✅ Test database cleaned$(RESET)"

# =============================================================================
# DATABASE INSPECTION
# =============================================================================

db-inspect-schema: ## 🔍 Inspect database schema
	@echo "$(BLUE)🔍 Database Schema Inspection$(RESET)"
	@echo "============================="
	@echo ""
	@echo "$(YELLOW)Available schemas:$(RESET)"
	@psql "$$DATABASE_URL" -c "\dn+"
	@echo ""
	@echo "$(YELLOW)Store schema tables:$(RESET)"
	@psql "$$DATABASE_URL" -c "\dt+ store.*"

db-inspect-tenants: ## 🔍 Inspect tenant databases
	@echo "$(BLUE)🔍 Tenant Database Inspection$(RESET)"
	@echo "============================="
	@echo "$(YELLOW)All oil & gas databases:$(RESET)"
	@psql "$$DATABASE_URL" -c "\l" | grep oilgas || echo "$(BLUE)💡 No tenant databases found$(RESET)"

db-inspect-indexes: ## 🔍 Inspect database indexes
	@echo "$(BLUE)🔍 Database Indexes$(RESET)"
	@echo "=================="
	@psql "$$DATABASE_URL" -c "\
		SELECT \
			schemaname, \
			tablename, \
			indexname, \
			indexdef \
		FROM pg_indexes \
		WHERE schemaname = 'store' \
		ORDER BY tablename, indexname;"

db-inspect-constraints: ## 🔍 Inspect database constraints
	@echo "$(BLUE)🔍 Database Constraints$(RESET)"
	@echo "======================="
	@psql "$$DATABASE_URL" -c "\
		SELECT \
			tc.table_schema, \
			tc.table_name, \
			tc.constraint_name, \
			tc.constraint_type \
		FROM information_schema.table_constraints tc \
		WHERE tc.table_schema = 'store' \
		ORDER BY tc.table_name, tc.constraint_type;"

# =============================================================================
# QUICK DIAGNOSTICS
# =============================================================================

db-quick-check: ## ⚡ Quick database health check
	@echo "$(GREEN)⚡ Quick Database Health Check$(RESET)"
	@echo "=============================="
	@connectivity=$$(psql "$$DATABASE_URL" -c "SELECT 'OK';" -t 2>/dev/null | tr -d ' ' || echo 'FAILED'); \
	if [ "$$connectivity" = "OK" ]; then \
		echo "$(GREEN)✅ Database: Connected$(RESET)"; \
	else \
		echo "$(RED)❌ Database: Connection failed$(RESET)"; \
		exit 1; \
	fi
	@table_count=$$(psql "$$DATABASE_URL" -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='store';" -t 2>/dev/null | tr -d ' ' || echo '0'); \
	echo "$(BLUE)📊 Tables: $$table_count in store schema$(RESET)"
	@tenant_count=$$(psql "$$DATABASE_URL" -c "\l" | grep oilgas | wc -l | tr -d ' '); \
	echo "$(BLUE)🏢 Tenants: $$tenant_count tenant databases$(RESET)"

db-debug-env: ## 🐛 Debug environment variables
	@echo "$(BLUE)🐛 Environment Debug$(RESET)"
	@echo "==================="
	@echo "DATABASE_URL: $${DATABASE_URL:-'❌ Not set'}"
	@echo "DATABASE_URL_TEST: $${DATABASE_URL_TEST:-'❌ Not set'}"
	@echo "APP_ENV: $${APP_ENV:-'❌ Not set'}"
	@echo "API_PORT: $${API_PORT:-'❌ Not set'}"
	@echo ""
	@echo "$(YELLOW)Current directory:$(RESET) $$(pwd)"
	@echo "$(YELLOW).env.local exists:$(RESET) $$([[ -f .env.local ]] && echo '✅ Yes' || echo '❌ No')"

# =============================================================================
# HELP ADDITIONS
# =============================================================================

help-database-testing: ## 📖 Show database testing help
	@echo "$(BLUE)Database Testing Commands$(RESET)"
	@echo "========================="
	@echo ""
	@echo "$(GREEN)🧪 TESTING:$(RESET)"
	@echo "  db-test-local         - Test local database setup"
	@echo "  db-test-tables        - Test table record counts"
	@echo "  db-test-tenant        - Test tenant creation"
	@echo "  db-test-connections   - Test connection pooling"
	@echo "  db-connection-limits  - Show connection limits and usage"
	@echo "  db-test-sample-data   - Test data operations"
	@echo "  db-test-comprehensive - Run all database tests"
	@echo ""
	@echo "$(BLUE)🔍 INSPECTION:$(RESET)"
	@echo "  db-inspect-schema     - Inspect database schema"
	@echo "  db-inspect-tenants    - Inspect tenant databases"
	@echo "  db-inspect-indexes    - Inspect database indexes"
	@echo "  db-inspect-constraints - Inspect constraints"
	@echo ""
	@echo "$(YELLOW)⚡ QUICK CHECKS:$(RESET)"
	@echo "  db-quick-check        - Quick health check"
	@echo "  db-debug-env          - Debug environment variables"
