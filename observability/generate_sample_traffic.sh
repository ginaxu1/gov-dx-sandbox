#!/bin/bash

# Script to generate sample traffic for Grafana dashboard
# This sends requests to various endpoints to populate metrics

BASE_URL="${PORTAL_BACKEND_URL:-http://localhost:3000}"
INTERVAL="${REQUEST_INTERVAL:-2}"  # seconds between requests
COUNT="${REQUEST_COUNT:-50}"        # number of requests per endpoint (0 = infinite)

echo "=========================================="
echo "Generating Sample Traffic for Grafana"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "Interval: ${INTERVAL}s"
echo "Count: ${COUNT} (0 = infinite)"
echo ""
echo "Press Ctrl+C to stop"
echo "=========================================="
echo ""

# Function to send a request and show status
send_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$BASE_URL$endpoint" 2>/dev/null)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        status="✓"
    elif [ "$http_code" -ge 400 ] && [ "$http_code" -lt 500 ]; then
        status="⚠"
    else
        status="✗"
    fi
    
    echo "[$status] $method $endpoint -> $http_code"
}

# Function to run requests in a loop
run_requests() {
    local i=0
    while [ $COUNT -eq 0 ] || [ $i -lt $COUNT ]; do
        echo ""
        echo "--- Batch $((i+1)) ---"
        
        # Health and metrics endpoints (should work)
        send_request "GET" "/health" "" "Health check"
        sleep 0.5
        send_request "GET" "/metrics" "" "Metrics endpoint"
        sleep 0.5
        
        # API endpoints (will generate 401s without auth, but that's useful metrics)
        send_request "GET" "/api/v1/members" "" "Get members"
        sleep 0.5
        send_request "GET" "/api/v1/schemas" "" "Get schemas"
        sleep 0.5
        send_request "GET" "/api/v1/applications" "" "Get applications"
        sleep 0.5
        send_request "GET" "/api/v1/schema-submissions" "" "Get schema submissions"
        sleep 0.5
        send_request "GET" "/api/v1/application-submissions" "" "Get application submissions"
        sleep 0.5
        
        # POST requests (will fail without auth/data, but generates metrics)
        send_request "POST" "/api/v1/members" '{"name":"Test User","email":"test@example.com"}' "Create member"
        sleep 0.5
        send_request "POST" "/api/v1/schemas" '{"schema_name":"Test Schema"}' "Create schema"
        sleep 0.5
        
        # Invalid endpoints (404s)
        send_request "GET" "/api/v1/unknown" "" "Unknown endpoint"
        sleep 0.5
        send_request "GET" "/nonexistent" "" "Non-existent path"
        sleep 0.5
        
        # Invalid JSON (400s)
        send_request "POST" "/api/v1/members" '{"invalid": }' "Invalid JSON"
        sleep 0.5
        
        i=$((i+1))
        
        if [ $COUNT -eq 0 ] || [ $i -lt $COUNT ]; then
            echo ""
            echo "Waiting ${INTERVAL}s before next batch..."
            sleep $INTERVAL
        fi
    done
}

# Check if portal-backend is accessible
echo "Checking if portal-backend is accessible..."
if ! curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
    echo "❌ ERROR: Cannot reach portal-backend at $BASE_URL"
    echo "   Make sure the service is running and accessible."
    echo ""
    echo "   Try: curl $BASE_URL/health"
    exit 1
fi

echo "✓ Portal-backend is accessible"
echo ""

# Start generating traffic
run_requests

echo ""
echo "=========================================="
echo "Traffic generation complete!"
echo "=========================================="
echo ""
echo "View metrics in Grafana:"
echo "  http://localhost:3002/d/go-services/go-services-metrics"
echo ""
echo "Or check Prometheus directly:"
echo "  http://localhost:9091"
echo ""



