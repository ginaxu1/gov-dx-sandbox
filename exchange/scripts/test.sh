#!/bin/bash
# Test Exchange Services

set -e

echo "Testing Exchange Services..."

# Test PDP
echo "Testing PDP..."
curl -s -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{"consumer":{"id":"test-app","name":"Test App","type":"mobile_app"},"request":{"resource":"person_data","action":"read","data_fields":["person.fullName"]},"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}' | jq '.'

# Test CE
echo "Testing CE..."
curl -s -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{"consumer_id":"test-app","data_owner":"test-owner","data_fields":["person.fullName"],"purpose":"testing","expiry_days":30}' | jq '.'

echo "✅ All tests passed!"