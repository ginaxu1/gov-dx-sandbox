#!/bin/bash

# Direct test of audit service endpoint
# This tests if the audit service can receive and store management events

set -e

API_SERVER_URL="${API_SERVER_URL:-http://localhost:3000}"
AUDIT_SERVICE_URL="${AUDIT_SERVICE_URL:-http://localhost:3001}"

echo "=== Testing Audit Service Directly ==="
echo ""

# Test 1: Check audit service health
echo "1. Checking audit service health..."
HEALTH=$(curl -s "$AUDIT_SERVICE_URL/health")
echo "$HEALTH" | jq '.' 2>/dev/null || echo "$HEALTH"
echo ""

# Test 2: Send a test management event directly to audit service
echo "2. Sending test management event to audit service..."
TEST_EVENT=$(cat <<EOF
{
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
}
EOF
)

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$AUDIT_SERVICE_URL/api/events" \
  -H "Content-Type: application/json" \
  -d "$TEST_EVENT")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | grep -v "HTTP_CODE:")

if [ "$HTTP_CODE" = "201" ]; then
  echo "✓ Test event created successfully (HTTP $HTTP_CODE)"
  echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
  echo "✗ Failed to create test event (HTTP $HTTP_CODE)"
  echo "Response: $BODY"
fi
echo ""

# Test 3: Query management events
echo "3. Querying management events..."
EVENTS=$(curl -s "$AUDIT_SERVICE_URL/api/events?limit=5")
echo "$EVENTS" | jq '.' 2>/dev/null || echo "$EVENTS"
echo ""

# Test 4: Check if API server can reach audit service
echo "4. Testing API server -> Audit service connection..."
echo "   (This requires making an actual API request with JWT token)"
echo "   To test: Make a POST request to $API_SERVER_URL/api/v1/members with valid JWT"
echo ""

echo "=== Test Complete ==="

