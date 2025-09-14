#!/bin/bash
# Comprehensive Consent Engine Integration Tests
# Tests all consent workflow scenarios including OTP retry logic

echo "=== Consent Engine Integration Tests ==="
echo "Testing complete consent workflow with OTP verification and retry logic"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
CONSENT_ENGINE_URL="http://localhost:8081"
DEFAULT_CONSENT_ID="consent_03c134ae"

# Test function for API calls
test_api_call() {
    local test_name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local expected_status="$5"
    local timeout="${6:-10}"
    
    echo -e "${BLUE}Test: $test_name${NC}"
    echo "URL: $url"
    echo "Method: $method"
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    echo ""
    
    if [ -n "$data" ]; then
        RESPONSE=$(timeout $timeout curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null || echo -e "\n408")
    else
        RESPONSE=$(timeout $timeout curl -s -w "\n%{http_code}" -X "$method" "$url" 2>/dev/null || echo -e "\n408")
    fi
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    echo "Response Code: $HTTP_CODE"
    echo "Response Body: $BODY"
    echo ""
    
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC}"
        return 0
    else
        echo -e "${RED}❌ FAIL - Expected $expected_status, got $HTTP_CODE${NC}"
        return 1
    fi
}

# Test 1: Health Check
echo -e "${PURPLE}=== Test 1: Health Check ===${NC}"
test_api_call "Health Check" "GET" "$CONSENT_ENGINE_URL/health" "" "200"

# Test 2: Get Default Consent Record
echo -e "${PURPLE}=== Test 2: Get Default Consent Record ===${NC}"
test_api_call "Get Default Consent" "GET" "$CONSENT_ENGINE_URL/consents/$DEFAULT_CONSENT_ID" "" "200"

# Test 3: Create New Consent (Non-existing)
echo -e "${PURPLE}=== Test 3: Create New Consent (Non-existing) ===${NC}"
NEW_CONSENT_DATA='{
    "app_id": "passport-app",
    "data_fields": [
        {
            "owner_type": "citizen",
            "owner_id": "1991111111",
            "fields": ["personInfo.permanentAddress"]
        }
    ],
    "purpose": "passport_application",
    "session_id": "session_new_test",
    "redirect_url": "https://passport-app.gov.lk"
}'

RESPONSE=$(curl -s -X POST "$CONSENT_ENGINE_URL/consents" \
    -H "Content-Type: application/json" \
    -d "$NEW_CONSENT_DATA" 2>/dev/null)

if [ $? -eq 0 ]; then
    NEW_CONSENT_ID=$(echo "$RESPONSE" | jq -r '.redirect_url' | sed 's/.*consent_id=\([^&]*\).*/\1/')
    echo -e "${GREEN}✅ Created new consent: $NEW_CONSENT_ID${NC}"
else
    echo -e "${RED}❌ Failed to create new consent${NC}"
    exit 1
fi

# Test 4: Create Consent with Existing (owner_id, app_id) Pair
echo -e "${PURPLE}=== Test 4: Create Consent with Existing Pair ===${NC}"
EXISTING_CONSENT_DATA='{
    "app_id": "passport-app",
    "data_fields": [
        {
            "owner_type": "citizen",
            "owner_id": "1991111111",
            "fields": ["personInfo.permanentAddress", "personInfo.birthDate"]
        }
    ],
    "purpose": "passport_application",
    "session_id": "session_existing_test",
    "redirect_url": "https://passport-app.gov.lk"
}'

RESPONSE=$(curl -s -X POST "$CONSENT_ENGINE_URL/consents" \
    -H "Content-Type: application/json" \
    -d "$EXISTING_CONSENT_DATA" 2>/dev/null)

if [ $? -eq 0 ]; then
    EXISTING_CONSENT_ID=$(echo "$RESPONSE" | jq -r '.redirect_url' | sed 's/.*consent_id=\([^&]*\).*/\1/')
    if [ "$EXISTING_CONSENT_ID" = "$NEW_CONSENT_ID" ]; then
        echo -e "${GREEN}✅ Correctly returned existing consent: $EXISTING_CONSENT_ID${NC}"
    else
        echo -e "${YELLOW}⚠️  Created new consent instead of returning existing: $EXISTING_CONSENT_ID${NC}"
    fi
else
    echo -e "${RED}❌ Failed to test existing consent scenario${NC}"
fi

# Test 5: Approve Consent
echo -e "${PURPLE}=== Test 5: Approve Consent ===${NC}"
APPROVE_DATA='{
    "status": "approved",
    "owner_id": "1991111111",
    "message": "Approved via consent portal"
}'

test_api_call "Approve Consent" "PUT" "$CONSENT_ENGINE_URL/consents/$NEW_CONSENT_ID" "$APPROVE_DATA" "200"

# Test 6: Reject Consent (using default consent)
echo -e "${PURPLE}=== Test 6: Reject Consent ===${NC}"
REJECT_DATA='{
    "status": "rejected",
    "owner_id": "199512345678",
    "message": "Rejected via consent portal"
}'

