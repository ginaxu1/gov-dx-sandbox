#!/bin/bash

# Comprehensive Consent Flow Test Script

echo "=== Comprehensive Consent Flow Test Suite ==="
echo "Testing the complete consent flow: App -> DataCustodian -> PDP -> ConsentEngine"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: No Consent Required Flow
echo -e "${BLUE}Test 1: No Consent Required Flow${NC}"
echo "Scenario: App requests data that doesn't require consent"
echo "Expected: Direct data access without consent flow"
echo ""

echo "Step 1: App requests data (person.fullName, person.nic, person.photo)"
PDP_RESPONSE=$(curl -s -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic", "person.photo"]
    },
    "context": {
      "ip_address": "192.168.1.100",
      "user_agent": "PassportApp/1.0"
    }
  }')

echo "PDP Decision:"
echo "$PDP_RESPONSE" | jq '.'

CONSENT_REQUIRED=$(echo "$PDP_RESPONSE" | jq -r '.consent_required // false')
ALLOW=$(echo "$PDP_RESPONSE" | jq -r '.allow // false')

if [ "$CONSENT_REQUIRED" = "false" ] && [ "$ALLOW" = "true" ]; then
    echo -e "${GREEN}✅ Test 1 PASSED: No consent required, direct access granted${NC}"
else
    echo -e "${RED}❌ Test 1 FAILED: Expected no consent required and access granted${NC}"
fi

echo "---"

# Test 2: Consent Required Flow
echo -e "${BLUE}Test 2: Consent Required Flow${NC}"
echo "Scenario: App requests data that requires consent"
echo "Expected: Consent flow triggered, consent required"
echo ""

echo "Step 1: App requests data with consent-required fields"
PDP_RESPONSE=$(curl -s -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.permanentAddress", "person.birthDate"]
    },
    "context": {
      "ip_address": "192.168.1.100",
      "user_agent": "PassportApp/1.0"
    }
  }')

echo "PDP Decision:"
echo "$PDP_RESPONSE" | jq '.'

CONSENT_REQUIRED=$(echo "$PDP_RESPONSE" | jq -r '.consent_required // false')
ALLOW=$(echo "$PDP_RESPONSE" | jq -r '.allow // false')
CONSENT_FIELDS=$(echo "$PDP_RESPONSE" | jq -r '.consent_required_fields // []')
DATA_OWNER=$(echo "$PDP_RESPONSE" | jq -r '.data_owner // ""')

