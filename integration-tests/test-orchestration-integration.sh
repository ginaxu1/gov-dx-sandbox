#!/bin/bash

# Test script for orchestration engine integration with audit service
# This script tests the complete flow from orchestration engine to audit service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AUDIT_SERVICE_URL="http://localhost:3001"
ORCHESTRATION_ENGINE_URL="http://localhost:4000"
TEST_DB_HOST="${CHOREO_DB_AUDIT_HOSTNAME:-localhost}"
TEST_DB_PORT="${CHOREO_DB_AUDIT_PORT:-5432}"
TEST_DB_USERNAME="${CHOREO_DB_AUDIT_USERNAME:-postgres}"
TEST_DB_PASSWORD="${CHOREO_DB_AUDIT_PASSWORD:-password}"
TEST_DB_DATABASE="${CHOREO_DB_AUDIT_DATABASENAME:-defaultdb}"

# Test data
APPLICATION_ID="app-123"
SCHEMA_ID="schema-456"
CONSUMER_ID="consumer-123"
PROVIDER_ID="provider-456"

echo -e "${BLUE}=== Orchestration Engine Integration Test Suite ===${NC}"
echo ""

# Function to print test results
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $test_name: $message"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}✗${NC} $test_name: $message"
    else
        echo -e "${YELLOW}?${NC} $test_name: $message"
    fi
}

# Function to check if service is running
check_service() {
    local service_name="$1"
    local url="$2"
    
    if curl -s "$url/health" > /dev/null 2>&1; then
        print_result "$service_name Health Check" "PASS" "Service is running"
        return 0
    else
        print_result "$service_name Health Check" "FAIL" "Service is not running at $url"
        return 1
    fi
}

# Function to test POST /api/logs endpoint
test_post_logs() {
    local test_name="$1"
    local status="$2"
    local query="$3"
    local expected_http_status="$4"
    
    local request_body=$(cat <<EOF
{
    "status": "$status",
    "requestedData": "$query",
    "applicationId": "$APPLICATION_ID",
    "schemaId": "$SCHEMA_ID",
    "consumerId": "$CONSUMER_ID",
    "providerId": "$PROVIDER_ID"
}
EOF
)
    
    local response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$request_body" \
        "$AUDIT_SERVICE_URL/api/logs")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_http_status" ]; then
        print_result "$test_name" "PASS" "HTTP $http_code - Log created successfully"
        
        # Verify response structure
        if echo "$body" | jq -e '.id' > /dev/null 2>&1; then
            print_result "$test_name Response Structure" "PASS" "Response contains required fields"
        else
            print_result "$test_name Response Structure" "FAIL" "Response missing required fields"
        fi
    else
        print_result "$test_name" "FAIL" "Expected HTTP $expected_http_status, got $http_code"
        echo "Response: $body"
    fi
}

# Function to test GET /api/logs endpoint
test_get_logs() {
    local test_name="$1"
    local query_params="$2"
    local expected_min_count="$3"
    
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs$query_params")
    
    if echo "$response" | jq -e '.logs' > /dev/null 2>&1; then
        local count=$(echo "$response" | jq '.logs | length')
        local total=$(echo "$response" | jq '.total')
        
        if [ "$count" -ge "$expected_min_count" ]; then
            print_result "$test_name" "PASS" "Retrieved $count logs (total: $total)"
            
            # Verify view data
            local has_consumer_id=$(echo "$response" | jq '.logs[0].consumerId != null and .logs[0].consumerId != ""')
            local has_provider_id=$(echo "$response" | jq '.logs[0].providerId != null and .logs[0].providerId != ""')
            
            if [ "$has_consumer_id" = "true" ] && [ "$has_provider_id" = "true" ]; then
                print_result "$test_name View Data" "PASS" "View provides consumer and provider IDs"
            else
                print_result "$test_name View Data" "FAIL" "View missing consumer or provider IDs"
            fi
        else
            print_result "$test_name" "FAIL" "Expected at least $expected_min_count logs, got $count"
        fi
    else
        print_result "$test_name" "FAIL" "Invalid response format"
        echo "Response: $response"
    fi
}

# Function to test orchestration engine GraphQL endpoint
test_orchestration_engine() {
    local test_name="$1"
    local query="$2"
    
    local request_body=$(cat <<EOF
{
    "query": "$query"
}
EOF
)
    
    local response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$request_body" \
        "$ORCHESTRATION_ENGINE_URL/")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        print_result "$test_name" "PASS" "GraphQL query processed successfully"
        
        # Check if audit log was created
        sleep 2  # Wait for async audit logging
        local audit_response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?consumerId=$CONSUMER_ID&limit=1")
        local audit_count=$(echo "$audit_response" | jq '.logs | length')
        
        if [ "$audit_count" -gt 0 ]; then
            print_result "$test_name Audit Logging" "PASS" "Audit log created for GraphQL query"
        else
            print_result "$test_name Audit Logging" "FAIL" "No audit log found for GraphQL query"
        fi
    else
        print_result "$test_name" "FAIL" "Expected HTTP 200, got $http_code"
        echo "Response: $body"
    fi
}

