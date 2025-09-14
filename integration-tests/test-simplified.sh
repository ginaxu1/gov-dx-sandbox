#!/bin/bash
# Simplified Integration Test Script
# This script demonstrates how to use the consolidated test utilities

# Source the common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/test-utils.sh"

# Run all tests using the consolidated functions
run_all_tests

echo ""
log_info "Test suite completed"
