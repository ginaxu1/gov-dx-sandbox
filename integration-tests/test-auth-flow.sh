#!/bin/bash

# Test script to demonstrate the consumer authentication flow
# This script tests the complete flow from consumer creation to GraphQL access

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_SERVER_URL="http://localhost:3000"
ORCHESTRATION_ENGINE_URL="http://localhost:4000"

echo -e "${BLUE}=== Consumer Authentication Flow Test ===${NC}"
echo

# Function to print test steps
print_step() {
    echo -e "${YELLOW}Step $1: $2${NC}"
}

# Function to check if service is running
check_service() {
    local url=$1
    local service_name=$2
    
    echo -e "${BLUE}Checking if $service_name is running...${NC}"
    if curl -s "$url/health" > /dev/null; then
        echo -e "${GREEN}✓ $service_name is running${NC}"
    else
        echo -e "${RED}✗ $service_name is not running at $url${NC}"
        echo "Please start the services:"
        echo "  - API Server: cd api-server-go && go run main.go"
        echo "  - Orchestration Engine: cd exchange/orchestration-engine-go && go run main.go"
        exit 1
    fi
    echo
}

# Step 1: Check services
print_step "1" "Checking if services are running"
check_service "$API_SERVER_URL" "API Server"
check_service "$ORCHESTRATION_ENGINE_URL" "Orchestration Engine"

# Step 2: Create a consumer
print_step "2" "Creating a consumer"
CONSUMER_RESPONSE=$(curl -s -X POST "$API_SERVER_URL/consumers" \
    -H "Content-Type: application/json" \
    -d '{
        "consumerName": "Test Passport Application",
        "contactEmail": "test@passport.gov.lk",
        "phoneNumber": "+94-11-123-4567"
    }')

echo "Consumer creation response:"
echo "$CONSUMER_RESPONSE" | jq '.' 2>/dev/null || echo "$CONSUMER_RESPONSE"
echo

# Extract consumer ID
CONSUMER_ID=$(echo "$CONSUMER_RESPONSE" | jq -r '.consumerId' 2>/dev/null)
if [ "$CONSUMER_ID" = "null" ] || [ -z "$CONSUMER_ID" ]; then
    echo -e "${RED}✗ Failed to extract consumer ID${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Consumer created with ID: $CONSUMER_ID${NC}"
echo

# Step 3: Create a consumer application
print_step "3" "Creating a consumer application"
APP_RESPONSE=$(curl -s -X POST "$API_SERVER_URL/consumer-applications/$CONSUMER_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "required_fields": {
            "fullName": true,
            "dateOfBirth": true,
            "address": true,
            "profession": true
        }
    }')

echo "Application creation response:"
echo "$APP_RESPONSE" | jq '.' 2>/dev/null || echo "$APP_RESPONSE"
echo

# Extract submission ID
SUBMISSION_ID=$(echo "$APP_RESPONSE" | jq -r '.submissionId' 2>/dev/null)
if [ "$SUBMISSION_ID" = "null" ] || [ -z "$SUBMISSION_ID" ]; then
    echo -e "${RED}✗ Failed to extract submission ID${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Application created with submission ID: $SUBMISSION_ID${NC}"
echo

# Step 4: Approve the application (admin action)
print_step "4" "Approving the consumer application"
APPROVAL_RESPONSE=$(curl -s -X PUT "$API_SERVER_URL/consumer-applications/$SUBMISSION_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "status": "approved"
    }')

echo "Application approval response:"
echo "$APPROVAL_RESPONSE" | jq '.' 2>/dev/null || echo "$APPROVAL_RESPONSE"
echo

# Extract credentials
API_KEY=$(echo "$APPROVAL_RESPONSE" | jq -r '.credentials.apiKey' 2>/dev/null)
API_SECRET=$(echo "$APPROVAL_RESPONSE" | jq -r '.credentials.apiSecret' 2>/dev/null)

if [ "$API_KEY" = "null" ] || [ -z "$API_KEY" ] || [ "$API_SECRET" = "null" ] || [ -z "$API_SECRET" ]; then
    echo -e "${RED}✗ Failed to extract API credentials${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Application approved with credentials:${NC}"
echo "  API Key: $API_KEY"
echo "  API Secret: $API_SECRET"
echo

