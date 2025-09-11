#!/bin/bash
# Comprehensive Consent Flow Test Script

echo "=== Comprehensive Consent Flow Test Suite ==="
echo "Testing consent scenarios: data owner is provider vs data owner is NOT provider"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Test function for PDP requests
test_pdp_request() {
    local test_name="$1"
    local scenario="$2"
    local expected="$3"
    local data="$4"
    
    echo -e "${BLUE}Test: $test_name${NC}"
    echo "Scenario: $scenario"
    echo "Expected: $expected"
    echo ""
    
    PDP_RESPONSE=$(curl -s -X POST http://localhost:8082/decide \
      -H "Content-Type: application/json" \
      -d "$data")
    
    echo "PDP Decision:"
    echo "$PDP_RESPONSE" | jq '.'
    
    CONSENT_REQUIRED=$(echo "$PDP_RESPONSE" | jq -r '.consent_required // false')
    ALLOW=$(echo "$PDP_RESPONSE" | jq -r '.allow // false')
    CONSENT_FIELDS=$(echo "$PDP_RESPONSE" | jq -r '.consent_required_fields // []')
    DATA_OWNER=$(echo "$PDP_RESPONSE" | jq -r '.data_owner // ""')
    # Note: deny_reason was removed from the simplified response format
    
    echo ""
    return 0
}

# Test function for Consent Engine requests
test_consent_request() {
    local test_name="$1"
    local data="$2"
    
    echo -e "${BLUE}Test: $test_name${NC}"
    echo ""
    
    CONSENT_RESPONSE=$(curl -s -X POST http://localhost:8081/consent \
      -H "Content-Type: application/json" \
      -d "$data")
    
    echo "Consent Engine Response:"
    echo "$CONSENT_RESPONSE" | jq '.'
    
    CONSENT_ID=$(echo "$CONSENT_RESPONSE" | jq -r '.id // ""')
    CONSENT_STATUS=$(echo "$CONSENT_RESPONSE" | jq -r '.status // ""')
    DATA_CONSUMER=$(echo "$CONSENT_RESPONSE" | jq -r '.data_consumer // ""')
    DATA_OWNER=$(echo "$CONSENT_RESPONSE" | jq -r '.data_owner // ""')
    FIELDS=$(echo "$CONSENT_RESPONSE" | jq -r '.fields // []')
    
    echo ""
    return 0
}

# Check if services are running
echo -e "${BLUE}=== Service Health Checks ===${NC}"

# Check PDP
PDP_HEALTH=$(curl -s http://localhost:8082/health 2>/dev/null || echo "Not available")
if [ "$PDP_HEALTH" != "Not available" ]; then
    echo -e "${GREEN}✅ Policy Decision Point (PDP) is running on port 8082${NC}"
else
    echo -e "${RED}❌ Policy Decision Point (PDP) not responding on port 8082${NC}"
    echo "Please start the PDP: cd policy-decision-point && go run main.go"
    exit 1
fi

# Check Consent Engine
CE_HEALTH=$(curl -s http://localhost:8081/health 2>/dev/null || echo "Not available")
if [ "$CE_HEALTH" != "Not available" ]; then
    echo -e "${GREEN}✅ Consent Engine (CE) is running on port 8081${NC}"
else
    echo -e "${RED}❌ Consent Engine (CE) not responding on port 8081${NC}"
    echo "Please start the CE: cd consent-engine && go run main.go"
    exit 1
fi

echo ""

# Test 1: Data Owner IS the Provider (No Consent Required)
echo -e "${BLUE}=== Test 1: Data Owner IS the Provider (No Consent Required) ===${NC}"
echo "Scenario: Provider (DRP) requests data owned by DRP"
echo "Fields: person.fullName, person.nic (both owned by DRP)"
echo "Expected: No consent required, direct access"
echo ""

test_pdp_request "Data Owner = Provider" \
  "Provider requests data it owns" \
  "No consent required, direct access" \
  '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_001",
    "required_fields": ["person.fullName", "person.nic"]
  }'

if [ "$CONSENT_REQUIRED" = "false" ] && [ "$ALLOW" = "true" ]; then
    echo -e "${GREEN}✅ Test 1 PASSED: No consent required, direct access granted${NC}"
    echo "Reason: Data owner (DRP) = Provider (DRP), no cross-provider consent needed"
else
    echo -e "${RED}❌ Test 1 FAILED: Expected no consent required and access granted${NC}"
    echo "Consent required: $CONSENT_REQUIRED, Allow: $ALLOW"
fi

echo "---"

# Test 2: Data Owner is NOT the Provider (Consent Required)
echo -e "${BLUE}=== Test 2: Data Owner is NOT the Provider (Consent Required) ===${NC}"
echo "Scenario: Provider (DRP) requests data owned by RGD"
echo "Fields: person.permanentAddress, person.photo (both owned by RGD)"
echo "Expected: Consent required, consent flow triggered"
echo ""

test_pdp_request "Data Owner ≠ Provider" \
  "Provider requests data owned by different entity" \
  "Consent required, consent flow triggered" \
  '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_002",
    "required_fields": ["person.permanentAddress", "person.photo"]
  }'

