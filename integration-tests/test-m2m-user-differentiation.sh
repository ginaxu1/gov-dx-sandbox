#!/bin/bash

# Test script to verify M2M vs User differentiation in audit logs
# This script sends different types of requests to test the classification

echo "🧪 Testing M2M vs User Differentiation in Audit Logs"
echo "=================================================="

# Set up environment variables
export REDIS_ADDR=localhost:6379
export CHOREO_DB_AUDIT_HOSTNAME=pg-41200aa141064e6cbabf311dce37c04a-opendifd1461627769-choreo-o.h.aivencloud.com
export CHOREO_DB_AUDIT_PORT=19847
export CHOREO_DB_AUDIT_USERNAME=avnadmin
export CHOREO_DB_AUDIT_PASSWORD=AVNS_HwUxELSQImHrLu9XnYD
export CHOREO_DB_AUDIT_DATABASENAME=defaultdb
export DB_SSLMODE=require

ORCHESTRATION_ENGINE_URL="http://localhost:4000"
AUDIT_SERVICE_URL="http://localhost:3001"

echo "📊 Checking current audit logs count..."
INITIAL_COUNT=$(curl -s "$AUDIT_SERVICE_URL/audit-logs?limit=1" | jq -r '.total // 0')
echo "Initial audit logs count: $INITIAL_COUNT"

echo ""
echo "🔧 Test 1: M2M Request (API Key Authentication)"
echo "----------------------------------------------"
curl -X POST "$ORCHESTRATION_ENGINE_URL/" \
  -H "Authorization: ApiKey test-api-key-12345" \
  -H "User-Agent: curl/7.68.0" \
  -H "X-Client-Type: system" \
  -H "X-Schema-ID: test-schema-m2m" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { types { name } } }"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "👤 Test 2: User Request (JWT Authentication)"
echo "--------------------------------------------"
curl -X POST "$ORCHESTRATION_ENGINE_URL/" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36" \
  -H "X-User-ID: user-12345" \
  -H "X-Session-ID: session-abc123" \
  -H "X-Schema-ID: test-schema-user" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { types { name } } }"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "🤖 Test 3: System Request (Service Account)"
echo "--------------------------------------------"
curl -X POST "$ORCHESTRATION_ENGINE_URL/" \
  -H "Authorization: Basic c2VydmljZS1hY2NvdW50OnBhc3N3b3Jk" \
  -H "User-Agent: systemd/247" \
  -H "X-System-Request: true" \
  -H "X-Schema-ID: test-schema-system" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { types { name } } }"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "📦 Test 4: Batch Job Request"
echo "----------------------------"
curl -X POST "$ORCHESTRATION_ENGINE_URL/" \
  -H "Authorization: Bearer batch-token" \
  -H "User-Agent: batch-processor/1.0" \
  -H "X-Batch-Job: true" \
  -H "X-Schema-ID: test-schema-batch" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { types { name } } }"}' \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "⏳ Waiting 5 seconds for audit processing..."
sleep 5

echo ""
echo "📊 Checking audit logs with new fields..."
curl -s "$AUDIT_SERVICE_URL/audit-logs?limit=10" | jq -r '
  .logs[] | 
  "ID: \(.id) | Type: \(.requestType // "N/A") | Auth: \(.authMethod // "N/A") | User: \(.userId // "N/A") | Session: \(.sessionId // "N/A") | App: \(.applicationId)"
'

echo ""
echo "🔍 Detailed analysis of request types:"
curl -s "$AUDIT_SERVICE_URL/audit-logs?limit=20" | jq -r '
  .logs[] | 
  select(.requestType != null) |
  "Request Type: \(.requestType) | Auth Method: \(.authMethod) | User ID: \(.userId // "none") | Session ID: \(.sessionId // "none")"
'

echo ""
echo "✅ Test completed! Check the audit logs above to verify M2M vs User differentiation."
