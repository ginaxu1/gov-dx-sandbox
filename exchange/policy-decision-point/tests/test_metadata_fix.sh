#!/bin/bash

# Test script to verify the metadata update fix
echo "=== Testing Policy Decision Point Metadata Update Fix ==="
echo ""

# Test the metadata update endpoint with the same payload that was failing
echo "Testing metadata update endpoint..."
echo ""

# Test data matching the failing request
TEST_DATA='{
  "application_id": "passport-app",
  "fields": [
    {
      "field_name": "person.fullName",
      "grant_duration": "P30D"
    }
  ]
}'

echo "Request payload:"
echo "$TEST_DATA" | jq .
echo ""

# Test the endpoint (assuming it's running on localhost:8082)
echo "Sending request to /metadata/update..."
echo ""

curl -X POST http://localhost:8082/metadata/update \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA" \
  -w "\nHTTP Status: %{http_code}\n" \
  -v

echo ""
echo "Test completed."
