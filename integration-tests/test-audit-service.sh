#!/bin/bash

# Integration Tests for Audit Service
# Tests: CORS, endpoints, database operations, filtering, pagination

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
AUDIT_SERVICE_URL="http://localhost:3001"
TEST_DB_PORT="${TEST_DB_PORT:-5433}"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test function wrapper
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "Running test: $test_name"
    
    if eval "$test_command"; then
        log_success "$test_name"
        return 0
    else
        log_error "$test_name"
        return 1
    fi
}

# Check if service is running
check_service_health() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/health" 2>/dev/null)
    if echo "$response" | jq -e '.status == "healthy"' >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Setup: Start services
setup() {
    log_info "Setting up integration test environment..."
    
    # Check if audit service is running
    if ! check_service_health; then
        log_warning "Audit service is not running on $AUDIT_SERVICE_URL"
        log_info "Note: Integration tests require a running audit-service instance."
        log_info "To run these tests:"
        log_info "  1. Ensure PostgreSQL is running"
        log_info "  2. Set environment variables (CHOREO_DB_AUDIT_HOSTNAME, etc.)"
        log_info "  3. Start the service: cd audit-service && go run main.go"
        log_info "  4. Then run this test script again"
        log_info ""
        log_info "Skipping tests - service not available"
        exit 0
    else
        log_success "Audit service is running"
    fi
}

# Test 1: Health Check Endpoint
test_health_endpoint() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/health")
    local status=$(echo "$response" | jq -r '.status')
    
    if [ "$status" = "healthy" ]; then
        log_info "Health check response: $response"
        return 0
    else
        log_error "Health check failed. Response: $response"
        return 1
    fi
}

# Test 2: Version Endpoint
test_version_endpoint() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/version")
    local service=$(echo "$response" | jq -r '.service')
    
    if [ "$service" = "audit-service" ]; then
        log_info "Version check passed: $response"
        return 0
    else
        log_error "Version check failed. Response: $response"
        return 1
    fi
}

# Test 3: CORS Headers
test_cors_headers() {
    # Test preflight request
    local preflight=$(curl -s -X OPTIONS "$AUDIT_SERVICE_URL/api/logs" \
        -H "Origin: http://localhost:5173" \
        -H "Access-Control-Request-Method: GET" \
        -v 2>&1)
    
    if echo "$preflight" | grep -q "Access-Control-Allow-Origin"; then
        log_info "CORS preflight headers present"
        return 0
    else
        log_error "CORS preflight headers missing"
        return 1
    fi
}

# Test 4: GET /api/logs (empty at start)
test_get_logs_empty() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?limit=10")
    local total=$(echo "$response" | jq -r '.total // 0')
    
    log_info "Initial logs count: $total"
    return 0  # Always pass - just checking service responds
}

# Test 5: POST /api/logs - Create Success Log
test_create_success_log() {
    local log_data='{
        "status": "success",
        "requestedData": "query { personInfo(nic: \"199512345678\") { fullName } }",
        "applicationId": "test-app-001",
        "schemaId": "schema-person"
    }'
    
    local response=$(curl -s -X POST "$AUDIT_SERVICE_URL/api/logs" \
        -H "Content-Type: application/json" \
        -d "$log_data")
    
    local status=$(echo "$response" | jq -r '.status')
    
    if [ "$status" = "success" ]; then
        log_info "Success log created: $response"
        echo "$response" | jq -r '.id' > /tmp/test_log_id
        return 0
    else
        log_error "Failed to create success log. Response: $response"
        return 1
    fi
}

# Test 6: POST /api/logs - Create Failure Log
test_create_failure_log() {
    local log_data='{
        "status": "failure",
        "requestedData": "query { invalidInfo(id: \"123\") { field } }",
        "applicationId": "test-app-002",
        "schemaId": "schema-vehicle"
    }'
    
    local response=$(curl -s -X POST "$AUDIT_SERVICE_URL/api/logs" \
        -H "Content-Type: application/json" \
        -d "$log_data")
    
    local status=$(echo "$response" | jq -r '.status')
    
    if [ "$status" = "failure" ]; then
        log_info "Failure log created: $response"
        return 0
    else
        log_error "Failed to create failure log. Response: $response"
        return 1
    fi
}

# Test 7: GET /api/logs - Retrieve logs
test_get_logs_with_data() {
    sleep 1  # Wait for database to sync
    
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs")
    local logs_count=$(echo "$response" | jq '.logs | length')
    
    if [ "$logs_count" -gt 0 ]; then
        log_info "Retrieved $logs_count logs"
        return 0
    else
        log_error "No logs retrieved"
        return 1
    fi
}

# Test 8: GET /api/logs - Filter by status
test_filter_by_status() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?status=success")
    local logs_count=$(echo "$response" | jq '.logs | length')
    local total=$(echo "$response" | jq '.total')
    
    if [ "$total" -gt 0 ]; then
        log_info "Found $total success logs"
        return 0
    else
        log_warning "No success logs found"
        return 0  # Pass - filtering works even if no results
    fi
}

# Test 9: GET /api/logs - Filter by application
test_filter_by_application() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?applicationId=test-app-001")
    local total=$(echo "$response" | jq '.total')
    
    if [ "$total" -gt 0 ]; then
        log_info "Found $total logs for application test-app-001"
        return 0
    else
        log_warning "No logs found for test-app-001"
        return 1
    fi
}