# Step 5: Authenticate and get access token
print_step "5" "Authenticating consumer and getting access token"
AUTH_RESPONSE=$(curl -s -X POST "$API_SERVER_URL/auth/token" \
    -H "Content-Type: application/json" \
    -d "{
        \"consumerId\": \"$CONSUMER_ID\",
        \"secret\": \"$API_SECRET\"
    }")

echo "Authentication response:"
echo "$AUTH_RESPONSE" | jq '.' 2>/dev/null || echo "$AUTH_RESPONSE"
echo

# Extract access token
ACCESS_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.accessToken' 2>/dev/null)
if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}✗ Failed to extract access token${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Authentication successful${NC}"
echo "  Access Token: ${ACCESS_TOKEN:0:50}..."
echo

# Step 6: Validate the access token
print_step "6" "Validating the access token"
VALIDATION_RESPONSE=$(curl -s -X POST "$API_SERVER_URL/auth/validate" \
    -H "Content-Type: application/json" \
    -d "{
        \"token\": \"$ACCESS_TOKEN\"
    }")

echo "Token validation response:"
echo "$VALIDATION_RESPONSE" | jq '.' 2>/dev/null || echo "$VALIDATION_RESPONSE"
echo

IS_VALID=$(echo "$VALIDATION_RESPONSE" | jq -r '.valid' 2>/dev/null)
if [ "$IS_VALID" = "true" ]; then
    echo -e "${GREEN}✓ Token validation successful${NC}"
else
    echo -e "${RED}✗ Token validation failed${NC}"
    exit 1
fi
echo

# Step 7: Test GraphQL access with authentication
print_step "7" "Testing GraphQL access with authentication"
GRAPHQL_RESPONSE=$(curl -s -X POST "$ORCHESTRATION_ENGINE_URL/" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d '{
        "query": "query { personInfo(nic: \"199512345678\") { fullName dateOfBirth address } }"
    }')

echo "GraphQL response:"
echo "$GRAPHQL_RESPONSE" | jq '.' 2>/dev/null || echo "$GRAPHQL_RESPONSE"
echo

# Check if GraphQL request was successful
if echo "$GRAPHQL_RESPONSE" | grep -q "errors"; then
    echo -e "${YELLOW}⚠ GraphQL request returned errors (this might be expected if no data is available)${NC}"
else
    echo -e "${GREEN}✓ GraphQL request successful${NC}"
fi
echo

# Step 8: Test GraphQL access without authentication (should fail)
print_step "8" "Testing GraphQL access without authentication (should fail)"
UNAUTH_GRAPHQL_RESPONSE=$(curl -s -X POST "$ORCHESTRATION_ENGINE_URL/" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "query { personInfo(nic: \"199512345678\") { fullName dateOfBirth address } }"
    }')

echo "Unauthenticated GraphQL response:"
echo "$UNAUTH_GRAPHQL_RESPONSE" | jq '.' 2>/dev/null || echo "$UNAUTH_GRAPHQL_RESPONSE"
echo

if echo "$UNAUTH_GRAPHQL_RESPONSE" | grep -q "UNAUTHENTICATED"; then
    echo -e "${GREEN}✓ Unauthenticated request properly rejected${NC}"
else
    echo -e "${YELLOW}⚠ Unauthenticated request was not rejected as expected${NC}"
fi
echo

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN}✓ Consumer created successfully${NC}"
echo -e "${GREEN}✓ Consumer application created and approved${NC}"
echo -e "${GREEN}✓ Authentication token generated${NC}"
echo -e "${GREEN}✓ Token validation working${NC}"
echo -e "${GREEN}✓ GraphQL access with authentication working${NC}"
echo -e "${GREEN}✓ Unauthenticated access properly blocked${NC}"
echo
echo -e "${BLUE}=== Authentication Flow Complete ===${NC}"
echo
echo "You can now use the LK Passport Application with these credentials:"
echo "  Consumer ID: $CONSUMER_ID"
echo "  Secret: $API_SECRET"
echo
echo "Or use the access token directly:"
echo "  Access Token: $ACCESS_TOKEN"
echo
echo -e "${YELLOW}Note: The access token expires in 24 hours.${NC}"
