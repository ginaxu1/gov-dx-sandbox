#!/bin/bash

# Integration Test Script for Audit Middleware
# Tests the complete flow: Orchestration Engine -> Redis -> Audit Service -> Database

# set -e  # Exit on any error - commented out for better error handling

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ORCHESTRATION_ENGINE_URL="http://localhost:4000"
AUDIT_SERVICE_URL="http://localhost:3001"
REDIS_HOST="localhost"
REDIS_PORT="6379"
TEST_STREAM="audit-events"
TEST_GROUP="audit-processors"

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

# Cleanup function
cleanup() {
    log_info "Cleaning up test data..."
    # Clear Redis stream
    redis-cli -h $REDIS_HOST -p $REDIS_PORT del $TEST_STREAM 2>/dev/null || true
    # Clear any pending messages
    redis-cli -h $REDIS_HOST -p $REDIS_PORT xgroup destroy $TEST_STREAM $TEST_GROUP 2>/dev/null || true
}

# Setup function
setup() {
    log_info "Setting up integration test environment..."
    
    # Check if Redis is running
    if ! redis-cli -h $REDIS_HOST -p $REDIS_PORT ping >/dev/null 2>&1; then
        log_error "Redis is not running on $REDIS_HOST:$REDIS_PORT"
        exit 1
    fi
    
    # Check if orchestration engine is running
    if ! curl -s "$ORCHESTRATION_ENGINE_URL/health" >/dev/null 2>&1; then
        log_error "Orchestration engine is not running on $ORCHESTRATION_ENGINE_URL"
        exit 1
    fi
    
    # Check if audit service is running
    if ! curl -s "$AUDIT_SERVICE_URL/health" >/dev/null 2>&1; then
        log_error "Audit service is not running on $AUDIT_SERVICE_URL"
        exit 1
    fi
    
    log_success "All services are running"
}

# Test 1: Verify services are healthy
test_services_health() {
    local orchestration_health=$(curl -s "$ORCHESTRATION_ENGINE_URL/health" | jq -r '.message // .status // "unknown"')
    local audit_health=$(curl -s "$AUDIT_SERVICE_URL/health" | jq -r '.status // "unknown"')
    
    if [[ "$orchestration_health" == *"Healthy"* ]] && [[ "$audit_health" == *"healthy"* ]]; then
        return 0
    else
        log_error "Services health check failed - Orchestration: $orchestration_health, Audit: $audit_health"
        return 1
    fi
}

# Test 2: Verify Redis connection and stream setup
test_redis_setup() {
    # Check if Redis is accessible
    if ! redis-cli -h $REDIS_HOST -p $REDIS_PORT ping >/dev/null 2>&1; then
        return 1
    fi
    
    # Check if audit stream exists
    local stream_exists=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT exists $TEST_STREAM)
    if [[ "$stream_exists" == "0" ]]; then
        log_warning "Audit stream does not exist yet, will be created on first message"
    fi
    
    return 0
}

# Test 3: Send test request to orchestration engine
test_orchestration_request() {
    local test_id=$(date +%s)
    local response=$(curl -s -X POST "$ORCHESTRATION_ENGINE_URL/" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-token-$test_id" \
        -d '{
            "query": "query { person(nic: \"123456789V\") { fullName address } }"
        }')
    
    # Check if we got a response (even if it's an error, that's expected)
    if [[ -n "$response" ]]; then
        log_info "Orchestration engine responded: $response"
        return 0
    else
        return 1
    fi
}

# Test 4: Verify audit message was sent to Redis
test_redis_message() {
    local max_attempts=10
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        local stream_length=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT xlen $TEST_STREAM)
        if [[ "$stream_length" -gt 0 ]]; then
            log_info "Found $stream_length messages in Redis stream"
            return 0
        fi
        
        attempt=$((attempt + 1))
        sleep 1
    done
    
    log_error "No messages found in Redis stream after $max_attempts attempts"
    return 1
}

# Test 5: Verify audit message content
test_audit_message_content() {
    local messages=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT xrange $TEST_STREAM - +)
    
    if [[ -z "$messages" ]]; then
        log_error "No messages found in Redis stream"
        return 1
    fi
    
    # Check if message contains expected fields
    local message_content=$(echo "$messages" | tail -n +2)  # Skip the message ID line
    
    if [[ "$message_content" == *"status"* ]] && [[ "$message_content" == *"requested_data"* ]]; then
        log_info "Audit message contains expected fields"
        return 0
    else
        log_error "Audit message missing expected fields: $message_content"
        return 1
    fi
}

