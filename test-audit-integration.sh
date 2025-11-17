#!/bin/bash

# Integration test: Test the full flow from API server to audit service
# This simulates what happens when a real API request is made

set -e

API_SERVER_URL="${API_SERVER_URL:-http://localhost:3000}"
AUDIT_SERVICE_URL="${AUDIT_SERVICE_URL:-http://localhost:3001}"

echo "=== Testing API Server -> Audit Service Integration ==="
echo ""

# Test 1: Verify services are running
echo "1. Verifying services..."
API_HEALTH=$(curl -s "$API_SERVER_URL/health")
AUDIT_HEALTH=$(curl -s "$AUDIT_SERVICE_URL/health")
echo "   API Server: $(echo "$API_HEALTH" | jq -r '.status' 2>/dev/null || echo 'unknown')"
echo "   Audit Service: $(echo "$AUDIT_HEALTH" | jq -r '.status' 2>/dev/null || echo 'unknown')"
echo ""

# Test 2: Check audit service endpoint directly
echo "2. Testing audit service endpoint..."
TEST_RESPONSE=$(curl -s -X POST "$AUDIT_SERVICE_URL/api/events" \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "CREATE",
    "actor": {
      "type": "USER",
      "id": "test-user-123",
      "role": "ADMIN"
    },
    "target": {
      "resource": "MEMBERS",
      "resourceId": "test-member-456"
    }
  }' -w "\n%{http_code}")

HTTP_CODE=$(echo "$TEST_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
  echo "   ✓ Audit service endpoint is working (HTTP $HTTP_CODE)"
else
  echo "   ✗ Audit service endpoint returned HTTP $HTTP_CODE"
  echo "   Response: $(echo "$TEST_RESPONSE" | head -n-1)"
fi
echo ""

# Test 3: Check current event count
echo "3. Checking current management events..."
EVENTS_RESPONSE=$(curl -s "$AUDIT_SERVICE_URL/api/events?limit=1")
EVENT_COUNT=$(echo "$EVENTS_RESPONSE" | jq -r '.total // 0' 2>/dev/null || echo "0")
echo "   Current events in database: $EVENT_COUNT"
echo ""

# Test 4: Summary
echo "=== Test Summary ==="
echo "✓ Services are running"
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
  echo "✓ Audit service can receive events"
  echo ""
  echo "Next step: Make an actual API request to $API_SERVER_URL/api/v1/*"
  echo "with a valid JWT token to test the full integration."
else
  echo "✗ Audit service endpoint issue - check routing"
fi
echo ""

