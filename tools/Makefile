.PHONY: help build test demo

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-15s %s\n", $$1, $$2}'

build: ## Build the processor
	@echo "🔨 Building..."
	@go mod tidy
	@mkdir -p bin
	@go build -o bin/mdb_processor ./cmd/mdb_processor.go
	@echo "✅ Build complete"

test: build ## Run test
	@echo "🧪 Testing..."
	@mkdir -p output
	@./bin/mdb_processor -file test/fixtures/basic_test.csv -company "Test" -output output -verbose
	@echo "✅ Test complete"

demo: test ## Run demo
	@echo "🎯 Demo complete! Check output/ directory"

clean: ## Clean artifacts
	@rm -rf bin output
