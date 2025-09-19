#!/bin/bash

# Common Functions and Utilities for API Server Test Scripts
# This file contains shared functions to eliminate code duplication

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default configuration
DEFAULT_API_BASE_URL="http://localhost:3000"
DEFAULT_ORCHESTRATION_ENGINE_URL="http://localhost:4000"
DEFAULT_CONSENT_ENGINE_URL="http://localhost:8081"
DEFAULT_PDP_URL="http://localhost:8082"
DEFAULT_ASGARDEO_BASE_URL="https://api.asgardeo.io/t/lankasoftwarefoundation"

# Global variables
API_BASE_URL=${API_BASE_URL:-$DEFAULT_API_BASE_URL}
ORCHESTRATION_ENGINE_URL=${ORCHESTRATION_ENGINE_URL:-$DEFAULT_ORCHESTRATION_ENGINE_URL}
CONSENT_ENGINE_URL=${CONSENT_ENGINE_URL:-$DEFAULT_CONSENT_ENGINE_URL}
PDP_URL=${PDP_URL:-$DEFAULT_PDP_URL}
ASGARDEO_BASE_URL=${ASGARDEO_BASE_URL:-$DEFAULT_ASGARDEO_BASE_URL}

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Print functions
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_success() {
    print_status $GREEN "✅ $1"
}

print_error() {
    print_status $RED "❌ $1"
}

print_warning() {
    print_status $YELLOW "⚠️  $1"
}

print_info() {
    print_status $BLUE "ℹ️  $1"
}

print_header() {
    local title=$1
    local color=${2:-$BLUE}
    echo ""
    print_status $color "=========================================="
    print_status $color "$title"
    print_status $color "=========================================="
    echo ""
}

print_step() {
    local step_num=$1
    local step_desc=$2
    print_status $PURPLE "STEP $step_num: $step_desc"
    echo ""
}

print_substep() {
    local step_num=$1
    local step_desc=$2
    print_status $CYAN "  $step_num. $step_desc"
}

# HTTP request functions
make_request() {
    local method=$1
    local url=$2
    local headers=$3
    local data=$4
    local timeout=${5:-10}
    
    if [ -n "$data" ]; then
        curl -s --max-time $timeout -X $method "$url" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data" 2>/dev/null || echo '{"error": "Connection failed"}'
    else
        curl -s --max-time $timeout -X $method "$url" \
            $headers 2>/dev/null || echo '{"error": "Connection failed"}'
    fi
}

make_post_request() {
    local url=$1
    local data=$2
    local headers=${3:-""}
    make_request "POST" "$url" "$headers" "$data"
}

make_get_request() {
    local url=$1
    local headers=${2:-""}
    make_request "GET" "$url" "$headers"
}

make_put_request() {
    local url=$1
    local data=$2
    local headers=${3:-""}
    make_request "PUT" "$url" "$headers" "$data"
}

make_delete_request() {
    local url=$1
    local headers=${2:-""}
    make_request "DELETE" "$url" "$headers"
}

# JSON parsing functions
extract_json_field() {
    local json=$1
    local field=$2
    echo "$json" | jq -r ".$field // empty" 2>/dev/null || echo ""
}

extract_json_array() {
    local json=$1
    local field=$2
    echo "$json" | jq -r ".$field[] // empty" 2>/dev/null || echo ""
}

is_json_valid() {
    local json=$1
    echo "$json" | jq . >/dev/null 2>&1
}

# Health check functions
check_service_health() {
    local service_name=$1
    local url=$2
    local required=${3:-true}
    
    print_substep "1" "Checking $service_name"
    
    local response=$(make_get_request "$url/health" 2>/dev/null)
    
    if echo "$response" | grep -q "healthy\|Healthy\|OK\|ok" 2>/dev/null; then
        print_success "$service_name is running"
        return 0
    else
        if [ "$required" = "true" ]; then
            print_error "$service_name is not running. Please start it first."
            return 1
        else
            print_warning "$service_name is not running (optional for this test)"
            return 0
        fi
    fi
}

