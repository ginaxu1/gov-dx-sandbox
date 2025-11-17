# Orchestration Engine to Audit Service Integration Test

## Overview

The Orchestration Engine sends **Data Exchange Events** to the Audit Service when it federates GraphQL queries to providers. These events are stored in the `audit_logs` table (separate from `management_events`).

## Flow

```
GraphQL Query → Orchestration Engine
  → Federates to Providers
  → For each provider response:
    → Looks up consumerId (member ID from consumerAppID)
    → Looks up providerId (member ID from providerSchemaID)
    → Sends DataExchangeEvent to POST /v1/audit/exchange
    → Stored in audit_logs table
```

## Prerequisites

1. **Start Audit Service**:
```bash
cd audit-service
export CHOREO_OPENDIF_DATABASE_HOSTNAME=pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com
export CHOREO_OPENDIF_DATABASE_PORT=19847
export CHOREO_OPENDIF_DATABASE_USERNAME=avnadmin
export CHOREO_OPENDIF_DATABASE_PASSWORD='AVNS_HwUxELSQImHrLu9XnYD'
export CHOREO_OPENDIF_DATABASE_DATABASENAME=testdb2
export DB_SSLMODE=require
go run . &
```

2. **Start API Server** (needed for member ID lookups):
```bash
cd api-server-go
export CHOREO_AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
# ... other required env vars ...
go run . &
```

3. **Start Orchestration Engine**:
```bash
cd exchange/orchestration-engine-go
export CHOREO_AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
export CHOREO_API_SERVER_URL=http://localhost:8080
# ... other required env vars ...
go run . &
```

## Test Commands

### 1. Verify Services are Running

```bash
# Check audit service
curl http://localhost:3001/health

# Check API server
curl http://localhost:8080/health

# Check orchestration engine
curl http://localhost:4000/health
```

### 2. Send GraphQL Query to Orchestration Engine

```bash
# Example GraphQL query
QUERY='{
  "query": "query { personInfo { name address } }"
}'

# Send to orchestration engine
curl -X POST http://localhost:4000/graphql \
  -H "Content-Type: application/json" \
  -d "$QUERY"
```

**Note**: The orchestration engine will:
1. Parse the query
2. Identify which providers to call
3. For each provider:
   - Look up consumerId from consumerAppID via API Server
   - Look up providerId from providerSchemaID via API Server
   - Send audit event to audit-service
   - Store in `audit_logs` table

### 3. Verify Data Exchange Event in Audit Service

```bash
# Get data exchange logs (if endpoint exists)
curl http://localhost:3001/api/logs

# Filter by consumer
curl "http://localhost:3001/api/logs?consumerId=member-consumer-123"

# Filter by provider
curl "http://localhost:3001/api/logs?providerId=member-provider-456"
```

### 4. Query Database Directly

```bash
export PGPASSWORD='AVNS_HwUxELSQImHrLu9XnYD'
psql -h pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com \
     -p 19847 -U avnadmin -d testdb2
```

```sql
-- View all data exchange events
SELECT 
    id,
    timestamp,
    status,
    application_id,
    schema_id,
    consumer_id,
    provider_id,
    requested_data
FROM audit_logs
ORDER BY timestamp DESC
LIMIT 10;

-- Count events by status
SELECT status, COUNT(*) 
FROM audit_logs 
GROUP BY status;

-- Count events by consumer
SELECT consumer_id, COUNT(*) 
FROM audit_logs 
WHERE consumer_id IS NOT NULL
GROUP BY consumer_id;

-- Count events by provider
SELECT provider_id, COUNT(*) 
FROM audit_logs 
WHERE provider_id IS NOT NULL
GROUP BY provider_id;

-- View events for specific consumer-provider pair
SELECT * 
FROM audit_logs 
WHERE consumer_id = 'member-consumer-123' 
  AND provider_id = 'member-provider-456'
ORDER BY timestamp DESC;

-- View recent SUCCESS events
SELECT * 
FROM audit_logs 
WHERE status = 'success' 
ORDER BY timestamp DESC 
LIMIT 5;

-- View recent FAILURE events
SELECT * 
FROM audit_logs 
WHERE status = 'failure' 
ORDER BY timestamp DESC 
LIMIT 5;
```

### 5. Test Direct Audit Service Endpoint

```bash
# Send a data exchange event directly to audit service
curl -X POST http://localhost:3001/v1/audit/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "consumerAppId": "passport-app",
    "consumerId": "member-consumer-123",
    "providerSchemaId": "drp-schema-v1",
    "providerId": "member-provider-456",
    "requestedFields": ["personInfo.name", "personInfo.address"],
    "status": "SUCCESS"
  }'
```

