# =============================================================================
# EVENTS MODULE - make/events.mk
# =============================================================================
# Event system management for cross-tenant coordination

.PHONY: events-setup events-start events-stop events-test events-monitor

# =============================================================================
# EVENT SYSTEM LIFECYCLE
# =============================================================================

events-setup: ## 🛠️  Setup event system (in-memory)
	@echo "$(GREEN)📡 Setting up event system...$(RESET)"
	@echo "$(BLUE)Phase 1: In-memory event bus$(RESET)"
	@echo "$(YELLOW)💡 NATS upgrade available in Week 3$(RESET)"
	@mkdir -p logs/events
	@echo "$(GREEN)✅ Event system ready$(RESET)"

events-start: ## 🛠️  Start event system
	@echo "$(YELLOW)📡 Starting event system...$(RESET)"
	@echo "$(BLUE)💡 In-memory event bus - no external dependencies$(RESET)"
	@echo "$(GREEN)✅ Event system active$(RESET)"

events-stop: ## 🛠️  Stop event system
	@echo "$(YELLOW)⏹️  Stopping event system...$(RESET)"
	@echo "$(GREEN)✅ Event system stopped$(RESET)"

# =============================================================================
# EVENT MONITORING
# =============================================================================

events-monitor: ## 📊 Monitor event activity
	@echo "$(BLUE)📊 Event System Monitoring$(RESET)"
	@echo "==========================="
	@go run cmd/tools/monitor-events.go

events-stats: ## 📊 Show event statistics
	@echo "$(BLUE)📊 Event Statistics$(RESET)"
	@echo "==================="
	@go run cmd/tools/event-stats.go

events-history: ## 📊 Show recent event history
	@echo "$(BLUE)📊 Recent Event History$(RESET)"
	@echo "======================="
	@read -p "Tenant ID (or 'all'): " tenant && \
	read -p "Hours to show (default 1): " hours && \
	hours=$${hours:-1} && \
	go run cmd/tools/event-history.go --tenant="$$tenant" --hours="$$hours"

# =============================================================================
# EVENT TESTING
# =============================================================================

events-test: ## 🧪 Run event system tests
	@echo "$(YELLOW)🧪 Running event system tests...$(RESET)"
	@go test -v ./internal/events/...
	@echo "$(GREEN)✅ Event tests complete$(RESET)"

events-test-publish: ## 🧪 Test event publishing
	@echo "$(YELLOW)🧪 Testing event publishing...$(RESET)"
	@read -p "Tenant ID: " tenant && \
	read -p "Event type (test/work_order/inventory): " event_type && \
	go run cmd/tools/test-publish.go \
		--tenant="$$tenant" \
		--type="$$event_type" \
		--data='{"test": true, "timestamp": "'$$(date -Iseconds)'"}' && \
	echo "$(GREEN)✅ Event published$(RESET)"

events-test-subscribe: ## 🧪 Test event subscription
	@echo "$(YELLOW)🧪 Testing event subscription...$(RESET)"
	@echo "$(BLUE)Listening for events (Ctrl+C to stop)...$(RESET)"
	@go run cmd/tools/test-subscribe.go

# =============================================================================
# NATS INTEGRATION (Week 3)
# =============================================================================

events-nats-setup: ## ⚡ Setup NATS event system
	@echo "$(YELLOW)⚡ Setting up NATS event system...$(RESET)"
	@if ! command -v nats-server >/dev/null 2>&1; then \
		echo "$(BLUE)📦 Installing NATS server...$(RESET)"; \
		go install github.com/nats-io/nats-server/v2@latest; \
	fi
	@mkdir -p data/nats logs/nats
	@echo "$(GREEN)✅ NATS event system ready$(RESET)"

events-nats-start: ## ⚡ Start NATS server
	@echo "$(YELLOW)🚀 Starting NATS server...$(RESET)"
	@nats-server \
		--store_dir=./data/nats \
		--log_file=./logs/nats/server.log \
		--pid_file=./tmp/nats.pid \
		--jetstream \
		--daemon && \
	echo "$(GREEN)✅ NATS server started$(RESET)" || \
	echo "$(RED)❌ Failed to start NATS server$(RESET)"

events-nats-stop: ## ⚡ Stop NATS server
	@echo "$(YELLOW)⏹️  Stopping NATS server...$(RESET)"
	@if [ -f ./tmp/nats.pid ]; then \
		kill $$(cat ./tmp/nats.pid) && \
		rm -f ./tmp/nats.pid && \
		echo "$(GREEN)✅ NATS server stopped$(RESET)"; \
	else \
		echo "$(YELLOW)💡 NATS server not running$(RESET)"; \
	fi

events-nats-status: ## ⚡ Check NATS server status
	@echo "$(BLUE)📊 NATS Server Status$(RESET)"
	@echo "====================="
	@if command -v nats >/dev/null 2>&1; then \
		nats server info; \
	else \
		echo "$(YELLOW)💡 Install NATS CLI: go install github.com/nats-io/natscli/nats@latest$(RESET)"; \
		curl -s http://localhost:8222/varz | jq . 2>/dev/null || echo "NATS server not accessible"; \
	fi

events-upgrade-nats: ## ⚡ Upgrade to NATS event system
	@echo "$(YELLOW)⚡ Upgrading to NATS event system...$(RESET)"
	@$(MAKE) events-nats-setup
	@go run cmd/tools/upgrade-events-nats.go
	@echo "$(GREEN)✅ Upgraded to NATS event system$(RESET)"
	@echo "$(BLUE)💡 Restart API server to use NATS events$(RESET)"

