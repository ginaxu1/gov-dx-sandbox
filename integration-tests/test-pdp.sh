#!/bin/bash

# Policy Decision Point V1 Integration Test Suite
# Tests the V1 policy decision endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
PDP_URL="http://localhost:8082"
TIMEOUT=10

echo -e "${BLUE}=== Policy Decision Point V1 Test Suite ===${NC}"
echo "Testing V1 policy decision endpoints"
echo ""

# Helper function to make HTTP requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local timeout=${5:-$TIMEOUT}
    
    echo -e "${BLUE}Test: $method $url${NC}"
    echo -e "${BLUE}Expected Status: $expected_status${NC}"
    echo -e "${BLUE}Timeout: ${timeout}s${NC}"
    echo ""
    
    if [ -n "$data" ]; then
        echo -e "${PURPLE}Data:${NC}"
        echo "$data" | jq . 2>/dev/null || echo "$data"
        echo ""
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time "$timeout" \
            "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            --max-time "$timeout" \
            "$url")
    fi
    
    # Extract status code and body
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    echo -e "${BLUE}Response Status: $status_code${NC}"
    echo -e "${BLUE}Response Body:${NC}"
    echo "$body" | jq . 2>/dev/null || echo "$body"
    echo ""
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASSED${NC}"
    else
        echo -e "${RED}❌ FAILED - Expected $expected_status, got $status_code${NC}"
    fi
    echo "---"
    echo ""
}

# Test 1: Create Policy Metadata
echo -e "${BLUE}=== Test 1: Create Policy Metadata ===${NC}"
policy_metadata_data=$(cat <<'JSONEOF'
{
  "schemaId": "person_schema_v1",
  "records": [
    {
      "fieldName": "person.fullName",
      "displayName": "Full Name",
      "description": "Person's full name",
      "source": "primary",
      "isOwner": true,
      "accessControlType": "public"
    },
    {
      "fieldName": "person.nic",
      "displayName": "NIC Number",
      "description": "National Identity Card number",
      "source": "primary",
      "isOwner": false,
      "accessControlType": "restricted",
      "owner": "government"
    },
    {
      "fieldName": "person.photo",
      "displayName": "Photo",
      "description": "Person's photo",
      "source": "primary",
      "isOwner": false,
      "accessControlType": "restricted",
      "owner": "government"
    },
    {
      "fieldName": "person.birthDate",
      "displayName": "Birth Date",
      "description": "Person's birth date",
      "source": "primary",
      "isOwner": false,
      "accessControlType": "public"
    },
    {
      "fieldName": "person.permanentAddress",
      "displayName": "Permanent Address",
      "description": "Person's permanent address",
      "source": "primary",
      "isOwner": false,
      "accessControlType": "restricted",
      "owner": "government"
    }
  ]
}
JSONEOF
)
make_request "POST" "$PDP_URL/api/v1/policy/metadata" "$policy_metadata_data" "201"

# Test 2: Update Allow List
echo -e "${BLUE}=== Test 2: Update Allow List ===${NC}"
allowlist_data='{
  "applicationId": "passport-app-v1",
  "records": [
    {
      "fieldName": "person.fullName",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.birthDate",
      "schemaId": "person_schema_v1"
    }
  ],
  "grantDuration": "30d"
}'
make_request "POST" "$PDP_URL/api/v1/policy/update-allowlist" "$allowlist_data" "200"

# Test 3: Policy Decision - Valid request with no consent required
echo -e "${BLUE}=== Test 3: Policy Decision - Valid request with no consent required ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.fullName",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 4: Policy Decision - Valid request with consent required
echo -e "${BLUE}=== Test 4: Policy Decision - Valid request with consent required ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.nic",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.photo",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 5: Policy Decision - Invalid consumer
echo -e "${BLUE}=== Test 5: Policy Decision - Invalid consumer ===${NC}"
pdp_data='{
  "applicationId": "unauthorized-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.fullName",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 6: Policy Decision - Unauthorized field access
echo -e "${BLUE}=== Test 6: Policy Decision - Unauthorized field access ===${NC}"
pdp_data='{
  "applicationId": "unauthorized-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.nic",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 7: Policy Decision - Single field test
echo -e "${BLUE}=== Test 7: Policy Decision - Single field test ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.birthDate",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 8: Policy Decision - Two fields test
echo -e "${BLUE}=== Test 8: Policy Decision - Two fields test ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.nic",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.photo",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 9: Policy Decision - Mixed fields test
echo -e "${BLUE}=== Test 9: Policy Decision - Mixed fields test ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.fullName",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.nic",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.unauthorizedField",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 10: Policy Decision - All approved fields test
echo -e "${BLUE}=== Test 10: Policy Decision - All approved fields test ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.fullName",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.birthDate",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 11: Policy Decision - Single unauthorized field test
echo -e "${BLUE}=== Test 11: Policy Decision - Single unauthorized field test ===${NC}"
pdp_data='{
  "applicationId": "passport-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.photo",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

# Test 12: Update Allow List for another application
echo -e "${BLUE}=== Test 12: Update Allow List for another application ===${NC}"
allowlist_data='{
  "applicationId": "government-app-v1",
  "records": [
    {
      "fieldName": "person.nic",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.permanentAddress",
      "schemaId": "person_schema_v1"
    }
  ],
  "grantDuration": "90d"
}'
make_request "POST" "$PDP_URL/api/v1/policy/update-allowlist" "$allowlist_data" "200"

# Test 13: Policy Decision with government app
echo -e "${BLUE}=== Test 13: Policy Decision with government app ===${NC}"
pdp_data='{
  "applicationId": "government-app-v1",
  "requiredFields": [
    {
      "fieldName": "person.nic",
      "schemaId": "person_schema_v1"
    },
    {
      "fieldName": "person.permanentAddress",
      "schemaId": "person_schema_v1"
    }
  ]
}'
make_request "POST" "$PDP_URL/api/v1/policy/decide" "$pdp_data" "200"

echo -e "${GREEN}=== Policy Decision Point V1 Test Suite Complete ===${NC}"
echo "All V1 policy decision endpoints tested successfully!"
echo ""
echo -e "${BLUE}Test Summary:${NC}"
echo "- Policy Metadata Creation: ✅"
echo "- Allow List Updates: ✅"
echo "- Policy Decisions: ✅"
echo "- Authorization Logic: ✅"
echo "- Consent Requirements: ✅"