### 6. Test with Missing Required Fields

```bash
# This should fail - missing consumerId
curl -X POST http://localhost:3001/v1/audit/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "consumerAppId": "passport-app",
    "providerSchemaId": "drp-schema-v1",
    "providerId": "member-provider-456",
    "requestedFields": ["personInfo.name"],
    "status": "SUCCESS"
  }'

# This should fail - missing providerId
curl -X POST http://localhost:3001/v1/audit/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "consumerAppId": "passport-app",
    "consumerId": "member-consumer-123",
    "providerSchemaId": "drp-schema-v1",
    "requestedFields": ["personInfo.name"],
    "status": "SUCCESS"
  }'
```

## Expected Results

1. **Data Exchange Event Structure**:
   - `consumerAppId` → stored as `application_id`
   - `providerSchemaId` → stored as `schema_id`
   - `consumerId` → stored as `consumer_id` (REQUIRED, member ID)
   - `providerId` → stored as `provider_id` (REQUIRED, member ID)
   - `requestedFields` → stored as `requested_data` (JSON array)
   - `status` → stored as `status` ("success" or "failure")

2. **Database Storage**:
   - Table: `audit_logs` (NOT `management_events`)
   - Both `consumer_id` and `provider_id` must be present (NOT NULL)
   - Enables member-to-member data exchange tracking

3. **Validation**:
   - Audit service validates that `consumerId` and `providerId` are required
   - Returns 400 Bad Request if missing

## Complete Test Script

```bash
#!/bin/bash

AUDIT_SERVICE_URL="http://localhost:3001"
ORCHESTRATION_ENGINE_URL="http://localhost:4000"
API_SERVER_URL="http://localhost:8080"

echo "=== Testing Orchestration Engine to Audit Service Integration ==="

# 1. Check services
echo "1. Checking services..."
curl -s $AUDIT_SERVICE_URL/health | jq .
curl -s $API_SERVER_URL/health | jq .
curl -s $ORCHESTRATION_ENGINE_URL/health | jq .

# 2. Get initial event count
echo "2. Initial event count..."
INITIAL_COUNT=$(export PGPASSWORD='AVNS_HwUxELSQImHrLu9XnYD' && \
  psql -h pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com \
       -p 19847 -U avnadmin -d testdb2 -t -c \
       "SELECT COUNT(*) FROM audit_logs;" 2>&1 | tr -d ' ')
echo "Initial count: $INITIAL_COUNT"

# 3. Send test data exchange event directly
echo "3. Sending test data exchange event..."
curl -s -X POST $AUDIT_SERVICE_URL/v1/audit/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "consumerAppId": "test-consumer-app",
    "consumerId": "test-consumer-member",
    "providerSchemaId": "test-provider-schema",
    "providerId": "test-provider-member",
    "requestedFields": ["field1", "field2"],
    "status": "SUCCESS"
  }' | jq .

sleep 2

# 4. Verify event was created
echo "4. Verifying event was created..."
FINAL_COUNT=$(export PGPASSWORD='AVNS_HwUxELSQImHrLu9XnYD' && \
  psql -h pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com \
       -p 19847 -U avnadmin -d testdb2 -t -c \
       "SELECT COUNT(*) FROM audit_logs;" 2>&1 | tr -d ' ')
echo "Final count: $FINAL_COUNT"

if [ "$FINAL_COUNT" -gt "$INITIAL_COUNT" ]; then
  echo "✅ Event was created successfully"
else
  echo "❌ Event was not created"
fi

# 5. View the created event
echo "5. Viewing created event..."
export PGPASSWORD='AVNS_HwUxELSQImHrLu9XnYD'
psql -h pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com \
     -p 19847 -U avnadmin -d testdb2 -c \
     "SELECT application_id, schema_id, consumer_id, provider_id, status, timestamp FROM audit_logs ORDER BY timestamp DESC LIMIT 1;"

echo "=== Test Complete ==="
```

## Troubleshooting

1. **No events in database**:
   - Check orchestration engine logs for audit client errors
   - Verify `CHOREO_AUDIT_CONNECTION_SERVICEURL` is set correctly
   - Check if member ID lookups are working (consumerId/providerId)

2. **Missing consumerId or providerId**:
   - Verify API Server is running and accessible
   - Check if `lookupMemberIDFromApplication()` and `lookupMemberIDFromSchema()` are working
   - Events will be skipped if member IDs cannot be found

3. **Events not appearing immediately**:
   - Audit logging is asynchronous (fire-and-forget)
   - Wait 1-2 seconds before querying database
   - Check audit service logs for any errors

4. **Validation errors**:
   - Ensure `consumerId` and `providerId` are included in the request
   - Check audit service logs for validation error messages

