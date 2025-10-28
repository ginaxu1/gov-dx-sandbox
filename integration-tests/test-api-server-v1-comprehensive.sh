#!/bin/bash

# V1 API Server Integration Tests
# Tests all V1 endpoints comprehensively

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_SERVER_URL="http://localhost:3000"
PDP_SERVICE_URL="http://localhost:8082"
TEST_TIMEOUT=30

# Test data
CONSUMER_ID=""
PROVIDER_ID=""
ENTITY_ID=""
APPLICATION_ID=""
SCHEMA_ID=""
APPLICATION_SUBMISSION_ID=""
SCHEMA_SUBMISSION_ID=""

# Utility functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if service is running
check_service() {
    local service_name=$1
    local service_url=$2
    
    log_info "Checking if $service_name is running at $service_url"
    
    if curl -s --max-time 5 "$service_url/health" > /dev/null 2>&1; then
        log_success "$service_name is running"
        return 0
    else
        log_error "$service_name is not running or not responding"
        return 1
    fi
}

# Make HTTP request and return response
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method"
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    curl_cmd="$curl_cmd '$url'"
    
    local response=$(eval $curl_cmd)
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$method $url -> $status_code"
        echo "$body"
        return 0
    else
        log_error "$method $url -> Expected $expected_status, got $status_code"
        echo "Response: $body"
        return 1
    fi
}

# Test Entity endpoints
test_entities() {
    log_info "=== Testing Entity Endpoints ==="
    
    # Create entity
    local entity_data='{
        "name": "Test Entity",
        "email": "test-entity@example.com",
        "phoneNumber": "1234567890",
        "idpUserID": "test-entity-user-123"
    }'
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/entities" "$entity_data" "201")
    ENTITY_ID=$(echo "$response" | jq -r '.entityID')
    log_success "Created entity with ID: $ENTITY_ID"
    
    # Get entity
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/entities/$ENTITY_ID" "" "200")
    log_success "Retrieved entity: $(echo "$get_response" | jq -r '.name')"
    
    # Update entity
    local update_data='{
        "name": "Updated Test Entity",
        "email": "updated-entity@example.com",
        "phoneNumber": "9876543210"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/entities/$ENTITY_ID" "$update_data" "200")
    log_success "Updated entity: $(echo "$update_response" | jq -r '.name')"
    
    # Get all entities
    local all_entities=$(make_request "GET" "$API_SERVER_URL/api/v1/entities" "" "200")
    local entity_count=$(echo "$all_entities" | jq 'length')
    log_success "Retrieved $entity_count entities"
}

# Test Consumer endpoints
test_consumers() {
    log_info "=== Testing Consumer Endpoints ==="
    
    # Create consumer
    local consumer_data='{
        "name": "Test Consumer",
        "email": "test-consumer@example.com",
        "phoneNumber": "1234567890",
        "idpUserID": "test-consumer-user-123"
    }'
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/consumers" "$consumer_data" "201")
    CONSUMER_ID=$(echo "$response" | jq -r '.consumerID')
    log_success "Created consumer with ID: $CONSUMER_ID"
    
    # Get consumer
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/consumers/$CONSUMER_ID" "" "200")
    log_success "Retrieved consumer: $(echo "$get_response" | jq -r '.name')"
    
    # Update consumer
    local update_data='{
        "name": "Updated Test Consumer",
        "email": "updated-consumer@example.com",
        "phoneNumber": "9876543210"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/consumers/$CONSUMER_ID" "$update_data" "200")
    log_success "Updated consumer: $(echo "$update_response" | jq -r '.name')"
    
    # Get all consumers
    local all_consumers=$(make_request "GET" "$API_SERVER_URL/api/v1/consumers" "" "200")
    local consumer_count=$(echo "$all_consumers" | jq 'length')
    log_success "Retrieved $consumer_count consumers"
}

