#!/bin/bash

# Test SCIM Integration Script
# This script demonstrates how to test the SCIM integration with Asgardeo

echo "🔧 Testing SCIM Integration with Asgardeo"
echo "========================================"

# Check if environment variables are set
if [ -z "$ASGARDEO_M2M_CLIENT_ID" ] || [ -z "$ASGARDEO_M2M_CLIENT_SECRET" ]; then
    echo "⚠️  M2M credentials not configured. The service will use hardcoded mappings."
    echo "   To enable SCIM integration, set:"
    echo "   export ASGARDEO_M2M_CLIENT_ID=your-m2m-client-id"
    echo "   export ASGARDEO_M2M_CLIENT_SECRET=your-m2m-client-secret"
    echo ""
fi

# Start the consent engine in the background
echo "🚀 Starting consent engine..."
go run *.go &
ENGINE_PID=$!

# Wait for the service to start
echo "⏳ Waiting for service to start..."
sleep 3

# Test 1: Create a consent with owner_id (should use SCIM lookup or fallback)
echo ""
echo "📝 Test 1: Creating consent with owner_id lookup"
echo "-----------------------------------------------"

curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "test-app",
    "data_fields": [{
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["personInfo.permanentAddress"]
    }],
    "purpose": "test_purpose",
    "session_id": "test_session_123"
  }' | jq '.'

echo ""

# Test 2: Create another consent with different owner_id
echo "📝 Test 2: Creating consent with different owner_id"
echo "--------------------------------------------------"

curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "test-app",
    "data_fields": [{
      "owner_type": "citizen",
      "owner_id": "dfdfd",
      "fields": ["personInfo.permanentAddress"]
    }],
    "purpose": "test_purpose",
    "session_id": "test_session_456"
  }' | jq '.'

echo ""

# Test 3: Test with unknown owner_id (should fail gracefully)
echo "📝 Test 3: Testing with unknown owner_id"
echo "---------------------------------------"

curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "test-app",
    "data_fields": [{
      "owner_type": "citizen",
      "owner_id": "unknown_nic_123",
      "fields": ["personInfo.permanentAddress"]
    }],
    "purpose": "test_purpose",
    "session_id": "test_session_789"
  }' | jq '.'

echo ""

# Cleanup
echo "🧹 Cleaning up..."
kill $ENGINE_PID 2>/dev/null

echo ""
echo "✅ SCIM integration test completed!"
echo ""
echo "📋 What to look for:"
echo "   - If M2M credentials are configured: Look for 'User found via SCIM' logs"
echo "   - If M2M credentials are NOT configured: Look for 'using hardcoded mapping' warnings"
echo "   - Successful consent creation with proper owner_email resolution"
echo "   - Graceful error handling for unknown owner_ids"