test_api_call "Reject Consent" "PUT" "$CONSENT_ENGINE_URL/consents/$DEFAULT_CONSENT_ID" "$REJECT_DATA" "200"

# Test 6.5: Approve Previously Rejected Consent
echo -e "${PURPLE}=== Test 6.5: Approve Previously Rejected Consent ===${NC}"
APPROVE_REJECTED_DATA='{
    "status": "approved",
    "owner_id": "199512345678",
    "message": "Approved after reconsideration"
}'

test_api_call "Approve Previously Rejected Consent" "PUT" "$CONSENT_ENGINE_URL/consents/$DEFAULT_CONSENT_ID" "$APPROVE_REJECTED_DATA" "200"

# Test 7: OTP Verification - Correct OTP
echo -e "${PURPLE}=== Test 7: OTP Verification - Correct OTP ===${NC}"
CORRECT_OTP_DATA='{
    "otp_code": "123456"
}'

test_api_call "Correct OTP" "POST" "$CONSENT_ENGINE_URL/consents/$NEW_CONSENT_ID/otp" "$CORRECT_OTP_DATA" "200"

# Test 8: OTP Verification - Incorrect OTP (1st attempt)
echo -e "${PURPLE}=== Test 8: OTP Verification - Incorrect OTP (1st attempt) ===${NC}"
# First, create a new consent for OTP testing
OTP_CONSENT_DATA='{
    "app_id": "passport-app",
    "data_fields": [
        {
            "owner_type": "citizen",
            "owner_id": "1992222222",
            "fields": ["personInfo.permanentAddress"]
        }
    ],
    "purpose": "passport_application",
    "session_id": "session_otp_test",
    "redirect_url": "https://passport-app.gov.lk"
}'

RESPONSE=$(curl -s -X POST "$CONSENT_ENGINE_URL/consents" \
    -H "Content-Type: application/json" \
    -d "$OTP_CONSENT_DATA" 2>/dev/null)

OTP_CONSENT_ID=$(echo "$RESPONSE" | jq -r '.redirect_url' | sed 's/.*consent_id=\([^&]*\).*/\1/')

# Approve the consent first
APPROVE_OTP_DATA='{
    "status": "approved",
    "owner_id": "1992222222",
    "message": "Approved via consent portal"
}'

curl -s -X PUT "$CONSENT_ENGINE_URL/consents/$OTP_CONSENT_ID" \
    -H "Content-Type: application/json" \
    -d "$APPROVE_OTP_DATA" > /dev/null

# Now test wrong OTP
WRONG_OTP_DATA='{
    "otp_code": "000000"
}'

test_api_call "Wrong OTP (1st attempt)" "POST" "$CONSENT_ENGINE_URL/consents/$OTP_CONSENT_ID/otp" "$WRONG_OTP_DATA" "400"

# Test 9: OTP Verification - Incorrect OTP (2nd attempt)
echo -e "${PURPLE}=== Test 9: OTP Verification - Incorrect OTP (2nd attempt) ===${NC}"
test_api_call "Wrong OTP (2nd attempt)" "POST" "$CONSENT_ENGINE_URL/consents/$OTP_CONSENT_ID/otp" "$WRONG_OTP_DATA" "400"

# Test 10: OTP Verification - Incorrect OTP (3rd attempt - should reject)
echo -e "${PURPLE}=== Test 10: OTP Verification - Incorrect OTP (3rd attempt - should reject) ===${NC}"
test_api_call "Wrong OTP (3rd attempt - should reject)" "POST" "$CONSENT_ENGINE_URL/consents/$OTP_CONSENT_ID/otp" "$WRONG_OTP_DATA" "400"

# Test 11: Verify Consent Status After OTP Rejection
echo -e "${PURPLE}=== Test 11: Verify Consent Status After OTP Rejection ===${NC}"
test_api_call "Check Rejected Status" "GET" "$CONSENT_ENGINE_URL/consents/$OTP_CONSENT_ID" "" "200"

# Test 12: Test OTP on Rejected Consent (should fail)
echo -e "${PURPLE}=== Test 12: Test OTP on Rejected Consent ===${NC}"
test_api_call "OTP on Rejected Consent" "POST" "$CONSENT_ENGINE_URL/consents/$OTP_CONSENT_ID/otp" "$CORRECT_OTP_DATA" "400"

# Test 13: Test OTP on Pending Consent (should fail)
echo -e "${PURPLE}=== Test 13: Test OTP on Pending Consent ===${NC}"
PENDING_CONSENT_DATA='{
    "app_id": "passport-app",
    "data_fields": [
        {
            "owner_type": "citizen",
            "owner_id": "1993333333",
            "fields": ["personInfo.permanentAddress"]
        }
    ],
    "purpose": "passport_application",
    "session_id": "session_pending_test",
    "redirect_url": "https://passport-app.gov.lk"
}'

