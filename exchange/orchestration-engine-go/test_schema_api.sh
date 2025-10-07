#!/bin/bash

# Test script for schema management API
BASE_URL="http://localhost:4000"

echo "🧪 Testing Schema Management API"
echo "================================"

# Test 1: Get active schema (should be empty initially)
echo "Test 1: Get active schema"
curl -s -X GET "$BASE_URL/sdl" | jq '.' || echo "No active schema (expected)"

# Test 2: Create a new schema
echo -e "\nTest 2: Create new schema"
curl -s -X POST "$BASE_URL/sdl" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0.0",
    "sdl": "type Query { hello: String }",
    "created_by": "test-user"
  }' | jq '.'

# Test 3: Get all schemas
echo -e "\nTest 3: Get all schemas"
curl -s -X GET "$BASE_URL/sdl/versions" | jq '.'

# Test 4: Get active schema (should now have the schema)
echo -e "\nTest 4: Get active schema"
curl -s -X GET "$BASE_URL/sdl" | jq '.'

# Test 5: Validate SDL
echo -e "\nTest 5: Validate SDL"
curl -s -X POST "$BASE_URL/sdl/validate" \
  -H "Content-Type: application/json" \
  -d '{
    "sdl": "type Query { hello: String world: String }"
  }' | jq '.'

# Test 6: Check compatibility
echo -e "\nTest 6: Check compatibility"
curl -s -X POST "$BASE_URL/sdl/check-compatibility" \
  -H "Content-Type: application/json" \
  -d '{
    "sdl": "type Query { hello: String world: String }"
  }' | jq '.'

# Test 7: Create another schema version
echo -e "\nTest 7: Create another schema version"
curl -s -X POST "$BASE_URL/sdl" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.1.0",
    "sdl": "type Query { hello: String world: String }",
    "created_by": "test-user"
  }' | jq '.'

# Test 8: Get all schemas again
echo -e "\nTest 8: Get all schemas"
curl -s -X GET "$BASE_URL/sdl/versions" | jq '.'

echo -e "\n✅ Schema API tests completed!"