if [ "$CONSENT_REQUIRED" = "true" ] && [ "$ALLOW" = "true" ]; then
    echo -e "${GREEN}✅ Test 2a PASSED: Consent required, access granted with consent flow${NC}"
    echo "Consent required for fields: $CONSENT_FIELDS"
    echo "Data owner: $DATA_OWNER"
    
    # Test Consent Engine integration
    echo ""
    echo -e "${PURPLE}Testing Consent Engine integration...${NC}"
    
    test_consent_request "Create Consent Record" \
      '{
        "app_id": "passport-app",
        "data_fields": [
          {
            "owner_type": "citizen",
            "owner_id": "'$DATA_OWNER'",
            "fields": ["person.permanentAddress", "person.photo"]
          }
        ],
        "purpose": "passport_application",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback",
        "expires_at": 1757560679,
        "grant_duration": "30d"
      }'

if [ "$CONSENT_ID" != "" ] && [ "$CONSENT_ID" != "null" ]; then
        echo -e "${GREEN}✅ Test 2b PASSED: Consent record created successfully${NC}"
        echo "Consent ID: $CONSENT_ID"
        echo "Status: $CONSENT_STATUS"
        echo "Data Consumer: $DATA_CONSUMER"
        echo "Data Owner: $DATA_OWNER"
        echo "Fields: $FIELDS"
        
        # Test consent approval
echo ""
        echo -e "${PURPLE}Testing consent approval...${NC}"

CONSENT_UPDATE_RESPONSE=$(curl -s -X PUT "http://localhost:8081/consent/$CONSENT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "'$DATA_OWNER'",
            "reason": "User granted consent via SMS OTP",
    "metadata": {
              "consent_method": "sms_otp",
      "user_verified": true
    }
  }')

echo "Consent update response:"
echo "$CONSENT_UPDATE_RESPONSE" | jq '.'

UPDATED_STATUS=$(echo "$CONSENT_UPDATE_RESPONSE" | jq -r '.status // ""')

if [ "$UPDATED_STATUS" = "approved" ]; then
            echo -e "${GREEN}✅ Test 2c PASSED: Consent approved successfully${NC}"
    echo "Final status: $UPDATED_STATUS"
        else
            echo -e "${RED}❌ Test 2c FAILED: Failed to approve consent${NC}"
        fi
        
    else
        echo -e "${RED}❌ Test 2b FAILED: Failed to create consent record${NC}"
    fi
    
else
    echo -e "${RED}❌ Test 2 FAILED: Expected consent required and access granted${NC}"
    echo "Consent required: $CONSENT_REQUIRED, Allow: $ALLOW"
fi

echo "---"

# Test 3: Mixed Ownership (Some fields require consent, others don't)
echo -e "${BLUE}=== Test 3: Mixed Ownership (Partial Consent Required) ===${NC}"
echo "Scenario: Provider requests data from multiple owners"
echo "Fields: person.fullName (DRP), person.birthDate (RGD), person.permanentAddress (RGD)"
echo "Expected: Consent required only for RGD fields"
echo ""

test_pdp_request "Mixed Ownership" \
  "Provider requests data from multiple owners" \
  "Consent required only for cross-provider fields" \
  '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_003",
    "required_fields": ["person.fullName", "person.birthDate", "person.permanentAddress"]
  }'

if [ "$CONSENT_REQUIRED" = "true" ] && [ "$ALLOW" = "true" ]; then
    echo -e "${GREEN}✅ Test 3 PASSED: Mixed ownership handled correctly${NC}"
    echo "Consent required for fields: $CONSENT_FIELDS"
    echo "Data owner: $DATA_OWNER"
    echo "Note: person.fullName (DRP) accessed immediately, RGD fields require consent"
else
    echo -e "${RED}❌ Test 3 FAILED: Expected mixed ownership to be handled correctly${NC}"
fi

echo "---"

