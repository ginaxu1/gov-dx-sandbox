#!/bin/bash

# Consent Engine Environment Setup Script
# This script helps you set up the required environment variables for local development

echo "🔧 Setting up Consent Engine environment..."

# Create .env.local file if it doesn't exist
if [ ! -f .env.local ]; then
    echo "📝 Creating .env.local file..."
    cat > .env.local << 'EOF'
# Required - Asgardeo Configuration
# Get these values from your Asgardeo application settings
ASGARDEO_BASE_URL=https://api.asgardeo.io/t/lankasoftwarefoundation
ASGARDEO_JWKS_URL=https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks
ASGARDEO_ISSUER=https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/token
ASGARDEO_AUDIENCE=YOUR_AUDIENCE
ASGARDEO_ORG_NAME=lankasoftwarefoundation

# Required - Asgardeo M2M Configuration for SCIM API
# Get these from your consumer-auth-backend application in Asgardeo
ASGARDEO_M2M_CLIENT_ID=your-m2m-client-id
ASGARDEO_M2M_CLIENT_SECRET=your-m2m-client-secret

# Required - Service Configuration
CONSENT_PORTAL_URL=https://your-frontend-domain.com
ORCHESTRATION_ENGINE_URL=https://your-orchestration-engine.com
M2M_API_KEY=your-m2m-api-key-here
ENVIRONMENT=production

# Optional - Service Configuration (with defaults)
PORT=8081
LOG_LEVEL=info
LOG_FORMAT=text
CORS=true
RATE_LIMIT=100

# Test Configuration (for running tests)
TEST_CONSENT_PORTAL_URL=http://localhost:5173
TEST_JWKS_URL=https://api.asgardeo.io/t/YOUR_TENANT/oauth2/jwks
TEST_ASGARDEO_ISSUER=https://api.asgardeo.io/t/YOUR_TENANT/oauth2/token
TEST_ASGARDEO_AUDIENCE=YOUR_AUDIENCE
TEST_ASGARDEO_ORG_NAME=YOUR_ORG_NAME
EOF
    echo "✅ Created .env.local file with template values"
else
    echo "⚠️  .env.local already exists, skipping creation"
fi

echo ""
echo "📋 Next steps:"
echo "1. Edit .env.local and replace the placeholder values with your actual Asgardeo credentials"
echo "2. Get your Asgardeo credentials from: https://console.asgardeo.io/"
echo "3. Run the service with: go run main.go engine.go jwt_verifier.go"
echo ""
echo "🔍 Required Asgardeo values:"
echo "   - ASGARDEO_BASE_URL: Your Asgardeo organization URL"
echo "   - ASGARDEO_JWKS_URL: JWKS endpoint URL"
echo "   - ASGARDEO_ISSUER: JWT issuer URL"
echo "   - ASGARDEO_AUDIENCE: JWT audience"
echo "   - ASGARDEO_ORG_NAME: Your organization name"
echo "   - ASGARDEO_M2M_CLIENT_ID: M2M client ID for SCIM API access"
echo "   - ASGARDEO_M2M_CLIENT_SECRET: M2M client secret for SCIM API access"
echo ""
echo "🔍 Required Service values:"
echo "   - CONSENT_PORTAL_URL: Your frontend URL"
echo "   - ORCHESTRATION_ENGINE_URL: Your orchestration engine URL"
echo "   - M2M_API_KEY: Your M2M API key"
echo ""
echo "📚 For more information, see README.md"