# Test 6: Verify audit service is processing messages
test_audit_service_processing() {
    local max_attempts=15
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        # Check pending messages
        local pending=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT xpending $TEST_STREAM $TEST_GROUP)
        
        if [[ "$pending" == "0" ]]; then
            log_info "All messages have been processed by audit service"
            return 0
        fi
        
        attempt=$((attempt + 1))
        sleep 2
    done
    
    log_warning "Some messages are still pending after $max_attempts attempts"
    return 1
}

# Test 7: Verify audit logs in database
test_audit_logs_database() {
    local max_attempts=10
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?limit=5")
        
        if [[ "$response" != *"Failed to retrieve audit logs"* ]] && [[ "$response" != *"error"* ]]; then
            log_info "Audit logs retrieved successfully"
            return 0
        fi
        
        attempt=$((attempt + 1))
        sleep 2
    done
    
    log_warning "Could not retrieve audit logs from database (this may be expected if database view is missing)"
    return 1
}

# Test 8: Test multiple requests
test_multiple_requests() {
    local success_count=0
    local total_requests=3
    
    for i in $(seq 1 $total_requests); do
        local test_id="multi-test-$i-$(date +%s)"
        local response=$(curl -s -X POST "$ORCHESTRATION_ENGINE_URL/" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $test_id" \
            -d '{
                "query": "query { test { id } }"
            }')
        
        if [[ -n "$response" ]]; then
            success_count=$((success_count + 1))
        fi
        
        sleep 1
    done
    
    if [[ $success_count -eq $total_requests ]]; then
        log_info "All $total_requests requests were processed"
        return 0
    else
        log_error "Only $success_count out of $total_requests requests were processed"
        return 1
    fi
}

# Test 9: Verify Redis stream statistics
test_redis_stream_stats() {
    local stream_info=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT xinfo stream $TEST_STREAM 2>/dev/null)
    
    if [[ -n "$stream_info" ]]; then
        log_info "Redis stream info retrieved successfully"
        return 0
    else
        log_error "Could not retrieve Redis stream info"
        return 1
    fi
}

# Test 10: Test error handling
test_error_handling() {
    # Send malformed request
    local response=$(curl -s -X POST "$ORCHESTRATION_ENGINE_URL/" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer error-test" \
        -d '{"invalid": "json"}' || true)
    
    # Even malformed requests should be processed by audit middleware
    if [[ -n "$response" ]]; then
        log_info "Error handling test completed"
        return 0
    else
        log_warning "Error handling test inconclusive"
        return 1
    fi
}

# Main test execution
main() {
    log_info "Starting Audit Integration Test Suite"
    log_info "====================================="
    
    # Setup
    setup
    
    # Run tests
    run_test "Services Health Check" "test_services_health"
    run_test "Redis Setup" "test_redis_setup"
    run_test "Orchestration Engine Request" "test_orchestration_request"
    run_test "Redis Message Creation" "test_redis_message"
    run_test "Audit Message Content" "test_audit_message_content"
    run_test "Audit Service Processing" "test_audit_service_processing"
    run_test "Audit Logs Database" "test_audit_logs_database"
    run_test "Multiple Requests" "test_multiple_requests"
    run_test "Redis Stream Statistics" "test_redis_stream_stats"
    run_test "Error Handling" "test_error_handling"
    
    # Results
    log_info "====================================="
    log_info "Test Results Summary:"
    log_info "Total Tests: $TOTAL_TESTS"
    log_success "Passed: $TESTS_PASSED"
    if [[ $TESTS_FAILED -gt 0 ]]; then
        log_error "Failed: $TESTS_FAILED"
    else
        log_info "Failed: $TESTS_FAILED"
    fi
    
    # Cleanup
    cleanup
    
    # Exit code
    if [[ $TESTS_FAILED -eq 0 ]]; then
        log_success "All tests passed! 🎉"
        exit 0
    else
        log_error "Some tests failed. Please check the logs above."
        exit 1
    fi
}

# Run main function
main "$@"