# Test Provider endpoints
test_providers() {
    log_info "=== Testing Provider Endpoints ==="
    
    # Create provider
    local provider_data='{
        "name": "Test Provider",
        "email": "test-provider@example.com",
        "phoneNumber": "1234567890",
        "idpUserID": "test-provider-user-123"
    }'
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/providers" "$provider_data" "201")
    PROVIDER_ID=$(echo "$response" | jq -r '.providerID')
    log_success "Created provider with ID: $PROVIDER_ID"
    
    # Get provider
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/providers/$PROVIDER_ID" "" "200")
    log_success "Retrieved provider: $(echo "$get_response" | jq -r '.name')"
    
    # Update provider
    local update_data='{
        "name": "Updated Test Provider",
        "email": "updated-provider@example.com",
        "phoneNumber": "9876543210"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/providers/$PROVIDER_ID" "$update_data" "200")
    log_success "Updated provider: $(echo "$update_response" | jq -r '.name')"
    
    # Get all providers
    local all_providers=$(make_request "GET" "$API_SERVER_URL/api/v1/providers" "" "200")
    local provider_count=$(echo "$all_providers" | jq 'length')
    log_success "Retrieved $provider_count providers"
}

# Test Schema endpoints
test_schemas() {
    log_info "=== Testing Schema Endpoints ==="
    
    if [ -z "$PROVIDER_ID" ]; then
        log_error "Provider ID not set. Run provider tests first."
        return 1
    fi
    
    # Create schema
    local schema_data="{
        \"schemaName\": \"Test Schema\",
        \"schemaDescription\": \"Test Schema Description\",
        \"sdl\": \"type Person { fullName: String email: String phoneNumber: String }\",
        \"endpoint\": \"http://example.com/graphql\",
        \"providerID\": \"$PROVIDER_ID\"
    }"
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/schemas" "$schema_data" "201")
    SCHEMA_ID=$(echo "$response" | jq -r '.schemaID')
    log_success "Created schema with ID: $SCHEMA_ID"
    
    # Get schema
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/schemas/$SCHEMA_ID" "" "200")
    log_success "Retrieved schema: $(echo "$get_response" | jq -r '.schemaName')"
    
    # Update schema
    local update_data='{
        "schemaName": "Updated Test Schema",
        "schemaDescription": "Updated Test Schema Description",
        "sdl": "type Person { fullName: String email: String phoneNumber: String address: String }",
        "endpoint": "http://updated-example.com/graphql"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/schemas/$SCHEMA_ID" "$update_data" "200")
    log_success "Updated schema: $(echo "$update_response" | jq -r '.schemaName')"
    
    # Get all schemas
    local all_schemas=$(make_request "GET" "$API_SERVER_URL/api/v1/schemas" "" "200")
    local schema_count=$(echo "$all_schemas" | jq 'length')
    log_success "Retrieved $schema_count schemas"
}

# Test Application endpoints
test_applications() {
    log_info "=== Testing Application Endpoints ==="
    
    if [ -z "$CONSUMER_ID" ]; then
        log_error "Consumer ID not set. Run consumer tests first."
        return 1
    fi
    
    # Create application
    local application_data="{
        \"applicationName\": \"Test Application\",
        \"applicationDescription\": \"Test Application Description\",
        \"selectedFields\": [
            {
                \"fieldName\": \"person.fullName\",
                \"schemaID\": \"$SCHEMA_ID\"
            },
            {
                \"fieldName\": \"person.email\",
                \"schemaID\": \"$SCHEMA_ID\"
            }
        ],
        \"consumerID\": \"$CONSUMER_ID\"
    }"
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/applications" "$application_data" "201")
    APPLICATION_ID=$(echo "$response" | jq -r '.applicationID')
    log_success "Created application with ID: $APPLICATION_ID"
    
    # Get application
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/applications/$APPLICATION_ID" "" "200")
    log_success "Retrieved application: $(echo "$get_response" | jq -r '.applicationName')"
    
    # Update application
    local update_data='{
        "applicationName": "Updated Test Application",
        "applicationDescription": "Updated Test Application Description"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/applications/$APPLICATION_ID" "$update_data" "200")
    log_success "Updated application: $(echo "$update_response" | jq -r '.applicationName')"
    
    # Get all applications
    local all_applications=$(make_request "GET" "$API_SERVER_URL/api/v1/applications" "" "200")
    local application_count=$(echo "$all_applications" | jq 'length')
    log_success "Retrieved $application_count applications"
}

