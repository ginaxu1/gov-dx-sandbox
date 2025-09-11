#!/bin/bash
# Policy Decision Point (PDP) Test Suite

echo "=== Policy Decision Point (PDP) Test Suite ==="
echo ""

# Test function
test_pdp() {
    local test_name="$1"
    local expected="$2"
    local data="$3"
    
    echo "Test: $test_name"
    echo "Expected: $expected"
    echo ""
    
    curl -X POST http://localhost:8082/decide \
      -H "Content-Type: application/json" \
      -d "$data" | jq '.'
    
    echo "---"
}

# Test 1: Valid request with no consent required
test_pdp "Valid request with no consent required" "allow=true, consent_required=false" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_001",
    "required_fields": ["person.fullName", "person.nic", "person.photo"]
  }'

# Test 2: Valid request with consent required
test_pdp "Valid request with consent required" "allow=true, consent_required=true" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_002",
    "required_fields": ["person.fullName", "person.permanentAddress", "person.birthDate"]
  }'

# Test 3: Invalid consumer
test_pdp "Invalid consumer" "allow=false" '{
    "consumer_id": "unknown-app",
    "app_id": "unknown-app",
    "request_id": "req_003",
    "required_fields": ["person.fullName"]
  }'

# Test 4: Unauthorized field access
test_pdp "Unauthorized field access" "allow=true, consent_required=true" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_004",
    "required_fields": ["person.fullName", "person.birthDate"]
  }'

# Test 5: Invalid action (not applicable with new format)
test_pdp "Single field test" "allow=true, consent_required=false" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_005",
    "required_fields": ["person.fullName"]
  }'

# Test 6: Two fields test
test_pdp "Two fields test" "allow=true, consent_required=false" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_006",
    "required_fields": ["person.fullName", "person.nic"]
  }'

# Test 7: Mixed fields test
test_pdp "Mixed fields test" "allow=false" '{
    "consumer_id": "unauthorized-app",
    "app_id": "unauthorized-app",
    "request_id": "req_007",
    "required_fields": ["person.fullName", "person.permanentAddress"]
  }'

# Test 8: All approved fields test
test_pdp "All approved fields test" "allow=true, consent_required=false" '{
    "consumer_id": "driver-app",
    "app_id": "driver-app",
    "request_id": "req_008",
    "required_fields": ["person.birthDate"]
  }'

# Test 9: Single unauthorized field test
test_pdp "Single unauthorized field test" "allow=true, consent_required=true" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_009",
    "required_fields": ["person.photo"]
  }'

echo ""
echo "PDP Test Suite Complete"