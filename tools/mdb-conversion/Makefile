# tools/mdb-conversion/Makefile
.PHONY: help setup status clean build test
.DEFAULT_GOAL := help

# Paths
TOOLS_ROOT := $(shell pwd)
PROJECT_ROOT := $(shell cd ../.. && pwd)
OUTPUT_DIR := $(PROJECT_ROOT)/output
BUILD_DIR := $(TOOLS_ROOT)/build
BIN_DIR := $(BUILD_DIR)/bin

help: ## Show available commands
	@echo "🛠️  MDB Conversion Tools"
	@echo "========================"
	@echo ""
	@echo "COMMANDS:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "EXAMPLES:"
	@echo "  make setup                    # Initial setup"
	@echo "  make status                   # Check status"
	@echo "  make build                    # Build tools"
	@echo "  make test                     # Test tools"

setup: setup-dirs build ## Complete setup
	@echo "✅ MDB conversion tools setup complete!"

setup-dirs: ## Create necessary directories
	@echo "📁 Creating directories..."
	@mkdir -p $(BUILD_DIR) $(BIN_DIR) $(OUTPUT_DIR)/conversion
	@echo "✅ Directories created"

build: ## Build Go tools
	@echo "🔨 Building tools..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/mdb_processor ./cmd/mdb_processor.go
	@go build -o $(BIN_DIR)/cf_analyzer ./cmd/cf_analyzer.go
	@go build -o $(BIN_DIR)/tester ./cmd/tester.go
	@echo "✅ Tools built"

status: ## Show tool status
	@echo "📊 MDB Conversion Tools Status"
	@echo "=============================="
	@echo "Tools Root: $(TOOLS_ROOT)"
	@echo "Project Root: $(PROJECT_ROOT)"
	@echo ""
	@echo "🔧 System Dependencies:"
	@if command -v go >/dev/null 2>&1; then \
		echo "  ✅ Go: $$(go version | awk '{print $$3}')"; \
	else \
		echo "  ❌ Go: not found"; \
	fi
	@if command -v mdb-ver >/dev/null 2>&1; then \
		echo "  ✅ mdb-tools: available"; \
	else \
		echo "  ⚠️  mdb-tools: not found (install: brew install mdb-tools)"; \
	fi
	@echo ""
	@echo "🔨 Built Tools:"
	@for tool in mdb_processor cf_analyzer tester; do \
		if [ -f "$(BIN_DIR)/$$tool" ]; then \
			echo "  ✅ $$tool: ready"; \
		else \
			echo "  ❌ $$tool: not built (run: make build)"; \
		fi; \
	done
	@echo ""
	@echo "📁 Directories:"
	@for dir in "$(BUILD_DIR)" "$(OUTPUT_DIR)" "config" "test"; do \
		if [ -d "$$dir" ]; then \
			echo "  ✅ $$dir: exists"; \
		else \
			echo "  ❌ $$dir: missing"; \
		fi; \
	done

test: build ## Test the tools
	@echo "🧪 Testing tools..."
	@$(BIN_DIR)/mdb_processor test
	@$(BIN_DIR)/cf_analyzer test  
	@$(BIN_DIR)/tester basic
	@echo "✅ All tests passed"

convert-mdb: ## Convert MDB file (placeholder - usage: make convert-mdb FILE=database.mdb)
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Usage: make convert-mdb FILE=path/to/database.mdb"; \
		exit 1; \
	fi
	@echo "🔄 Converting MDB: $(FILE)"
	@$(BIN_DIR)/mdb_processor convert "$(FILE)"

analyze-cf: ## Analyze ColdFusion app (placeholder - usage: make analyze-cf DIR=cf_app)
	@if [ -z "$(DIR)" ]; then \
		echo "❌ Usage: make analyze-cf DIR=path/to/cf_app"; \
		exit 1; \
	fi
	@echo "🔍 Analyzing CF: $(DIR)"
	@$(BIN_DIR)/cf_analyzer analyze "$(DIR)"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

install-deps: ## Install system dependencies
	@echo "🔧 Installing dependencies..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install mdb-tools || echo "⚠️  mdb-tools installation failed or already installed"; \
	elif command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y mdb-tools || echo "⚠️  apt installation failed"; \
	else \
		echo "⚠️  Please install mdb-tools manually"; \
	fi

# Quick development commands
quick-test: ## Quick validation test
	@echo "⚡ Quick test..."
	@$(MAKE) build
	@$(MAKE) test
	@echo "✅ Quick test complete"

dev-status: ## Development status check
	@echo "👨‍💻 Development Status:"
	@echo "Go module: $$(if [ -f go.mod ]; then echo '✅'; else echo '❌'; fi)"
	@echo "Tools built: $$(ls $(BIN_DIR) 2>/dev/null | wc -l || echo 0)/3"
	@echo "Config files: $$(ls config/*.json 2>/dev/null | wc -l || echo 0)"
