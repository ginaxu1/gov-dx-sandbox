#!/bin/sh
set -e

# Generate config.js from environment variables at runtime
cat > /usr/share/nginx/html/config.js << EOF
window.configs = {
  VITE_API_URL: '${VITE_API_URL:-}',
  VITE_LOGS_URL: '${VITE_LOGS_URL:-}',
  VITE_IDP_CLIENT_ID: '${VITE_IDP_CLIENT_ID:-}',
  VITE_IDP_BASE_URL: '${VITE_IDP_BASE_URL:-}',
  VITE_IDP_SCOPE: '${VITE_IDP_SCOPE:-}',
  VITE_IDP_ADMIN_ROLE: '${VITE_IDP_ADMIN_ROLE:-}',
  VITE_SIGN_IN_REDIRECT_URL: '${VITE_SIGN_IN_REDIRECT_URL:-}',
  VITE_SIGN_OUT_REDIRECT_URL: '${VITE_SIGN_OUT_REDIRECT_URL:-}'
};
EOF

echo "Configuration file generated at runtime with current environment variables"

# Start nginx
exec nginx -g 'daemon off;'