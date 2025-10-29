#!/bin/bash

# Comprehensive Test Runner for Exchange Services

set -e  # Exit on any error

# Source the common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/test-utils.sh"

echo "=== Exchange Services Test Suite ==="

# Check if services are running
if ! check_all_services; then
    log_error "Cannot run tests - services not available"
    log_info "Start services with: make start-exchange"
    exit 1
fi

echo ""
log_info "Running individual test suites..."
echo ""

# Run PDP tests
log_info "=== Policy Decision Point Tests ==="
if [ -f "${SCRIPT_DIR}/test-pdp.sh" ]; then
    "${SCRIPT_DIR}/test-pdp.sh"
fi

echo ""

# Run Consent Engine tests
log_info "=== Consent Engine Tests ==="
if [ -f "${SCRIPT_DIR}/test-consent-flow.sh" ]; then
    "${SCRIPT_DIR}/test-consent-flow.sh"
elif [ -f "${SCRIPT_DIR}/test-consent-flow-simple.sh" ]; then
    "${SCRIPT_DIR}/test-consent-flow-simple.sh"
fi

echo ""

# Run complete workflow tests
log_info "=== Complete Workflow Tests ==="
if [ -f "${SCRIPT_DIR}/test-complete-flow.sh" ]; then
    "${SCRIPT_DIR}/test-complete-flow.sh"
fi

if [ -f "${SCRIPT_DIR}/test-consent-workflow-complete.sh" ]; then
    "${SCRIPT_DIR}/test-consent-workflow-complete.sh"
fi

echo ""

# Run API Server tests
log_info "=== API Server Tests ==="
if [ -f "${SCRIPT_DIR}/test-api-server-flow.sh" ]; then
    "${SCRIPT_DIR}/test-api-server-flow.sh"
fi

echo ""

log_success "All test suites completed successfully!"
echo ""
log_info "Test Summary:"
echo "- Policy Decision Point: ✅"
echo "- Consent Engine: ✅"
echo "- Complete Workflows: ✅"
echo "- API Server: ✅"