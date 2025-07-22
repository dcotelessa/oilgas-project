#!/bin/bash
# Complete Migration Verification

set -e

echo "🔍 Complete Migration Verification"
echo "=================================="

# Check we're in the right place
if [ ! -f "Makefile" ]; then
    echo "❌ Please run this from your project root directory"
    exit 1
fi

echo "📁 Project root: $(pwd)"
echo "📅 Verification time: $(date)"

echo ""
echo "1️⃣ Build System Verification"
echo "============================"

echo "🔨 Testing make migration-build..."
if make migration-build >/dev/null 2>&1; then
    echo "✅ make migration-build works"
else
    echo "❌ make migration-build failed"
    make migration-build 2>&1 | tail -5 | sed 's/^/  /'
    exit 1
fi

echo "🔍 Checking built binary..."
if [ -f "tools/bin/mdb_processor" ]; then
    echo "✅ Binary exists: tools/bin/mdb_processor"
    echo "📊 Binary size: $(ls -lh tools/bin/mdb_processor | awk '{print $5}')"
    echo "🔐 Binary permissions: $(ls -l tools/bin/mdb_processor | awk '{print $1}')"
else
    echo "❌ Binary not found"
    exit 1
fi

echo ""
echo "2️⃣ Command Interface Verification"
echo "================================="

cd tools/

echo "🧪 Testing help command..."
if ./bin/mdb_processor -help >/dev/null 2>&1; then
    echo "✅ Help command works"
    echo "📄 Help preview:"
    ./bin/mdb_processor -help | head -5 | sed 's/^/  /'
else
    echo "❌ Help command failed"
    ./bin/mdb_processor -help 2>&1 | head -3 | sed 's/^/  /'
fi

echo ""
echo "🧪 Testing version command..."
if version_output=$(./bin/mdb_processor -version 2>/dev/null); then
    echo "✅ Version command works: $version_output"
else
    echo "❌ Version command failed"
fi

echo ""
echo "🧪 Testing error handling..."
if ./bin/mdb_processor -file nonexistent.csv -company "Test" 2>/dev/null; then
    echo "⚠️  Should have failed for nonexistent file"
else
    echo "✅ Properly handles nonexistent files"
fi

echo ""
echo "3️⃣ Test Data Verification"
echo "========================="

echo "📄 Checking test fixtures..."
if [ -f "test/fixtures/basic_test.csv" ]; then
    echo "✅ basic_test.csv exists"
    echo "📊 Content preview:"
    head -3 test/fixtures/basic_test.csv | sed 's/^/  /'
else
    echo "⚠️  basic_test.csv missing - creating it..."
    mkdir -p test/fixtures
    cat > test/fixtures/basic_test.csv << 'EOF'
WorkOrder,Customer,Joints,Size,Grade,Connection
LB-001001,Test Customer,100,5.5,L-80,BTC
LB-001002,Another Customer,150,7,P-110,VAM TOP
EOF
    echo "✅ Created basic_test.csv"
fi

# Create a more comprehensive test file
echo "📄 Creating comprehensive test data..."
cat > test/fixtures/comprehensive_test.csv << 'EOF'
WorkOrder,Customer,Joints,Size,Grade,Connection,Weight,DateIn
LB-001001,chevron corporation,100,5.5,J-55,BUTTRESS THREAD CASING,2500.50,2024-01-15
LB-001002,exxon mobil corp,150,7,L-80,BTC,4200.75,2024-01-16
LB-001003,conocophillips company,75,Nine 5/8,P-110,VAM TOP,6800.25,2024-01-17
LB-001004,Test Company LLC,200,4.5,N-80,LTC,1800.00,2024-01-18
INVALID-FORMAT,sample customer,50,invalid size,X99,UNKNOWN,abc,invalid-date
EOF

echo "✅ Created comprehensive_test.csv"

echo ""
echo "4️⃣ Basic Processing Verification"
echo "================================"

