#!/bin/bash

# Test script for Audit Service endpoints
# Make sure the audit service is running on port 8082

BASE_URL="http://localhost:8082"

echo "Testing Audit Service Endpoints..."
echo "=================================="

# Test health check
echo "1. Testing health check..."
curl -s "$BASE_URL/health" | jq . || echo "Health check failed"
echo ""

# Test admin endpoint (no auth required)
echo "2. Testing admin endpoint (GET /audit/events)..."
curl -s "$BASE_URL/audit/events?limit=5" | jq . || echo "Admin endpoint failed"
echo ""

# Test provider endpoint (with mock auth)
echo "3. Testing provider endpoint (GET /audit/provider/events)..."
curl -s -H "Authorization: Bearer mock-token" "$BASE_URL/audit/provider/events?provider_id=test-provider&limit=5" | jq . || echo "Provider endpoint failed"
echo ""

# Test consumer endpoint (with mock auth)
echo "4. Testing consumer endpoint (GET /audit/consumer/events)..."
curl -s -H "Authorization: Bearer mock-token" "$BASE_URL/audit/consumer/events?consumer_id=test-consumer&limit=5" | jq . || echo "Consumer endpoint failed"
echo ""

# Test filtering
echo "5. Testing filtering (admin endpoint with filters)..."
curl -s "$BASE_URL/audit/events?transaction_status=SUCCESS&limit=3" | jq . || echo "Filtering test failed"
echo ""

echo "Audit Service endpoint tests completed!"
