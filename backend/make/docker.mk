# Docker and Infrastructure
.PHONY: docker-up docker-down docker-logs

docker-up: ## Start PostgreSQL with Docker Compose
	@echo "$(GREEN)Starting PostgreSQL...$(RESET)"
	docker-compose up -d postgres
	@echo "$(GREEN)✅ PostgreSQL started on localhost:5432$(RESET)"

docker-down: ## Stop all Docker services
	@echo "$(YELLOW)Stopping Docker services...$(RESET)"
	docker-compose down
	@echo "$(GREEN)✅ Docker services stopped$(RESET)"

docker-logs: ## Show Docker logs
	@echo "$(GREEN)Showing PostgreSQL logs...$(RESET)"
	docker-compose logs -f postgres
