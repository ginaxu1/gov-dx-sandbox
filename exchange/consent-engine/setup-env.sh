#!/bin/bash

# Consent Engine Environment Setup Script
# This script helps you set up the required environment variables for local development

echo "🔧 Setting up Consent Engine environment..."

# Create .env.local file if it doesn't exist
if [ ! -f .env.local ]; then
    echo "📝 Creating .env.local file..."
    cat > .env.local << 'EOF'
# Asgardeo Configuration
# Get these values from your Asgardeo application settings
ASGARDEO_BASE_URL=https://api.asgardeo.io/t/lankasoftwarefoundation
ASGARDEO_CLIENT_ID=your_client_id_here
ASGARDEO_CLIENT_SECRET=your_client_secret_here

# JWT Configuration (optional - defaults are provided)
ASGARDEO_JWKS_URL=https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks
ASGARDEO_ISSUER=https://api.asgardeo.io/t/lankasoftwarefoundation
ASGARDEO_AUDIENCE=lankasoftwarefoundation

# Service Configuration (optional - defaults are provided)
PORT=8081
CONSENT_PORTAL_URL=http://localhost:5173
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
echo "   - ASGARDEO_CLIENT_ID: Your application's client ID"
echo "   - ASGARDEO_CLIENT_SECRET: Your application's client secret"
echo ""
echo "📚 For more information, see README.md"
