# scripts/setup-dev.sh
#!/bin/bash
# Development Environment Setup Script

set -e

echo "🚀 Setting up Oil & Gas Development Environment"

# Check required tools
command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed."; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed."; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed."; exit 1; }

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p backups
mkdir -p logs
mkdir -p bin

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "⚠️  Please update .env with your preferred settings"
fi

# Start databases
echo "🗄️  Starting databases..."
make db-up

# Wait for databases to be ready
echo "⏳ Waiting for databases to initialize..."
sleep 20

# Check database status
echo "🔍 Checking database status..."
make db-status

# Run migrations
echo "🚀 Running migrations..."
make migrate-auth-up
make migrate-lb-up

# Seed development data
echo "🌱 Seeding development data..."
make dev-seed

echo "✅ Development environment setup complete!"
echo ""
echo "🎯 Next steps:"
echo "  1. Update .env file with your settings"
echo "  2. Run 'make app-run' to start the application"
echo "  3. Visit http://localhost:8080 to access the API"
echo "  4. Visit http://localhost:8081 for PgAdmin (optional)"
echo ""
echo "📚 Available commands:"
echo "  make help           - Show all available commands"
echo "  make db-status      - Check database status"
echo "  make db-shell-auth  - Access auth database shell"
echo "  make db-shell-longbeach - Access Long Beach database shell"

