#!/bin/sh
set -e

# Generate config.js from environment variables at runtime
cat > /usr/share/nginx/html/config.js << EOF
window.configs = {
  apiUrl: '${VITE_API_URL:-}',
  logsUrl: '${VITE_LOGS_URL:-}',
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