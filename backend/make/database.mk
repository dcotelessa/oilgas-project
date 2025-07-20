# Database Operations
.PHONY: migrate seed migrate-reset db-status backup-db

migrate: ## Run database migrations
	@echo "$(GREEN)Running migrations for $(ENV)...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set. Check your .env.local file.$(RESET)"; \
		exit 1; \
	fi
	go run migrator.go migrate $(ENV)
	@echo "$(GREEN)✅ Migrations completed$(RESET)"

seed: ## Seed database with test data
	@echo "$(GREEN)Seeding database for $(ENV)...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set. Check your .env.local file.$(RESET)"; \
		exit 1; \
	fi
	go run migrator.go seed $(ENV)
	@echo "$(GREEN)✅ Database seeded$(RESET)"

migrate-reset: ## Reset database (DESTRUCTIVE)
	@echo "$(RED)⚠️  Resetting database for $(ENV)...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@read -p "Are you sure? This will delete all data! (y/N): " confirm && [ "$$confirm" = "y" ]
	go run migrator.go reset $(ENV)
	@echo "$(GREEN)✅ Database reset. Run 'make migrate && make seed' to restore$(RESET)"

db-status: ## Show database status
	@echo "$(GREEN)Database status for $(ENV):$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	go run migrator.go status $(ENV)

backup-db: ## Backup database
	@echo "$(GREEN)Creating database backup...$(RESET)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)❌ DATABASE_URL not set$(RESET)"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	mkdir -p backups; \
	pg_dump "$(DATABASE_URL)" > backups/backup_$$timestamp.sql; \
	echo "$(GREEN)✅ Backup created: backups/backup_$$timestamp.sql$(RESET)"
