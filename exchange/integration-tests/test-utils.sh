#!/bin/bash
# Shared test utilities for integration tests

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Configuration
API_SERVER_URL="http://localhost:3000"
PDP_URL="http://localhost:8082"
CE_URL="http://localhost:8081"

# Common test data
TEST_CONSUMER_DATA='{
  "consumerName": "Test Consumer",
  "contactEmail": "test@example.com",
  "phoneNumber": "123-456-7890"
}'

TEST_PROVIDER_DATA='{
  "providerName": "Test Provider",
  "contactEmail": "provider@example.com",
  "phoneNumber": "987-654-3210",
  "providerType": "government"
}'

TEST_SCHEMA_DATA='{
  "sdl": "directive @accessControl(type: String!) on FIELD_DEFINITION\n\ndirective @source(value: String!) on FIELD_DEFINITION\n\ndirective @isOwner(value: Boolean!) on FIELD_DEFINITION\n\ndirective @description(value: String!) on FIELD_DEFINITION\n\ntype User {\n  id: ID! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  name: String! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  email: String! @accessControl(type: \"restricted\") @source(value: \"authoritative\") @isOwner(value: false)\n  phone: String! @accessControl(type: \"restricted\") @source(value: \"authoritative\") @isOwner(value: false)\n}\n\ntype Query {\n  getUser(id: ID!): User @description(value: \"Get user by ID\")\n  listUsers: [User!]! @description(value: \"List all users\")\n}"
}'

# Utility functions
log_info() {
    echo -e "${BLUE}$1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

log_test() {
    echo -e "${PURPLE}Test: $1${NC}"
}

# Test API endpoint with full response handling
test_api_endpoint() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    local extract_id="$6"  # Optional: field to extract from response
    
    log_test "$test_name"
    echo "Method: $method $endpoint"
    echo "Expected Status: $expected_status"
    echo ""
    
    if [ -n "$data" ]; then
        RESPONSE=$(curl -s -w "\n%{http_code}" -X "$method" "$API_SERVER_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        RESPONSE=$(curl -s -w "\n%{http_code}" -X "$method" "$API_SERVER_URL$endpoint")
    fi
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    echo "Response Status: $HTTP_CODE"
    echo "Response Body:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        log_success "PASSED"
        if [ -n "$extract_id" ]; then
            EXTRACTED_ID=$(echo "$BODY" | jq -r ".$extract_id")
            if [ "$EXTRACTED_ID" != "null" ] && [ -n "$EXTRACTED_ID" ]; then
                echo "$EXTRACTED_ID"
                return 0
            else
                log_error "Failed to extract $extract_id from response"
                return 1
            fi
        fi
        return 0
    else
        log_error "FAILED - Expected $expected_status, got $HTTP_CODE"
        return 1
    fi
}

# Test PDP endpoint
test_pdp_endpoint() {
    local test_name="$1"
    local expected="$2"
    local data="$3"
    
    log_test "$test_name"
    echo "Expected: $expected"
    echo ""
    
    RESPONSE=$(curl -s -X POST "$PDP_URL/decide" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    echo "PDP Decision:"
    echo "$RESPONSE" | jq '.'
    echo "---"
}

# Test Consent Engine endpoint
test_ce_endpoint() {
    local test_name="$1"
    local data="$2"
    
    log_test "$test_name"
    echo ""
    
    RESPONSE=$(curl -s -X POST "$CE_URL/consent" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    echo "Consent Engine Response:"
    echo "$RESPONSE" | jq '.'
    echo "---"
}

# Health check functions
check_service_health() {
    local service_name="$1"
    local url="$2"
    local port="$3"
    
    log_info "Checking $service_name on port $port..."
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$url/health" 2>/dev/null || echo "000")
    
    if [ "$STATUS" = "200" ]; then
        log_success "$service_name is running (HTTP $STATUS)"
        return 0
    else
        log_error "$service_name is not responding (HTTP $STATUS)"
        return 1
    fi
}

# Create test consumer and return ID
create_test_consumer() {
    local consumer_name="${1:-Test Consumer}"
    local email="${2:-test@example.com}"
    local phone="${3:-123-456-7890}"
    
    local data="{
        \"consumerName\": \"$consumer_name\",
        \"contactEmail\": \"$email\",
        \"phoneNumber\": \"$phone\"
    }"
    
    RESPONSE=$(curl -s -X POST "$API_SERVER_URL/consumers" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    echo "$RESPONSE" | jq -r '.consumerId'
}

# Create test provider and return ID
create_test_provider() {
    local provider_name="${1:-Test Provider}"
    local email="${2:-provider@example.com}"
    local phone="${3:-987-654-3210}"
    local type="${4:-government}"
    
    # Create submission
    local submission_data="{
        \"providerName\": \"$provider_name\",
        \"contactEmail\": \"$email\",
        \"phoneNumber\": \"$phone\",
        \"providerType\": \"$type\"
    }"
    
    SUBMISSION_RESPONSE=$(curl -s -X POST "$API_SERVER_URL/provider-submissions" \
        -H "Content-Type: application/json" \
        -d "$submission_data")
    
    SUBMISSION_ID=$(echo "$SUBMISSION_RESPONSE" | jq -r '.submissionId')
    
    if [ "$SUBMISSION_ID" != "null" ] && [ -n "$SUBMISSION_ID" ]; then
        # Approve submission
        local approval_data='{"status": "approved"}'
        APPROVAL_RESPONSE=$(curl -s -X PUT "$API_SERVER_URL/provider-submissions/$SUBMISSION_ID" \
            -H "Content-Type: application/json" \
            -d "$approval_data")
        
        echo "$APPROVAL_RESPONSE" | jq -r '.providerId'
    else
        echo "null"
    fi
}

# Create test schema and return ID
create_test_schema() {
    local provider_id="$1"
    local sdl="${2:-$TEST_SCHEMA_DATA}"
    
    # Create schema submission
    local schema_data="{\"sdl\": \"$sdl\"}"
    
    SCHEMA_RESPONSE=$(curl -s -X POST "$API_SERVER_URL/providers/$provider_id/schema-submissions" \
        -H "Content-Type: application/json" \
        -d "$schema_data")
    
    SCHEMA_ID=$(echo "$SCHEMA_RESPONSE" | jq -r '.submissionId')
    
    if [ "$SCHEMA_ID" != "null" ] && [ -n "$SCHEMA_ID" ]; then
        # Submit for review
        curl -s -X PUT "$API_SERVER_URL/providers/$provider_id/schema-submissions/$SCHEMA_ID" \
            -H "Content-Type: application/json" \
            -d '{"status": "pending"}' > /dev/null
        
        # Approve schema
        curl -s -X PUT "$API_SERVER_URL/providers/$provider_id/schema-submissions/$SCHEMA_ID" \
            -H "Content-Type: application/json" \
            -d '{"status": "approved"}' > /dev/null
        
        echo "$SCHEMA_ID"
    else
        echo "null"
    fi
}

# Cleanup function
cleanup_test_data() {
    log_info "Cleaning up test data..."
    # This would implement cleanup logic if needed
    log_success "Cleanup complete"
}
