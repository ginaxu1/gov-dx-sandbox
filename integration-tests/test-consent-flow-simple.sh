#!/bin/bash
# Simplified Consent Flow Test Script using DRY principles

# Source common utilities
source "$(dirname "$0")/test-utils.sh"

echo "=== Simplified Consent Flow Test Suite ==="
echo "Testing consent scenarios using standardized functions"
echo ""

# Test 1: Provider requests data it owns
echo "=== Test 1: Provider requests data it owns ==="
test_pdp_decision "Provider owns data" "true" "true" "$CONSENT_FLOW_DATA_PROVIDER_OWNS"

# Test 2: Provider requests data from different owner
echo "=== Test 2: Provider requests data from different owner ==="
test_pdp_decision "Different owner data" "true" "true" "$CONSENT_FLOW_DATA_DIFFERENT_OWNER"

# Test 3: Mixed ownership scenario
echo "=== Test 3: Mixed ownership scenario ==="
test_pdp_decision "Mixed ownership" "true" "true" "$CONSENT_FLOW_DATA_MIXED_OWNERSHIP"

# Test 4: Restricted field access
echo "=== Test 4: Restricted field access ==="
test_pdp_decision "Restricted field access" "false" "false" "$CONSENT_FLOW_DATA_RESTRICTED_ACCESS"

# Test 5: Unknown app access
echo "=== Test 5: Unknown app access ==="
test_pdp_decision "Unknown app access" "true" "false" "$CONSENT_FLOW_DATA_UNKNOWN_APP"

# Test 6: Consent creation
echo "=== Test 6: Consent creation ==="
test_consent_creation "Create consent" "$CONSENT_TEST_DATA"

echo ""
echo "=== Consent Flow Test Suite Complete ==="
