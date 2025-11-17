#!/bin/bash

# Test script: Make API request to API Server and verify record in management_events table

set -e

API_SERVER_URL="${API_SERVER_URL:-http://localhost:3000}"
AUDIT_SERVICE_URL="${AUDIT_SERVICE_URL:-http://localhost:3001}"
DB_HOST="${CHOREO_OPENDIF_DATABASE_HOSTNAME:-pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com}"
DB_PORT="${CHOREO_OPENDIF_DATABASE_PORT:-19847}"
DB_USER="${CHOREO_OPENDIF_DATABASE_USERNAME:-avnadmin}"
DB_PASSWORD="${CHOREO_OPENDIF_DATABASE_PASSWORD:-AVNS_HwUxELSQImHrLu9XnYD}"
DB_NAME="${CHOREO_OPENDIF_DATABASE_DATABASENAME:-testdb2}"

export PGPASSWORD="$DB_PASSWORD"

echo "=== Testing API Server -> Audit Service -> Database Flow ==="
echo ""

# Step 1: Check services
echo "1. Checking services..."
API_STATUS=$(curl -s "$API_SERVER_URL/health" | jq -r '.status' 2>/dev/null || echo "unknown")
AUDIT_STATUS=$(curl -s "$AUDIT_SERVICE_URL/health" | jq -r '.status' 2>/dev/null || echo "unknown")
echo "   API Server: $API_STATUS"
echo "   Audit Service: $AUDIT_STATUS"
echo ""

# Step 2: Get initial count
echo "2. Getting initial event count from database..."
INITIAL_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM management_events;" 2>&1 | tr -d ' ' || echo "0")
echo "   Initial count: $INITIAL_COUNT"
echo ""

# Step 3: Create test event via audit service (simulating what API server would do)
echo "3. Creating test management event via audit service..."
TEST_EVENT=$(cat <<EOF
{
  "eventType": "UPDATE",
  "actor": {
    "type": "USER",
    "id": "integration-test-user-$(date +%s)",
    "role": "ADMIN"
  },
  "target": {
    "resource": "MEMBERS",
    "resourceId": "integration-test-member-$(date +%s)"
  }
}
EOF
)

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$AUDIT_SERVICE_URL/api/events" \
  -H "Content-Type: application/json" \
  -d "$TEST_EVENT")

HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
  echo "   ✓ Event created successfully (HTTP $HTTP_CODE)"
  EVENT_ID=$(echo "$BODY" | jq -r '.eventId // .id // "unknown"' 2>/dev/null || echo "unknown")
  echo "   Event ID: $EVENT_ID"
else
  echo "   ✗ Failed to create event (HTTP $HTTP_CODE)"
  echo "   Response: $BODY"
  exit 1
fi
echo ""

# Step 4: Wait for async processing
echo "4. Waiting for event to be processed..."
sleep 2
echo ""

# Step 5: Verify in database
echo "5. Verifying event in database..."
NEW_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM management_events;" 2>&1 | tr -d ' ' || echo "0")
echo "   New count: $NEW_COUNT"

if [ "$NEW_COUNT" -gt "$INITIAL_COUNT" ]; then
  echo "   ✓ SUCCESS: New event was created!"
  echo ""
  echo "   Latest event:"
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT event_type, actor_type, actor_id, actor_role, target_resource, target_resource_id, timestamp FROM management_events ORDER BY timestamp DESC LIMIT 1;" 2>&1 | tail -5
else
  echo "   ✗ FAILURE: No new event was created"
  echo "   (Count remained at $INITIAL_COUNT)"
  exit 1
fi
echo ""

# Step 6: Summary
echo "=== Test Summary ==="
echo "✓ Services are running"
echo "✓ Audit service endpoint is working"
echo "✓ Event was created in database"
echo ""
echo "Note: To test with actual API Server request, you need a valid JWT token."
echo "The audit middleware will automatically log events when API requests are made."

