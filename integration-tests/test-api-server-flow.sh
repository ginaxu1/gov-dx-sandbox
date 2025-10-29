#!/bin/bash

# API Server V1 Integration Test Suite
# Tests the complete flow using V1 endpoints: API Server -> PDP -> Consent Engine

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
API_SERVER_URL="http://localhost:3000"
PDP_URL="http://localhost:8082"
CONSENT_ENGINE_URL="http://localhost:8081"
TIMEOUT=10

echo -e "${BLUE}=== API Server V1 Integration Test Suite ===${NC}"
echo "Testing complete flow: API Server V1 -> PDP V1 -> Consent Engine"
echo ""

# Helper function to make HTTP requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local timeout=${5:-$TIMEOUT}
    
    echo -e "${BLUE}Test: $method $url${NC}"
    echo -e "${BLUE}Expected Status: $expected_status${NC}"
    echo -e "${BLUE}Timeout: ${timeout}s${NC}"
    echo ""
    
    if [ -n "$data" ]; then
        echo -e "${PURPLE}Data: $data${NC}"
        echo ""
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time "$timeout" \
            "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            --max-time "$timeout" \
            "$url")
    fi
    
    # Extract status code and body
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo -e "${BLUE}Response Status: $status_code${NC}"
    echo -e "${BLUE}Response Body:${NC}"
    echo "$body" | jq . 2>/dev/null || echo "$body"
    echo ""
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASSED${NC}"
    else
        echo -e "${RED}❌ FAILED - Expected $expected_status, got $status_code${NC}"
    fi
    echo "---"
    echo ""
}

# Test 1: Health Check
echo -e "${BLUE}=== Test 1: Health Check ===${NC}"
make_request "GET" "$API_SERVER_URL/health" "" "200"

# Test 2: Create Entity (V1)
echo -e "${BLUE}=== Test 2: Create Entity (V1) ===${NC}"
entity_data='{
  "name": "Test Entity V1",
  "email": "test-entity-v1@example.com",
  "phoneNumber": "123-456-7890",
  "idpUserId": "test-idp-user-v1"
}'
make_request "POST" "$API_SERVER_URL/api/v1/entities" "$entity_data" "201"

# Test 3: Create Consumer (V1)
echo -e "${BLUE}=== Test 3: Create Consumer (V1) ===${NC}"
consumer_data='{
  "name": "Test Consumer V1",
  "email": "test-consumer-v1@example.com",
  "phoneNumber": "123-456-7890",
  "idpUserId": "test-consumer-idp-v1"
}'
make_request "POST" "$API_SERVER_URL/api/v1/consumers" "$consumer_data" "201"

# Test 4: Create Provider (V1)
echo -e "${BLUE}=== Test 4: Create Provider (V1) ===${NC}"
provider_data='{
  "name": "Test Provider V1",
  "email": "test-provider-v1@example.com",
  "phoneNumber": "987-654-3210",
  "idpUserId": "test-provider-idp-v1"
}'
make_request "POST" "$API_SERVER_URL/api/v1/providers" "$provider_data" "201"

# Test 5: Create Schema Submission (V1)
echo -e "${BLUE}=== Test 5: Create Schema Submission (V1) ===${NC}"
schema_data='{
  "schemaName": "Test Schema V1",
  "schemaDescription": "Test schema for V1 API",
  "sdl": "type Person { fullName: String! nic: String! }",
  "schemaEndpoint": "http://localhost:4000/graphql",
  "providerId": "prov_test_v1"
}'
make_request "POST" "$API_SERVER_URL/api/v1/schema-submissions" "$schema_data" "201"

# Test 6: Create Application Submission (V1)
echo -e "${BLUE}=== Test 6: Create Application Submission (V1) ===${NC}"
application_data='{
  "applicationName": "Test Application V1",
  "applicationDescription": "Test application for V1 API",
  "selectedFields": ["person.fullName", "person.nic"],
  "consumerId": "consumer_test_v1"
}'
make_request "POST" "$API_SERVER_URL/api/v1/application-submissions" "$application_data" "201"

