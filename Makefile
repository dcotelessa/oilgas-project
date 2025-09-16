# Makefile for Oil & Gas Multi-Database Setup

# Load environment variables
include .env
export

# Default target
.PHONY: help
help:
	@echo "Oil & Gas Multi-Database Management"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-up              Start both databases"
	@echo "  db-down            Stop both databases"
	@echo "  db-reset           Reset both databases (DESTRUCTIVE)"
	@echo "  db-logs            Show database logs"
	@echo "  db-status          Show database status"
	@echo ""
	@echo "Migration Commands:"
	@echo "  migrate-auth-up    Run auth database migrations"
	@echo "  migrate-lb-up      Run Long Beach migrations"
	@echo "  migrate-status     Show migration status"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev-setup          Complete development setup"
	@echo "  dev-seed           Seed databases with test data"
	@echo "  dev-clean          Clean development environment"
	@echo ""
	@echo "Application Commands:"
	@echo "  app-build          Build application"
	@echo "  app-run            Run application"
	@echo "  app-test           Run tests"

# Database Management
.PHONY: db-up
db-up:
	@echo "Starting databases..."
	docker-compose up -d auth-db longbeach-db
	@echo "Waiting for databases to be ready..."
	@sleep 10
	@docker-compose exec auth-db pg_isready -U $(AUTH_DB_USER) -d auth_central || echo "Auth DB not ready yet..."
	@docker-compose exec longbeach-db pg_isready -U $(LONGBEACH_DB_USER) -d location_longbeach || echo "Long Beach DB not ready yet..."

.PHONY: db-down
db-down:
	@echo "Stopping databases..."
	docker-compose down

.PHONY: db-reset
db-reset:
	@echo "WARNING: This will destroy all data!"
	@read -p "Are you sure? (y/N) " confirm && [ "$confirm" = "y" ] || exit 1
	docker-compose down -v
	docker volume rm $(shell docker volume ls -q | grep -E "(auth_db_data|longbeach_db_data)") 2>/dev/null || true
	$(MAKE) db-up

.PHONY: db-logs
db-logs:
	docker-compose logs -f auth-db longbeach-db

.PHONY: db-status
db-status:
	@echo "=== Database Status ==="
	@echo "Auth Database:"
	@docker-compose exec auth-db pg_isready -U $(AUTH_DB_USER) -d auth_central && echo "✓ Ready" || echo "✗ Not Ready"
	@echo "Long Beach Database:"
	@docker-compose exec longbeach-db pg_isready -U $(LONGBEACH_DB_USER) -d location_longbeach && echo "✓ Ready" || echo "✗ Not Ready"

# Migration Management
.PHONY: migrate-auth-up
migrate-auth-up:
	@echo "Running auth database migrations..."
	cd database/scripts && go run migrate.go auth up

.PHONY: migrate-lb-up
migrate-lb-up:
	@echo "Running Long Beach database migrations..."
	cd database/scripts && go run migrate.go longbeach up

.PHONY: migrate-auth-down
migrate-auth-down:
	@echo "Rolling back auth database migrations..."
	cd database/scripts && go run migrate.go auth down

.PHONY: migrate-lb-down
migrate-lb-down:
	@echo "Rolling back Long Beach database migrations..."
	cd database/scripts && go run migrate.go longbeach down

.PHONY: migrate-status
migrate-status:
	@echo "=== Migration Status ==="
	@echo "Auth Database Migrations:"
	@docker-compose exec auth-db psql -U $(AUTH_DB_USER) -d auth_central -c "SELECT version, applied_at FROM schema_migrations ORDER BY version;" 2>/dev/null || echo "No migrations table found"
	@echo ""
	@echo "Long Beach Database Migrations:"
	@docker-compose exec longbeach-db psql -U $(LONGBEACH_DB_USER) -d location_longbeach -c "SELECT version, applied_at FROM schema_migrations ORDER BY version;" 2>/dev/null || echo "No migrations table found"

# Development Environment
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	$(MAKE) db-up
	@sleep 15
	@echo "Databases should be ready. Checking status..."
	$(MAKE) db-status
	@echo "Running migrations..."
	$(MAKE) migrate-auth-up
	$(MAKE) migrate-lb-up
	@echo "Seeding test data..."
	$(MAKE) dev-seed
	@echo "Development environment ready!"

.PHONY: dev-seed
dev-seed:
	@echo "Seeding development data..."
	@echo "Creating test customers..."
	cd backend && go run cmd/tools/seed/main.go --tenant=longbeach --customers=10
	@echo "Development data seeded!"

.PHONY: dev-clean
dev-clean:
	@echo "Cleaning development environment..."
	docker-compose down -v
	docker system prune -f
	@echo "Development environment cleaned!"

# Application Management
.PHONY: app-build
app-build:
	@echo "Building application..."
	mkdir -p bin
	cd backend && go build -o ../bin/oil-gas-app cmd/longbeach/main.go

.PHONY: app-run
app-run: app-build
	@echo "Starting application..."
	./bin/oil-gas-app

.PHONY: app-test
app-test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: app-test-integration
app-test-integration:
	@echo "Running integration tests..."
	$(MAKE) db-up
	@sleep 10
	go test -v ./test/integration/...

# Database Access
.PHONY: db-shell-auth
db-shell-auth:
	docker-compose exec auth-db psql -U $(AUTH_DB_USER) -d auth_central

.PHONY: db-shell-longbeach
db-shell-longbeach:
	docker-compose exec longbeach-db psql -U $(LONGBEACH_DB_USER) -d location_longbeach

# Backup and Restore
.PHONY: backup-auth
backup-auth:
	@echo "Backing up auth database..."
	docker-compose exec auth-db pg_dump -U $(AUTH_DB_USER) -d auth_central > backups/auth_$(shell date +%Y%m%d_%H%M%S).sql

.PHONY: backup-longbeach
backup-longbeach:
	@echo "Backing up Long Beach database..."
	docker-compose exec longbeach-db pg_dump -U $(LONGBEACH_DB_USER) -d location_longbeach > backups/longbeach_$(shell date +%Y%m%d_%H%M%S).sql

.PHONY: backup-all
backup-all: backup-auth backup-longbeach

# Customer Migration from Access
.PHONY: migrate-customers
migrate-customers:
	@echo "Migrating customers from Access database export..."
	@read -p "Enter path to customer export file: " filepath && \
	go run cmd/tools/migrate-customers/main.go --tenant=longbeach --file="$filepath"

# Production Commands
.PHONY: prod-deploy
prod-deploy:
	@echo "Deploying to production..."
	@echo "This would run production deployment scripts"

.PHONY: prod-backup
prod-backup:
	@echo "Creating production backup..."
	@echo "This would run production backup scripts"