check_all_services() {
    print_header "Service Health Checks"
    
    local all_healthy=true
    
    check_service_health "API Server" "$API_BASE_URL" true || all_healthy=false
    check_service_health "Orchestration Engine" "$ORCHESTRATION_ENGINE_URL" true || all_healthy=false
    check_service_health "Consent Engine" "$CONSENT_ENGINE_URL" false
    check_service_health "Policy Decision Point" "$PDP_URL" false
    
    echo ""
    
    if [ "$all_healthy" = "false" ]; then
        print_error "Required services are not running. Please start them first."
        exit 1
    fi
}

# Test result tracking
increment_test_count() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

record_test_success() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    increment_test_count
}

record_test_failure() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    increment_test_count
}

record_test_skip() {
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
    increment_test_count
}

# Test assertion functions
assert_equals() {
    local expected=$1
    local actual=$2
    local message=$3
    
    if [ "$expected" = "$actual" ]; then
        print_success "$message"
        return 0
    else
        print_error "$message (expected: $expected, got: $actual)"
        return 1
    fi
}

assert_contains() {
    local text=$1
    local pattern=$2
    local message=$3
    
    if echo "$text" | grep -q "$pattern" 2>/dev/null; then
        print_success "$message"
        return 0
    else
        print_error "$message (pattern not found: $pattern)"
        return 1
    fi
}

assert_json_field() {
    local json=$1
    local field=$2
    local expected=$3
    local message=$4
    
    local actual=$(extract_json_field "$json" "$field")
    assert_equals "$expected" "$actual" "$message"
}

assert_json_contains() {
    local json=$1
    local field=$2
    local pattern=$3
    local message=$4
    
    local value=$(extract_json_field "$json" "$field")
    assert_contains "$value" "$pattern" "$message"
}

# Consumer management functions
create_test_consumer() {
    local consumer_name=${1:-"Test Consumer $(date +%s)"}
    local email=${2:-"test@example.com"}
    local phone=${3:-"+1-555-0123"}
    
    print_substep "1" "Creating test consumer: $consumer_name"
    
    local consumer_data=$(cat <<EOF
{
    "consumerName": "$consumer_name",
    "contactEmail": "$email",
    "phoneNumber": "$phone"
}
EOF
)
    
    local response=$(make_post_request "$API_BASE_URL/consumers" "$consumer_data")
    local consumer_id=$(extract_json_field "$response" "consumerId")
    
    if [ -n "$consumer_id" ] && [ "$consumer_id" != "null" ]; then
        print_success "Consumer created: $consumer_id"
        echo "$consumer_id"
    else
        print_error "Failed to create consumer"
        echo "Response: $response"
        return 1
    fi
}

create_consumer_application() {
    local consumer_id=$1
    local app_name=${2:-"Test Application"}
    local description=${3:-"Test application for integration testing"}
    
    print_substep "2" "Creating consumer application: $app_name"
    
    local app_data=$(cat <<EOF
{
    "consumerId": "$consumer_id",
    "requiredFields": {
        "person.fullName": true,
        "person.email": true
    }
}
EOF
)
    
    local response=$(make_post_request "$API_BASE_URL/consumer-applications" "$app_data")
    local submission_id=$(extract_json_field "$response" "submissionId")
    
    if [ -n "$submission_id" ] && [ "$submission_id" != "null" ]; then
        print_success "Application created: $submission_id"
        echo "$submission_id"
    else
        print_error "Failed to create consumer application"
        echo "Response: $response"
        return 1
    fi
}

approve_consumer_application() {
    local submission_id=$1
    
    print_substep "3" "Approving consumer application: $submission_id"
    
    local approval_data='{"status": "approved"}'
    local response=$(make_put_request "$API_BASE_URL/consumer-applications/$submission_id" "$approval_data")
    
    local api_key=$(extract_json_field "$response" "credentials.apiKey")
    local api_secret=$(extract_json_field "$response" "credentials.apiSecret")
    
    if [ -n "$api_key" ] && [ -n "$api_secret" ] && [ "$api_key" != "null" ] && [ "$api_secret" != "null" ]; then
        print_success "Application approved and credentials generated"
        echo "API_KEY:$api_key"
        echo "API_SECRET:$api_secret"
    else
        print_error "Failed to approve application or generate credentials"
        echo "Response: $response"
        return 1
    fi
}