# Test Application Submission endpoints
test_application_submissions() {
    log_info "=== Testing Application Submission Endpoints ==="
    
    if [ -z "$CONSUMER_ID" ]; then
        log_error "Consumer ID not set. Run consumer tests first."
        return 1
    fi
    
    # Create application submission
    local submission_data="{
        \"applicationName\": \"Test Application Submission\",
        \"applicationDescription\": \"Test Application Submission Description\",
        \"selectedFields\": [
            {
                \"fieldName\": \"person.fullName\",
                \"schemaID\": \"$SCHEMA_ID\"
            }
        ],
        \"consumerID\": \"$CONSUMER_ID\"
    }"
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/application-submissions" "$submission_data" "201")
    APPLICATION_SUBMISSION_ID=$(echo "$response" | jq -r '.submissionID')
    log_success "Created application submission with ID: $APPLICATION_SUBMISSION_ID"
    
    # Get application submission
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/application-submissions/$APPLICATION_SUBMISSION_ID" "" "200")
    log_success "Retrieved application submission: $(echo "$get_response" | jq -r '.applicationName')"
    
    # Update application submission
    local update_data='{
        "status": "approved"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/application-submissions/$APPLISSION_SUBMISSION_ID" "$update_data" "200")
    log_success "Updated application submission status"
    
    # Get all application submissions
    local all_submissions=$(make_request "GET" "$API_SERVER_URL/api/v1/application-submissions" "" "200")
    local submission_count=$(echo "$all_submissions" | jq 'length')
    log_success "Retrieved $submission_count application submissions"
}

# Test Schema Submission endpoints
test_schema_submissions() {
    log_info "=== Testing Schema Submission Endpoints ==="
    
    if [ -z "$PROVIDER_ID" ]; then
        log_error "Provider ID not set. Run provider tests first."
        return 1
    fi
    
    # Create schema submission
    local submission_data="{
        \"schemaName\": \"Test Schema Submission\",
        \"schemaDescription\": \"Test Schema Submission Description\",
        \"sdl\": \"type Person { fullName: String email: String }\",
        \"endpoint\": \"http://submission-example.com/graphql\",
        \"providerID\": \"$PROVIDER_ID\"
    }"
    
    local response=$(make_request "POST" "$API_SERVER_URL/api/v1/schema-submissions" "$submission_data" "201")
    SCHEMA_SUBMISSION_ID=$(echo "$response" | jq -r '.submissionID')
    log_success "Created schema submission with ID: $SCHEMA_SUBMISSION_ID"
    
    # Get schema submission
    local get_response=$(make_request "GET" "$API_SERVER_URL/api/v1/schema-submissions/$SCHEMA_SUBMISSION_ID" "" "200")
    log_success "Retrieved schema submission: $(echo "$get_response" | jq -r '.schemaName')"
    
    # Update schema submission
    local update_data='{
        "status": "approved"
    }'
    
    local update_response=$(make_request "PUT" "$API_SERVER_URL/api/v1/schema-submissions/$SCHEMA_SUBMISSION_ID" "$update_data" "200")
    log_success "Updated schema submission status"
    
    # Get all schema submissions
    local all_submissions=$(make_request "GET" "$API_SERVER_URL/api/v1/schema-submissions" "" "200")
    local submission_count=$(echo "$all_submissions" | jq 'length')
    log_success "Retrieved $submission_count schema submissions"
}

