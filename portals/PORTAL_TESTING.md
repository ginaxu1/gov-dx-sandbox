# Portal Testing Guide

## Overview

This document provides comprehensive testing instructions for all three portals (Admin Portal, Consent Portal, Member Portal) to ensure `config.js` variables are properly loaded.

## Portal Configuration

Each portal uses a `config.js` file that is:
1. **Generated at runtime** in Docker (via `entrypoint.sh`)
2. **Created manually** for local development (in `public/` directory)
3. **Loaded in HTML** before the main application script
4. **Accessed via `window.configs`** in the application code

## Config.js File Locations

| Portal | Config Path | HTML Reference |
|--------|------------|----------------|
| Admin Portal | `admin-portal/public/config.js` | `./config.js` |
| Consent Portal | `consent-portal/public/config.js` | `./config.js` |
| Member Portal | `member-portal/public/config.js` | `./config.js` |

**Note**: Vite serves files from `public/` at the root (`/`), so `public/config.js` is accessible as `/config.js`.

## Required Configuration Variables

### Admin Portal
```javascript
window.configs = {
  VITE_API_URL: string,
  VITE_LOGS_URL: string,
  VITE_IDP_CLIENT_ID: string,
  VITE_IDP_BASE_URL: string,
  VITE_IDP_SCOPE: string,
  VITE_IDP_ADMIN_ROLE: string,
  VITE_SIGN_IN_REDIRECT_URL: string,
  VITE_SIGN_OUT_REDIRECT_URL: string
};
```

### Consent Portal
```javascript
window.configs = {
  apiUrl: string,
  VITE_CLIENT_ID: string,
  VITE_BASE_URL: string,
  VITE_SCOPE: string,
  signInRedirectURL: string,
  signOutRedirectURL: string
};
```

### Member Portal
```javascript
window.configs = {
  apiUrl: string,
  logsUrl: string,
  VITE_CLIENT_ID: string,
  VITE_BASE_URL: string,
  VITE_SCOPE: string,
  signInRedirectURL: string,
  signOutRedirectURL: string
};
```

## Why We Need `setup-portals.sh`

The `setup-portals.sh` script is essential because:

1. **Config files are not committed**: `config.js` files are in `.gitignore` (they contain environment-specific values)
2. **Docker generates at runtime**: In production/Docker, `entrypoint.sh` generates config.js from environment variables
3. **Local development needs manual setup**: For local development, config.js must be created manually
4. **Consistency**: Ensures all portals have the correct config structure and required variables
5. **Validation**: Verifies HTML references, dependencies, and variable presence before testing
6. **Time-saving**: Automates what would otherwise be manual, error-prone setup

**Without this script**, developers would need to:
- Manually create 3 config.js files
- Remember the exact variable names for each portal
- Ensure correct file paths and HTML references
- Validate everything works correctly

## Testing Steps

### 1. Create Config Files

Run the setup script to create all config.js files:
```bash
cd portals
./setup-portals.sh
```

This will:
- Create `config.js` files for all portals with test values
- Verify files exist and are readable
- Check HTML references are correct
- Validate required variables are present
- Check if dependencies are installed
- Provide portal-specific port numbers and test commands

### 2. Test Each Portal

#### Admin Portal
```bash
cd portals/admin-portal
VITE_PORT=5174 npm run dev
```

**Verification:**
1. Open http://localhost:5174
2. Open Browser DevTools (F12)
3. Go to Console tab
4. Look for logs:
   - `"Auth config:"` - Should show auth configuration
   - `"Window configs:"` - Should show all config variables
5. Verify no `undefined` values in config objects
6. Check Network tab - `config.js` should load with status 200

#### Consent Portal
```bash
cd portals/consent-portal
VITE_PORT=5175 npm run dev
```

**Verification:**
1. Open http://localhost:5175
2. Open Browser DevTools (F12)
3. Go to Console tab
4. Look for logs:
   - `"Auth config:"` - Should show auth configuration
   - `"Window configs:"` - Should show all config variables
5. Verify no `undefined` values in config objects
6. Check Network tab - `config.js` should load with status 200

#### Member Portal
```bash
cd portals/member-portal
VITE_PORT=5176 npm run dev
```

**Verification:**
1. Open http://localhost:5176
2. Open Browser DevTools (F12)
3. Go to Console tab
4. Look for logs:
   - `"Auth config:"` - Should show auth configuration
   - `"Window configs:"` - Should show all config variables
5. Verify no `undefined` values in config objects
6. Check Network tab - `config.js` should load with status 200

## Expected Console Output

Each portal should log something like:

```javascript
Auth config: {
  signInRedirectURL: "http://localhost:5173",
  signOutRedirectURL: "http://localhost:5173",
  clientID: "test-client-id-123",
  baseUrl: "https://api.asgardeo.io/t/test-org",
  scope: ["openid", "profile"],
  endpoints: { ... }
}

Window configs: {
  apiUrl: "http://localhost:3000",
  VITE_CLIENT_ID: "test-client-id-123",
  VITE_BASE_URL: "https://api.asgardeo.io/t/test-org",
  // ... other variables
}
```

## Troubleshooting

### Issue: Config.js not loading
**Symptoms:**
- Console shows `window.configs is undefined`
- Network tab shows 404 for `config.js`

**Solutions:**
1. Verify `config.js` exists in `public/` directory
2. Check HTML file references `./config.js` correctly
3. Restart the dev server
4. Clear browser cache

### Issue: Undefined values in config
**Symptoms:**
- Console shows `undefined` for some config values
- Auth fails to initialize

**Solutions:**
1. Check `config.js` file has all required variables
2. Verify variable names match exactly (case-sensitive)
3. Ensure no typos in variable names
4. Check for missing quotes around string values

### Issue: Config.js loads but variables are wrong
**Symptoms:**
- Config loads but values don't match expectations

**Solutions:**
1. Verify `config.js` content matches expected format
2. Check for environment variable overrides
3. Ensure no caching issues (hard refresh: Cmd+Shift+R / Ctrl+Shift+R)

## Test Checklist

- [ ] Admin Portal config.js created
- [ ] Consent Portal config.js created
- [ ] Member Portal config.js created
- [ ] Admin Portal loads without errors
- [ ] Consent Portal loads without errors
- [ ] Member Portal loads without errors
- [ ] All portals show 'Window configs:' in console
- [ ] All expected variables are present
- [ ] No undefined values in config objects
- [ ] config.js loads with HTTP 200 status
- [ ] Auth configuration initializes correctly

## Automated Setup

Use the setup script to create and verify config files:

```bash
# Create and verify config files for all portals
./setup-portals.sh
```

This single script replaces the need for multiple scripts and handles:
- Config file creation
- HTML reference validation
- Dependency checking
- Variable validation
- Port assignment

## Notes

- Config files are in `.gitignore` (not committed)
- In Docker, `entrypoint.sh` generates config.js at runtime
- For local dev, create config.js manually or use the test script
- Each portal expects different variable names (see above)
- Vite serves `public/` directory at root (`/`)

