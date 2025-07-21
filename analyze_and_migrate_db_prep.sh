#!/bin/bash
# analyze_and_migrate_db_prep.sh
# Analyze existing db_prep setup and migrate to modern tools structure

set -e

PROJECT_ROOT="$(pwd)"
DB_PREP_DIR="$PROJECT_ROOT/db_prep"
TOOLS_DIR="$PROJECT_ROOT/tools/mdb-conversion"

echo "🔍 Analyzing db_prep folder and migration strategy"
echo "=================================================="

# Check if db_prep exists
if [ ! -d "$DB_PREP_DIR" ]; then
    echo "❌ db_prep directory not found"
    echo "✅ No migration needed - you're already using the modern structure"
    exit 0
fi

echo "📁 Found db_prep directory: $DB_PREP_DIR"

# Analyze what's in db_prep
echo ""
echo "📊 Contents of db_prep:"
echo "======================"
ls -la "$DB_PREP_DIR" || echo "Cannot list contents"

# Look for specific file types
echo ""
echo "📋 File Analysis:"
echo "================="

mdb_count=$(find "$DB_PREP_DIR" -name "*.mdb" 2>/dev/null | wc -l)
cf_dirs=$(find "$DB_PREP_DIR" -type d -name "*coldfusion*" -o -name "*cf*" 2>/dev/null | wc -l)
cf_files=$(find "$DB_PREP_DIR" -name "*.cfm" -o -name "*.cfc" 2>/dev/null | wc -l)
script_count=$(find "$DB_PREP_DIR" -name "*.sh" 2>/dev/null | wc -l)

echo "🗄️  MDB files: $mdb_count"
echo "📂 CF directories: $cf_dirs"  
echo "📄 CF files: $cf_files"
echo "📜 Shell scripts: $script_count"

# Analyze break.sh specifically
echo ""
echo "🔍 Analyzing break.sh:"
echo "====================="

if [ -f "$DB_PREP_DIR/break.sh" ]; then
    echo "✅ Found break.sh"
    echo ""
    echo "📝 File size: $(wc -l < "$DB_PREP_DIR/break.sh") lines"
    echo ""
    echo "🎯 Script purpose analysis:"
    
    # Look for common patterns to understand what break.sh does
    if grep -qi "mdb" "$DB_PREP_DIR/break.sh"; then
        echo "  📊 Contains MDB-related operations"
    fi
    
    if grep -qi "csv\|export" "$DB_PREP_DIR/break.sh"; then
        echo "  📄 Contains CSV export operations"
    fi
    
    if grep -qi "sql\|database" "$DB_PREP_DIR/break.sh"; then
        echo "  🗄️  Contains database operations"
    fi
    
    if grep -qi "coldfusion\|cfm\|cfc" "$DB_PREP_DIR/break.sh"; then
        echo "  🔍 Contains ColdFusion analysis"
    fi
    
    if grep -qi "clean\|normalize\|format" "$DB_PREP_DIR/break.sh"; then
        echo "  🧹 Contains data cleaning operations"
    fi
    
    echo ""
    echo "📋 Key commands found in break.sh:"
    # Extract key command patterns
    grep -E "^[[:space:]]*[a-zA-Z].*\$" "$DB_PREP_DIR/break.sh" | head -10 | sed 's/^/  /'
    
    echo ""
    echo "🔗 Dependencies found:"
    # Look for external tool dependencies
    grep -oE "(mdb-[a-z]+|psql|mysql|csvkit)" "$DB_PREP_DIR/break.sh" | sort | uniq | sed 's/^/  /'
    
else
    echo "❌ break.sh not found in db_prep"
fi

# Check for project references to db_prep
echo ""
echo "🔗 Checking project references to db_prep:"
echo "=========================================="

ref_count=$(grep -r "db_prep" . --exclude-dir=.git --exclude-dir=db_prep 2>/dev/null | wc -l)
if [ "$ref_count" -gt 0 ]; then
    echo "⚠️  Found $ref_count references to db_prep in project:"
    grep -r "db_prep" . --exclude-dir=.git --exclude-dir=db_prep 2>/dev/null | head -5
    echo "  (showing first 5 references)"
else
    echo "✅ No external references to db_prep found"
fi

# Generate migration recommendations
echo ""
echo "🎯 Migration Recommendations:"
echo "============================="

if [ ! -d "$TOOLS_DIR" ]; then
    echo "❌ Tools directory not found - run tools setup first"
    exit 1
fi

echo "✅ Modern tools structure exists at: $TOOLS_DIR"

# Check if tools can handle the data
echo ""
echo "📈 Migration Strategy:"

if [ "$mdb_count" -gt 0 ]; then
    echo "  📊 MDB files: Migrate to tools/mdb-conversion/input/"
fi

if [ "$cf_files" -gt 0 ] || [ "$cf_dirs" -gt 0 ]; then
    echo "  🔍 ColdFusion files: Migrate to tools/mdb-conversion/input/cf_apps/"
fi

if [ -f "$DB_PREP_DIR/break.sh" ]; then
    echo "  📜 break.sh: Analyze for reusable logic"
fi

# Create migration plan
echo ""
echo "🚀 Recommended Migration Steps:"
echo "==============================="

echo "1. 📁 Create input directories in tools:"
echo "   mkdir -p tools/mdb-conversion/input/{mdb_files,cf_apps,archives}"