# Test error handling
test_error_handling() {
    log_info "=== Testing Error Handling ==="
    
    # Test 404 for non-existent resource
    make_request "GET" "$API_SERVER_URL/api/v1/consumers/non-existent-id" "" "404"
    
    # Test 400 for invalid data
    local invalid_data='{"invalid": "data"}'
    make_request "POST" "$API_SERVER_URL/api/v1/consumers" "$invalid_data" "400"
    
    # Test 405 for unsupported method
    make_request "DELETE" "$API_SERVER_URL/api/v1/consumers" "" "405"
    
    log_success "Error handling tests completed"
}

# Test complete workflow
test_complete_workflow() {
    log_info "=== Testing Complete Workflow ==="
    
    # 1. Create provider
    local provider_data='{
        "name": "Workflow Provider",
        "email": "workflow-provider@example.com",
        "phoneNumber": "1111111111",
        "idpUserID": "workflow-provider-user"
    }'
    
    local provider_response=$(make_request "POST" "$API_SERVER_URL/api/v1/providers" "$provider_data" "201")
    local workflow_provider_id=$(echo "$provider_response" | jq -r '.providerID')
    
    # 2. Create schema for provider
    local schema_data="{
        \"schemaName\": \"Workflow Schema\",
        \"schemaDescription\": \"Schema for workflow testing\",
        \"sdl\": \"type Person { fullName: String email: String phoneNumber: String }\",
        \"endpoint\": \"http://workflow-provider.com/graphql\",
        \"providerID\": \"$workflow_provider_id\"
    }"
    
    local schema_response=$(make_request "POST" "$API_SERVER_URL/api/v1/schemas" "$schema_data" "201")
    local workflow_schema_id=$(echo "$schema_response" | jq -r '.schemaID')
    
    # 3. Create consumer
    local consumer_data='{
        "name": "Workflow Consumer",
        "email": "workflow-consumer@example.com",
        "phoneNumber": "2222222222",
        "idpUserID": "workflow-consumer-user"
    }'
    
    local consumer_response=$(make_request "POST" "$API_SERVER_URL/api/v1/consumers" "$consumer_data" "201")
    local workflow_consumer_id=$(echo "$consumer_response" | jq -r '.consumerID')
    
    # 4. Create application
    local application_data="{
        \"applicationName\": \"Workflow Application\",
        \"applicationDescription\": \"Application for workflow testing\",
        \"selectedFields\": [
            {
                \"fieldName\": \"person.fullName\",
                \"schemaID\": \"$workflow_schema_id\"
            },
            {
                \"fieldName\": \"person.email\",
                \"schemaID\": \"$workflow_schema_id\"
            }
        ],
        \"consumerID\": \"$workflow_consumer_id\"
    }"
    
    local application_response=$(make_request "POST" "$API_SERVER_URL/api/v1/applications" "$application_data" "201")
    local workflow_application_id=$(echo "$application_response" | jq -r '.applicationID')
    
    log_success "Complete workflow test completed successfully"
    log_info "Created: Provider($workflow_provider_id), Schema($workflow_schema_id), Consumer($workflow_consumer_id), Application($workflow_application_id)"
}

# Main test execution
main() {
    log_info "Starting V1 API Server Integration Tests"
    log_info "API Server URL: $API_SERVER_URL"
    log_info "PDP Service URL: $PDP_SERVICE_URL"
    
    # Check if services are running
    if ! check_service "API Server" "$API_SERVER_URL"; then
        log_error "API Server is not running. Please start it first."
        exit 1
    fi
    
    if ! check_service "PDP Service" "$PDP_SERVICE_URL"; then
        log_warning "PDP Service is not running. Some tests may fail."
    fi
    
    # Run all tests
    test_entities
    test_consumers
    test_providers
    test_schemas
    test_applications
    test_application_submissions
    test_schema_submissions
    test_error_handling
    test_complete_workflow
    
    log_success "All V1 API Server integration tests completed successfully!"
}

# Run main function
main "$@"
