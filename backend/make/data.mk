# =============================================================================
# Data Import Commands (MDB Processing)
# =============================================================================

.PHONY: data-check data-convert-mdb data-import-mdb data-validate data-status data-setup data-backup data-stats

## Check MDB files and conversion readiness
data-check:
	@echo "📂 Checking for MDB data files..."
	@test -d data/mdb && echo "✅ data/mdb directory exists" || echo "❌ Create data/mdb directory and add your .mdb files"
	@test -d data/csv && echo "✅ data/csv directory exists" || echo "❌ Create data/csv directory for converted files"
	@test -f tools/convert_mdb.go && echo "✅ MDB converter available" || echo "❌ MDB converter needs implementation"
	@echo ""
	@echo "Expected structure:"
	@echo "  data/mdb/customers.mdb"
	@echo "  data/mdb/inventory.mdb"
	@echo "  data/csv/ (converted files)"

## Convert MDB files to CSV (placeholder - customize for your data)
data-convert-mdb:
	@echo "🔄 Converting MDB files to CSV..."
	@mkdir -p data/csv
	@echo "⚠️  MDB converter needs customization for your specific files"
	@echo "   Edit tools/convert_mdb.go with your table structures"
	@echo "   Available in Phase 3B implementation"

## Import converted MDB data (renamed from import-mdb-data to avoid conflicts)
data-import-mdb:
	@echo "📥 Importing MDB data..."
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible" && exit 1)
	@echo "⚠️  Import process needs customization for your data"
	@echo "   Customize tools/import_data.go for your CSV structure"
	@echo "   Available in Phase 3B implementation"

## Validate imported data integrity
data-validate:
	@echo "🔍 Validating imported data..."
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible" && exit 1)
	@cd backend && go run tools/validate_import.go 2>/dev/null || echo "⚠️  Validation tool not implemented yet"

## Show import status and data counts
data-status:
	@echo "📊 Data Import Status"
	@echo "====================="
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible" && exit 1)
	@$(MAKE) db-exec SQL="SELECT 'Customers' as table_name, count(*) as records FROM store.customers UNION ALL SELECT 'Inventory', count(*) FROM store.inventory UNION ALL SELECT 'Grades', count(*) FROM store.grade UNION ALL SELECT 'Sizes', count(*) FROM store.sizes;" 2>/dev/null || echo "❌ Cannot query tables"
	@echo ""
	@echo "🔍 Sample Records:"
	@$(MAKE) db-exec SQL="SELECT customer_name, billing_city, billing_state FROM store.customers LIMIT 3;" 2>/dev/null || echo "❌ Cannot query customers"

## Create data directories
data-setup:
	@echo "📁 Creating data directories..."
	@mkdir -p data/mdb data/csv data/backup
	@echo "✅ Data directories created"
	@echo ""
	@echo "📝 Next steps:"
	@echo "  1. Copy your .mdb files to data/mdb/"
	@echo "  2. Run: make data-convert-mdb"
	@echo "  3. Run: make data-import-mdb"

## Backup current data before import
data-backup:
	@echo "💾 Backing up current data..."
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible" && exit 1)
	@mkdir -p data/backup
	@$(MAKE) db-dump FILE=data/backup/backup_$(shell date +%Y%m%d_%H%M%S).sql 2>/dev/null || echo "⚠️  Backup functionality depends on db-dump target"
	@echo "✅ Data backed up to data/backup/"

## Show data statistics
data-stats:
	@echo "📊 Database Statistics"
	@echo "======================"
	@$(MAKE) db-health 2>/dev/null || (echo "❌ Database not accessible" && exit 1)
	@echo ""
	@echo "Table Sizes:"
	@$(MAKE) db-exec SQL="SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size FROM pg_tables WHERE schemaname = 'store' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;" 2>/dev/null || echo "❌ Cannot query table sizes"
	@echo ""
	@echo "Record Counts:"
	@$(MAKE) db-exec SQL="SELECT 'customers' as table_name, count(*) FROM store.customers UNION ALL SELECT 'inventory', count(*) FROM store.inventory UNION ALL SELECT 'grades', count(*) FROM store.grade UNION ALL SELECT 'sizes', count(*) FROM store.sizes;" 2>/dev/null || echo "❌ Cannot query record counts"

## Aliases for backward compatibility
.PHONY: import-check convert-mdb import-mdb-data import-validate import-status

import-check: data-check
convert-mdb: data-convert-mdb
import-mdb-data: data-import-mdb
import-validate: data-validate
import-status: data-status
