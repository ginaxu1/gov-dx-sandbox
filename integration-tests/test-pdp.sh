#!/bin/bash
# Policy Decision Point (PDP) Test Suite

# Source common utilities
source "$(dirname "$0")/test-utils.sh"

echo "=== Policy Decision Point (PDP) Test Suite ==="
echo ""

# Test 1: Valid request with no consent required
test_pdp_decision "Valid request with no consent required" "true" "false" "$PDP_TEST_DATA_PUBLIC_FIELD"

# Test 2: Valid request with consent required
test_pdp_decision "Valid request with consent required" "true" "true" "$PDP_TEST_DATA_CONSENT_REQUIRED"

# Test 3: Invalid consumer
test_pdp_decision "Invalid consumer" "false" "false" "$PDP_TEST_DATA_RESTRICTED_FIELD"

# Test 4: Unauthorized field access
test_pdp_decision "Unauthorized field access" "false" "false" "$PDP_TEST_DATA_RESTRICTED_FIELD"

# Test 5: Single field test
test_pdp_decision "Single field test" "true" "false" "$PDP_TEST_DATA_PUBLIC_FIELD"

# Test 6: Two fields test
test_pdp_decision "Two fields test" "true" "true" "$PDP_TEST_DATA_CONSENT_REQUIRED"

# Test 7: Mixed fields test
test_pdp_decision "Mixed fields test" "false" "false" "$PDP_TEST_DATA_RESTRICTED_FIELD"

# Test 8: All approved fields test
test_pdp_decision "All approved fields test" "true" "false" "$PDP_TEST_DATA_AUTHORIZED_RESTRICTED"

# Test 9: Single unauthorized field test
test_pdp_decision "Single unauthorized field test" "true" "true" '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_009",
    "required_fields": ["person.photo"]
  }'

echo ""
echo "PDP Test Suite Complete"