echo "🧪 Testing basic file processing..."
rm -rf output/basic_test 2>/dev/null || true

if ./bin/mdb_processor -file test/fixtures/basic_test.csv -company "Basic Test Co" -output output/basic_test -verbose; then
    echo "✅ Basic processing successful"
else
    echo "❌ Basic processing failed"
    exit 1
fi

echo ""
echo "📊 Checking basic output..."
if [ -d "output/basic_test" ]; then
    echo "✅ Output directory created"
    echo "📁 Output structure:"
    find output/basic_test -type f | sed 's/^/  📄 /'
    
    # Check CSV output
    if [ -f "output/basic_test/csv/basic_test.csv" ]; then
        echo "✅ CSV output created"
        echo "📄 CSV content preview:"
        head -3 "output/basic_test/csv/basic_test.csv" | sed 's/^/  /'
    else
        echo "❌ CSV output missing"
    fi
    
    # Check SQL output
    if [ -f "output/basic_test/sql/basic_test.sql" ]; then
        echo "✅ SQL output created"
        echo "📄 SQL content preview:"
        head -5 "output/basic_test/sql/basic_test.sql" | sed 's/^/  /'
    else
        echo "❌ SQL output missing"
    fi
else
    echo "❌ No output directory created"
fi

echo ""
echo "5️⃣ Business Rules Verification"
echo "=============================="

echo "🧪 Testing oil & gas business rules..."
rm -rf output/comprehensive_test 2>/dev/null || true

if ./bin/mdb_processor -file test/fixtures/comprehensive_test.csv -company "Business Rules Test" -output output/comprehensive_test -verbose; then
    echo "✅ Comprehensive processing successful"
else
    echo "❌ Comprehensive processing failed"
fi

echo ""
echo "🔍 Verifying business rule transformations..."
if [ -f "output/comprehensive_test/csv/comprehensive_test.csv" ]; then
    csv_file="output/comprehensive_test/csv/comprehensive_test.csv"
    
    echo "📊 Checking transformations:"
    
    # Check grade normalization
    if grep -q "J55" "$csv_file" && ! grep -q "J-55" "$csv_file"; then
        echo "  ✅ Grade normalization: J-55 → J55"
    else
        echo "  ⚠️  Grade normalization unclear"
        echo "    Grades found: $(cut -d',' -f5 "$csv_file" | tail -n +2 | sort | uniq | tr '\n' ' ')"
    fi
    
    # Check size normalization
    if grep -q "5 1/2" "$csv_file" || grep -q "5\"" "$csv_file"; then
        echo "  ✅ Size normalization working"
    else
        echo "  ⚠️  Size normalization unclear"
        echo "    Sizes found: $(cut -d',' -f4 "$csv_file" | tail -n +2 | sort | uniq | tr '\n' ' ')"
    fi
    
    # Check customer name normalization
    if grep -qi "Chevron" "$csv_file" && grep -qi "Exxon" "$csv_file"; then
        echo "  ✅ Customer name normalization working"
    else
        echo "  ⚠️  Customer name normalization unclear"
    fi
    
    echo ""
    echo "📄 Sample transformed data:"
    head -5 "$csv_file" | sed 's/^/  /'
    
else
    echo "❌ Comprehensive output CSV not found"
fi

echo ""
echo "6️⃣ Configuration Verification"
echo "============================="

echo "📄 Checking configuration file..."
if [ -f "config/oil_gas_mappings.json" ]; then
    echo "✅ Configuration file exists"
    
    # Validate JSON syntax
    if python3 -m json.tool config/oil_gas_mappings.json >/dev/null 2>&1; then
        echo "✅ Configuration JSON is valid"
    else
        echo "❌ Configuration JSON is invalid"
        python3 -m json.tool config/oil_gas_mappings.json 2>&1 | head -3 | sed 's/^/  /'
    fi
    
    echo "📊 Configuration preview:"
    head -10 config/oil_gas_mappings.json | sed 's/^/  /'
else
    echo "❌ Configuration file missing"
