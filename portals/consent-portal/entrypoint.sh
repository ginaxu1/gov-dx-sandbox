#!/bin/sh
set -e

# Ensure the public directory exists
mkdir -p /usr/share/nginx/html/public

# Generate config.js from environment variables at runtime
# Note: Consent portal expects config.js at ./public/config.js (different from other portals)
cat > /usr/share/nginx/html/public/config.js << EOF
window.configs = {
  apiUrl: '${VITE_API_URL:-}',
  VITE_CLIENT_ID: '${VITE_CLIENT_ID:-}',
  VITE_BASE_URL: '${VITE_BASE_URL:-}',
  VITE_SCOPE: '${VITE_SCOPE:-}',
  signInRedirectURL: '${VITE_SIGN_IN_REDIRECT_URL:-}',
  signOutRedirectURL: '${VITE_SIGN_OUT_REDIRECT_URL:-}'
};
EOF

echo "Configuration file generated at runtime with current environment variables"

# Start nginx
exec nginx -g 'daemon off;'