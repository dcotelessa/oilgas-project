#!/bin/bash
# Test API integration after repository wiring

set -e

API_BASE="http://localhost:8000"
TIMEOUT=10

echo "🧪 Testing API Integration"
echo "================================"

# Check if server is running
echo -n "Checking if API server is running... "
if curl -s --max-time 5 "$API_BASE/health" > /dev/null; then
    echo "✅ Server responding"
else
    echo "❌ Server not responding"
    echo ""
    echo "Start the server first:"
    echo "  make api-start"
    exit 1
fi

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
            count=$(echo "$body" | jq -r '.count // empty' 2>/dev/null || echo "")
            if [ -n "$count" ]; then
                echo "   📊 Records: $count"
            fi
        fi
    else
        echo "❌ ($http_code)"
        echo "   Expected: $expected_status"
        return 1
    fi
}

# Test endpoints
echo ""
echo "1. Health Checks"
echo "----------------"
test_endpoint "GET" "/health" "200" "Health check"
test_endpoint "GET" "/api/v1/status" "200" "API status"

echo ""
echo "2. Customer Endpoints"
echo "---------------------"
test_endpoint "GET" "/api/v1/customers" "200" "Get all customers"
test_endpoint "GET" "/api/v1/customers/search?q=oil" "200" "Search customers"

echo ""
echo "3. Inventory Endpoints"
echo "----------------------"
test_endpoint "GET" "/api/v1/inventory" "200" "Get all inventory"

echo ""
echo "4. Reference Data"
echo "-----------------"
test_endpoint "GET" "/api/v1/grades" "200" "Get all grades"
test_endpoint "GET" "/api/v1/sizes" "200" "Get all sizes"

echo ""
echo "🎯 API Integration Test Summary"
echo "================================"
echo "✅ Server running and responding"
echo "✅ Repository layer connected"
echo "✅ Basic endpoints working"
echo ""
echo "🚀 API Integration Complete!"
echo ""
echo "Next steps:"
echo "1. Import your MDB data: make data-convert && make data-import"
echo "2. View API examples: make api-examples"
echo "3. Start development: make api-dev"
