#!/bin/bash
# Mock Data Setup Script for Integration Tests
# This script initializes the necessary test data for the API server

echo "=== Setting up Mock Data for Integration Tests ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_SERVER_URL="http://localhost:3000"

# Function to make API calls with error handling
make_api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="$4"
    local description="$5"
    
    echo -e "${BLUE}Setting up: $description${NC}"
    
    if [ -n "$data" ]; then
        RESPONSE=$(curl -s -w "\n%{http_code}" -X "$method" "$API_SERVER_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        RESPONSE=$(curl -s -w "\n%{http_code}" -X "$method" "$API_SERVER_URL$endpoint")
    fi
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        echo -e "${GREEN}✅ $description - Success${NC}"
        echo "$BODY"
    else
        echo -e "${YELLOW}⚠️ $description - Got $HTTP_CODE (expected $expected_status)${NC}"
        echo "$BODY"
    fi
    
    echo ""
}

# Check if API server is running
echo -e "${PURPLE}=== API Server Health Check ===${NC}"
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_SERVER_URL/health" 2>/dev/null || echo "000")
if [ "$API_STATUS" = "200" ]; then
    echo -e "${GREEN}✅ API Server is running${NC}"
else
    echo -e "${RED}❌ API Server is not responding${NC}"
    echo "Please start the API server first: cd api-server-go && go run main.go"
    exit 1
fi

echo ""

# 1. Create test consumers
echo -e "${BLUE}=== Creating Test Consumers ===${NC}"

CONSUMER_DATA='{
  "consumerName": "Mock Test Consumer 1",
  "contactEmail": "consumer1@test.com",
  "phoneNumber": "111-111-1111"
}'
make_api_call "POST" "/consumers" "$CONSUMER_DATA" "201" "Test Consumer 1"

CONSUMER_DATA2='{
  "consumerName": "Mock Test Consumer 2", 
  "contactEmail": "consumer2@test.com",
  "phoneNumber": "222-222-2222"
}'
make_api_call "POST" "/consumers" "$CONSUMER_DATA2" "201" "Test Consumer 2"

# 2. Create test provider submissions
echo -e "${BLUE}=== Creating Test Provider Submissions ===${NC}"

TIMESTAMP=$(date +%s)
PROVIDER_DATA1="{
  \"providerName\": \"Mock Test Provider $TIMESTAMP\",
  \"contactEmail\": \"provider1@test.com\",
  \"phoneNumber\": \"333-333-3333\",
  \"providerType\": \"government\"
}"
make_api_call "POST" "/provider-submissions" "$PROVIDER_DATA1" "201" "Test Provider 1"

# 3. Create allow list data for testing
echo -e "${BLUE}=== Creating Allow List Test Data ===${NC}"

# First, let's try to create a field allow list entry
ALLOW_LIST_DATA='{
  "consumer_id": "test-consumer-123",
  "expires_at": 1757560679,
  "grant_duration": "30d", 
  "reason": "Mock test data",
  "updated_by": "test-admin"
}'
make_api_call "POST" "/admin/fields/user.email/allow-list" "$ALLOW_LIST_DATA" "201" "Allow List Entry for user.email"

# 4. Create consumer applications
echo -e "${BLUE}=== Creating Test Consumer Applications ===${NC}"

# Get a consumer ID first
CONSUMERS_RESPONSE=$(curl -s -X GET "$API_SERVER_URL/consumers")
CONSUMER_ID=$(echo "$CONSUMERS_RESPONSE" | jq -r '.items[0].consumerId // empty')

if [ -n "$CONSUMER_ID" ] && [ "$CONSUMER_ID" != "null" ]; then
    APP_DATA='{
      "requiredFields": {
        "person.fullName": true,
        "person.email": true,
        "person.phone": false
      },
      "purpose": "Mock test application",
      "description": "Test application for integration testing"
    }'
    make_api_call "POST" "/consumer-applications/$CONSUMER_ID" "$APP_DATA" "201" "Consumer Application"
else
    echo -e "${YELLOW}⚠️ No consumer ID available for application creation${NC}"
fi

echo -e "${GREEN}=== Mock Data Setup Complete ===${NC}"
echo "Test data has been created for:"
echo "✅ Test consumers"
echo "✅ Test provider submissions" 
echo "✅ Allow list entries"
echo "✅ Consumer applications"
echo ""
echo "You can now run the integration tests: ./run-all-tests.sh"