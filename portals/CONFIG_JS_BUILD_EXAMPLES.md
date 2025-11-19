# Config.js Build Examples

## Overview
This document provides examples of how to build portal Docker images with `config.js` injection using build arguments.

## Admin Portal

### Build Command
```bash
docker build \
  --build-arg VITE_API_URL="http://api.example.com" \
  --build-arg VITE_LOGS_URL="http://logs.example.com" \
  --build-arg VITE_IDP_CLIENT_ID="your-client-id" \
  --build-arg VITE_IDP_BASE_URL="https://api.asgardeo.io/t/your-tenant" \
  --build-arg VITE_IDP_SCOPE="openid profile" \
  --build-arg VITE_IDP_ADMIN_ROLE="admin" \
  --build-arg VITE_SIGN_IN_REDIRECT_URL="https://admin.example.com" \
  --build-arg VITE_SIGN_OUT_REDIRECT_URL="https://admin.example.com" \
  -t admin-portal:latest \
  ./admin-portal
```

### Generated config.js
```javascript
window.configs = {
  VITE_API_URL: 'http://api.example.com',
  VITE_LOGS_URL: 'http://logs.example.com',
  VITE_IDP_CLIENT_ID: 'your-client-id',
  VITE_IDP_BASE_URL: 'https://api.asgardeo.io/t/your-tenant',
  VITE_IDP_SCOPE: 'openid profile',
  VITE_IDP_ADMIN_ROLE: 'admin',
  VITE_SIGN_IN_REDIRECT_URL: 'https://admin.example.com',
  VITE_SIGN_OUT_REDIRECT_URL: 'https://admin.example.com'
};
```

## Member Portal

### Build Command
```bash
docker build \
  --build-arg VITE_API_URL="http://api.example.com" \
  --build-arg VITE_LOGS_URL="http://logs.example.com" \
  --build-arg VITE_CLIENT_ID="your-client-id" \
  --build-arg VITE_BASE_URL="https://api.asgardeo.io/t/your-tenant" \
  --build-arg VITE_SCOPE="openid profile" \
  --build-arg VITE_SIGN_IN_REDIRECT_URL="https://member.example.com" \
  --build-arg VITE_SIGN_OUT_REDIRECT_URL="https://member.example.com" \
  -t member-portal:latest \
  ./member-portal
```

### Generated config.js
```javascript
window.configs = {
  apiUrl: 'http://api.example.com',
  logsUrl: 'http://logs.example.com',
  VITE_CLIENT_ID: 'your-client-id',
  VITE_BASE_URL: 'https://api.asgardeo.io/t/your-tenant',
  VITE_SCOPE: 'openid profile',
  signInRedirectURL: 'https://member.example.com',
  signOutRedirectURL: 'https://member.example.com'
};
```

## Consent Portal

### Build Command
```bash
docker build \
  --build-arg VITE_API_URL="http://api.example.com" \
  --build-arg VITE_CLIENT_ID="your-client-id" \
  --build-arg VITE_BASE_URL="https://api.asgardeo.io/t/your-tenant" \
  --build-arg VITE_SCOPE="openid profile" \
  --build-arg VITE_SIGN_IN_REDIRECT_URL="https://consent.example.com" \
  --build-arg VITE_SIGN_OUT_REDIRECT_URL="https://consent.example.com" \
  -t consent-portal:latest \
  ./consent-portal
```

### Generated config.js
**Note**: Consent portal expects config.js at `./public/config.js` (different path!)

```javascript
window.configs = {
  apiUrl: 'http://api.example.com',
  VITE_CLIENT_ID: 'your-client-id',
  VITE_BASE_URL: 'https://api.asgardeo.io/t/your-tenant',
  VITE_SCOPE: 'openid profile',
  signInRedirectURL: 'https://consent.example.com',
  signOutRedirectURL: 'https://consent.example.com'
};
```

## Default Values

If build arguments are not provided, empty strings will be used:

```javascript
window.configs = {
  VITE_API_URL: '',
  VITE_LOGS_URL: '',
  // ... etc
};
```

## Verification

### Check config.js in built image
```bash
# Admin/Member Portal
docker run --rm <image-name> cat /usr/share/nginx/html/config.js

# Consent Portal (different path)
docker run --rm <image-name> cat /usr/share/nginx/html/public/config.js
```

### Test in running container
```bash
# Start container
docker run -d -p 8080:80 <image-name>

# Check config.js is accessible
curl http://localhost:8080/config.js

# For Consent Portal
curl http://localhost:8080/public/config.js
```

## Environment-Specific Builds

### Development
```bash
docker build \
  --build-arg VITE_API_URL="http://localhost:8080" \
  --build-arg VITE_LOGS_URL="http://localhost:3001" \
  # ... other dev values
  -t admin-portal:dev \
  ./admin-portal
```

### Staging
```bash
docker build \
  --build-arg VITE_API_URL="https://api.staging.example.com" \
  --build-arg VITE_LOGS_URL="https://logs.staging.example.com" \
  # ... other staging values
  -t admin-portal:staging \
  ./admin-portal
```

### Production
```bash
docker build \
  --build-arg VITE_API_URL="https://api.example.com" \
  --build-arg VITE_LOGS_URL="https://logs.example.com" \
  # ... other production values
  -t admin-portal:prod \
  ./admin-portal
```

## Docker Compose Example

```yaml
version: '3.8'

services:
  admin-portal:
    build:
      context: ./admin-portal
      args:
        VITE_API_URL: ${VITE_API_URL}
        VITE_LOGS_URL: ${VITE_LOGS_URL}
        VITE_IDP_CLIENT_ID: ${VITE_IDP_CLIENT_ID}
        VITE_IDP_BASE_URL: ${VITE_IDP_BASE_URL}
        VITE_IDP_SCOPE: ${VITE_IDP_SCOPE}
        VITE_IDP_ADMIN_ROLE: ${VITE_IDP_ADMIN_ROLE}
        VITE_SIGN_IN_REDIRECT_URL: ${VITE_SIGN_IN_REDIRECT_URL}
        VITE_SIGN_OUT_REDIRECT_URL: ${VITE_SIGN_OUT_REDIRECT_URL}
    ports:
      - "8080:80"
```

## Notes

1. **No Code Changes Required**: The application code remains unchanged. It reads from `window.configs` which is populated by the injected `config.js`.

2. **Build-Time Injection**: `config.js` is generated during Docker build, not at runtime. This means:
   - Values are baked into the image
   - No runtime environment variables needed
   - Secure for production use

3. **Path Differences**: 
   - Admin & Member: `./config.js` (root)
   - Consent: `./public/config.js` (subdirectory)

4. **Empty Values**: If arguments are not provided, empty strings are used. Applications should handle this gracefully.
