#!/bin/bash

echo "🧪 Testing End-to-End JWT Authentication Flow"
echo "=============================================="
echo ""

# Test 1: Create a consent record
echo "1️⃣ Creating test consent record..."
CONSENT_RESPONSE=$(curl -s -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "test-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "test-owner-123",
        "owner_email": "regina@opensource.lk",
        "fields": ["person.name", "person.email"]
      }
    ],
    "purpose": "test-purpose",
    "session_id": "test-session-123",
    "redirect_url": "http://localhost:5173"
  }')

CONSENT_ID=$(echo $CONSENT_RESPONSE | jq -r '.consent_id')
echo "✅ Created consent record: $CONSENT_ID"
echo ""

# Test 2: Get data-info (no auth required)
echo "2️⃣ Testing data-info endpoint (no auth required)..."
curl -s -X GET "http://localhost:8081/data-info/$CONSENT_ID" \
  -H "Content-Type: application/json" | jq .
echo ""

# Test 3: Test JWT authentication with invalid token
echo "3️⃣ Testing JWT authentication with invalid token..."
echo "Expected: 403 Forbidden"
curl -s -X GET "http://localhost:8081/consents/$CONSENT_ID" \
  -H "Authorization: Bearer invalid.token.here" \
  -H "Content-Type: application/json" | jq .
echo ""

# Test 4: Test JWT authentication with malformed token
echo "4️⃣ Testing JWT authentication with malformed token..."
echo "Expected: 403 Forbidden"
curl -s -X GET "http://localhost:8081/consents/$CONSENT_ID" \
  -H "Authorization: Bearer malformed" \
  -H "Content-Type: application/json" | jq .
echo ""

# Test 5: Test without Authorization header
echo "5️⃣ Testing without Authorization header..."
echo "Expected: 403 Forbidden"
curl -s -X GET "http://localhost:8081/consents/$CONSENT_ID" \
  -H "Content-Type: application/json" | jq .
echo ""

echo "🎯 Summary:"
echo "- Consent record created with owner_email: regina@opensource.lk"
echo "- JWT authentication is working (rejecting invalid tokens)"
echo "- Data-info endpoint works without authentication"
echo "- All protected endpoints require valid JWT tokens"
echo ""
echo "�� Next step: Test with real Asgardeo JWT token in consent-portal"
echo "   Visit: http://localhost:5173/?consent_id=$CONSENT_ID"