if [ "$CONSENT_REQUIRED" = "true" ] && [ "$ALLOW" = "true" ]; then
    echo -e "${GREEN}✅ Test 2a PASSED: Consent required, access granted with consent flow${NC}"
    echo "Consent required for fields: $CONSENT_FIELDS"
    echo "Data owner: $DATA_OWNER"
    
    # Test 2b: Check Consent Engine
    echo ""
    echo "Step 2: Check Consent Engine availability"
    CE_RESPONSE=$(curl -s -X GET http://localhost:8081/health 2>/dev/null || echo "Not available")
    
    if [ "$CE_RESPONSE" != "Not available" ]; then
        echo -e "${GREEN}✅ Test 2b PASSED: Consent Engine is available${NC}"
        echo "Consent Engine Response: $CE_RESPONSE"
    else
        echo -e "${YELLOW}⚠️  Test 2b WARNING: Consent Engine not responding to /health${NC}"
        
        # Try other endpoints
        echo "Trying alternative Consent Engine endpoints..."
        CE_ALT_RESPONSE=$(curl -s -X GET http://localhost:8081/ 2>/dev/null || echo "Not available")
        echo "Consent Engine Root Response: $CE_ALT_RESPONSE"
    fi
    
else
    echo -e "${RED}❌ Test 2 FAILED: Expected consent required and access granted${NC}"
fi

echo "---"

# Test 3: Unauthorized Access
echo -e "${BLUE}Test 3: Unauthorized Access${NC}"
echo "Scenario: App requests data it's not authorized to access"
echo "Expected: Access denied without consent flow"
echo ""

echo "Step 1: App requests unauthorized data"
PDP_RESPONSE=$(curl -s -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.ssn"]
    },
    "context": {
      "ip_address": "192.168.1.100",
      "user_agent": "PassportApp/1.0"
    }
  }')

echo "PDP Decision:"
echo "$PDP_RESPONSE" | jq '.'

ALLOW=$(echo "$PDP_RESPONSE" | jq -r '.allow // false')
DENY_REASON=$(echo "$PDP_RESPONSE" | jq -r '.deny_reason // ""')

if [ "$ALLOW" = "false" ] && [[ "$DENY_REASON" == *"not authorized"* ]]; then
    echo -e "${GREEN}✅ Test 3 PASSED: Unauthorized access correctly denied${NC}"
    echo "Deny reason: $DENY_REASON"
else
    echo -e "${RED}❌ Test 3 FAILED: Expected unauthorized access to be denied${NC}"
fi

echo "---"

# Test 4: Invalid Consumer
echo -e "${BLUE}Test 4: Invalid Consumer${NC}"
echo "Scenario: Unknown consumer requests data"
echo "Expected: Access denied, consumer not found"
echo ""

echo "Step 1: Unknown consumer requests data"
PDP_RESPONSE=$(curl -s -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "unknown-app",
      "name": "Unknown Application"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName"]
    },
    "context": {
      "ip_address": "192.168.1.100"
    }
  }')

echo "PDP Decision:"
echo "$PDP_RESPONSE" | jq '.'

ALLOW=$(echo "$PDP_RESPONSE" | jq -r '.allow // false')
DENY_REASON=$(echo "$PDP_RESPONSE" | jq -r '.deny_reason // ""')

if [ "$ALLOW" = "false" ] && [[ "$DENY_REASON" == *"not found"* ]]; then
    echo -e "${GREEN}✅ Test 4 PASSED: Invalid consumer correctly denied${NC}"
    echo "Deny reason: $DENY_REASON"
else
    echo -e "${RED}❌ Test 4 FAILED: Expected invalid consumer to be denied${NC}"
fi

echo "---"

# Test 5: Consent Engine Integration Test
echo -e "${BLUE}Test 5: Consent Engine Integration Test${NC}"
echo "Scenario: Test Consent Engine endpoints and functionality"
echo ""

echo "Step 1: Test Consent Engine endpoints"
echo "Testing various Consent Engine endpoints..."

# Test different endpoints
ENDPOINTS=("/" "/health" "/status" "/consent" "/api/consent" "/consent/check")

for endpoint in "${ENDPOINTS[@]}"; do
    echo "Testing endpoint: http://localhost:8081$endpoint"
    RESPONSE=$(curl -s -X GET "http://localhost:8081$endpoint" 2>/dev/null || echo "Not available")
    if [ "$RESPONSE" != "Not available" ] && [ "$RESPONSE" != "404 page not found" ]; then
        echo -e "${GREEN}✅ Endpoint $endpoint responded: $RESPONSE${NC}"
    else
        echo -e "${YELLOW}⚠️  Endpoint $endpoint: $RESPONSE${NC}"
    fi
done

echo "---"

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "This test suite verifies the consent flow components:"
echo "1. Policy Decision Point (PDP) - Port 8080"
echo "2. Consent Engine (CE) - Port 8081"
echo ""
echo "The flow follows the diagram:"
echo "App -> DataCustodian -> PDP -> ConsentEngine (if consent required)"
echo ""
echo "Key scenarios tested:"
echo "- No consent required: Direct data access"
echo "- Consent required: Consent flow triggered"
echo "- Unauthorized access: Proper denial"
echo "- Invalid consumer: Proper denial"
echo "- Consent Engine availability: Service health check"
echo ""
echo -e "${GREEN}Consent Flow Test Suite Complete${NC}"
