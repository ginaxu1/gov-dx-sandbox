#!/bin/bash

# Test manual processing of audit messages

echo "🔍 Testing Manual Audit Processing"
echo "================================="

# Send a test request to create a new message
echo "1. Sending test request..."
curl -X POST http://localhost:4000/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer manual-test-$(date +%s)" \
  -d '{
    "query": "query { test { id } }"
  }' >/dev/null 2>&1

# Wait for message to be created
sleep 2

# Check stream length
echo "2. Stream length after request:"
redis-cli xlen audit-events

# Check pending messages
echo "3. Pending messages:"
redis-cli xpending audit-events audit-processors

# Get the latest message
echo "4. Latest message details:"
redis-cli xrange audit-events - + COUNT 1

# Wait for processing
echo "5. Waiting for processing (10 seconds)..."
sleep 10

# Check if message was processed
echo "6. Pending messages after waiting:"
redis-cli xpending audit-events audit-processors

# Check database count
echo "7. Database count:"
curl -s "http://localhost:3001/api/logs" | jq '.total'

echo "Test completed."