# =============================================================================
# EVENT STREAMS (NATS JetStream)
# =============================================================================

events-streams-create: ## ⚡ Create event streams
	@echo "$(YELLOW)📡 Creating event streams...$(RESET)"
	@go run cmd/tools/create-streams.go
	@echo "$(GREEN)✅ Event streams created$(RESET)"

events-streams-list: ## ⚡ List event streams
	@echo "$(BLUE)📡 Event Streams$(RESET)"
	@echo "================"
	@if command -v nats >/dev/null 2>&1; then \
		nats stream list; \
	else \
		go run cmd/tools/list-streams.go; \
	fi

events-streams-info: ## ⚡ Show stream information
	@echo "$(BLUE)📡 Stream Information$(RESET)"
	@echo "===================="
	@read -p "Stream name: " stream && \
	if command -v nats >/dev/null 2>&1; then \
		nats stream info "$$stream"; \
	else \
		go run cmd/tools/stream-info.go --stream="$$stream"; \
	fi

# =============================================================================
# CROSS-TENANT ANALYTICS
# =============================================================================

events-analytics: ## 📊 Cross-tenant analytics dashboard
	@echo "$(BLUE)📊 Cross-Tenant Analytics$(RESET)"
	@echo "=========================="
	@go run cmd/tools/analytics-dashboard.go

events-kpis: ## 📊 Show business KPIs
	@echo "$(BLUE)📊 Business KPIs$(RESET)"
	@echo "================"
	@go run cmd/tools/business-kpis.go

events-tenant-activity: ## 📊 Show tenant activity summary
	@echo "$(BLUE)📊 Tenant Activity Summary$(RESET)"
	@echo "=========================="
	@go run cmd/tools/tenant-activity.go

# =============================================================================
# EVENT DEBUGGING
# =============================================================================

events-debug: ## 🐛 Debug event system
	@echo "$(BLUE)🐛 Event System Debug$(RESET)"
	@echo "======================"
	@echo "Event bus status:"
	@go run cmd/tools/debug-events.go
	@echo ""
	@echo "Recent events:"
	@tail -20 logs/events/app.log 2>/dev/null || echo "No event logs found"

events-replay: ## 🔄 Replay events for tenant
	@echo "$(YELLOW)🔄 Replaying events...$(RESET)"
	@read -p "Tenant ID: " tenant && \
	read -p "Start time (YYYY-MM-DD HH:MM): " start_time && \
	read -p "End time (YYYY-MM-DD HH:MM): " end_time && \
	go run cmd/tools/replay-events.go \
		--tenant="$$tenant" \
		--start="$$start_time" \
		--end="$$end_time" && \
	echo "$(GREEN)✅ Events replayed$(RESET)"

# =============================================================================
# CLEANUP
# =============================================================================

events-clean: ## 🛠️  Clean event artifacts
	@echo "$(YELLOW)🧹 Cleaning event artifacts...$(RESET)"
	@rm -rf logs/events/*.log
	@rm -rf data/nats/*
	@rm -rf tmp/nats.pid
	@echo "$(GREEN)✅ Event cleanup complete$(RESET)"

events-clean-streams: ## 🧹 Clean NATS streams (DANGEROUS)
	@echo "$(RED)⚠️  WARNING: This will delete all event streams!$(RESET)"
	@read -p "Are you sure? Type 'DELETE STREAMS' to confirm: " confirm && \
	[ "$$confirm" = "DELETE STREAMS" ] && \
	go run cmd/tools/clean-streams.go && \
	echo "$(GREEN)✅ Event streams cleaned$(RESET)"

# =============================================================================
# HELP
# =============================================================================

help-events: ## 📖 Show events commands help
	@echo "$(BLUE)Events Module Commands$(RESET)"
	@echo "======================="
	@echo ""
	@echo "$(GREEN)🛠️  LIFECYCLE:$(RESET)"
	@echo "  events-setup          - Setup event system (in-memory)"
	@echo "  events-start          - Start event system"
	@echo "  events-stop           - Stop event system"
	@echo ""
	@echo "$(BLUE)📊 MONITORING:$(RESET)"
	@echo "  events-monitor        - Monitor event activity"
	@echo "  events-stats          - Show event statistics"
	@echo "  events-history        - Show recent event history"
	@echo ""
	@echo "$(YELLOW)🧪 TESTING:$(RESET)"
	@echo "  events-test           - Run event system tests"
	@echo "  events-test-publish   - Test event publishing"
	@echo "  events-test-subscribe - Test event subscription"
	@echo ""
	@echo "$(RED)⚡ NATS INTEGRATION:$(RESET)"
	@echo "  events-nats-setup     - Setup NATS event system"
	@echo "  events-nats-start     - Start NATS server"
	@echo "  events-nats-stop      - Stop NATS server"
	@echo "  events-nats-status    - Check NATS status"
	@echo "  events-upgrade-nats   - Upgrade to NATS"
	@echo ""
	@echo "$(GREEN)📡 STREAMS:$(RESET)"
	@echo "  events-streams-create - Create event streams"
	@echo "  events-streams-list   - List event streams"
	@echo "  events-streams-info   - Show stream information"
	@echo ""
	@echo "$(BLUE)📈 ANALYTICS:$(RESET)"
	@echo "  events-analytics      - Cross-tenant analytics"
	@echo "  events-kpis           - Show business KPIs"
	@echo "  events-tenant-activity - Tenant activity summary"
