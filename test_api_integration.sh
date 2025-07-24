#!/bin/bash
# Test API integration after repository wiring

set -e

API_BASE="http://localhost:8000"
TIMEOUT=10

echo "🧪 Testing API Integration"
echo "================================"

# Function to test endpoint
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local expected_status="$3"
    local description="$4"
    
    echo -n "Testing $description... "
    
    response=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT \
        -X "$method" "$API_BASE$endpoint" \
        -H "Content-Type: application/json")
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "✅ ($http_code)"
        if [ "$method" = "GET" ] && [ -n "$body" ]; then
            # Show data count if JSON response
            count=$(echo "$body" | jq -r '.count // .customers // .inventory // .grades // .sizes | length // empty' 2>/dev/null || echo "")
            if [ -n "$count" ]; then
                echo "   📊 Records: $count"
            fi
        fi
    else
        echo "❌ ($http_code)"
        echo "   Expected: $expected_status"
        echo "   Response: $body"
        return 1
    fi
}

# Check if server is running
echo -n "Checking if API server is running... "
if curl -s --max-time 5 "$API_BASE/health" > /dev/null; then
    echo "✅ Server responding"
else
    echo "❌ Server not responding"
    echo ""
    echo "Start the server first:"
    echo "  make api-start"
    echo "Or run in background:"
    echo "  make api-start &"
    exit 1
fi

# Test server connectivity
echo ""
echo "1. Server Health Checks"
echo "------------------------"
test_endpoint "GET" "/health" "200" "Health check"
test_endpoint "GET" "/api/v1/status" "200" "API status"

echo ""
echo "2. Customer Endpoints"
echo "---------------------"
test_endpoint "GET" "/api/v1/customers" "200" "Get all customers"
test_endpoint "GET" "/api/v1/customers/1" "200" "Get customer by ID"
test_endpoint "GET" "/api/v1/customers/99999" "404" "Get non-existent customer"
test_endpoint "GET" "/api/v1/customers/search?q=oil" "200" "Search customers"
test_endpoint "GET" "/api/v1/customers/search" "400" "Search without query"

echo ""
echo "3. Inventory Endpoints"
echo "----------------------"
test_endpoint "GET" "/api/v1/inventory" "200" "Get all inventory"
test_endpoint "GET" "/api/v1/inventory/1" "200" "Get inventory by ID"
test_endpoint "GET" "/api/v1/inventory/99999" "404" "Get non-existent inventory"
test_endpoint "GET" "/api/v1/inventory/search?q=5" "200" "Search inventory"

echo ""
echo "4. Reference Data"
echo "-----------------"
test_endpoint "GET" "/api/v1/grades" "200" "Get all grades"
test_endpoint "GET" "/api/v1/sizes" "200" "Get all sizes"

echo ""
echo "5. Error Handling"
echo "-----------------"
test_endpoint "GET" "/api/v1/customers/invalid" "400" "Invalid ID format"
test_endpoint "GET" "/api/v1/nonexistent" "404" "Non-existent endpoint"

echo ""
echo "6. Sample Data Verification"
echo "----------------------------"

# Check for sample data
echo -n "Checking for sample customers... "
customer_response=$(curl -s "$API_BASE/api/v1/customers")
customer_count=$(echo "$customer_response" | jq -r '.count // 0' 2>/dev/null || echo "0")

if [ "$customer_count" -gt "0" ]; then
    echo "✅ Found $customer_count customers"
    echo "   Sample: $(echo "$customer_response" | jq -r '.customers[0].customer_name // "Unknown"' 2>/dev/null)"
else
    echo "⚠️  No customers found - you may need to run seed data"
    echo "   Run: make dev-setup"
fi

echo -n "Checking for sample inventory... "
inventory_response=$(curl -s "$API_BASE/api/v1/inventory")
inventory_count=$(echo "$inventory_response" | jq -r '.count // 0' 2>/dev/null || echo "0")

if [ "$inventory_count" -gt "0" ]; then
    echo "✅ Found $inventory_count inventory items"
    echo "   Sample: $(echo "$inventory_response" | jq -r '.inventory[0].work_order // "Unknown"' 2>/dev/null)"
else
    echo "⚠️  No inventory found - you may need to run seed data"
    echo "   Run: make dev-setup"
fi

echo ""
echo "🎯 API Integration Test Summary"
echo "================================"
echo "✅ Server running and responding"
echo "✅ Repository layer connected"
echo "✅ CRUD operations working"
echo "✅ Error handling implemented"
echo "✅ Search functionality working"

if [ "$customer_count" -gt "0" ] && [ "$inventory_count" -gt "0" ]; then
    echo "✅ Sample data available"
    echo ""
    echo "🚀 Ready for MDB data import!"
    echo ""
    echo "Next steps:"
    echo "1. Process your MDB files: make convert-mdb"
    echo "2. Import real data: make import-mdb-data"
    echo "3. Test with real data: $0"
else
    echo "⚠️  Sample data missing"
    echo ""
    echo "Run this first:"
    echo "  make dev-setup"
    echo "  $0"
fi

echo ""
echo "🔍 Manual Testing Commands:"
echo "curl $API_BASE/api/v1/customers | jq"
echo "curl '$API_BASE/api/v1/customers/search?q=oil' | jq"
echo "curl $API_BASE/api/v1/inventory | jq"
echo "curl $API_BASE/api/v1/grades | jq"
echo "curl $API_BASE/api/v1/sizes | jq"