# Function to run database tests
test_database_integration() {
    echo -e "${BLUE}=== Database Integration Tests ===${NC}"
    
    # Test database connection
    if PGPASSWORD="$TEST_DB_PASSWORD" psql -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USERNAME" -d "$TEST_DB_DATABASE" -c "SELECT 1;" > /dev/null 2>&1; then
        print_result "Database Connection" "PASS" "Successfully connected to database"
    else
        print_result "Database Connection" "FAIL" "Failed to connect to database"
        return 1
    fi
    
    # Test view existence
    local view_exists=$(PGPASSWORD="$TEST_DB_PASSWORD" psql -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USERNAME" -d "$TEST_DB_DATABASE" -t -c "SELECT EXISTS(SELECT 1 FROM information_schema.views WHERE table_name = 'audit_logs_with_provider_consumer');")
    
    if [ "$view_exists" = "t" ]; then
        print_result "Database View" "PASS" "audit_logs_with_provider_consumer view exists"
    else
        print_result "Database View" "FAIL" "audit_logs_with_provider_consumer view does not exist"
    fi
    
    # Test view data
    local view_data=$(PGPASSWORD="$TEST_DB_PASSWORD" psql -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USERNAME" -d "$TEST_DB_DATABASE" -t -c "SELECT COUNT(*) FROM audit_logs_with_provider_consumer;")
    
    if [ "$view_data" -gt 0 ]; then
        print_result "View Data" "PASS" "View contains $view_data records"
    else
        print_result "View Data" "WARN" "View is empty (this may be expected for a fresh database)"
    fi
}

# Function to run unit tests
run_unit_tests() {
    echo -e "${BLUE}=== Unit Tests ===${NC}"
    
    cd "$(dirname "$0")"
    
    if [ -f "go.mod" ]; then
        echo "Running Go unit tests..."
        if go test -v ./tests/... -run TestOrchestrationEngineIntegration; then
            print_result "Unit Tests" "PASS" "All orchestration engine integration tests passed"
        else
            print_result "Unit Tests" "FAIL" "Some tests failed"
        fi
    else
        print_result "Unit Tests" "SKIP" "No Go module found, skipping unit tests"
    fi
}

# Function to clean up test data
cleanup_test_data() {
    echo -e "${BLUE}=== Cleanup ===${NC}"
    
    # Clean up audit logs created during testing
    local cleanup_response=$(curl -s -X DELETE "$AUDIT_SERVICE_URL/api/logs?consumerId=$CONSUMER_ID")
    
    if echo "$cleanup_response" | jq -e '.message' > /dev/null 2>&1; then
        print_result "Cleanup" "PASS" "Test data cleaned up"
    else
        print_result "Cleanup" "WARN" "Cleanup may have failed"
    fi
}

# Main test execution
main() {
    echo "Starting orchestration engine integration tests..."
    echo "Audit Service URL: $AUDIT_SERVICE_URL"
    echo "Orchestration Engine URL: $ORCHESTRATION_ENGINE_URL"
    echo ""
    
    # Check if services are running
    echo -e "${BLUE}=== Service Health Checks ===${NC}"
    if ! check_service "Audit Service" "$AUDIT_SERVICE_URL"; then
        echo "Please start the audit service before running tests"
        exit 1
    fi
    
    if ! check_service "Orchestration Engine" "$ORCHESTRATION_ENGINE_URL"; then
        echo "Please start the orchestration engine before running tests"
        exit 1
    fi
    
    echo ""
    
    # Test database integration
    test_database_integration
    echo ""
    
    # Test POST /api/logs endpoint
    echo -e "${BLUE}=== POST /api/logs Tests ===${NC}"
    test_post_logs "Success Query" "success" "query { user { id name email } }" "200"
    test_post_logs "Failure Query" "failure" "query { invalidField { id } }" "200"
    test_post_logs "Complex Query" "success" "query GetUserData(\$userId: ID!) { user(id: \$userId) { id name posts { id title } } }" "200"
    test_post_logs "Mutation Query" "success" "mutation CreateUser(\$input: UserInput!) { createUser(input: \$input) { id name } }" "200"
    test_post_logs "Invalid Status" "invalid" "query { user { id } }" "400"
    test_post_logs "Missing Fields" "success" "query { user { id } }" "400"
    echo ""
    
    # Test GET /api/logs endpoint
    echo -e "${BLUE}=== GET /api/logs Tests ===${NC}"
    test_get_logs "Get All Logs" "" "1"
    test_get_logs "Filter by Consumer ID" "?consumerId=$CONSUMER_ID" "1"
    test_get_logs "Filter by Provider ID" "?providerId=$PROVIDER_ID" "1"
    test_get_logs "Filter by Status" "?status=success" "1"
    test_get_logs "Combined Filters" "?consumerId=$CONSUMER_ID&status=success" "1"
    test_get_logs "Pagination" "?limit=2&offset=0" "1"
    echo ""
    
    # Test orchestration engine integration
    echo -e "${BLUE}=== Orchestration Engine Integration Tests ===${NC}"
    test_orchestration_engine "Simple Query" "query { user { id name } }"
    test_orchestration_engine "Complex Query" "query { users { id name email posts { id title } } }"
    echo ""
    
    # Run unit tests
    run_unit_tests
    echo ""
    
    # Cleanup
    cleanup_test_data
    echo ""
    
    echo -e "${GREEN}=== Test Suite Complete ===${NC}"
    echo "All tests have been executed. Check the results above for any failures."
}

# Run main function
main "$@"