# GraphQL testing functions
test_graphql_query() {
    local query=$1
    local token=$2
    local expected_pattern=${3:-"errors\\|data"}
    
    print_substep "1" "Testing GraphQL query"
    
    local graphql_data=$(cat <<EOF
{
    "query": "$query"
}
EOF
)
    
    local auth_header=""
    if [ -n "$token" ]; then
        auth_header="-H \"Authorization: Bearer $token\""
    fi
    
    local response=$(make_post_request "$ORCHESTRATION_ENGINE_URL/" "$graphql_data" "$auth_header")
    
    if echo "$response" | grep -q "$expected_pattern" 2>/dev/null; then
        print_success "GraphQL request processed successfully"
    else
        print_warning "GraphQL request may have failed"
    fi
    
    echo "RESPONSE:$response"
}

# Security testing functions
test_security_headers() {
    local url=$1
    local headers=(
        "X-Content-Type-Options: nosniff"
        "X-Frame-Options: DENY"
        "X-XSS-Protection: 1; mode=block"
        "Referrer-Policy: strict-origin-when-cross-origin"
        "Content-Security-Policy: default-src 'self'"
    )
    
    print_substep "1" "Testing security headers"
    
    local response=$(curl -s -I "$url" 2>/dev/null)
    
    for header in "${headers[@]}"; do
        if echo "$response" | grep -q "$header" 2>/dev/null; then
            print_success "$header"
        else
            print_error "Missing: $header"
        fi
    done
}

test_input_validation() {
    local base_url=$1
    
    print_substep "2" "Testing input validation"
    
    # Test path traversal
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$base_url/../etc/passwd" 2>/dev/null)
    if [ "$response" = "400" ]; then
        print_success "Path traversal blocked"
    else
        print_error "Path traversal not blocked (status: $response)"
    fi
    
    # Test XSS attempt
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$base_url/<script>alert('xss')</script>" 2>/dev/null)
    if [ "$response" = "400" ]; then
        print_success "XSS attempt blocked"
    else
        print_error "XSS attempt not blocked (status: $response)"
    fi
    
    # Test invalid content type
    local response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$base_url/consumers" \
        -H "Content-Type: text/plain" \
        -d "test" 2>/dev/null)
    if [ "$response" = "400" ]; then
        print_success "Invalid content type blocked"
    else
        print_error "Invalid content type not blocked (status: $response)"
    fi
}

# Cleanup functions
cleanup_test_data() {
    local consumer_id=$1
    
    if [ -n "$consumer_id" ]; then
        print_info "Cleaning up test data: $consumer_id"
        make_delete_request "$API_BASE_URL/consumers/$consumer_id" >/dev/null 2>&1 || true
        print_success "Test data cleaned up"
    fi
}

# Summary functions
show_test_summary() {
    local test_name=$1
    
    print_header "Test Summary: $test_name"
    print_info "Total Tests: $TOTAL_TESTS"
    print_success "Passed: $PASSED_TESTS"
    print_error "Failed: $FAILED_TESTS"
    print_warning "Skipped: $SKIPPED_TESTS"
    
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        print_success "All tests completed successfully!"
        return 0
    else
        print_error "Some tests failed. Please check the output above."
        return 1
    fi
}

# Utility functions
generate_test_id() {
    echo "test-$(date +%s)-$$"
}

wait_for_service() {
    local url=$1
    local max_attempts=${2:-30}
    local delay=${3:-1}
    
    print_info "Waiting for service at $url..."
    
    for i in $(seq 1 $max_attempts); do
        if curl -s "$url/health" >/dev/null 2>&1; then
            print_success "Service is ready"
            return 0
        fi
        sleep $delay
    done
    
    print_error "Service did not start within $((max_attempts * delay)) seconds"
    return 1
}

# Export functions for use in other scripts
export -f print_status print_success print_error print_warning print_info print_header print_step print_substep
export -f make_request make_post_request make_get_request make_put_request make_delete_request
export -f extract_json_field extract_json_array is_json_valid
export -f check_service_health check_all_services
export -f increment_test_count record_test_success record_test_failure record_test_skip
export -f assert_equals assert_contains assert_json_field assert_json_contains
export -f create_test_consumer create_consumer_application approve_consumer_application
export -f exchange_credentials_for_token validate_token
export -f test_graphql_query test_security_headers test_input_validation
export -f cleanup_test_data show_test_summary generate_test_id wait_for_service
