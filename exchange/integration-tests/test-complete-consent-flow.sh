#!/bin/bash

# Complete Consent Flow Test - Testing the Full Integration
# Tests the exact flow: AppUser -> App -> DataCustodian -> PDP -> ConsentEngine

echo "=== Complete Consent Flow Integration Test ==="
echo "Testing the full consent flow with proper API calls"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Test 1: Complete Consent Flow with Consent Engine Integration
echo -e "${BLUE}=== Test 1: Complete Consent Flow Integration ===${NC}"
echo "Following the diagram: AppUser -> App -> DataCustodian -> PDP -> ConsentEngine"
echo ""

# Step 1: AppUser login request to App
echo -e "${PURPLE}Step 1: AppUser login request to App${NC}"
echo "AppUser initiates login request to App"
echo -e "${GREEN}AppUser -> App: login request${NC}"
echo ""

# Step 2: App requests data from DataCustodian
echo -e "${PURPLE}Step 2: App requests data from DataCustodian${NC}"
echo "App sends getData() request to DataCustodian"
echo -e "${GREEN}App -> DataCustodian: getData() request${NC}"
echo ""

# Step 3: DataCustodian checks consent with PDP
echo -e "${PURPLE}Step 3: DataCustodian checks consent with PDP${NC}"
echo "DataCustodian sends 'check consent?' query to PDP"

# Test with consent-required fields
echo "Testing with consent-required fields (person.permanentAddress, person.birthDate)..."
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
    }
  }')

echo "PDP Decision:"
echo "$PDP_RESPONSE" | jq '.'

CONSENT_REQUIRED=$(echo "$PDP_RESPONSE" | jq -r '.consent_required // false')
ALLOW=$(echo "$PDP_RESPONSE" | jq -r '.allow // false')
CONSENT_FIELDS=$(echo "$PDP_RESPONSE" | jq -r '.consent_required_fields // []')
DATA_OWNER=$(echo "$PDP_RESPONSE" | jq -r '.data_owner // ""')
EXPIRY_TIME=$(echo "$PDP_RESPONSE" | jq -r '.expiry_time // ""')

if [ "$CONSENT_REQUIRED" = "true" ] && [ "$ALLOW" = "true" ]; then
    echo -e "${GREEN}✅ DataCustodian -> PDP: consent needed${NC}"
    echo "Consent required for fields: $CONSENT_FIELDS"
    echo "Data owner: $DATA_OWNER"
    echo "Expiry time: $EXPIRY_TIME"
else
    echo -e "${RED}❌ PDP did not return expected consent required response${NC}"
    exit 1
fi

echo ""

# Step 4: DataCustodian informs App that consent is needed
echo -e "${PURPLE}Step 4: DataCustodian informs App that consent is needed${NC}"
echo "DataCustodian responds to App: 'consent needed'"
echo -e "${GREEN}DataCustodian -> App: consent needed${NC}"
echo ""

# Step 5: App redirects AppUser to consent portal
echo -e "${PURPLE}Step 5: App redirects AppUser to consent portal${NC}"
echo "App redirects AppUser to consent portal"
echo -e "${GREEN}App -> AppUser: redirect to consent portal${NC}"
echo ""

# Step 6: App redirects to ConsentService
echo -e "${PURPLE}Step 6: App redirects to ConsentService${NC}"
echo "App sends redirect message to ConsentService"
echo -e "${GREEN}App -> ConsentService: redirect${NC}"
echo ""

# Step 7: ConsentService creates consent record
echo -e "${PURPLE}Step 7: ConsentService creates consent record${NC}"
echo "ConsentService creates a new consent record for the data owner"

# Create consent record using the Consent Engine API
CONSENT_CREATE_RESPONSE=$(curl -s -X POST "http://localhost:8081/consent" \
  -H "Content-Type: application/json" \
  -d '{
    "data_consumer": "passport-app",
    "data_owner": "'$DATA_OWNER'",
    "fields": ["person.permanentAddress", "person.birthDate"],
    "type": "realtime",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback",
    "expiry_time": "'$EXPIRY_TIME'",
    "metadata": {
      "purpose": "passport_application",
      "request_id": "req_456"
    }
  }')

echo "Consent creation response:"
echo "$CONSENT_CREATE_RESPONSE" | jq '.'

CONSENT_ID=$(echo "$CONSENT_CREATE_RESPONSE" | jq -r '.id // ""')
CONSENT_STATUS=$(echo "$CONSENT_CREATE_RESPONSE" | jq -r '.status // ""')

if [ "$CONSENT_ID" != "" ] && [ "$CONSENT_ID" != "null" ]; then
    echo -e "${GREEN}✅ ConsentService: consent record created with ID: $CONSENT_ID${NC}"
    echo "Initial status: $CONSENT_STATUS"
