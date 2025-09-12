#!/bin/bash

# Test script for the complete Consent Management Workflow
# This script demonstrates the full flow from passport app request to data retrieval

echo "=== Consent Management Workflow Test ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if services are running
print_status "Checking if services are running..."

# Check Consent Engine (port 8081)
if curl -s http://localhost:8081/health > /dev/null; then
    print_success "Consent Engine is running on port 8081"
else
    print_error "Consent Engine is not running on port 8081"
    print_warning "Please start the Consent Engine first: cd exchange/consent-engine && go run main.go"
    exit 1
fi

# Check Orchestration Engine (port 8000)
if curl -s http://localhost:8000/health > /dev/null; then
    print_success "Orchestration Engine is running on port 8000"
else
    print_error "Orchestration Engine is not running on port 8000"
    print_warning "Please start the Orchestration Engine first: cd exchange/orchestration-engine-go && go run main.go"
    exit 1
fi

# Check Policy Decision Point (port 8082)
if curl -s http://localhost:8082/health > /dev/null; then
    print_success "Policy Decision Point is running on port 8082"
else
    print_warning "Policy Decision Point is not running on port 8082"
    print_warning "This test will mock the PDP response"
fi

echo ""
print_status "Starting Consent Management Workflow Test..."
echo ""

# Step 1: Passport App makes data request to Orchestration Engine
print_status "Step 1: Passport App requests data from Orchestration Engine"

DATA_REQUEST='{
  "consumer_id": "passport-app",
  "data_owner_id": "199512345678",
  "required_fields": [
    "personInfo.permanentAddress",
    "personInfo.fullName",
    "personInfo.nic"
  ],
  "purpose": "passport_application",
  "request_id": "req_12345",
  "session_id": "session_123"
}'

echo "Request payload:"
echo "$DATA_REQUEST" | jq .
echo ""

print_status "Sending request to Orchestration Engine..."
RESPONSE=$(curl -s -X POST http://localhost:8000/data \
  -H "Content-Type: application/json" \
  -d "$DATA_REQUEST")

if [ $? -eq 0 ]; then
    print_success "Request sent successfully"
    echo "Response:"
    echo "$RESPONSE" | jq .
else
    print_error "Failed to send request to Orchestration Engine"
    exit 1
fi

echo ""

# Check if consent is required
STATUS=$(echo "$RESPONSE" | jq -r '.status // "unknown"')
REDIRECT_URL=$(echo "$RESPONSE" | jq -r '.redirect_url // ""')
CONSENT_ID=$(echo "$RESPONSE" | jq -r '.consent_id // ""')

if [ "$STATUS" = "pending" ]; then
    print_status "Step 2: Consent is required, redirecting to consent website"
    print_warning "Please open the consent website in your browser:"
    print_warning "http://localhost:8081/consent-website.html?consent_id=$CONSENT_ID"
    echo ""
    print_warning "In the consent website:"
    print_warning "1. Review the consent details"
    print_warning "2. Enter OTP: 123456"
    print_warning "3. Click 'Yes, Approve' or 'No, Deny'"
    echo ""
    print_status "Waiting for consent to be processed..."
    
    # Wait for user to process consent
    read -p "Press Enter after you have processed the consent in the browser..."
    
    # Step 3: Check consent status
    print_status "Step 3: Checking consent status..."
    
    CONSENT_STATUS_RESPONSE=$(curl -s "http://localhost:8081/consent-portal/?consent_id=$CONSENT_ID")
    CONSENT_STATUS=$(echo "$CONSENT_STATUS_RESPONSE" | jq -r '.status // "unknown"')
    
    echo "Consent status response:"
    echo "$CONSENT_STATUS_RESPONSE" | jq .
    echo ""
    
    if [ "$CONSENT_STATUS" = "approved" ]; then
        print_success "Consent has been approved!"
        
        # Step 4: Make another data request to get the actual data
        print_status "Step 4: Making another data request to retrieve the data..."
        
        FINAL_RESPONSE=$(curl -s -X POST http://localhost:8000/data \
          -H "Content-Type: application/json" \
          -d "$DATA_REQUEST")
        
        if [ $? -eq 0 ]; then
            print_success "Data retrieved successfully!"
            echo "Final response:"
            echo "$FINAL_RESPONSE" | jq .
        else
            print_error "Failed to retrieve data"
        fi
        
    elif [ "$CONSENT_STATUS" = "denied" ]; then
        print_error "Consent has been denied. Access will be blocked."
    else
        print_warning "Consent status is: $CONSENT_STATUS"
    fi
    
elif [ "$STATUS" = "success" ]; then
    print_success "No consent required, data retrieved directly!"
    echo "Data:"
    echo "$RESPONSE" | jq .
    
elif [ "$STATUS" = "denied" ]; then
    print_error "Access denied by policy"
    
else
    print_warning "Unknown response status: $STATUS"
fi

echo ""
print_status "=== Consent Management Workflow Test Complete ==="