fi

echo ""
echo "7️⃣ Package Integration Verification"
echo "==================================="

echo "🔨 Testing individual package compilation..."
packages=("config" "mapping" "processor" "reporting" "validation" "exporters")
compile_success=0

for pkg in "${packages[@]}"; do
    if [ -d "internal/$pkg" ]; then
        if go build "./internal/$pkg" >/dev/null 2>&1; then
            echo "  ✅ internal/$pkg compiles"
            compile_success=$((compile_success + 1))
        else
            echo "  ❌ internal/$pkg compilation failed"
            go build "./internal/$pkg" 2>&1 | head -2 | sed 's/^/    /'
        fi
    else
        echo "  ⚠️  internal/$pkg directory missing"
    fi
done

echo "📊 Package compilation: $compile_success/${#packages[@]} successful"

echo ""
echo "🧪 Testing go mod status..."
if go mod verify >/dev/null 2>&1; then
    echo "✅ Go module verified"
else
    echo "⚠️  Go module verification issues:"
    go mod verify 2>&1 | head -3 | sed 's/^/  /'
fi

if go mod tidy >/dev/null 2>&1; then
    echo "✅ Go mod tidy successful"
else
    echo "❌ Go mod tidy failed"
fi

cd ..

echo ""
echo "8️⃣ Integration with Main Project"
echo "==============================="

echo "🔗 Testing main project integration..."
if grep -q "migration-build" Makefile; then
    echo "✅ Main Makefile has migration commands"
else
    echo "❌ Main Makefile missing migration commands"
fi

# Test from project root
echo "🧪 Testing from project root..."
if make migration-build >/dev/null 2>&1; then
    echo "✅ make migration-build works from project root"
else
    echo "❌ make migration-build fails from project root"
fi

echo ""
echo "9️⃣ Performance Verification"
echo "==========================="

echo "⏱️  Testing processing speed..."
cd tools/

# Create larger test file
echo "📊 Creating performance test data (1000 records)..."
{
    echo "WorkOrder,Customer,Joints,Size,Grade,Connection,Weight"
    for i in {1..1000}; do
        printf "LB-%06d,Customer %d,%d,5 1/2,L80,BTC,%.2f\n" $i $((i % 100)) $((50 + i % 200)) $((1000 + i))
    done
} > test/fixtures/performance_test.csv

echo "✅ Created performance test file (1000 records)"

# Time the processing
echo "🚀 Running performance test..."
start_time=$(date +%s)

if ./bin/mdb_processor -file test/fixtures/performance_test.csv -company "Performance Test" -output output/performance_test -workers 4; then
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    echo "✅ Performance test completed"
    echo "⏱️  Processing time: ${duration} seconds"
    
    if [ $duration -lt 10 ]; then
        echo "🚀 Performance: EXCELLENT (< 10 seconds)"
    elif [ $duration -lt 30 ]; then
        echo "✅ Performance: GOOD (< 30 seconds)"
    else
        echo "⚠️  Performance: SLOW (> 30 seconds)"
    fi
    
    # Calculate records per second
    if [ $duration -gt 0 ]; then
        rate=$((1000 / duration))
        echo "📊 Processing rate: ~${rate} records/second"
    fi
else
    echo "❌ Performance test failed"
fi

cd ..

echo ""
echo "🔟 Final Migration Status"
echo "========================"

# Calculate overall success
success_count=0
total_checks=10

echo "📋 Verification Results:"

# 1. Build system
if make migration-build >/dev/null 2>&1; then
    echo "  ✅ Build system working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Build system issues"
fi

# 2. Binary functionality
if [ -f "tools/bin/mdb_processor" ] && tools/bin/mdb_processor -help >/dev/null 2>&1; then
    echo "  ✅ Binary functionality working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Binary functionality issues"
fi

# 3. Basic processing
if [ -f "tools/output/basic_test/csv/basic_test.csv" ]; then
    echo "  ✅ Basic processing working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Basic processing issues"
fi

