#!/bin/bash

# Master Test Script - Runs All Workflows Once
# This script executes all test workflows in the correct order

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Script configuration
SCRIPT_NAME="Master Test Suite"
SCRIPT_DESCRIPTION="Runs all test workflows once in the correct order"

# Set default values
API_BASE_URL=${API_BASE_URL:-$DEFAULT_API_BASE_URL}
ASGARDEO_BASE_URL=${ASGARDEO_BASE_URL:-$DEFAULT_ASGARDEO_BASE_URL}

# Show script info
print_header "$SCRIPT_NAME" $GREEN
print_info "$SCRIPT_DESCRIPTION"
echo ""

# Test execution order
TESTS=(
    "basic:Basic Functionality Tests"
    "security:Security Control Tests"
    "asgardeo:Asgardeo Integration Tests"
    "workflow:Complete Workflow Tests"
)

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Run a single test category
run_test_category() {
    local test_type=$1
    local test_name=$2
    
    print_header "$test_name"
    
    case $test_type in
        "basic")
            ./scripts/test.sh --basic --quiet
            ;;
        "security")
            ./scripts/test.sh --security --quiet
            ;;
        "asgardeo")
            ./scripts/test.sh --asgardeo --quiet
            ;;
        "workflow")
            ./scripts/test.sh --workflow --quiet
            ;;
        *)
            print_error "Unknown test type: $test_type"
            return 1
            ;;
    esac
    
    local exit_code=$?
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ $exit_code -eq 0 ]; then
        print_success "$test_name completed successfully"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_error "$test_name failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    echo ""
    return $exit_code
}

# Run all tests
run_all_tests() {
    print_info "Starting comprehensive test suite..."
    print_info "This will run all test categories in sequence"
    echo ""
    
    local start_time=$(date +%s)
    
    for test in "${TESTS[@]}"; do
        local test_type=$(echo "$test" | cut -d':' -f1)
        local test_name=$(echo "$test" | cut -d':' -f2)
        
        run_test_category "$test_type" "$test_name"
    done
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_header "Test Suite Summary"
    print_info "Total Test Categories: $TOTAL_TESTS"
    print_success "Passed: $PASSED_TESTS"
    print_error "Failed: $FAILED_TESTS"
    print_warning "Skipped: $SKIPPED_TESTS"
    print_info "Total Duration: ${duration}s"
    
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        print_success "All test categories completed successfully!"
        return 0
    else
        print_error "Some test categories failed. Please check the output above."
        return 1
    fi
}

# Show help
show_help() {
    echo "Master Test Suite"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Description: Runs all test workflows once in the correct order"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -u, --url URL           API server URL (default: $DEFAULT_API_BASE_URL)"
    echo "  -a, --asgardeo URL      Asgardeo base URL (default: $DEFAULT_ASGARDEO_BASE_URL)"
    echo "  -v, --verbose           Enable verbose output"
    echo "  -q, --quiet             Suppress output except errors"
    echo ""
    echo "Test Categories:"
    echo "  1. Basic Functionality Tests"
    echo "  2. Security Control Tests"
    echo "  3. Asgardeo Integration Tests"
    echo "  4. Complete Workflow Tests"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all tests"
    echo "  $0 --verbose            # Run all tests with verbose output"
    echo "  $0 --url http://localhost:8080  # Run tests against different server"
    echo ""
}

# Handle help
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# Run all tests
run_all_tests

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi
