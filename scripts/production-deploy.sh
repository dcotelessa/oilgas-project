# scripts/production-deploy.sh
#!/bin/bash
# Production Deployment Script

set -e

echo "🚀 Deploying Oil & Gas System to Production"

# Configuration
ENVIRONMENT=${1:-production}
VERSION=${2:-latest}

if [ "$ENVIRONMENT" != "production" ] && [ "$ENVIRONMENT" != "staging" ]; then
    echo "❌ Invalid environment. Use: production or staging"
    exit 1
fi

echo "📋 Deployment Configuration:"
echo "  Environment: $ENVIRONMENT"
echo "  Version: $VERSION"
echo ""

# Pre-deployment checks
echo "🔍 Pre-deployment checks..."

# Check if required environment variables are set
required_vars=(
    "CENTRAL_AUTH_DB_URL"
    "LONGBEACH_DB_URL"  
    "JWT_SECRET"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required environment variable $var is not set"
        exit 1
    fi
done

# Database backup
echo "💾 Creating database backup..."
mkdir -p backups/pre-deploy
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
make backup-all

# Run database migrations
echo "🔄 Running database migrations..."
make migrate-auth-up
make migrate-lb-up

# Build application
echo "🔨 Building application..."
docker build -t oil-gas-app:$VERSION .

# Health check function
health_check() {
    local url=$1
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url/health" > /dev/null; then
            echo "✅ Health check passed"
            return 0
        fi
        echo "⏳ Attempt $attempt/$max_attempts - waiting for service..."
        sleep 10
        ((attempt++))
    done
    
    echo "❌ Health check failed after $max_attempts attempts"
    return 1
}

# Deploy based on environment
if [ "$ENVIRONMENT" = "production" ]; then
    echo "🏭 Production deployment..."
    
    # Stop old containers gracefully
    docker-compose -f docker-compose.prod.yml down --timeout 30
    
    # Start new containers
    docker-compose -f docker-compose.prod.yml up -d
    
    # Health check
    health_check "http://localhost:8080"
    
elif [ "$ENVIRONMENT" = "staging" ]; then
    echo "🧪 Staging deployment..."
    
    # Stop old containers
    docker-compose -f docker-compose.staging.yml down
    
    # Start new containers
    docker-compose -f docker-compose.staging.yml up -d
    
    # Health check
    health_check "http://localhost:8080"
fi

echo ""
echo "✅ Deployment completed successfully!"
echo "🔗 Application URL: http://localhost:8080"
echo "📊 PgAdmin URL: http://localhost:8081"
echo "📝 Logs: docker-compose logs -f app"