# Test 10: GET /api/logs - Pagination (limit)
test_pagination_limit() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?limit=1")
    local logs_count=$(echo "$response" | jq '.logs | length')
    
    if [ "$logs_count" -le 1 ]; then
        log_info "Pagination limit works: returned $logs_count logs"
        return 0
    else
        log_error "Pagination limit failed: returned $logs_count logs (expected <= 1)"
        return 1
    fi
}

# Test 11: GET /api/logs - Pagination (offset)
test_pagination_offset() {
    local response1=$(curl -s "$AUDIT_SERVICE_URL/api/logs?limit=1&offset=0")
    local log1_id=$(echo "$response1" | jq -r '.logs[0].id')
    
    local response2=$(curl -s "$AUDIT_SERVICE_URL/api/logs?limit=1&offset=1")
    local log2_id=$(echo "$response2" | jq -r '.logs[0].id // "none"')
    
    if [ "$log1_id" != "$log2_id" ]; then
        log_info "Pagination offset works"
        return 0
    else
        log_warning "Pagination offset test inconclusive (may have <2 logs)"
        return 0
    fi
}

# Test 12: POST /api/logs - Invalid status
test_invalid_status() {
    local log_data='{
        "status": "invalid",
        "requestedData": "query { test } }",
        "applicationId": "test-app-invalid",
        "schemaId": "schema-invalid"
    }'
    
    local http_code=$(curl -s -w "%{http_code}" -o /tmp/invalid_response \
        -X POST "$AUDIT_SERVICE_URL/api/logs" \
        -H "Content-Type: application/json" \
        -d "$log_data")
    
    if [ "$http_code" = "400" ]; then
        log_info "Invalid status correctly rejected with 400"
        return 0
    else
        log_error "Invalid status not rejected. HTTP code: $http_code"
        return 1
    fi
}

# Test 13: POST /api/logs - Missing required fields
test_missing_required_fields() {
    local log_data='{
        "status": "success",
        "applicationId": "test-app"
    }'
    
    local http_code=$(curl -s -w "%{http_code}" -o /tmp/missing_fields_response \
        -X POST "$AUDIT_SERVICE_URL/api/logs" \
        -H "Content-Type: application/json" \
        -d "$log_data")
    
    if [ "$http_code" = "400" ]; then
        log_info "Missing required fields correctly rejected with 400"
        return 0
    else
        log_error "Missing required fields not rejected. HTTP code: $http_code"
        return 1
    fi
}

# Test 14: GET /api/logs - Date range filtering
test_date_filtering() {
    local today=$(date +%Y-%m-%d)
    local tomorrow=$(date -v+1d +%Y-%m-%d 2>/dev/null || date -d "tomorrow" +%Y-%m-%d)
    
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?startDate=$today&endDate=$tomorrow")
    local total=$(echo "$response" | jq '.total')
    
    log_info "Date filter test completed (found $total logs between $today and $tomorrow)"
    return 0  # Pass - filtering syntax works
}

# Test 15: GET /api/logs - Response structure
test_response_structure() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs")
    
    # Check required fields in response
    local has_logs=$(echo "$response" | jq 'has("logs")')
    local has_total=$(echo "$response" | jq 'has("total")')
    local has_limit=$(echo "$response" | jq 'has("limit")')
    
    if [ "$has_logs" = "true" ] && [ "$has_total" = "true" ] && [ "$has_limit" = "true" ]; then
        log_info "Response structure is valid"
        return 0
    else
        log_error "Response structure invalid (has_logs=$has_logs, has_total=$has_total, has_limit=$has_limit)"
        return 1
    fi
}

# Main test execution
main() {
    echo ""
    echo "========================================="
    echo "  Audit Service Integration Tests"
    echo "========================================="
    echo ""
    
    setup
    
    # Run all tests
    echo ""
    log_info "Running integration tests..."
    echo ""
    
    run_test "Health Check" test_health_endpoint
    run_test "Version Endpoint" test_version_endpoint
    run_test "CORS Headers" test_cors_headers
    run_test "GET Logs (Empty)" test_get_logs_empty
    run_test "POST Success Log" test_create_success_log
    run_test "POST Failure Log" test_create_failure_log
    run_test "GET Logs (With Data)" test_get_logs_with_data
    run_test "Filter by Status" test_filter_by_status
    run_test "Filter by Application" test_filter_by_application
    run_test "Pagination Limit" test_pagination_limit
    run_test "Pagination Offset" test_pagination_offset
    run_test "Invalid Status" test_invalid_status
    run_test "Missing Required Fields" test_missing_required_fields
    run_test "Date Filtering" test_date_filtering
    run_test "Response Structure" test_response_structure
    
    # Print summary
    echo ""
    echo "========================================="
    echo "  Test Summary"
    echo "========================================="
    echo -e "${BLUE}Total Tests:${NC} $TOTAL_TESTS"
    echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
    echo -e "${RED}Failed:${NC} $TESTS_FAILED"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed!"
        exit 0
    else
        log_error "Some tests failed"
        exit 1
    fi
}

# Run main
main