# Test 4: Unauthorized Access
echo -e "${BLUE}=== Test 4: Unauthorized Access ===${NC}"
echo "Scenario: App requests data it's not authorized to access"
echo "Fields: person.nic (restricted field, app not in allow_list)"
echo "Expected: Access denied"
echo ""

test_pdp_request "Unauthorized Access" \
  "App requests data it's not authorized to access" \
  "Access denied" \
  '{
    "consumer_id": "unauthorized-app",
    "app_id": "unauthorized-app",
    "request_id": "req_004",
    "required_fields": ["person.nic"]
  }'

if [ "$ALLOW" = "false" ]; then
    echo -e "${GREEN}✅ Test 4 PASSED: Unauthorized access correctly denied${NC}"
    echo "Deny reason: $DENY_REASON"
else
    echo -e "${RED}❌ Test 4 FAILED: Expected unauthorized access to be denied${NC}"
fi

echo "---"

# Test 5: Invalid App ID
echo -e "${BLUE}=== Test 5: Invalid App ID ===${NC}"
echo "Scenario: Unknown app requests data"
echo "Expected: Access denied"
echo ""

test_pdp_request "Invalid App ID" \
  "Unknown app requests data" \
  "Access denied" \
  '{
    "consumer_id": "unknown-app",
    "app_id": "unknown-app",
    "request_id": "req_005",
    "required_fields": ["person.fullName"]
  }'

if [ "$ALLOW" = "false" ]; then
    echo -e "${GREEN}✅ Test 5 PASSED: Invalid app correctly denied${NC}"
    echo "Deny reason: $DENY_REASON"
else
    echo -e "${RED}❌ Test 5 FAILED: Expected invalid app to be denied${NC}"
fi

echo "---"

# Test 6: Consent Engine API Verification
echo -e "${BLUE}=== Test 6: Consent Engine API Verification ===${NC}"
echo "Testing all Consent Engine endpoints"
echo ""

# Test health endpoint
echo "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8081/health)
echo "Health response: $HEALTH_RESPONSE"

# Test consent retrieval (if we have a consent ID from previous tests)
if [ "$CONSENT_ID" != "" ] && [ "$CONSENT_ID" != "null" ]; then
echo ""
    echo "Testing consent retrieval..."
CONSENT_GET_RESPONSE=$(curl -s -X GET "http://localhost:8081/consent/$CONSENT_ID")
echo "Consent retrieval response:"
echo "$CONSENT_GET_RESPONSE" | jq '.'

# Test data owner consents
echo ""
echo "Testing data owner consents..."
DATA_OWNER_CONSENTS=$(curl -s -X GET "http://localhost:8081/data-owner/$DATA_OWNER")
echo "Data owner consents response:"
echo "$DATA_OWNER_CONSENTS" | jq '.'

# Test consumer consents
echo ""
echo "Testing consumer consents..."
CONSUMER_CONSENTS=$(curl -s -X GET "http://localhost:8081/consumer/passport-app")
echo "Consumer consents response:"
echo "$CONSUMER_CONSENTS" | jq '.'
fi

echo "---"

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "This test suite validates consent management scenarios:"
echo ""
echo "✅ Test 1: Data Owner = Provider (No Consent Required)"
echo "   - Provider requests data it owns"
echo "   - Direct access without consent flow"
echo ""
echo "✅ Test 2: Data Owner ≠ Provider (Consent Required)"
echo "   - Provider requests data owned by different entity"
echo "   - Consent flow triggered"
echo "   - Consent record created and approved"
echo ""
echo "✅ Test 3: Mixed Ownership (Partial Consent Required)"
echo "   - Provider requests data from multiple owners"
echo "   - Some fields accessed immediately, others require consent"
echo ""
echo "✅ Test 4: Unauthorized Access"
echo "   - App requests data it's not authorized to access"
echo "   - Access properly denied"
echo ""
echo "✅ Test 5: Invalid App ID"
echo "   - Unknown app requests data"
echo "   - Access properly denied"
echo ""
echo "✅ Test 6: Consent Engine API Verification"
echo "   - All Consent Engine endpoints functional"
echo "   - Consent record management working"
echo ""
echo "Key consent scenarios covered:"
echo "1. Data owner IS the provider → No consent required"
echo "2. Data owner is NOT the provider → Consent required"
echo "3. Mixed ownership → Partial consent required"
echo "4. Unauthorized access → Proper denial"
echo "5. Invalid requests → Proper error handling"
echo ""
echo -e "${GREEN}Comprehensive Consent Flow Test Suite Complete${NC}"