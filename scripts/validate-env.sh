#!/bin/bash
# scripts/validate-env.sh
# Validate environment variable configuration for two-database setup

set -e

echo "🔍 Validating environment configuration for two-database setup..."

# Load .env if it exists
if [ -f .env ]; then
    source .env
    echo "✅ Found and loaded .env file"
else
    echo "❌ .env file not found. Please copy .env.example to .env"
    exit 1
fi

# Function to check if variable is set
check_var() {
    local var_name=$1
    local var_value=${!var_name}
    
    if [ -z "$var_value" ]; then
        echo "❌ $var_name is not set"
        return 1
    else
        echo "✅ $var_name is set"
        return 0
    fi
}

# Function to check database URL format
check_db_url() {
    local var_name=$1
    local var_value=${!var_name}
    
    if [[ $var_value =~ ^postgresql://[^:]+:[^@]+@[^:/]+:[0-9]+/[^?]+(\?.+)?$ ]]; then
        echo "✅ $var_name has valid format"
        return 0
    else
        echo "❌ $var_name has invalid format: $var_value"
        return 1
    fi
}

echo ""
echo "📋 Checking required environment variables..."

# Check required variables
ERRORS=0

# Auth database variables
check_var "AUTH_DB_USER" || ERRORS=$((ERRORS + 1))
check_var "AUTH_DB_PASSWORD" || ERRORS=$((ERRORS + 1))
check_var "AUTH_DB_PORT" || ERRORS=$((ERRORS + 1))
check_var "CENTRAL_AUTH_DB_URL" || ERRORS=$((ERRORS + 1))

# Long Beach database variables
check_var "LONGBEACH_DB_USER" || ERRORS=$((ERRORS + 1))
check_var "LONGBEACH_DB_PASSWORD" || ERRORS=$((ERRORS + 1))
check_var "LONGBEACH_DB_PORT" || ERRORS=$((ERRORS + 1))
check_var "LONGBEACH_DB_URL" || ERRORS=$((ERRORS + 1))

# Application variables
check_var "APP_PORT" || ERRORS=$((ERRORS + 1))
check_var "JWT_SECRET" || ERRORS=$((ERRORS + 1))
check_var "DEFAULT_TENANT" || ERRORS=$((ERRORS + 1))

echo ""
echo "🔗 Checking database URL formats..."

# Check database URL formats
if [ ! -z "$CENTRAL_AUTH_DB_URL" ]; then
    check_db_url "CENTRAL_AUTH_DB_URL" || ERRORS=$((ERRORS + 1))
fi

if [ ! -z "$LONGBEACH_DB_URL" ]; then
    check_db_url "LONGBEACH_DB_URL" || ERRORS=$((ERRORS + 1))
fi

echo ""
echo "🚨 Checking for legacy configurations..."

# Check for legacy DATABASE_URL
if [ ! -z "$DATABASE_URL" ]; then
    echo "⚠️  Legacy DATABASE_URL found - consider removing this in favor of specific database URLs"
fi

# Check for mismatched ports
if [ "$AUTH_DB_PORT" == "$LONGBEACH_DB_PORT" ]; then
    echo "❌ AUTH_DB_PORT and LONGBEACH_DB_PORT should be different (got $AUTH_DB_PORT)"
    ERRORS=$((ERRORS + 1))
fi

echo ""
if [ $ERRORS -eq 0 ]; then
    echo "🎉 Environment configuration validation passed!"
    echo "✅ Two-database setup is properly configured"
else
    echo "💥 Environment configuration validation failed with $ERRORS errors"
    echo "Please fix the issues above before proceeding"
    exit 1
fi