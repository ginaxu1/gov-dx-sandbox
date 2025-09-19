#!/bin/bash

# Environment Setup Script for API Server Go
# This script helps developers set up their local environment securely

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}🔧 API Server Go - Environment Setup${NC}"
echo "=================================="
echo ""

# Check if .env.local already exists
if [ -f "$PROJECT_ROOT/.env.local" ]; then
    echo -e "${YELLOW}⚠️  .env.local already exists${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}ℹ️  Keeping existing .env.local${NC}"
        exit 0
    fi
fi

# Create template if it doesn't exist
if [ ! -f "$PROJECT_ROOT/.env.template" ]; then
    echo -e "${BLUE}📋 Creating environment template...${NC}"
    cat > "$PROJECT_ROOT/.env.template" << 'EOF'
# Environment Configuration Template for API Server Go
# Copy this file to .env.local and update the values

# Asgardeo Configuration (optional for local development)
# If not provided, the system will use mock tokens for testing
ASGARDEO_BASE_URL=https://your-instance.asgardeo.io
ASGARDEO_CLIENT_ID=your_client_id_here
ASGARDEO_CLIENT_SECRET=your_client_secret_here

# JWT Configuration (optional)
# If not provided, a default key will be generated
JWT_SECRET_KEY=your_jwt_secret_key_here

# Server Configuration
PORT=3000
LOG_LEVEL=debug

# CORS Configuration (optional)
CORS_ORIGINS=http://localhost:3000,http://localhost:4000

# Rate Limiting (optional)
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
EOF
    echo -e "${GREEN}✅ Created .env.template${NC}"
fi

# Copy template to .env.local
echo -e "${BLUE}📋 Copying environment template...${NC}"
cp "$PROJECT_ROOT/.env.template" "$PROJECT_ROOT/.env.local"

echo -e "${GREEN}✅ Created .env.local from template${NC}"
echo ""

# Prompt for values
echo -e "${YELLOW}🔐 Please provide the following values:${NC}"
echo ""

# ASGARDEO_CLIENT_ID
read -p "ASGARDEO_CLIENT_ID: " asgardeo_client_id
if [ -n "$asgardeo_client_id" ]; then
    sed -i.bak "s/your_client_id_here/$asgardeo_client_id/g" "$PROJECT_ROOT/.env.local"
    echo -e "${GREEN}✅ ASGARDEO_CLIENT_ID set${NC}"
else
    echo -e "${YELLOW}⚠️  ASGARDEO_CLIENT_ID not provided, using placeholder${NC}"
fi

# ASGARDEO_CLIENT_SECRET
read -s -p "ASGARDEO_CLIENT_SECRET: " asgardeo_client_secret
echo
if [ -n "$asgardeo_client_secret" ]; then
    sed -i.bak "s/your_client_secret_here/$asgardeo_client_secret/g" "$PROJECT_ROOT/.env.local"
    echo -e "${GREEN}✅ ASGARDEO_CLIENT_SECRET set${NC}"
else
    echo -e "${YELLOW}⚠️  ASGARDEO_CLIENT_SECRET not provided, using placeholder${NC}"
fi


# Clean up backup files
rm -f "$PROJECT_ROOT/.env.local.bak"

echo ""
echo -e "${GREEN}🎉 Environment setup complete!${NC}"
echo ""
echo -e "${BLUE}📝 Next steps:${NC}"
echo "1. Review .env.local to ensure all values are correct"
echo "2. Run the application: go run main.go"
echo "3. Or use Docker: docker-compose up"
echo ""
echo -e "${YELLOW}⚠️  Security reminder:${NC}"
echo "- Never commit .env.local to git"
echo "- Keep your credentials secure"
echo "- Use different credentials for different environments"
echo ""

# Verify .env.local is in .gitignore
if grep -q "\.env\.local" "$PROJECT_ROOT/../.gitignore"; then
    echo -e "${GREEN}✅ .env.local is properly ignored by git${NC}"
else
    echo -e "${RED}❌ WARNING: .env.local is not in .gitignore${NC}"
    echo "Please add '**/.env.local' to your .gitignore file"
fi
