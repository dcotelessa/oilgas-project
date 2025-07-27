# =============================================================================
# MONITORING MODULE - make/monitoring.mk
# =============================================================================
# System monitoring, metrics, and performance tracking

.PHONY: monitor-setup monitor-connections monitor-events monitor-api

# =============================================================================
# MONITORING SETUP
# =============================================================================

monitor-setup: ## 🛠️  Setup monitoring
	@echo "$(GREEN)📊 Setting up monitoring...$(RESET)"
	@mkdir -p logs/monitoring data/metrics
	@echo "$(GREEN)✅ Monitoring ready$(RESET)"

# =============================================================================
# CONNECTION MONITORING
# =============================================================================

monitor-connections: ## 📊 Monitor database connections
	@echo "$(BLUE)📊 Database Connection Monitoring$(RESET)"
	@echo "=================================="
	@go run cmd/tools/monitor-connections.go

monitor-pool-stats: ## 📊 Show connection pool statistics
	@echo "$(BLUE)📊 Connection Pool Statistics$(RESET)"
	@echo "==============================="
	@go run cmd/tools/connection-stats.go

monitor-tenant-usage: ## 📊 Monitor tenant resource usage
	@echo "$(BLUE)📊 Tenant Resource Usage$(RESET)"
	@echo "========================="
	@go run cmd/tools/tenant-usage.go

# =============================================================================
# API MONITORING
# =============================================================================

monitor-api: ## 📊 Monitor API performance
	@echo "$(BLUE)📊 API Performance Monitoring$(RESET)"
	@echo "=============================="
	@go run cmd/tools/monitor-api.go

monitor-response-times: ## 📊 Show API response times
	@echo "$(BLUE)📊 API Response Times$(RESET)"
	@echo "====================="
	@go run cmd/tools/response-times.go

monitor-error-rates: ## 📊 Show API error rates
	@echo "$(BLUE)📊 API Error Rates$(RESET)"
	@echo "=================="
	@go run cmd/tools/error-rates.go

# =============================================================================
# SYSTEM HEALTH
# =============================================================================

monitor-health: ## 📊 Complete system health check
	@echo "$(BLUE)📊 System Health Check$(RESET)"
	@echo "======================="
	@go run cmd/tools/health-check.go

monitor-resources: ## 📊 Monitor system resources
	@echo "$(BLUE)📊 System Resources$(RESET)"
	@echo "==================="
	@echo "Memory usage:"
	@ps aux | grep "go\|api-server" | head -5
	@echo ""
	@echo "Disk usage:"
	@df -h | grep -E "(data|logs)"

# =============================================================================
# CLEANUP
# =============================================================================

monitor-clean: ## 🛠️  Clean monitoring artifacts
	@echo "$(YELLOW)🧹 Cleaning monitoring artifacts...$(RESET)"
	@rm -rf logs/monitoring/*.log
	@rm -rf data/metrics/*
	@echo "$(GREEN)✅ Monitoring cleanup complete$(RESET)"

help-monitoring: ## 📖 Show monitoring commands help
	@echo "$(BLUE)Monitoring Module Commands$(RESET)"
	@echo "=========================="
	@echo ""
	@echo "$(GREEN)📊 CONNECTION MONITORING:$(RESET)"
	@echo "  monitor-connections   - Database connections"
	@echo "  monitor-pool-stats    - Connection pool stats"
	@echo "  monitor-tenant-usage  - Tenant resource usage"
	@echo ""
	@echo "$(YELLOW)📊 API MONITORING:$(RESET)"
	@echo "  monitor-api           - API performance"
	@echo "  monitor-response-times - Response times"
	@echo "  monitor-error-rates   - Error rates"
	@echo ""
	@echo "$(BLUE)📊 SYSTEM HEALTH:$(RESET)"
	@echo "  monitor-health        - Complete health check"
	@echo "  monitor-resources     - System resources"

