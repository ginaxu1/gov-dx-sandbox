#!/bin/bash

# Test runner script for orchestration-engine-go
# This script runs all unit tests and provides detailed output

set -e

echo "🚀 Running Orchestration Engine Go Tests"
echo "========================================"

# Change to the project directory
cd "$(dirname "$0")"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in a Go module
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Not in a Go module directory"
    exit 1
fi

echo "📦 Installing dependencies..."
go mod tidy

echo ""
echo "🧪 Running Unit Tests..."
echo "========================"

# Run tests with verbose output
go test -v ./federator/... -timeout 30s

echo ""
echo "🔍 Running Specific Test Categories..."
echo "======================================"

echo ""
echo "1️⃣  Query Parsing Tests..."
go test -v ./federator/ -run "TestQueryParsing" -timeout 10s

echo ""
echo "2️⃣  Schema Collection Tests..."
go test -v ./federator/ -run "TestSchemaCollection" -timeout 10s

echo ""
echo "3️⃣  Array Response Tests..."
go test -v ./federator/ -run "TestArrayResponseHandling" -timeout 15s

echo ""
echo "4️⃣  Integration Tests..."
go test -v ./federator/ -run "TestCompleteFederationFlow" -timeout 20s

echo ""
echo "5️⃣  Error Handling Tests..."
go test -v ./federator/ -run "TestFederationErrorHandling" -timeout 15s

echo ""
echo "📊 Test Coverage Report..."
echo "=========================="
go test -v ./federator/... -coverprofile=coverage.out -timeout 30s
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "✅ All tests completed successfully!"
echo "📈 Coverage report generated: coverage.html"
echo ""
echo "🎯 Test Summary:"
echo "   - Query Parsing: ✅"
echo "   - Schema Collection: ✅"
echo "   - Array Response Handling: ✅"
echo "   - Integration Flow: ✅"
echo "   - Error Handling: ✅"
echo ""
echo "🚀 Ready for implementation of bulk array support!"
