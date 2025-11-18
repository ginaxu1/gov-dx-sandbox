#/bin/bash

# Complete integration test: API Server -> Audit Service -> Database

set -e

DB_HOST="pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com"
DB_PORT="19847"
DB_USER="avnadmin"
DB_PASSWORD="AVNS_HwUxELSQImHrLu9XnYD"
DB_NAME="testdb2"

export PGPASSWORD="$DB_PASSWORD"

echo "=========================================="
echo "Audit Integration Test"
echo "=========================================="
echo ""

# Step 1: Kill any existing audit service
echo "1. Stopping any existing audit service..."
pkill -f "audit-service|go run.*audit" 2>/dev/null || true
sleep 2
echo "   ✓ Done"
echo ""

# Step 2: Start audit service with correct DB config
echo "2. Starting audit service with testdb2 configuration..."
cd /Users/tmp/opendif-mvp/audit-service

export CHOREO_OPENDIF_DATABASE_HOSTNAME="$DB_HOST"
export CHOREO_OPENDIF_DATABASE_PORT="$DB_PORT"
export CHOREO_OPENDIF_DATABASE_USERNAME="$DB_USER"
export CHOREO_OPENDIF_DATABASE_PASSWORD="$DB_PASSWORD"
export CHOREO_OPENDIF_DATABASE_DATABASENAME="$DB_NAME"
export DB_SSLMODE="require"

nohup go run . > /tmp/audit-service-test.log 2>&1 &
AUDIT_PID=$!
echo "   Audit service PID: $AUDIT_PID"
echo "   Waiting for service to start..."
sleep 8

# Check if service is running
if ! ps -p $AUDIT_PID > /dev/null 2>&1; then
    echo "   ✗ Service failed to start"
    tail -30 /tmp/audit-service-test.log
    exit 1
fi

# Check health endpoint
HEALTH=$(curl -s http://localhost:3001/health 2>/dev/null | jq -r '.status' 2>/dev/null || echo "unknown")
if [ "$HEALTH" != "healthy" ]; then
    echo "   ✗ Service health check failed (status: $HEALTH)"
    tail -30 /tmp/audit-service-test.log
    exit 1
fi
echo "   ✓ Service is healthy"
echo ""

# Step 3: Get initial event count
echo "3. Getting initial event count from database..."
INITIAL_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM management_events;" 2>&1 | tr -d ' ' || echo "0")
echo "   Initial count: $INITIAL_COUNT"
echo ""

# Step 4: Create test event via audit service API (simulating API server call)
echo "4. Creating test management event..."
TEST_TIMESTAMP=$(date +%s)
TEST_EVENT=$(cat <<EOF
{
  "eventType": "UPDATE",
  "actor": {
    "type": "USER",
    "id": "test-user-$TEST_TIMESTAMP",
    "role": "ADMIN"
  },
  "target": {
    "resource": "MEMBERS",
    "resourceId": "test-member-$TEST_TIMESTAMP"
  }
}
EOF
)

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST http://localhost:3001/api/events \
  -H "Content-Type: application/json" \
  -d "$TEST_EVENT" 2>&1)

HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2 || echo "000")
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//' | head -1)

if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "   ✓ Event created successfully (HTTP $HTTP_CODE)"
    EVENT_ID=$(echo "$BODY" | jq -r '.eventId // .id // "unknown"' 2>/dev/null || echo "unknown")
    echo "   Event ID: $EVENT_ID"
else
    echo "   ✗ Failed to create event (HTTP $HTTP_CODE)"
    echo "   Response: $BODY"
    echo ""
    echo "   Service logs:"
    tail -20 /tmp/audit-service-test.log
    pkill -P $AUDIT_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Step 5: Wait for async processing
echo "5. Waiting for event to be processed..."
sleep 3
echo ""

# Step 6: Verify in database
echo "6. Verifying event in database..."
NEW_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM management_events;" 2>&1 | tr -d ' ' || echo "0")
echo "   New count: $NEW_COUNT"

if [ "$NEW_COUNT" -gt "$INITIAL_COUNT" ]; then
    echo "   ✓ SUCCESS: New event was created!"
    echo ""
    echo "   Latest event details:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT event_type, actor_type, actor_id, actor_role, target_resource, target_resource_id, timestamp FROM management_events WHERE actor_id = 'test-user-$TEST_TIMESTAMP' ORDER BY timestamp DESC LIMIT 1;" 2>&1 | tail -5
else
    echo "   ✗ FAILURE: No new event was created"
    echo "   (Count remained at $INITIAL_COUNT)"
    echo ""
    echo "   Service logs:"
    tail -30 /tmp/audit-service-test.log
    pkill -P $AUDIT_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Step 7: Cleanup
echo "7. Cleaning up..."
pkill -P $AUDIT_PID 2>/dev/null || true
sleep 1
echo "   ✓ Done"
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "✓ Audit service started successfully"
echo "✓ Service health check passed"
echo "✓ Event created via API endpoint"
echo "✓ Event verified in database (testdb2.management_events)"
echo ""
echo "The audit service is correctly configured and working!"
echo ""
echo "Next step: Make an actual API request to the API Server"
echo "with a valid JWT token to test the full flow:"
echo "  API Server -> Audit Middleware -> Audit Service -> Database"
echo ""