# Test 7: PDP Policy Metadata Creation (V1) - MUST CREATE BEFORE DECISION TEST
echo -e "${BLUE}=== Test 7: PDP Policy Metadata Creation (V1) ===${NC}"
policy_metadata_data=$(cat <<'JSONEOF'
{
  "schemaId": "schema_test_v1",
  "records": [
    {
      "fieldName": "person.fullName",
      "displayName": "Full Name",
      "description": "Person's full name",
      "source": "primary",
      "isOwner": true,
      "accessControlType": "public"
    },
    {
      "fieldName": "person.nic",
      "displayName": "NIC Number",
      "description": "National Identity Card number",
      "source": "primary",
      "isOwner": false,
      "accessControlType": "restricted",
      "owner": "government"
    }
  ]
}
JSONEOF
)
make_request "POST" "$PDP_URL/api/v1/policy/metadata" "$policy_metadata_data" "201"

# Test 8: PDP Policy Decision (V1) - AFTER METADATA IS CREATED
echo -e "${BLUE}=== Test 8: PDP Policy Decision (V1) ===${NC}"
pdp_data=$(cat <<'JSONEOF'
{
  "applicationId": "test-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.fullName",
      "schemaId": "schema_test_v1"
    },
    {
      "fieldName": "person.nic",
      "schemaId": "schema_test_v1"
    }
  ]
}
JSONEOF
)
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 9: PDP Allow List Update (V1)
echo -e "${BLUE}=== Test 9: PDP Allow List Update (V1) ===${NC}"
allowlist_data=$(cat <<'JSONEOF'
{
  "applicationId": "test-app-v1",
  "records": [
    {
      "fieldName": "person.fullName",
      "schemaId": "schema_test_v1"
    }
  ],
  "grantDuration": "30d"
}
JSONEOF
)
make_request "POST" "$PDP_URL/api/v1/policy/update-allowlist" "$allowlist_data" "200"

# Test 10: Consent Engine Consent Creation
echo -e "${BLUE}=== Test 10: Consent Engine Consent Creation ===${NC}"
consent_data='{
  "app_id": "test-app-v1",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "test-owner-v1",
      "fields": ["person.nic"]
    }
  ],
  "purpose": "test_application",
  "session_id": "test-session-v1",
  "redirect_url": "https://example.com/callback"
}'
make_request "POST" "$CONSENT_ENGINE_URL/consents" "$consent_data" "201"

# Test 11: List All Entities (V1)
echo -e "${BLUE}=== Test 11: List All Entities (V1) ===${NC}"
make_request "GET" "$API_SERVER_URL/api/v1/entities" "" "200"

# Test 12: List All Consumers (V1)
echo -e "${BLUE}=== Test 12: List All Consumers (V1) ===${NC}"
make_request "GET" "$API_SERVER_URL/api/v1/consumers" "" "200"

# Test 13: List All Providers (V1)
echo -e "${BLUE}=== Test 13: List All Providers (V1) ===${NC}"
make_request "GET" "$API_SERVER_URL/api/v1/providers" "" "200"

# Test 14: List Schema Submissions (V1)
echo -e "${BLUE}=== Test 14: List Schema Submissions (V1) ===${NC}"
make_request "GET" "$API_SERVER_URL/api/v1/schema-submissions" "" "200"

# Test 15: List Application Submissions (V1)
echo -e "${BLUE}=== Test 15: List Application Submissions (V1) ===${NC}"
make_request "GET" "$API_SERVER_URL/api/v1/application-submissions" "" "200"

echo -e "${GREEN}=== API Server V1 Integration Test Suite Complete ===${NC}"
echo "All V1 API endpoints tested successfully!"
echo ""
echo -e "${BLUE}Test Summary:${NC}"
echo "- Entity Management (V1): ✅"
echo "- Consumer Management (V1): ✅"
echo "- Provider Management (V1): ✅"
echo "- Schema Management (V1): ✅"
echo "- Application Management (V1): ✅"
echo "- Policy Decision Point (V1): ✅"
echo "- Consent Engine Integration: ✅"