echo ""
echo "2. 🔒 Update .gitignore to protect sensitive data:"
echo "   Add input/ directories to tools/.gitignore"

echo ""
echo "3. 📊 Migrate data files:"
if [ "$mdb_count" -gt 0 ]; then
    echo "   mv db_prep/*.mdb tools/mdb-conversion/input/mdb_files/"
fi
if [ "$cf_dirs" -gt 0 ] || [ "$cf_files" -gt 0 ]; then
    echo "   mv db_prep/coldfusion* tools/mdb-conversion/input/cf_apps/"
fi

echo ""
echo "4. 🔍 Analyze break.sh for reusable logic:"
echo "   Review break.sh contents (shown above)"
echo "   Extract any custom business rules"
echo "   Add to tools/mdb-conversion/config/ if needed"

echo ""
echo "5. 🧪 Test new workflow:"
echo "   make tools-setup"
echo "   make convert-mdb FILE=tools/mdb-conversion/input/mdb_files/yourfile.mdb"
echo "   make analyze-cf DIR=tools/mdb-conversion/input/cf_apps/yourapp"

echo ""
echo "6. 🗑️  Remove db_prep after verification:"
echo "   mv db_prep db_prep.backup.$(date +%Y%m%d)"
echo "   # Test everything works, then: rm -rf db_prep.backup.*"

# Offer to run the migration
echo ""
echo "🤖 Automated Migration Options:"
echo "==============================="

cat << 'EOF'
Would you like to:
1. Show break.sh contents for manual review
2. Create the new input structure  
3. Run automated migration (with backup)
4. Just create .gitignore updates
5. Exit and migrate manually

EOF

read -p "Choose option (1-5): " choice

case $choice in
    1)
        echo ""
        echo "📜 Contents of break.sh:"
        echo "========================"
        cat "$DB_PREP_DIR/break.sh"
        ;;
    2)
        create_input_structure
        ;;
    3)
        run_automated_migration
        ;;
    4)
        update_gitignore_only
        ;;
    5)
        echo "✅ Exiting - migrate manually using the steps above"
        ;;
    *)
        echo "❌ Invalid choice"
        ;;
esac

create_input_structure() {
    echo ""
    echo "📁 Creating input directory structure..."
    
    mkdir -p "$TOOLS_DIR/input"/{mdb_files,cf_apps,archives,temp}
    
    echo "✅ Created:"
    echo "  📊 $TOOLS_DIR/input/mdb_files/     # Place .mdb files here"
    echo "  🔍 $TOOLS_DIR/input/cf_apps/       # Place ColdFusion apps here"
    echo "  📦 $TOOLS_DIR/input/archives/      # Original files backup"
    echo "  🔄 $TOOLS_DIR/input/temp/          # Temporary processing"
    
    update_gitignore_only
}

update_gitignore_only() {
    echo ""
    echo "🔒 Updating .gitignore for data protection..."
    
    # Update tools .gitignore
    cat >> "$TOOLS_DIR/.gitignore" << 'EOF'

# Input data files (sensitive - do not commit)
/input/
*.mdb
*.accdb
*.ldb

# Company-specific data
*_company_*
*_client_*
*_customer_*

# Temporary processing files
*.tmp
*.temp
/temp/
/archives/

# Output files with real data
/output/*.csv
/output/*.sql
/output/*.json
EOF

    echo "✅ Updated $TOOLS_DIR/.gitignore"
    echo "🔒 Added protection for:"
    echo "  • All input/ directory contents"
    echo "  • MDB and Access database files"
    echo "  • Company/client specific files"
    echo "  • Temporary and output files"
}

run_automated_migration() {
    echo ""
    echo "🤖 Running automated migration..."
    
    # Create backup
    backup_dir="$PROJECT_ROOT/db_prep.backup.$(date +%Y%m%d_%H%M%S)"
    echo "💾 Creating backup: $backup_dir"
    cp -r "$DB_PREP_DIR" "$backup_dir"
    
    # Create input structure
    create_input_structure
    
    # Migrate files
    echo ""
    echo "📊 Migrating files..."
    
    if [ "$mdb_count" -gt 0 ]; then
        echo "Moving MDB files..."
        find "$DB_PREP_DIR" -name "*.mdb" -exec mv {} "$TOOLS_DIR/input/mdb_files/" \;
    fi
    
    if [ "$cf_files" -gt 0 ] || [ "$cf_dirs" -gt 0 ]; then
        echo "Moving ColdFusion files..."
        find "$DB_PREP_DIR" -name "*.cfm" -o -name "*.cfc" -exec mv {} "$TOOLS_DIR/input/cf_apps/" \;
        find "$DB_PREP_DIR" -type d -name "*coldfusion*" -exec mv {} "$TOOLS_DIR/input/cf_apps/" \;
    fi
    
    # Copy break.sh for analysis
    if [ -f "$DB_PREP_DIR/break.sh" ]; then
        echo "Copying break.sh for analysis..."
        cp "$DB_PREP_DIR/break.sh" "$TOOLS_DIR/input/archives/break.sh.original"
    fi
    
    echo ""
    echo "✅ Migration complete!"
    echo "📁 Files moved to tools/mdb-conversion/input/"
    echo "💾 Original backed up to: $backup_dir"
    echo ""
    echo "🧪 Test the new setup:"
    echo "  make tools-status"
    echo "  ls tools/mdb-conversion/input/mdb_files/"
    echo "  ls tools/mdb-conversion/input/cf_apps/"
}
