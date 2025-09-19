#!/bin/bash

# Asgardeo Authentication Example (Refactored)
# This script demonstrates how to use Asgardeo authentication with the API Server
# This script uses shared functions to eliminate code duplication

set -e

# Source common functions and test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"
source "$SCRIPT_DIR/test-utils.sh"

# Configuration
CONSUMER_ID="consumer_acaedddca9bb5708"
API_KEY="6fc219ff801ff9a7eb610081abecfa78"
API_SECRET="2457b03fc75bc1bd415653491a3313585f1259e709d904f9ac8efaaa2cd53223"

# Show script info
print_header "Asgardeo Authentication Example"
print_info "Demonstrates how to use Asgardeo authentication with the API Server"
echo ""

# Main execution
main() {
    print_info "Starting Asgardeo authentication example..."
    echo ""
    
    # Check if API server is running
    print_step "1" "Checking API server health"
    if ! check_service_health "API Server" "$API_BASE_URL" true; then
        print_error "API server is not running. Please start it first."
        exit 1
    fi
    echo ""
    
    # Exchange Asgardeo token for internal JWT
    print_step "2" "Exchanging Asgardeo token for internal JWT"
    local exchange_result=$(exchange_credentials_for_token "$API_KEY" "$API_SECRET")
    
    local access_token=$(echo "$exchange_result" | grep "ACCESS_TOKEN:" | cut -d: -f2-)
    local consumer_id=$(echo "$exchange_result" | grep "CONSUMER_ID:" | cut -d: -f2-)
    
    if [ -z "$access_token" ] || [ "$access_token" = "null" ]; then
        print_error "Failed to get access token"
        echo "Response: $exchange_result"
        exit 1
    fi
    
    print_success "Successfully obtained access token"
    echo ""
    
    # Validate the token
    print_step "3" "Validating the access token"
    local validation_result=$(validate_token "$access_token")
    
    local is_valid=$(echo "$validation_result" | grep "VALID:" | cut -d: -f2-)
    local error=$(echo "$validation_result" | grep "ERROR:" | cut -d: -f2-)
    
    if [ "$is_valid" = "true" ]; then
        print_success "Token is valid"
    else
        print_warning "Token validation failed (expected if Asgardeo not configured)"
        print_warning "Error: $error"
        print_warning "This is expected if ASGARDEO_CLIENT_ID and ASGARDEO_CLIENT_SECRET are not set"
    fi
    echo ""
    
    # Example: Use the token to access a protected endpoint
    print_step "4" "Using token to access protected endpoints"
    print_info "Example: Accessing consumer information..."
    
    local consumer_response=$(make_get_request "$API_BASE_URL/consumers/$CONSUMER_ID" '-H "Authorization: Bearer '"$access_token"'"')
    
    echo "Consumer Response:"
    echo "$consumer_response" | jq '.' 2>/dev/null || echo "$consumer_response"
    echo ""
    
    print_success "Example completed successfully! 🎉"
    echo ""
    print_info "You can now use the access token for API requests:"
    echo "Authorization: Bearer $access_token"
    echo ""
}

# Run main function
main "$@"
