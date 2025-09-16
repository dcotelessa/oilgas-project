# scripts/migrate-from-access.sh
#!/bin/bash
# Customer Migration from Access Database

set -e

echo "🔄 Migrating customers from Access database"

# Check if export file is provided
if [ -z "$1" ]; then
    echo "❌ Usage: $0 <path-to-customer-export.json>"
    echo "   Export your Access customers table to JSON format first"
    exit 1
fi

EXPORT_FILE="$1"

# Validate export file exists
if [ ! -f "$EXPORT_FILE" ]; then
    echo "❌ Export file not found: $EXPORT_FILE"
    exit 1
fi

# Check if databases are running
echo "🔍 Checking database status..."
if ! make db-status > /dev/null 2>&1; then
    echo "❌ Databases are not running. Starting them..."
    make db-up
    sleep 10
fi

# Backup existing data
echo "💾 Creating backup of existing data..."
make backup-longbeach

# Run migration
echo "🚀 Running customer migration..."
go run cmd/tools/migrate-customers/main.go \
    --tenant=longbeach \
    --file="$EXPORT_FILE" \
    --dry-run=false \
    --batch-size=10

echo "✅ Customer migration completed!"
echo ""
echo "📊 To verify migration:"
echo "  make db-shell-longbeach"
echo "  SELECT COUNT(*) FROM customers;"

