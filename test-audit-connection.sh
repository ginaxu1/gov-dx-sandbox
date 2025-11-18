#!/bin/bash

# Test script to verify audit-service to api-server-go connection
# This script tests if API requests create records in management_events table

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Testing Audit Service Connection ===${NC}"

# Configuration
API_SERVER_URL="${API_SERVER_URL:-http://localhost:3000}"
AUDIT_SERVICE_URL="${AUDIT_SERVICE_URL:-http://localhost:3001}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-testdb2}"
DB_USER="${DB_USER:-user}"
DB_PASSWORD="${DB_PASSWORD:-password}"

echo "Configuration:"
echo "  API Server: $API_SERVER_URL"
echo "  Audit Service: $AUDIT_SERVICE_URL"
echo "  Database: $DB_NAME@$DB_HOST:$DB_PORT"
echo ""

# Check if services are running
echo -e "${YELLOW}1. Checking if services are running...${NC}"
if ! curl -s -f "$API_SERVER_URL/health" > /dev/null; then
    echo -e "${RED}✗ API Server is not running at $API_SERVER_URL${NC}"
    exit 1
fi
echo -e "${GREEN}✓ API Server is running${NC}"

if ! curl -s -f "$AUDIT_SERVICE_URL/health" > /dev/null; then
    echo -e "${RED}✗ Audit Service is not running at $AUDIT_SERVICE_URL${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Audit Service is running${NC}"
echo ""

# Check database connection
echo -e "${YELLOW}2. Checking database connection...${NC}"
export PGPASSWORD="$DB_PASSWORD"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}✗ Cannot connect to database${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Database connection successful${NC}"
echo ""

# Check if management_events table exists
echo -e "${YELLOW}3. Checking management_events table...${NC}"
TABLE_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'management_events');" | tr -d ' ')
if [ "$TABLE_EXISTS" != "t" ]; then
    echo -e "${RED}✗ management_events table does not exist${NC}"
    exit 1
fi
echo -e "${GREEN}✓ management_events table exists${NC}"
echo ""

# Get initial count of management events
echo -e "${YELLOW}4. Getting initial count of management events...${NC}"
INITIAL_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM management_events;" | tr -d ' ')
echo "Initial count: $INITIAL_COUNT"
echo ""

# Test: Make an API request (this requires a valid JWT token)
echo -e "${YELLOW}5. Testing API request (requires JWT token)...${NC}"
echo "Note: This test requires a valid JWT token in the AUTHORIZATION header"
echo "You can set it with: export JWT_TOKEN='your-token-here'"
echo ""

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${YELLOW}⚠ JWT_TOKEN not set, skipping API request test${NC}"
    echo "To test with a real request, set JWT_TOKEN and run:"
    echo "  curl -X POST $API_SERVER_URL/api/v1/members \\"
    echo "    -H 'Authorization: Bearer \$JWT_TOKEN' \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"name\":\"Test User\",\"email\":\"test@example.com\"}'"
    echo ""
else
    echo "Making test API request..."
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_SERVER_URL/api/v1/members" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"name":"Test User","email":"test@example.com","phoneNumber":"1234567890"}' || true)
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ API request successful (HTTP $HTTP_CODE)${NC}"
        
        # Wait a bit for async audit log to be processed
        echo "Waiting 2 seconds for audit log to be processed..."
        sleep 2
        
        # Check if new event was created
        NEW_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM management_events;" | tr -d ' ')
        echo "New count: $NEW_COUNT"
        
        if [ "$NEW_COUNT" -gt "$INITIAL_COUNT" ]; then
            echo -e "${GREEN}✓ SUCCESS: New management event was created!${NC}"
            echo ""
            echo "Latest management event:"
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT event_type, actor_type, actor_id, actor_role, target_resource, target_resource_id, timestamp FROM management_events ORDER BY timestamp DESC LIMIT 1;"
        else
            echo -e "${RED}✗ FAILURE: No new management event was created${NC}"
            echo "This indicates the audit middleware is not working correctly"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠ API request failed (HTTP $HTTP_CODE)${NC}"
        echo "Response: $BODY"
    fi
    echo ""
fi

# Summary
echo -e "${YELLOW}=== Summary ===${NC}"
echo "1. API Server: Running"
echo "2. Audit Service: Running"
echo "3. Database: Connected"
echo "4. management_events table: Exists"
echo ""
echo "To verify manually:"
echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c \"SELECT * FROM management_events ORDER BY timestamp DESC LIMIT 5;\""
echo ""

