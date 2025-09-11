#!/bin/bash

# Comprehensive API Endpoint Testing Script
# This script tests all endpoints and documents their payloads and responses

BASE_URL="http://localhost:3000"
echo "=== COMPREHENSIVE API ENDPOINT DOCUMENTATION ==="
echo "Base URL: $BASE_URL"
echo ""

# Test data variables
CONSUMER_ID=""
PROVIDER_ID=""
SUBMISSION_ID=""
SCHEMA_ID=""

# Helper function to make requests and format output
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo "--- $description ---"
    echo "Method: $method"
    echo "Endpoint: $endpoint"
    if [ -n "$data" ]; then
        echo "Payload: $data"
    fi
    echo "Response:"
    
    if [ -n "$data" ]; then
        curl -s -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json" -d "$data" | jq . 2>/dev/null || curl -s -X $method "$BASE_URL$endpoint" -H "Content-Type: application/json" -d "$data"
    else
        curl -s -X $method "$BASE_URL$endpoint" | jq . 2>/dev/null || curl -s -X $method "$BASE_URL$endpoint"
    fi
    echo ""
}

# 1. Health Check
test_endpoint "GET" "/health" "" "Health Check"

# 2. Consumer Management
echo "=== CONSUMER MANAGEMENT ==="

# Create Consumer
test_endpoint "POST" "/consumers" '{"consumerName": "Test Consumer", "contactEmail": "consumer@example.com", "phoneNumber": "1234567890"}' "Create Consumer"
CONSUMER_ID="consumer_$(date +%s)"

# List Consumers
test_endpoint "GET" "/consumers" "" "List All Consumers"

# Get Specific Consumer (using a known ID from previous response)
test_endpoint "GET" "/consumers/consumer_82806caa2585d5d1" "" "Get Specific Consumer"

# Update Consumer
test_endpoint "PUT" "/consumers/consumer_82806caa2585d5d1" '{"consumerName": "Updated Consumer", "contactEmail": "updated@example.com", "phoneNumber": "9876543210"}' "Update Consumer"

# Delete Consumer
test_endpoint "DELETE" "/consumers/consumer_82806caa2585d5d1" "" "Delete Consumer"

# 3. Consumer Applications
echo "=== CONSUMER APPLICATIONS ==="

# Create Consumer Application
test_endpoint "POST" "/consumer-applications" '{"consumerId": "consumer_82806caa2585d5d1", "required_fields": {"name": true, "email": true}}' "Create Consumer Application"

# List Consumer Applications
test_endpoint "GET" "/consumer-applications" "" "List All Consumer Applications"

# Get Consumer Applications by Consumer ID
test_endpoint "GET" "/consumer-applications?consumerId=consumer_82806caa2585d5d1" "" "Get Consumer Applications by Consumer ID"

# 4. Provider Submissions
echo "=== PROVIDER SUBMISSIONS ==="

# Create Provider Submission
test_endpoint "POST" "/provider-submissions" '{"providerName": "Test Provider", "contactEmail": "provider@example.com", "phoneNumber": "0987654321", "providerType": "government"}' "Create Provider Submission"
SUBMISSION_ID="sub_prov_$(date +%s)"

# List Provider Submissions
test_endpoint "GET" "/provider-submissions" "" "List All Provider Submissions"

# Get Specific Provider Submission
test_endpoint "GET" "/provider-submissions/sub_prov_900eecc7c253eecee060b527" "" "Get Specific Provider Submission"

# Update Provider Submission (Approve)
test_endpoint "PUT" "/provider-submissions/sub_prov_900eecc7c253eecee060b527" '{"status": "approved"}' "Update Provider Submission (Approve)"
PROVIDER_ID="prov_$(date +%s)"

# 5. Provider Profiles
echo "=== PROVIDER PROFILES ==="

# List Provider Profiles
test_endpoint "GET" "/provider-profiles" "" "List All Provider Profiles"

# Get Specific Provider Profile
test_endpoint "GET" "/provider-profiles/prov_d87a095d798d3713" "" "Get Specific Provider Profile"

# 6. RESTful Provider Schema Submissions
echo "=== RESTFUL PROVIDER SCHEMA SUBMISSIONS ==="

# List Provider's Schema Submissions
test_endpoint "GET" "/providers/prov_d87a095d798d3713/schema-submissions" "" "List Provider's Schema Submissions"

# Create New Schema Submission
test_endpoint "POST" "/providers/prov_d87a095d798d3713/schema-submissions" '{"sdl": "type User { id: ID! name: String! email: String! }"}' "Create New Schema Submission"
SCHEMA_ID="schema_$(date +%s)"

# Get Specific Schema Submission
test_endpoint "GET" "/providers/prov_d87a095d798d3713/schema-submissions/schema_8836f02475093fba612a8442" "" "Get Specific Schema Submission"

# Update Schema Submission (Admin Approval)
test_endpoint "PUT" "/providers/prov_d87a095d798d3713/schema-submissions/schema_8836f02475093fba612a8442" '{"status": "approved"}' "Update Schema Submission (Admin Approval)"

# Modify Existing Schema
test_endpoint "POST" "/providers/prov_d87a095d798d3713/schema-submissions" '{"sdl": "type User { id: ID! name: String! email: String! age: Int }", "schema_id": "schema_8836f02475093fba612a8442"}' "Modify Existing Schema"

# 7. Admin Dashboard
echo "=== ADMIN DASHBOARD ==="

# Get Admin Dashboard
test_endpoint "GET" "/admin/dashboard" "" "Get Admin Dashboard"

echo "=== END OF API DOCUMENTATION ==="
