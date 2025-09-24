#!/bin/bash

# Local Testing Environment Setup Script
# This script sets up the environment for local testing of hybrid JWT authentication

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up local testing environment for Consent Engine${NC}"
echo "=============================================================="
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    local item="$1"
    local status="$2"
    
    if [ "$status" = "OK" ]; then
        echo -e "${GREEN}✅ $item${NC}"
    else
        echo -e "${RED}❌ $item${NC}"
    fi
}

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"

# Check Go
if command_exists go; then
    print_status "Go $(go version | cut -d' ' -f3)" "OK"
else
    print_status "Go (not found)" "MISSING"
    echo -e "${RED}❌ Go is required. Please install Go 1.21 or later.${NC}"
    exit 1
fi

# Check Docker
if command_exists docker; then
    print_status "Docker $(docker --version | cut -d' ' -f3 | cut -d',' -f1)" "OK"
else
    print_status "Docker (not found)" "MISSING"
    echo -e "${RED}❌ Docker is required for PostgreSQL testing.${NC}"
    exit 1
fi

# Check Docker Compose
if command_exists docker-compose; then
    print_status "Docker Compose $(docker-compose --version | cut -d' ' -f3 | cut -d',' -f1)" "OK"
else
    print_status "Docker Compose (not found)" "MISSING"
    echo -e "${RED}❌ Docker Compose is required for PostgreSQL testing.${NC}"
    exit 1
fi

# Check jq
if command_exists jq; then
    print_status "jq $(jq --version)" "OK"
else
    print_status "jq (not found)" "MISSING"
    echo -e "${YELLOW}⚠️  jq is recommended for JSON parsing in tests${NC}"
fi

# Check Node.js (optional, for JWT token generation)
if command_exists node; then
    print_status "Node.js $(node --version)" "OK"
else
    print_status "Node.js (not found)" "MISSING"
    echo -e "${YELLOW}⚠️  Node.js is optional but recommended for JWT token generation${NC}"
fi

echo ""

# Create .env.local file for testing
echo -e "${YELLOW}📝 Creating .env.local for testing...${NC}"

cat > .env.local << 'EOF'
# Local Testing Environment Variables
# Copy this file and update with your actual values

# Database Configuration (for local PostgreSQL)
CHOREO_OPENDIF_DB_HOSTNAME=localhost
CHOREO_OPENDIF_DB_PORT=5433
CHOREO_OPENDIF_DB_USERNAME=test_user
CHOREO_OPENDIF_DB_PASSWORD=test_password
CHOREO_OPENDIF_DB_DATABASENAME=consent_engine_test
DB_SSLMODE=disable

# User JWT Configuration (Asgardeo)
# Replace with your actual Asgardeo values
ASGARDEO_JWKS_URL=https://api.asgardeo.io/t/yourorg/oauth2/jwks
ASGARDEO_ISSUER=https://api.asgardeo.io/t/yourorg/oauth2/token
ASGARDEO_AUDIENCE=your_audience
ASGARDEO_ORG_NAME=your_org_name

# M2M JWT Configuration (Choreo)
# Replace with your actual Choreo values
CHOREO_JWKS_URL=https://sts.choreo.dev/oauth2/jwks
CHOREO_ISSUER=https://sts.choreo.dev/oauth2/token
CHOREO_AUDIENCE=choreo_audience

# Service Configuration
PORT=8081
CONSENT_PORTAL_URL=http://localhost:5173
ORCHESTRATION_ENGINE_URL=http://localhost:4000
ENVIRONMENT=development
LOG_LEVEL=info
LOG_FORMAT=text
CORS=true
RATE_LIMIT=100

# Test Configuration
TEST_CONSENT_PORTAL_URL=http://localhost:5173
TEST_ASGARDEO_JWKS_URL=https://api.asgardeo.io/t/yourorg/oauth2/jwks
TEST_ASGARDEO_ISSUER=https://api.asgardeo.io/t/yourorg/oauth2/token
TEST_ASGARDEO_AUDIENCE=your_audience
TEST_ASGARDEO_ORG_NAME=your_org_name
TEST_CHOREO_JWKS_URL=https://sts.choreo.dev/oauth2/jwks
TEST_CHOREO_ISSUER=https://sts.choreo.dev/oauth2/token
TEST_CHOREO_AUDIENCE=choreo_audience
EOF

print_status ".env.local created" "OK"

echo ""

# Start PostgreSQL container
echo -e "${YELLOW}🐘 Starting PostgreSQL container...${NC}"
make setup-test-db
print_status "PostgreSQL container started" "OK"

echo ""

# Run tests to verify setup
echo -e "${YELLOW}🧪 Running basic tests to verify setup...${NC}"
if make test; then
    print_status "Basic tests passed" "OK"
else
    print_status "Basic tests failed" "FAIL"
    echo -e "${YELLOW}⚠️  Some tests may fail due to missing JWT configuration${NC}"
fi

echo ""

# Show next steps
echo -e "${GREEN}🎉 Local testing environment setup complete!${NC}"
echo ""
echo -e "${BLUE}📋 Next steps:${NC}"
echo "1. Edit .env.local with your actual JWT configuration values"
echo "2. Start the consent engine: make run"
echo "3. Run the local testing script: ./test_local_auth.sh"
echo "4. Or run specific tests: make test-local"
echo ""
echo -e "${BLUE}🔧 Useful commands:${NC}"
echo "make run              # Start consent engine with test database"
echo "make test             # Run tests with in-memory engine"
echo "make test-local       # Run tests with PostgreSQL"
echo "make clean            # Clean up test databases"
echo "./test_local_auth.sh  # Run comprehensive local testing"
echo ""
echo -e "${YELLOW}💡 For JWT token testing, you'll need to:${NC}"
echo "- Configure your Asgardeo JWT settings in .env.local"
echo "- Configure your Choreo M2M JWT settings in .env.local"
echo "- Or use the manual testing examples in the test script"
