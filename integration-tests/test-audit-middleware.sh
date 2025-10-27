#!/bin/bash

# Simple test script to verify audit middleware is working

echo "Testing Audit Middleware Integration"
echo "==================================="

# Clear Redis stream
echo "Clearing Redis stream..."
redis-cli del audit-events 2>/dev/null || true

# Check initial state
echo "Initial Redis stream length: $(redis-cli xlen audit-events)"

# Make a test request
echo "Making test request to orchestration engine..."
curl -X POST http://localhost:4000/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-$(date +%s)" \
  -d '{
    "query": "query { person(nic: \"123456789V\") { fullName address } }"
  }' >/dev/null 2>&1

# Wait a moment for processing
sleep 2

# Check if message was created
echo "Redis stream length after request: $(redis-cli xlen audit-events)"

# Check stream contents
echo "Stream contents:"
redis-cli xrange audit-events - +

# Check if audit service is processing
echo "Pending messages:"
redis-cli xpending audit-events audit-processors 2>/dev/null || echo "No consumer group or stream"

echo "Test completed."
