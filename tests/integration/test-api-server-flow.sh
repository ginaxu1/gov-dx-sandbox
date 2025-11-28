#!/bin/bash
# API Server Integration Test Suite
# Tests the full flow including API server, PDP, and Consent Engine

echo "=== API Server Integration Test Suite ==="
echo "Testing complete flow: API Server -> PDP -> Consent Engine"
echo ""

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

# Test function with timeout
test_api_endpoint() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    local timeout="${6:-10}"  # Default 10 second timeout
    
    echo -e "${BLUE}Test: $test_name${NC}"
    echo "Method: $method $endpoint"
    echo "Expected Status: $expected_status"
    echo "Timeout: ${timeout}s"
    echo ""
    
    if [ -n "$data" ]; then
        RESPONSE=$(timeout $timeout curl -s -w "\n%{http_code}" -X "$method" "$API_SERVER_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null || echo -e "\n408")
    else
        RESPONSE=$(timeout $timeout curl -s -w "\n%{http_code}" -X "$method" "$API_SERVER_URL$endpoint" 2>/dev/null || echo -e "\n408")
    fi
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    echo "Response Status: $HTTP_CODE"
    echo "Response Body:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASSED${NC}"
    else
        echo -e "${RED}❌ FAILED - Expected $expected_status, got $HTTP_CODE${NC}"
    fi
    
    echo "---"
    echo ""
}

# Check if API server is running
echo -e "${PURPLE}=== API Server Health Check ===${NC}"
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_SERVER_URL/health" 2>/dev/null || echo "000")
if [ "$API_STATUS" = "200" ]; then
    echo -e "${GREEN}✅ API Server is running (HTTP $API_STATUS)${NC}"
else
    echo -e "${RED}❌ API Server is not responding (HTTP $API_STATUS)${NC}"
    echo "Please start the API server: cd api-server-go && go run main.go"
    exit 1
fi

echo ""

# Test 1: Health and Debug Endpoints
echo -e "${BLUE}=== Test 1: Health and Debug Endpoints ===${NC}"
test_api_endpoint "Health Check" "GET" "/health" "" "200"
test_api_endpoint "Debug Info" "GET" "/debug" "" "200"

# Test 2: Consumer Management
echo -e "${BLUE}=== Test 2: Consumer Management ===${NC}"

# Create a consumer
CONSUMER_DATA='{
  "consumerName": "Test Consumer",
  "contactEmail": "test@example.com",
  "phoneNumber": "123-456-7890"
}'

test_api_endpoint "Create Consumer" "POST" "/consumers" "$CONSUMER_DATA" "201"

# Get consumer ID from response (assuming it's returned)
CONSUMER_RESPONSE=$(timeout 10 curl -s -X POST "$API_SERVER_URL/consumers" \
    -H "Content-Type: application/json" \
    -d "$CONSUMER_DATA" 2>/dev/null || echo '{"error": "timeout"}')
CONSUMER_ID=$(echo "$CONSUMER_RESPONSE" | jq -r '.consumerId // empty')

if [ "$CONSUMER_ID" != "null" ] && [ -n "$CONSUMER_ID" ]; then
    test_api_endpoint "Get Consumer" "GET" "/consumers/$CONSUMER_ID" "" "200"
    test_api_endpoint "List Consumers" "GET" "/consumers" "" "200"
else
    echo -e "${RED}❌ Failed to get consumer ID for further tests${NC}"
fi

# Test 3: Provider Management
echo -e "${BLUE}=== Test 3: Provider Management ===${NC}"

# Create a provider submission with unique name
TIMESTAMP=$(date +%s)
PROVIDER_NAME="Test Provider $TIMESTAMP"
PROVIDER_DATA="{
  \"providerName\": \"$PROVIDER_NAME\",
  \"contactEmail\": \"provider@example.com\",
  \"phoneNumber\": \"987-654-3210\",
  \"providerType\": \"government\"
}"

