#!/bin/bash

# Test script for Hybrid Authentication in Consent Engine
# This script tests both M2M and User JWT authentication flows

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONSENT_ENGINE_URL="http://localhost:8081"
CONSENT_ID="test_consent_123"

echo -e "${BLUE}🧪 Testing Hybrid Authentication for Consent Engine${NC}"
echo "=================================================="

# Function to print test results
print_test_result() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name: PASS${NC}"
    else
        echo -e "${RED}❌ $test_name: FAIL${NC}"
        echo -e "   Details: $details"
    fi
}

# Function to make HTTP request and check response
test_endpoint() {
    local method="$1"
    local url="$2"
    local headers="$3"
    local expected_status="$4"
    local test_name="$5"
    
    echo -e "${YELLOW}Testing: $test_name${NC}"
    echo "Request: $method $url"
    echo "Headers: $headers"
    
    response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" $headers)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    echo "Response Code: $http_code"
    echo "Response Body: $body"
    echo "---"
    
    if [ "$http_code" = "$expected_status" ]; then
        print_test_result "$test_name" "PASS" "Expected status $expected_status"
    else
        print_test_result "$test_name" "FAIL" "Expected status $expected_status, got $http_code"
    fi
}

# Test 1: Health Check (No Auth Required)
echo -e "\n${BLUE}1. Testing Health Check (No Authentication Required)${NC}"
test_endpoint "GET" "$CONSENT_ENGINE_URL/health" "" "200" "Health Check"

# Test 2: Create Consent (No Auth Required)
echo -e "\n${BLUE}2. Testing Consent Creation (No Authentication Required)${NC}"
create_payload='{
  "app_id": "test-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["personInfo.name", "personInfo.email"]
    }
  ],
  "purpose": "testing",
  "session_id": "test-session-123"
}'

echo "Creating test consent..."
response=$(curl -s -X POST "$CONSENT_ENGINE_URL/consents" \
  -H "Content-Type: application/json" \
  -d "$create_payload")

echo "Create Response: $response"

# Extract consent_id from response (assuming it's in the response)
CONSENT_ID=$(echo "$response" | grep -o '"consent_id":"[^"]*"' | cut -d'"' -f4)
if [ -z "$CONSENT_ID" ]; then
    CONSENT_ID="test_consent_123"  # Fallback for testing
fi

echo "Using Consent ID: $CONSENT_ID"

# Test 3: Get Consent without Authentication (Should Fail)
echo -e "\n${BLUE}3. Testing GET Consent without Authentication (Should Fail)${NC}"
test_endpoint "GET" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "" "401" "Get Consent - No Auth"

# Test 4: Get Consent with Invalid Token (Should Fail)
echo -e "\n${BLUE}4. Testing GET Consent with Invalid Token (Should Fail)${NC}"
test_endpoint "GET" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "-H 'Authorization: Bearer invalid_token'" "401" "Get Consent - Invalid Token"

# Test 5: Get Consent with Malformed Token (Should Fail)
echo -e "\n${BLUE}5. Testing GET Consent with Malformed Token (Should Fail)${NC}"
test_endpoint "GET" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "-H 'Authorization: Bearer malformed.token'" "401" "Get Consent - Malformed Token"

# Test 6: Get Consent with Valid User JWT (Should Pass - if you have a valid token)
echo -e "\n${BLUE}6. Testing GET Consent with Valid User JWT${NC}"
echo -e "${YELLOW}Note: This test requires a valid Asgardeo user JWT token${NC}"
echo -e "${YELLOW}To test this, you need to:${NC}"
echo -e "${YELLOW}1. Get a valid JWT token from Asgardeo${NC}"
echo -e "${YELLOW}2. Replace 'YOUR_USER_JWT_TOKEN' with the actual token${NC}"
echo -e "${YELLOW}3. Ensure the token's email matches the consent owner email${NC}"

# Uncomment and replace with actual token for testing:
# test_endpoint "GET" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "-H 'Authorization: Bearer YOUR_USER_JWT_TOKEN'" "200" "Get Consent - Valid User JWT"

# Test 7: Get Consent with Valid M2M Token (Should Pass - if you have a valid token)
echo -e "\n${BLUE}7. Testing GET Consent with Valid M2M Token${NC}"
echo -e "${YELLOW}Note: This test requires a valid Choreo M2M JWT token${NC}"
echo -e "${YELLOW}To test this, you need to:${NC}"
echo -e "${YELLOW}1. Get a valid M2M token from Choreo${NC}"
echo -e "${YELLOW}2. Replace 'YOUR_M2M_JWT_TOKEN' with the actual token${NC}"

# Uncomment and replace with actual token for testing:
# test_endpoint "GET" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "-H 'Authorization: Bearer YOUR_M2M_JWT_TOKEN'" "200" "Get Consent - Valid M2M JWT"

# Test 8: Update Consent without Authentication (Should Fail)
echo -e "\n${BLUE}8. Testing PUT Consent without Authentication (Should Fail)${NC}"
update_payload='{"status": "approved", "reason": "Test approval"}'
test_endpoint "PUT" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "-H 'Content-Type: application/json' -d '$update_payload'" "401" "Update Consent - No Auth"

# Test 9: Update Consent with Invalid Token (Should Fail)
echo -e "\n${BLUE}9. Testing PUT Consent with Invalid Token (Should Fail)${NC}"
test_endpoint "PUT" "$CONSENT_ENGINE_URL/consents/$CONSENT_ID" "-H 'Content-Type: application/json' -H 'Authorization: Bearer invalid_token' -d '$update_payload'" "401" "Update Consent - Invalid Token"

echo -e "\n${BLUE}🎯 Test Summary${NC}"
echo "=================="
echo -e "${GREEN}✅ Basic functionality tests completed${NC}"
echo -e "${YELLOW}⚠️  JWT token tests require valid tokens${NC}"
echo -e "${BLUE}📝 To complete testing:${NC}"
echo "   1. Add the missing environment variables in Choreo"
echo "   2. Get valid JWT tokens for testing"
echo "   3. Run the commented test cases with real tokens"

echo -e "\n${BLUE}🔧 Environment Variables to Add in Choreo:${NC}"
echo "   - ASGARDEO_AUDIENCE"
echo "   - CHOREO_JWKS_URL"
echo "   - CHOREO_ISSUER" 
echo "   - CHOREO_AUDIENCE"

echo -e "\n${GREEN}✨ Hybrid Authentication Implementation Ready for Testing!${NC}"
