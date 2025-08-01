# Docker and Infrastructure Commands (Dual Database Setup)
.PHONY: docker-up docker-down docker-logs docker-reset docker-status

docker-up: ## 🐳 Start both PostgreSQL databases
	@echo "🐳 Starting PostgreSQL databases..."
	@docker-compose up -d postgres-local postgres-test
	@echo "⏳ Waiting for databases to be ready..."
	@sleep 5
	@echo "✅ Databases ready:"
	@echo "  📍 Local:  localhost:5433 (oilgas_inventory_local)"
	@echo "  📍 Test:   localhost:5434 (oilgas_inventory_test)"

docker-down: ## 🛑 Stop all Docker services
	@echo "🛑 Stopping Docker services..."
	@docker-compose down
	@echo "✅ Docker services stopped"

docker-logs: ## 📋 Show database logs
	@echo "📋 Local database logs:"
	@docker-compose logs postgres-local | tail -20
	@echo ""
	@echo "📋 Test database logs:"
	@docker-compose logs postgres-test | tail -20

docker-status: ## 📊 Show container status
	@echo "📊 Docker container status:"
	@docker-compose ps

docker-reset: ## 🔄 Reset both databases (destructive!)
	@echo "⚠️  This will destroy all data in both databases!"
	@read -p "Are you sure? [y/N] " -n 1 -r; echo; \
	if [[ $REPLY =~ ^[Yy]$ ]]; then \
		docker-compose down; \
		docker volume rm $(docker volume ls -q | grep postgres) 2>/dev/null || true; \
		docker-compose up -d postgres-local postgres-test; \
		sleep 5; \
		echo "✅ Databases reset complete"; \
	else \
		echo "❌ Reset cancelled"; \
	fi