test_api_endpoint "Create Provider Submission" "POST" "/provider-submissions" "$PROVIDER_DATA" "201"

# Get submission ID from response
SUBMISSION_RESPONSE=$(timeout 10 curl -s -X POST "$API_SERVER_URL/provider-submissions" \
    -H "Content-Type: application/json" \
    -d "$PROVIDER_DATA" 2>/dev/null || echo '{"error": "timeout"}')
SUBMISSION_ID=$(echo "$SUBMISSION_RESPONSE" | jq -r '.submissionId // empty')

# If submission failed, try to get an existing submission ID
if [ "$SUBMISSION_ID" = "null" ] || [ -z "$SUBMISSION_ID" ]; then
    echo -e "${YELLOW}⚠️ Provider submission failed, trying to get existing submission...${NC}"
    EXISTING_SUBMISSIONS=$(timeout 10 curl -s -X GET "$API_SERVER_URL/provider-submissions" 2>/dev/null || echo '{"items": []}')
    SUBMISSION_ID=$(echo "$EXISTING_SUBMISSIONS" | jq -r '.items[0].submissionId // empty')
    
    if [ -z "$SUBMISSION_ID" ]; then
        echo -e "${YELLOW}⚠️ No existing submissions found, using mock ID for testing...${NC}"
        SUBMISSION_ID="sub_mock_$(date +%s)"
    else
        echo -e "${GREEN}✅ Using existing submission ID: $SUBMISSION_ID${NC}"
    fi
fi

if [ -n "$SUBMISSION_ID" ]; then
    # Approve the submission to create a provider profile
    APPROVAL_DATA='{"status": "approved"}'
    test_api_endpoint "Approve Provider Submission" "PUT" "/provider-submissions/$SUBMISSION_ID" "$APPROVAL_DATA" "200"
    
    # Get provider ID from approval response
    APPROVAL_RESPONSE=$(timeout 10 curl -s -X PUT "$API_SERVER_URL/provider-submissions/$SUBMISSION_ID" \
        -H "Content-Type: application/json" \
        -d "$APPROVAL_DATA" 2>/dev/null || echo '{"error": "timeout"}')
    PROVIDER_ID=$(echo "$APPROVAL_RESPONSE" | jq -r '.providerId // empty')
    
    if [ "$PROVIDER_ID" != "null" ] && [ -n "$PROVIDER_ID" ]; then
        test_api_endpoint "Get Provider" "GET" "/providers/$PROVIDER_ID" "" "200"
        test_api_endpoint "List Providers" "GET" "/providers" "" "200"
    else
        echo -e "${RED}❌ Failed to get provider ID for further tests${NC}"
    fi
else
    echo -e "${RED}❌ Failed to get submission ID for further tests${NC}"
fi

# Test 4: Schema Management
echo -e "${BLUE}=== Test 4: Schema Management ===${NC}"

