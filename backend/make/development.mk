# =============================================================================
# Enhanced Development Commands
# =============================================================================

.PHONY: dev-full-setup dev-quick-test dev-reset-with-data dev-api-ready

## Complete development setup (clean slate)
dev-full-setup: docker-down docker-up dev-wait dev-setup api-test
	@echo "🎉 Full development environment ready!"
	@echo ""
	@echo "🚀 Next steps:"
	@echo "  1. Start API: make api-start"
	@echo "  2. Test API: make api-test"
	@echo "  3. Import MDB: make convert-mdb && make import-mdb-data"

## Quick test of entire development stack
dev-quick-test: dev-ensure-db test-unit api-test-quick
	@echo "✅ Development stack working correctly"

## Reset database and reload with fresh data
dev-reset-with-data: dev-db-reset dev-setup import-status
	@echo "🔄 Database reset with fresh sample data"

## Setup development environment optimized for API work
dev-api-ready: dev-ensure-db dev-setup
	@echo "🔧 Development environment ready for API development"
	@echo ""
	@echo "📊 Current Data Status:"
	@$(MAKE) import-status
	@echo ""
	@echo "🚀 Ready to start:"
	@echo "  make api-start     # Start API server"
	@echo "  make api-test      # Test API integration"
	@echo "  make api-dev       # Start with auto-reload"

## Show development environment status
dev-status:
	@echo "🔍 Development Environment Status"
	@echo "================================="
	@echo ""
	@echo "Docker Services:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=oil-gas" 2>/dev/null || echo "  No containers running"
	@echo ""
	@echo "Database Connection:"
	@$(MAKE) db-health || echo "  ❌ Database not accessible"
	@echo ""
	@echo "API Server:"
	@curl -s http://localhost:8000/health > /dev/null && echo "  ✅ API responding on port 8000" || echo "  ❌ API not responding"
	@echo ""
	@echo "Go Environment:"
	@cd backend && go version 2>/dev/null || echo "  ❌ Go not available"
	@echo ""
	@echo "Required Tools:"
	@which jq > /dev/null && echo "  ✅ jq available" || echo "  ❌ jq missing (needed for API testing)"
	@which curl > /dev/null && echo "  ✅ curl available" || echo "  ❌ curl missing"
	@which docker > /dev/null && echo "  ✅ docker available" || echo "  ❌ docker missing"

## Install development dependencies
dev-install-deps:
	@echo "📦 Installing development dependencies..."
	@echo ""
	@echo "Go tools:"
	@cd backend && go install github.com/cosmtrek/air@latest && echo "  ✅ air (auto-reload)"
	@echo ""
	@echo "System tools (you may need sudo):"
	@which jq > /dev/null || echo "  Install jq: brew install jq (macOS) or apt install jq (Ubuntu)"
	@which curl > /dev/null || echo "  Install curl: usually pre-installed"
	@echo ""
	@echo "✅ Development dependencies check complete"

## Create .air.toml for auto-reload (if using air)
dev-create-air-config:
	@echo "🔄 Creating air configuration for auto-reload..."
	@cd backend && if [ ! -f .air.toml ]; then \
		cat > .air.toml << 'EOF' && \
root = "." \
testdata_dir = "testdata" \
tmp_dir = "tmp" \
\
[build] \
args_bin = [] \
bin = "./tmp/main" \
cmd = "go build -o ./tmp/main ./cmd/server" \
delay = 1000 \
exclude_dir = ["assets", "tmp", "vendor", "testdata"] \
exclude_file = [] \
exclude_regex = ["_test.go"] \
exclude_unchanged = false \
follow_symlink = false \
full_bin = "" \
include_dir = [] \
include_ext = ["go", "tpl", "tmpl", "html"] \
kill_delay = "0s" \
log = "build-errors.log" \
send_interrupt = false \
stop_on_root = false \
\
[color] \
app = "" \
build = "yellow" \
main = "magenta" \
runner = "green" \
watcher = "cyan" \
\
[log] \
time = false \
\
[misc] \
clean_on_exit = false \
\
[screen] \
clear_on_rebuild = false \
EOF \
		echo "✅ Air configuration created at backend/.air.toml"; \
	else \
		echo "ℹ️  Air configuration already exists"; \
	fi
	@echo "💡 Use 'make api-dev' to start with auto-reload"

## Verify air configuration and structure
dev-check-air:
	@echo "🔍 Checking Air Configuration"
	@echo "=============================="
	@echo ""
	@echo "Project Structure:"
	@test -f go.mod && echo "  ✅ go.mod in root" || echo "  ❌ go.mod not in root"
	@test -f backend/go.mod && echo "  ✅ go.mod in backend/" || echo "  ❌ go.mod not in backend/"
	@echo ""
	@echo "Go Main File:"
	@test -f cmd/server/main.go && echo "  ✅ main.go at root/cmd/server/" || echo "  ❌ main.go not at root/cmd/server/"
	@test -f backend/cmd/server/main.go && echo "  ✅ main.go at backend/cmd/server/" || echo "  ❌ main.go not at backend/cmd/server/"
	@echo ""
	@echo "Air Installation:"
	@which air > /dev/null && echo "  ✅ air installed" || echo "  ❌ air not installed (run: go install github.com/cosmtrek/air@latest)"
	@echo ""
	@echo "Air Configuration:"
	@test -f .air.toml && echo "  ✅ .air.toml in root" || echo "  ❌ .air.toml not in root"
	@test -f backend/.air.toml && echo "  ✅ .air.toml in backend/" || echo "  ❌ .air.toml not in backend/"
	@echo ""
	@echo "🎯 Recommendation:"
	@if [ -f backend/go.mod ]; then \
		echo "  Structure: backend/go.mod detected"; \
		echo "  Air config should be: backend/.air.toml"; \
		echo "  Run: make dev-create-air-config"; \
	elif [ -f go.mod ]; then \
		echo "  Structure: root/go.mod detected"; \
		echo "  Air config should be: .air.toml"; \
		echo "  Update make/api.mk to remove 'cd backend &&'"; \
	else \
		echo "  ❌ No go.mod found - check your Go setup"; \
	fi

## Fix air configuration based on project structure
dev-fix-air:
	@echo "🔧 Fixing Air Configuration"
	@echo "============================"
	@if [ -f backend/go.mod ]; then \
		echo "Backend structure detected - creating backend/.air.toml"; \
		$(MAKE) dev-create-air-config; \
	elif [ -f go.mod ]; then \
		echo "Root structure detected - creating .air.toml"; \
		cat > .air.toml << 'EOF' && \
root = "." \
tmp_dir = "tmp" \
\
[build] \
bin = "./tmp/main" \
cmd = "go build -o ./tmp/main ./cmd/server" \
delay = 1000 \
exclude_dir = ["assets", "tmp", "vendor", "testdata"] \
include_ext = ["go", "tpl", "tmpl", "html"] \
\
[color] \
build = "yellow" \
main = "magenta" \
runner = "green" \
\
[misc] \
clean_on_exit = false \
EOF \
		echo "✅ Air configuration updated"; \
	else \
		echo "❌ No go.mod found - cannot create air config"; \
	fi
