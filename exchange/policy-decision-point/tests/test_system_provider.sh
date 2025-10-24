#!/bin/bash

# Test script to verify system-default provider is used
echo "=== Testing System Default Provider Implementation ==="
echo ""

# Test data with a field that doesn't have a specific provider
TEST_DATA='{
  "application_id": "system-test-app",
  "fields": [
    {
      "field_name": "system.generatedField",
      "grant_duration": "P30D"
    }
  ]
}'

echo "Request payload:"
echo "$TEST_DATA" | jq .
echo ""

# Test the endpoint
echo "Sending request to /metadata/update..."
echo ""

curl -X POST http://localhost:8082/metadata/update \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA" \
  -w "\nHTTP Status: %{http_code}\n" \
  -v

echo ""
echo "Test completed."