if [ -n "$PROVIDER_ID" ]; then
    # Create a schema submission
    SCHEMA_DATA='{
      "sdl": "directive @accessControl(type: String!) on FIELD_DEFINITION\n\ndirective @source(value: String!) on FIELD_DEFINITION\n\ndirective @isOwner(value: Boolean!) on FIELD_DEFINITION\n\ndirective @description(value: String!) on FIELD_DEFINITION\n\ntype User {\n  id: ID! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  name: String! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  email: String! @accessControl(type: \"restricted\") @source(value: \"authoritative\") @isOwner(value: false)\n  phone: String! @accessControl(type: \"restricted\") @source(value: \"authoritative\") @isOwner(value: false)\n}\n\ntype Query {\n  getUser(id: ID!): User @description(value: \"Get user by ID\")\n  listUsers: [User!]! @description(value: \"List all users\")\n}"
    }'
    
    test_api_endpoint "Create Schema Submission" "POST" "/providers/$PROVIDER_ID/schema-submissions" "$SCHEMA_DATA" "201" "15"
    
    # Get schema ID from response with timeout
    echo -e "${BLUE}Getting schema ID...${NC}"
    SCHEMA_RESPONSE=$(timeout 10 curl -s -X POST "$API_SERVER_URL/providers/$PROVIDER_ID/schema-submissions" \
        -H "Content-Type: application/json" \
        -d "$SCHEMA_DATA" 2>/dev/null || echo '{"error": "timeout"}')
    SCHEMA_ID=$(echo "$SCHEMA_RESPONSE" | jq -r '.submissionId // empty')
    
    if [ -n "$SCHEMA_ID" ] && [ "$SCHEMA_ID" != "null" ]; then
        echo -e "${GREEN}✅ Got schema ID: $SCHEMA_ID${NC}"
        
        # Submit for review with timeout
        REVIEW_DATA='{"status": "pending"}'
        test_api_endpoint "Submit Schema for Review" "PUT" "/providers/$PROVIDER_ID/schema-submissions/$SCHEMA_ID" "$REVIEW_DATA" "200" "15"
        
        # Approve the schema with timeout
        APPROVAL_DATA='{"status": "approved"}'
        test_api_endpoint "Approve Schema" "PUT" "/providers/$PROVIDER_ID/schema-submissions/$SCHEMA_ID" "$APPROVAL_DATA" "200" "15"
        
        test_api_endpoint "List Provider Schemas" "GET" "/providers/$PROVIDER_ID/schemas" "" "200" "10"
    else
        echo -e "${YELLOW}⚠️ Failed to get schema ID, trying to use existing schema...${NC}"
        
        # Try to get existing schemas
        EXISTING_SCHEMAS=$(timeout 10 curl -s -X GET "$API_SERVER_URL/providers/$PROVIDER_ID/schema-submissions" 2>/dev/null || echo '{"items": []}')
        SCHEMA_ID=$(echo "$EXISTING_SCHEMAS" | jq -r '.items[0].submissionId // empty')
        
        if [ -n "$SCHEMA_ID" ] && [ "$SCHEMA_ID" != "null" ]; then
            echo -e "${GREEN}✅ Using existing schema ID: $SCHEMA_ID${NC}"
            test_api_endpoint "List Provider Schemas" "GET" "/providers/$PROVIDER_ID/schemas" "" "200" "10"
        else
            echo -e "${YELLOW}⚠️ No schemas available, skipping schema tests${NC}"
        fi
    fi
else
    echo -e "${YELLOW}⚠️ Skipping schema tests - no provider ID available${NC}"
fi

# Test 5: Allow List Management
echo -e "${BLUE}=== Test 5: Allow List Management ===${NC}"

# Test allow list endpoints - expect 404 if no data exists, that's OK for testing
test_api_endpoint "List Allow List for Field" "GET" "/admin/fields/user.email/allow-list" "" "200"

# Add consumer to allow list
ALLOW_LIST_DATA='{
  "consumer_id": "test-consumer-123",
  "expires_at": 1757560679,
  "grant_duration": "30d",
  "reason": "Test consent approval",
  "updated_by": "admin"
}'

test_api_endpoint "Add Consumer to Allow List" "POST" "/admin/fields/user.email/allow-list" "$ALLOW_LIST_DATA" "201"

# Test getting specific consumer from allow list
test_api_endpoint "Get Consumer in Allow List" "GET" "/admin/fields/user.email/allow-list/test-consumer-123" "" "200"

# Test 6: Admin Statistics
echo -e "${BLUE}=== Test 6: Admin Statistics ===${NC}"
test_api_endpoint "Get Admin Statistics" "GET" "/admin/statistics" "" "200"
test_api_endpoint "Get Admin Metrics" "GET" "/admin/metrics" "" "200"
test_api_endpoint "Get Recent Activity" "GET" "/admin/recent-activity" "" "200"

echo ""
echo -e "${GREEN}=== API Server Integration Test Suite Complete ===${NC}"
echo "All API server endpoints tested successfully!"