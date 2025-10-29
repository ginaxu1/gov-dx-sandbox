#!/bin/bash

# Test all /api/v1 endpoints and clean up created records

API_URL="http://localhost:3000"
TEST_PREFIX="test-$(date +%s)"

echo "=== Testing API Server V1 Endpoints ==="

# Track created IDs for cleanup
CONSUMER_ID=""
PROVIDER_ID=""
ENTITY_ID=""
APPLICATION_ID=""
SCHEMA_ID=""

# 1. Health Check
echo -e "\n1. Testing /health"
curl -s "$API_URL/health" | jq .
echo ""

# 2. Create Consumer
echo "2. Creating Consumer"
CONSUMER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/consumers" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$TEST_PREFIX Consumer\",
    \"email\": \"${TEST_PREFIX}consumer@test.com\",
    \"phoneNumber\": \"1234567890\",
    \"idpUserID\": \"test-user-123\"
  }")
echo "$CONSUMER_RESPONSE" | jq .
CONSUMER_ID=$(echo "$CONSUMER_RESPONSE" | jq -r '.consumerId // empty')
ENTITY_ID=$(echo "$CONSUMER_RESPONSE" | jq -r '.entityId // empty')
echo "Created Consumer ID: $CONSUMER_ID"
echo ""

# 3. Get Consumer
if [ ! -z "$CONSUMER_ID" ]; then
  echo "3. Getting Consumer $CONSUMER_ID"
  curl -s "$API_URL/api/v1/consumers/$CONSUMER_ID" | jq .
  echo ""
fi

# 4. List All Consumers
echo "4. Listing All Consumers"
curl -s "$API_URL/api/v1/consumers" | jq '.count'
echo ""

# 5. Create Provider
echo "5. Creating Provider"
PROVIDER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/providers" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$TEST_PREFIX Provider\",
    \"email\": \"${TEST_PREFIX}provider@test.com\",
    \"phoneNumber\": \"9876543210\",
    \"idpUserID\": \"test-provider-123\"
  }")
echo "$PROVIDER_RESPONSE" | jq .
PROVIDER_ID=$(echo "$PROVIDER_RESPONSE" | jq -r '.providerId // empty')
echo "Created Provider ID: $PROVIDER_ID"
echo ""

# 6. Get Provider
if [ ! -z "$PROVIDER_ID" ]; then
  echo "6. Getting Provider $PROVIDER_ID"
  curl -s "$API_URL/api/v1/providers/$PROVIDER_ID" | jq .
  echo ""
fi

# 7. List All Providers
echo "7. Listing All Providers"
curl -s "$API_URL/api/v1/providers" | jq 'length'
echo ""

# 8. Create Schema (requires provider)
if [ ! -z "$PROVIDER_ID" ]; then
  echo "8. Creating Schema"
  SCHEMA_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/schemas" \
    -H "Content-Type: application/json" \
    -d "{
      \"schemaName\": \"$TEST_PREFIX Schema\",
      \"schemaDescription\": \"Test Schema\",
      \"sdl\": \"type Person { fullName: String }\",
      \"endpoint\": \"http://example.com/graphql\",
      \"providerId\": \"$PROVIDER_ID\"
    }")
  echo "$SCHEMA_RESPONSE" | jq .
  SCHEMA_ID=$(echo "$SCHEMA_RESPONSE" | jq -r '.schemaId // empty')
  echo "Created Schema ID: $SCHEMA_ID"
  echo ""
fi

# 9. List All Schemas
echo "9. Listing All Schemas"
curl -s "$API_URL/api/v1/schemas" | jq '.count'
echo ""

# 10. Create Application (requires consumer)
if [ ! -z "$CONSUMER_ID" ]; then
  echo "10. Creating Application"
  APP_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/applications" \
    -H "Content-Type: application/json" \
    -d "{
      \"applicationName\": \"$TEST_PREFIX Application\",
      \"applicationDescription\": \"Test Application\",
      \"selectedFields\": [
        {
          \"fieldName\": \"person.fullName\",
          \"schemaId\": \"${SCHEMA_ID:-test-schema-1}\"
        }
      ],
      \"consumerId\": \"$CONSUMER_ID\"
    }")
  echo "$APP_RESPONSE" | jq .
  APPLICATION_ID=$(echo "$APP_RESPONSE" | jq -r '.applicationId // empty')
  echo "Created Application ID: $APPLICATION_ID"
  echo ""
fi

# 11. List All Applications
echo "11. Listing All Applications"
curl -s "$API_URL/api/v1/applications" | jq '.count'
echo ""

# 12. Get Entity
if [ ! -z "$ENTITY_ID" ]; then
  echo "12. Getting Entity $ENTITY_ID"
  curl -s "$API_URL/api/v1/entities/$ENTITY_ID" | jq .
  echo ""
fi

# Cleanup
echo "=== Cleaning Up Test Records ==="

if [ ! -z "$APPLICATION_ID" ]; then
  echo "Deleting Application $APPLICATION_ID"
  curl -s -X DELETE "$API_URL/api/v1/applications/$APPLICATION_ID" || echo "Failed to delete application"
fi

if [ ! -z "$SCHEMA_ID" ]; then
  echo "Deleting Schema $SCHEMA_ID"
  curl -s -X DELETE "$API_URL/api/v1/schemas/$SCHEMA_ID" || echo "Failed to delete schema"
fi

if [ ! -z "$PROVIDER_ID" ]; then
  echo "Deleting Provider $PROVIDER_ID"
  curl -s -X DELETE "$API_URL/api/v1/providers/$PROVIDER_ID" || echo "Failed to delete provider"
fi

if [ ! -z "$CONSUMER_ID" ]; then
  echo "Deleting Consumer $CONSUMER_ID"
  curl -s -X DELETE "$API_URL/api/v1/consumers/$CONSUMER_ID" || echo "Failed to delete consumer"
fi

echo ""
echo "=== Test Complete ==="