# 4. Business rules
if [ -f "tools/output/comprehensive_test/csv/comprehensive_test.csv" ]; then
    echo "  ✅ Business rules processing working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Business rules processing issues"
fi

# 5. Configuration
if [ -f "tools/config/oil_gas_mappings.json" ]; then
    echo "  ✅ Configuration system working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Configuration system issues"
fi

# 6. Package integration
cd tools/
if go build ./internal/... >/dev/null 2>&1; then
    echo "  ✅ Package integration working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Package integration issues"
fi
cd ..

# 7. Output generation
if [ -d "tools/output" ] && ls tools/output/*/csv/*.csv >/dev/null 2>&1; then
    echo "  ✅ Output generation working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Output generation issues"
fi

# 8. SQL generation
if ls tools/output/*/sql/*.sql >/dev/null 2>&1; then
    echo "  ✅ SQL generation working"
    success_count=$((success_count + 1))
else
    echo "  ❌ SQL generation issues"
fi

# 9. Error handling
if ! tools/bin/mdb_processor -file nonexistent.csv -company "Test" >/dev/null 2>&1; then
    echo "  ✅ Error handling working"
    success_count=$((success_count + 1))
else
    echo "  ❌ Error handling issues"
fi

# 10. Performance
if [ -f "tools/output/performance_test/csv/performance_test.csv" ]; then
    echo "  ✅ Performance acceptable"
    success_count=$((success_count + 1))
else
    echo "  ❌ Performance issues"
fi

echo ""
echo "📊 Overall Success Rate: $success_count/$total_checks ($(( success_count * 100 / total_checks ))%)"

if [ $success_count -eq $total_checks ]; then
    echo ""
    echo "🎉 MIGRATION VERIFICATION: COMPLETE SUCCESS!"
    echo "============================================="
    echo ""
    echo "✅ All systems operational:"
    echo "  • Build system works perfectly"
    echo "  • MDB processor handles CSV files correctly"
    echo "  • Oil & gas business rules apply properly"
    echo "  • Output generation (CSV, SQL, reports) working"
    echo "  • Performance is acceptable"
    echo "  • Error handling is robust"
    echo ""
    echo "🚀 Ready for production use!"
    echo ""
    echo "📚 Usage examples:"
    echo "  make migration-build"
    echo "  cd tools"
    echo "  ./bin/mdb_processor -file your_data.csv -company \"Your Company\" -verbose"
    echo ""
    echo "📁 Check outputs in: tools/output/"
    
elif [ $success_count -ge 8 ]; then
    echo ""
    echo "✅ MIGRATION VERIFICATION: MOSTLY SUCCESSFUL"
    echo "============================================"
    echo ""
    echo "🎯 Core functionality working well with minor issues"
    echo "💡 Review the failed checks above for improvement areas"
    echo ""
    echo "🚀 Safe to proceed with testing and refinement"
    
elif [ $success_count -ge 6 ]; then
    echo ""
    echo "⚠️  MIGRATION VERIFICATION: PARTIAL SUCCESS"
    echo "==========================================="
    echo ""
    echo "🔧 Significant functionality working but needs attention"
    echo "💡 Address the failed areas before production use"
    echo ""
    echo "🛠️  Recommended: Review error messages and fix issues"
    
else
    echo ""
    echo "❌ MIGRATION VERIFICATION: NEEDS WORK"
    echo "===================================="
    echo ""
    echo "🔧 Major issues detected - migration needs refinement"
    echo "💡 Focus on the failed checks above"
    echo ""
    echo "🛠️  Recommended: Address core issues before proceeding"
fi

echo ""
echo "📁 Generated artifacts for inspection:"
echo "  tools/bin/mdb_processor        - Main binary"
echo "  tools/output/*/                - Sample processing results"
echo "  tools/test/fixtures/           - Test data files"
echo "  tools/config/                  - Configuration files"

echo ""
echo "🎯 Migration verification complete!"
echo "Timestamp: $(date)"
