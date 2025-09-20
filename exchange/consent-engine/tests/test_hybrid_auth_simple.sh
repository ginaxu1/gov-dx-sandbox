#!/bin/bash

# Simple test script to demonstrate hybrid authentication behavior
# This test uses curl to test the actual running server

echo "🧪 Testing Hybrid Authentication Behavior"
echo "=========================================="

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Server is not running. Please start the consent engine server first:"
    echo "   go run *.go"
    echo ""
    echo "Then run this test again."
    exit 1
fi

echo "✅ Server is running"

# Test 1: M2M call without JWT (should work)
echo ""
echo "📡 Test 1: M2M call without JWT (should work)"
echo "----------------------------------------------"

# First create a consent
echo "Creating a test consent..."
CONSENT_RESPONSE=$(curl -s -X POST http://localhost:8080/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "test-app",
    "data_fields": [{
      "owner_type": "citizen",
      "owner_id": "test-owner-123",
      "owner_email": "test@example.com",
      "fields": ["person.name", "person.email"]
    }],
    "purpose": "testing",
    "session_id": "test-session-123"
  }')

echo "Consent creation response: $CONSENT_RESPONSE"

# Extract consent ID
CONSENT_ID=$(echo $CONSENT_RESPONSE | jq -r '.consent_id')
echo "Consent ID: $CONSENT_ID"

# Test M2M call without JWT (should work)
echo ""
echo "Testing M2M call without JWT..."
M2M_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "http://localhost:8080/consents/$CONSENT_ID")
echo "M2M response: $M2M_RESPONSE"

# Test 2: Frontend call without JWT (should fail)
echo ""
echo "🌐 Test 2: Frontend call without JWT (should fail)"
echo "------------------------------------------------"

FRONTEND_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "http://localhost:8080/consents/$CONSENT_ID" \
  -H "X-Requested-With: XMLHttpRequest" \
  -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
echo "Frontend response: $FRONTEND_RESPONSE"

# Test 3: M2M call with JWT (should work)
echo ""
echo "🔐 Test 3: M2M call with JWT (should work)"
echo "------------------------------------------"

# Create a mock M2M token (this would normally come from your OAuth provider)
M2M_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjbGllbnQxMjMiLCJjbGllbnRfaWQiOiJjbGllbnQxMjMiLCJzY29wZSI6ImNvbnNlbnQ6cmVhZCBjb25zZW50OndyaXRlIiwiaXNzIjoiaHR0cHM6Ly9hcGkuYXNnYXJkZW8uaW8vdC90ZXN0b3JnL29hdXRoMi90b2tlbiIsImF1ZCI6ImNvbnNlbnQtYXBpIiwiaWF0IjoxMjM0NTY3ODkwLCJleHAiOjEyMzQ1NzE0OTB9.invalid_signature"

M2M_WITH_JWT_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "http://localhost:8080/consents/$CONSENT_ID" \
  -H "Authorization: Bearer $M2M_TOKEN")
echo "M2M with JWT response: $M2M_WITH_JWT_RESPONSE"

# Test 4: Frontend call with JWT (should work)
echo ""
echo "🌐🔐 Test 4: Frontend call with JWT (should work)"
echo "------------------------------------------------"

# Create a mock user token (this would normally come from your OAuth provider)
USER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwiaXNzIjoiaHR0cHM6Ly9hcGkuYXNnYXJkZW8uaW8vdC90ZXN0b3JnL29hdXRoMi90b2tlbiIsImF1ZCI6ImNvbnNlbnQtYXBpIiwiaWF0IjoxMjM0NTY3ODkwLCJleHAiOjEyMzQ1NzE0OTB9.invalid_signature"

FRONTEND_WITH_JWT_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X GET "http://localhost:8080/consents/$CONSENT_ID" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Requested-With: XMLHttpRequest" \
  -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
echo "Frontend with JWT response: $FRONTEND_WITH_JWT_RESPONSE"

echo ""
echo "✅ Test completed!"
echo ""
echo "Summary:"
echo "- M2M calls without JWT: Should work (HTTP 200)"
echo "- Frontend calls without JWT: Should fail (HTTP 401)"
echo "- M2M calls with JWT: Should work (HTTP 200)"
echo "- Frontend calls with JWT: Should work (HTTP 200)"
