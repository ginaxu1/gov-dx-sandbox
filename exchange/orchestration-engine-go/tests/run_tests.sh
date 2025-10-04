#!/bin/bash

# Schema Implementation Test Runner
# This script runs all tests for the unified schema implementation

set -e

echo "🧪 Running Unified Schema Implementation Tests"
echo "=============================================="

# Change to the orchestration-engine-go directory
cd "$(dirname "$0")/.."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go to run tests."
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Not in the orchestration-engine-go directory. Please run from the correct location."
    exit 1
fi

echo "📁 Working directory: $(pwd)"
echo ""

# Run tests with verbose output
echo "🔍 Running Schema Service Tests..."
go test ./tests -v -run TestSchemaService

echo ""
echo "🔍 Running Schema Compatibility Tests..."
go test ./tests -v -run TestSchemaService_CheckCompatibility

echo ""
echo "🔍 Running Contract Tester Tests..."
go test ./tests -v -run TestContractTestSuite

echo ""
echo "🔍 Running Schema Handler Tests..."
go test ./tests -v -run TestSchemaHandler

echo ""
echo "🔍 Running GraphQL Handler Tests..."
go test ./tests -v -run TestGraphQLHandler

echo ""
echo "🔍 Running Integration Tests..."
go test ./tests -v -run TestSchemaIntegration

echo ""
echo "🎯 Running All Tests..."
go test ./tests -v

echo ""
echo "✅ All tests completed successfully!"
echo ""
echo "📊 Test Summary:"
echo "- Schema Service: Unit tests for core service logic"
echo "- Schema Compatibility: Version compatibility checking"
echo "- Contract Tester: Backward compatibility testing framework"
echo "- Schema Handler: API endpoint handlers"
echo "- GraphQL Handler: GraphQL query processing with version support"
echo "- Integration Tests: End-to-end workflow testing"
echo ""
echo "🚀 Ready for implementation!"