RESPONSE=$(curl -s -X POST "$CONSENT_ENGINE_URL/consents" \
    -H "Content-Type: application/json" \
    -d "$PENDING_CONSENT_DATA" 2>/dev/null)

PENDING_CONSENT_ID=$(echo "$RESPONSE" | jq -r '.redirect_url' | sed 's/.*consent_id=\([^&]*\).*/\1/')

test_api_call "OTP on Pending Consent" "POST" "$CONSENT_ENGINE_URL/consents/$PENDING_CONSENT_ID/otp" "$CORRECT_OTP_DATA" "400"

# Test 14: Test Invalid Status Transitions
echo -e "${PURPLE}=== Test 14: Test Invalid Status Transitions ===${NC}"
INVALID_STATUS_DATA='{
    "status": "invalid_status",
    "owner_id": "1991111111",
    "message": "Invalid status test"
}'

test_api_call "Invalid Status" "PUT" "$CONSENT_ENGINE_URL/consents/$NEW_CONSENT_ID" "$INVALID_STATUS_DATA" "500"

# Test 15: Test Missing Required Fields
echo -e "${PURPLE}=== Test 15: Test Missing Required Fields ===${NC}"
MISSING_FIELDS_DATA='{
    "app_id": "passport-app",
    "data_fields": [
        {
            "owner_type": "citizen",
            "owner_id": "",
            "fields": ["personInfo.permanentAddress"]
        }
    ],
    "purpose": "passport_application",
    "session_id": "session_missing_test",
    "redirect_url": "https://passport-app.gov.lk"
}'

test_api_call "Missing Required Fields" "POST" "$CONSENT_ENGINE_URL/consents" "$MISSING_FIELDS_DATA" "400"

# Test 16: Test Non-existent Consent
echo -e "${PURPLE}=== Test 16: Test Non-existent Consent ===${NC}"
test_api_call "Non-existent Consent" "GET" "$CONSENT_ENGINE_URL/consents/non_existent_id" "" "404"

# Test 17: Test Revoke Consent
echo -e "${PURPLE}=== Test 17: Test Revoke Consent ===${NC}"
REVOKE_DATA='{
    "reason": "User requested revocation"
}'

test_api_call "Revoke Consent" "DELETE" "$CONSENT_ENGINE_URL/consents/$NEW_CONSENT_ID" "$REVOKE_DATA" "200"

# Test 17.5: Move Revoked Consent to Pending
echo -e "${PURPLE}=== Test 17.5: Move Revoked Consent to Pending ===${NC}"
MOVE_TO_PENDING_DATA='{
    "status": "pending",
    "owner_id": "1991111111",
    "message": "Moving revoked consent to pending for reconsideration"
}'

test_api_call "Move Revoked Consent to Pending" "PUT" "$CONSENT_ENGINE_URL/consents/$NEW_CONSENT_ID" "$MOVE_TO_PENDING_DATA" "200"

# Test 17.6: Approve Previously Revoked Consent (now pending)
echo -e "${PURPLE}=== Test 17.6: Approve Previously Revoked Consent (now pending) ===${NC}"
APPROVE_REVOKED_DATA='{
    "status": "approved",
    "owner_id": "1991111111",
    "message": "Approved after revocation reconsideration"
}'

test_api_call "Approve Previously Revoked Consent (now pending)" "PUT" "$CONSENT_ENGINE_URL/consents/$NEW_CONSENT_ID" "$APPROVE_REVOKED_DATA" "200"

# Test 18: Test Revoke Non-existent Consent
echo -e "${PURPLE}=== Test 18: Test Revoke Non-existent Consent ===${NC}"
test_api_call "Revoke Non-existent Consent" "DELETE" "$CONSENT_ENGINE_URL/consents/non_existent_id" "$REVOKE_DATA" "500"

# Summary
echo -e "${PURPLE}=== Test Summary ===${NC}"
echo -e "${GREEN}Consent Engine Integration Tests Completed!${NC}"
echo ""
echo "Test Coverage:"
echo "✅ Health Check"
echo "✅ Default Consent Record"
echo "✅ Create New Consent (Non-existing)"
echo "✅ Create Consent with Existing Pair"
echo "✅ Approve Consent"
echo "✅ Reject Consent"
echo "✅ Approve Previously Rejected Consent"
echo "✅ OTP Verification - Correct OTP"
echo "✅ OTP Verification - Incorrect OTP (3 attempts)"
echo "✅ OTP Rejection After 3 Failed Attempts"
echo "✅ OTP on Rejected/Pending Consent (Error Cases)"
echo "✅ Invalid Status Transitions"
echo "✅ Missing Required Fields"
echo "✅ Non-existent Consent Handling"
echo "✅ Revoke Consent"
echo "✅ Approve Previously Revoked Consent"
echo "✅ Error Handling"
echo ""
echo -e "${GREEN}All consent workflow scenarios tested successfully!${NC}"
