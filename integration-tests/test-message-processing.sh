#!/bin/bash

# Test script to verify complete message processing flow
# Orchestration Engine -> Redis -> Audit Service -> Database

set -e

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

# Test 1: Verify services are running
test_services_running() {
    local orchestration_health=$(curl -s "$ORCHESTRATION_ENGINE_URL/health" | jq -r '.message // .status // "unknown"')
    local audit_health=$(curl -s "$AUDIT_SERVICE_URL/health" | jq -r '.status // "unknown"')
    
    if [[ "$orchestration_health" == *"Healthy"* ]] && [[ "$audit_health" == *"healthy"* ]]; then
        return 0
    else
        log_error "Services health check failed - Orchestration: $orchestration_health, Audit: $audit_health"
        return 1
    fi
}

# Test 2: Get initial database count
get_initial_db_count() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs")
    local total=$(echo "$response" | jq -r '.total // 0')
    echo "$total"
}

# Test 3: Send request to orchestration engine
send_test_request() {
    local test_id="processing-test-$(date +%s)"
    local response=$(curl -s -X POST "$ORCHESTRATION_ENGINE_URL/" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $test_id" \
        -d '{
            "query": "query { person(nic: \"123456789V\") { fullName address } }"
        }')
    
    if [[ -n "$response" ]]; then
        log_info "Request sent successfully"
        return 0
    else
        return 1
    fi
}

# Test 4: Verify message appears in Redis
test_redis_message_created() {
    local max_attempts=5
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

# Test 5: Wait for message processing and verify database update
test_message_processing() {
    local initial_count=$1
    local max_attempts=20
    local attempt=0
    
    log_info "Initial database count: $initial_count"
    
    while [[ $attempt -lt $max_attempts ]]; do
        sleep 2
        
        # Check if Redis stream is empty (messages processed)
        local stream_length=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT xlen $TEST_STREAM)
        local pending_count=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT xpending $TEST_STREAM $TEST_GROUP 2>/dev/null | head -n1 || echo "0")
        
        # Get current database count
        local current_count=$(get_initial_db_count)
        
        log_info "Attempt $((attempt + 1))/$max_attempts - Stream: $stream_length, Pending: $pending_count, DB: $current_count"
        
        # Check if new records were added to database
        if [[ "$current_count" -gt "$initial_count" ]]; then
            local new_records=$((current_count - initial_count))
            log_info "Found $new_records new audit records in database!"
            return 0
        fi
        
        # If Redis stream is empty and no pending messages, processing is complete
        if [[ "$stream_length" -eq 0 ]] && [[ "$pending_count" -eq 0 ]]; then
            log_info "Redis stream is empty and no pending messages"
            # Give it one more check for database
            sleep 2
            local final_count=$(get_initial_db_count)
            if [[ "$final_count" -gt "$initial_count" ]]; then
                local new_records=$((final_count - initial_count))
                log_info "Found $new_records new audit records in database!"
                return 0
            fi
        fi
        
        attempt=$((attempt + 1))
    done
    
    log_warning "Message processing may not be working - no new database records found"
    return 1
}

# Test 6: Verify audit log content
test_audit_log_content() {
    local response=$(curl -s "$AUDIT_SERVICE_URL/api/logs?limit=1")
    local has_required_fields=$(echo "$response" | jq -r '.logs[0] | has("id") and has("timestamp") and has("status") and has("requestedData")')
    
    if [[ "$has_required_fields" == "true" ]]; then
        log_info "Audit log contains required fields"
        return 0
    else
        log_error "Audit log missing required fields"
        return 1
    fi
}

# Test 7: Test multiple requests processing
test_multiple_requests() {
    local initial_count=$(get_initial_db_count)
    local success_count=0
    local total_requests=3
    
    log_info "Sending $total_requests requests..."
    
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
        log_info "All $total_requests requests sent successfully"
        
        # Wait for processing
        sleep 5
        
        # Check if new records were added
        local final_count=$(get_initial_db_count)
        if [[ "$final_count" -gt "$initial_count" ]]; then
            local new_records=$((final_count - initial_count))
            log_info "Found $new_records new records from multiple requests"
            return 0
        else
            log_warning "No new records found from multiple requests"
            return 1
        fi
    else
        log_error "Only $success_count out of $total_requests requests were sent"
        return 1
    fi
}

# Main test execution
main() {
    log_info "Starting Message Processing Verification Test"
    log_info "=============================================="
    
    # Get initial database count
    local initial_count=$(get_initial_db_count)
    log_info "Initial database count: $initial_count"
    
    # Run tests
    run_test "Services Running" "test_services_running"
    run_test "Send Test Request" "send_test_request"
    run_test "Redis Message Created" "test_redis_message_created"
    run_test "Message Processing" "test_message_processing $initial_count"
    run_test "Audit Log Content" "test_audit_log_content"
    run_test "Multiple Requests" "test_multiple_requests"
    
    # Results
    log_info "=============================================="
    log_info "Test Results Summary:"
    log_info "Total Tests: $TOTAL_TESTS"
    log_success "Passed: $TESTS_PASSED"
    if [[ $TESTS_FAILED -gt 0 ]]; then
        log_error "Failed: $TESTS_FAILED"
    else
        log_info "Failed: $TESTS_FAILED"
    fi
    
    # Final database count
    local final_count=$(get_initial_db_count)
    local total_new_records=$((final_count - initial_count))
    log_info "Final database count: $final_count"
    log_info "Total new records created: $total_new_records"
    
    # Exit code
    if [[ $TESTS_FAILED -eq 0 ]]; then
        log_success "All message processing tests passed! 🎉"
        exit 0
    else
        log_error "Some message processing tests failed."
        exit 1
    fi
}

# Run main function
main "$@"
