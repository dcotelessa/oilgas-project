# Docker Infrastructure Management
.PHONY: docker-up docker-up-admin docker-down docker-clean docker-logs docker-status docker-restart

docker-up: ## Start both local and test databases
	@echo "$(GREEN)Starting PostgreSQL databases...$(RESET)"
	@docker-compose up -d postgres-local postgres-test --remove-orphans
	@echo "$(BLUE)Waiting for databases to be ready...$(RESET)"
	@sleep 5
	@docker-compose exec postgres-local pg_isready -U postgres || echo "Local DB starting..."
	@docker-compose exec postgres-test pg_isready -U postgres || echo "Test DB starting..."
	@echo "$(GREEN)✅ Databases started$(RESET)"
	@echo "$(BLUE)Local DB: localhost:5433$(RESET)"
	@echo "$(BLUE)Test DB:  localhost:5434$(RESET)"

docker-up-admin: ## Start databases + PgAdmin
	@echo "$(GREEN)Starting PostgreSQL databases with PgAdmin...$(RESET)"
	@docker-compose --profile admin up -d
	@echo "$(BLUE)Waiting for services to be ready...$(RESET)"
	@sleep 8
	@docker-compose exec postgres-local pg_isready -U postgres || echo "Local DB starting..."
	@docker-compose exec postgres-test pg_isready -U postgres || echo "Test DB starting..."
	@echo "$(GREEN)✅ All services started$(RESET)"
	@echo "$(BLUE)Local DB:  localhost:5433$(RESET)"
	@echo "$(BLUE)Test DB:   localhost:5434$(RESET)"
	@echo "$(BLUE)PgAdmin:   http://localhost:8080$(RESET)"
	@echo "$(YELLOW)PgAdmin Login: admin@oilgas.local / admin123$(RESET)"

docker-logs: ## Show database logs
	@echo "$(GREEN)Showing database logs (Ctrl+C to exit)...$(RESET)"
	@docker-compose logs -f postgres-local postgres-test

docker-status: ## Show status of all containers
	@echo "$(GREEN)Docker Container Status:$(RESET)"
	@docker-compose ps
	@echo ""
	@echo "$(BLUE)Database Health Check:$(RESET)"
	@docker-compose exec postgres-local pg_isready -U postgres -d oilgas_inventory_local 2>/dev/null && echo "✅ Local DB: Ready" || echo "❌ Local DB: Not ready"
	@docker-compose exec postgres-test pg_isready -U postgres -d oilgas_inventory_test 2>/dev/null && echo "✅ Test DB: Ready" || echo "❌ Test DB: Not ready"

docker-restart: ## Restart both databases
	@echo "$(YELLOW)Restarting PostgreSQL databases...$(RESET)"
	@docker-compose restart postgres-local postgres-test
	@sleep 3
	@make docker-status

docker-down: ## Stop all services
	@echo "$(YELLOW)Stopping all services...$(RESET)"
	@docker-compose down
	@echo "$(GREEN)✅ All services stopped$(RESET)"

docker-clean: ## Remove all data (destructive)
	@echo "$(RED)⚠️  This will destroy all database data!$(RESET)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ]
	@echo "$(YELLOW)Removing all containers and data...$(RESET)"
	@docker-compose down -v
	@docker system prune -f
	@echo "$(GREEN)✅ All data removed$(RESET)"

# Database-specific Docker commands
docker-local-only: ## Start only local database
	@echo "$(GREEN)Starting local database only...$(RESET)"
	@docker-compose up -d postgres-local
	@sleep 3
	@docker-compose exec postgres-local pg_isready -U postgres || echo "Local DB starting..."

docker-test-only: ## Start only test database
	@echo "$(GREEN)Starting test database only...$(RESET)"
	@docker-compose up -d postgres-test
	@sleep 3
	@docker-compose exec postgres-test pg_isready -U postgres || echo "Test DB starting..."

# Debugging and maintenance
docker-shell-local: ## Open shell in local database container
	@echo "$(GREEN)Opening shell in local database container...$(RESET)"
	@docker-compose exec postgres-local bash

docker-shell-test: ## Open shell in test database container
	@echo "$(GREEN)Opening shell in test database container...$(RESET)"
	@docker-compose exec postgres-test bash

docker-backup-local: ## Backup local database
	@echo "$(GREEN)Creating backup of local database...$(RESET)"
	@mkdir -p backups
	@docker-compose exec postgres-local pg_dump -U postgres oilgas_inventory_local > backups/local_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✅ Backup created in backups/ directory$(RESET)"

docker-restore-local: ## Restore local database from backup
	@echo "$(GREEN)Available backups:$(RESET)"
	@ls -la backups/*.sql 2>/dev/null || echo "No backups found"
	@read -p "Enter backup filename: " backup && \
	docker-compose exec -T postgres-local psql -U postgres -d oilgas_inventory_local < backups/$$backup

# Network and volume management
docker-network-info: ## Show Docker network information
	@echo "$(GREEN)Docker Network Information:$(RESET)"
	@docker network ls | grep oilgas || echo "No oilgas networks found"
	@echo ""
	@echo "$(GREEN)Container Network Details:$(RESET)"
	@docker-compose exec postgres-local hostname -I 2>/dev/null || echo "Local DB not running"
	@docker-compose exec postgres-test hostname -I 2>/dev/null || echo "Test DB not running"

docker-volume-info: ## Show Docker volume information
	@echo "$(GREEN)Docker Volume Information:$(RESET)"
	@docker volume ls | grep postgres || echo "No postgres volumes found"
	@echo ""
	@echo "$(GREEN)Volume Usage:$(RESET)"
	@docker system df

# Quick development workflows
docker-dev-start: docker-up ## Start development environment
	@echo "$(GREEN)🚀 Development environment starting...$(RESET)"
	@echo "$(BLUE)Waiting for databases to be fully ready...$(RESET)"
	@sleep 2
	@make docker-status

docker-test-start: docker-up ## Start testing environment
	@echo "$(GREEN)🧪 Testing environment starting...$(RESET)"
	@echo "$(BLUE)Waiting for databases to be fully ready...$(RESET)"
	@sleep 2
	@make docker-status
