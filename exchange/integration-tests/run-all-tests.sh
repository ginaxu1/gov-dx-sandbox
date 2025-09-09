#!/bin/bash

# Comprehensive Test Runner for Exchange Services
# Runs all test suites for the Policy Decision Point and Consent Engine

set -e  # Exit on any error

echo "=== Exchange Services Test Suite ==="
echo "Running comprehensive tests for PDP and Consent Engine"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up test environment..."
    # Add any cleanup logic here if needed
}

# Set trap for cleanup on exit
trap cleanup EXIT

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXCHANGE_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}Test Directory: $SCRIPT_DIR${NC}"
echo -e "${BLUE}Exchange Directory: $EXCHANGE_DIR${NC}"
echo ""

# Check if services are running
echo -e "${PURPLE}=== Service Health Check ===${NC}"

# Check PDP
echo "Checking Policy Decision Point (PDP) on port 8082..."
PDP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/health 2>/dev/null || echo "000")
if [ "$PDP_STATUS" = "200" ]; then
    echo -e "${GREEN}✅ PDP is running (HTTP $PDP_STATUS)${NC}"
else
    echo -e "${RED}❌ PDP is not responding (HTTP $PDP_STATUS)${NC}"
    echo "Please start the PDP service: cd $EXCHANGE_DIR && docker-compose up -d policy-decision-point"
    exit 1
fi

# Check Consent Engine
echo "Checking Consent Engine (CE) on port 8081..."
CE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health 2>/dev/null || echo "000")
if [ "$CE_STATUS" = "200" ]; then
    echo -e "${GREEN}✅ Consent Engine is running (HTTP $CE_STATUS)${NC}"
else
    echo -e "${RED}❌ Consent Engine is not responding (HTTP $CE_STATUS)${NC}"
    echo "Please start the Consent Engine: cd $EXCHANGE_DIR && docker-compose up -d consent-engine"
    exit 1
fi

echo ""

# Test 1: Policy Decision Point Tests
echo -e "${BLUE}=== Test 1: Policy Decision Point (PDP) Tests ===${NC}"
echo "Running PDP policy logic tests..."
echo ""

cd "$SCRIPT_DIR"
if ./test-pdp.sh; then
    echo -e "${GREEN}✅ PDP Tests PASSED${NC}"
else
    echo -e "${RED}❌ PDP Tests FAILED${NC}"
    exit 1
fi

echo ""

# Test 2: Consent Flow Tests
echo -e "${BLUE}=== Test 2: Consent Flow Tests ===${NC}"
echo "Running consent flow integration tests..."
echo ""

if ./test-consent-flow.sh; then
    echo -e "${GREEN}✅ Consent Flow Tests PASSED${NC}"
else
    echo -e "${RED}❌ Consent Flow Tests FAILED${NC}"
    exit 1
fi

echo ""

# Test 3: Complete Flow Tests
echo -e "${BLUE}=== Test 3: Complete Flow Tests ===${NC}"
echo "Running complete flow integration tests..."
echo ""

if ./test-complete-flow.sh; then
    echo -e "${GREEN}✅ Complete Flow Tests PASSED${NC}"
else
    echo -e "${RED}❌ Complete Flow Tests FAILED${NC}"
    exit 1
fi

echo ""

# Test 4: Complete Consent Flow Tests
echo -e "${BLUE}=== Test 4: Complete Consent Flow Tests ===${NC}"
echo "Running complete consent flow integration tests..."
echo ""

if ./test-complete-consent-flow.sh; then
    echo -e "${GREEN}✅ Complete Consent Flow Tests PASSED${NC}"
else
    echo -e "${RED}❌ Complete Consent Flow Tests FAILED${NC}"
    exit 1
fi

echo ""

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN}All test suites completed successfully!${NC}"
echo ""
echo "Test Coverage:"
echo "✅ Policy Decision Point (PDP) - Policy logic and authorization"
echo "✅ Consent Flow - Basic consent flow integration"
echo "✅ Complete Flow - End-to-end flow simulation"
echo "✅ Complete Consent Flow - Full consent flow with Consent Engine"
echo ""
echo "Services Tested:"
echo "✅ Policy Decision Point (Port 8082)"
echo "✅ Consent Engine (Port 8081)"
echo ""
echo -e "${GREEN}Exchange Services Test Suite Complete!${NC}"