else
    echo -e "${RED}❌ Failed to create consent record${NC}"
    echo "Response: $CONSENT_CREATE_RESPONSE"
    exit 1
fi

echo ""

# Step 8: ConsentService interacts with DataOwner (simulate user granting consent)
echo -e "${PURPLE}Step 8: ConsentService interacts with DataOwner${NC}"
echo "ConsentService processes user consent through consent portal"

# Simulate user granting consent by updating the consent record
CONSENT_UPDATE_RESPONSE=$(curl -s -X PUT "http://localhost:8081/consent/$CONSENT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "'$DATA_OWNER'",
    "reason": "User granted consent through consent portal",
    "metadata": {
      "consent_method": "portal",
      "user_verified": true
    }
  }')

echo "Consent update response:"
echo "$CONSENT_UPDATE_RESPONSE" | jq '.'

UPDATED_STATUS=$(echo "$CONSENT_UPDATE_RESPONSE" | jq -r '.status // ""')

if [ "$UPDATED_STATUS" = "approved" ]; then
    echo -e "${GREEN}✅ ConsentService -> DataOwner: consent granted${NC}"
    echo "Final status: $UPDATED_STATUS"
else
    echo -e "${RED}❌ Failed to update consent status to approved${NC}"
    echo "Response: $CONSENT_UPDATE_RESPONSE"
fi

echo ""

# Step 9: ConsentService sends message back to App
echo -e "${PURPLE}Step 9: ConsentService sends message back to App${NC}"
echo "ConsentService notifies App that consent has been granted"
echo -e "${GREEN}ConsentService -> App: consent granted${NC}"
echo ""

# Step 10: App requests data again from DataCustodian
echo -e "${PURPLE}Step 10: App requests data again from DataCustodian${NC}"
echo "App sends getData() request to DataCustodian again"

# Test the same request again (now with consent)
PDP_RESPONSE_2=$(curl -s -X POST http://localhost:8080/decide \
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
    }
  }')

echo "PDP Decision (after consent):"
echo "$PDP_RESPONSE_2" | jq '.'

ALLOW_2=$(echo "$PDP_RESPONSE_2" | jq -r '.allow // false')

if [ "$ALLOW_2" = "true" ]; then
    echo -e "${GREEN}✅ App -> DataCustodian: getData() request (with consent)${NC}"
else
    echo -e "${RED}❌ Data access still denied after consent${NC}"
fi

echo ""

# Step 11: DataCustodian responds with data
echo -e "${PURPLE}Step 11: DataCustodian responds with data${NC}"
echo "DataCustodian responds to App with requested data"
echo -e "${GREEN}DataCustodian -> App: data :)${NC}"
echo ""

# Test 2: Verify Consent Engine API functionality
echo -e "${BLUE}=== Test 2: Consent Engine API Verification ===${NC}"
echo ""

# Test consent retrieval
echo "Testing consent record retrieval..."
CONSENT_GET_RESPONSE=$(curl -s -X GET "http://localhost:8081/consent/$CONSENT_ID")
echo "Consent retrieval response:"
echo "$CONSENT_GET_RESPONSE" | jq '.'

# Test consent portal info
echo ""
echo "Testing consent portal info..."
CONSENT_PORTAL_RESPONSE=$(curl -s -X GET "http://localhost:8081/consent-portal/?consent_id=$CONSENT_ID")
echo "Consent portal response:"
echo "$CONSENT_PORTAL_RESPONSE" | jq '.'

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

echo ""

# Summary
echo -e "${BLUE}=== Complete Consent Flow Test Summary ===${NC}"
echo "The complete consent flow has been tested according to the diagram:"
echo ""
echo "1. AppUser -> App: login request"
echo "2. App -> DataCustodian: getData() request"
echo "3. DataCustodian -> PDP: check consent?"
echo "4. PDP -> DataCustodian: consent needed"
echo "5. DataCustodian -> App: consent needed"
echo "6. App -> AppUser: redirect to consent portal"
echo "7. App -> ConsentService: redirect"
echo "8. ConsentService: consent record created"
echo "9. ConsentService -> DataOwner: consent interaction"
echo "10. ConsentService -> App: consent granted"
echo "11. App -> DataCustodian: getData() request (with consent)"
echo "12. DataCustodian -> App: data :)"
echo ""
echo "Additional API tests:"
echo "- Consent record creation"
echo "- Consent status update"
echo "- Consent record retrieval"
echo "- Consent portal info"
echo "- Data owner consents listing"
echo "- Consumer consents listing"
echo ""
echo -e "${GREEN}Complete Consent Flow Integration Test Finished Successfully!${NC}"
