#!/bin/bash

# Base URL for the consent engine
BASE_URL="http://localhost:8081"

echo "🧪 Testing Hybrid Authentication Debug"
echo "====================================="

# Check if server is running
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ Server is not running. Please start the consent engine server first:"
    echo "   ./consent-engine"
    echo ""
    echo "Then run this test again."
    exit 1
fi

echo "✅ Server is running."

# Create a test consent first
echo ""
echo "Creating a test consent..."
CONSENT_CREATE_PAYLOAD='{
    "app_id": "test-app",
    "data_fields": [
        {
            "owner_type": "citizen",
            "owner_id": "test-owner-123",
            "owner_email": "regina@opensource.lk",
            "fields": ["person.name", "person.email"]
        }
    ],
    "purpose": "testing hybrid auth",
    "session_id": "test-session-123"
}'

CONSENT_RESPONSE=$(curl -s -X POST "$BASE_URL/consents" \
    -H "Content-Type: application/json" \
    -d "$CONSENT_CREATE_PAYLOAD")

CONSENT_ID=$(echo "$CONSENT_RESPONSE" | jq -r '.consent_id')

echo "Consent creation response: $CONSENT_RESPONSE"
echo "Consent ID: $CONSENT_ID"
echo ""

if [ -z "$CONSENT_ID" ] || [ "$CONSENT_ID" == "null" ]; then
    echo "❌ Failed to create test consent. Exiting."
    exit 1
fi

# Test 1: M2M call without JWT (should work)
echo "📡 Test 1: M2M call without JWT (should work)"
RESPONSE=$(curl -s -X GET "$BASE_URL/consents/$CONSENT_ID" \
    -H "Content-Type: application/json" \
    -w "\nHTTP_CODE:%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1 | sed 's/HTTP_CODE://')
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
echo "HTTP_CODE: $HTTP_CODE"
echo ""

# Test 2: Frontend call without JWT (should fail)
echo "🌐 Test 2: Frontend call without JWT (should fail)"
RESPONSE=$(curl -s -X GET "$BASE_URL/consents/$CONSENT_ID" \
    -H "Content-Type: application/json" \
    -H "X-Requested-With: XMLHttpRequest" \
    -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36" \
    -w "\nHTTP_CODE:%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1 | sed 's/HTTP_CODE://')
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
echo "HTTP_CODE: $HTTP_CODE"
echo ""

# Test 3: Frontend call with invalid JWT (should fail)
echo "🔐 Test 3: Frontend call with invalid JWT (should fail)"
RESPONSE=$(curl -s -X GET "$BASE_URL/consents/$CONSENT_ID" \
    -H "Content-Type: application/json" \
    -H "X-Requested-With: XMLHttpRequest" \
    -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36" \
    -H "Authorization: Bearer invalid-token" \
    -w "\nHTTP_CODE:%{http_code}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1 | sed 's/HTTP_CODE://')
BODY=$(echo "$RESPONSE" | sed '$d')

echo "Response: $BODY"
echo "HTTP_CODE: $HTTP_CODE"
echo ""

echo "✅ Debug test completed!"
echo ""
echo "Summary:"
echo "- M2M calls without JWT: Should work (HTTP 200)"
echo "- Frontend calls without JWT: Should fail (HTTP 401)"
echo "- Frontend calls with invalid JWT: Should fail (HTTP 401)"
