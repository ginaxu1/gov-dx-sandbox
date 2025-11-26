# Portals

React-based portals for the OpenDIF platform.

## Portals

- **Admin Portal** - Administrative interface for managing the platform
- **Consent Portal** - User-facing consent management interface
- **Member Portal** - Member/provider portal for data exchange

## Quick Start

### Setup Configuration

All portals require a `config.js` file for runtime configuration. Use the setup script:

```bash
cd portals
./setup-portals.sh
```

This script will:
- Create `config.js` files for all portals
- Validate configuration
- Check dependencies
- Provide test commands

### Run a Portal

```bash
# Admin Portal
cd admin-portal
VITE_PORT=5174 npm run dev

# Consent Portal
cd consent-portal
VITE_PORT=5175 npm run dev

# Member Portal
cd member-portal
VITE_PORT=5176 npm run dev
```

## Configuration

Each portal uses a `config.js` file that is:
- **Generated at runtime** in Docker (via `entrypoint.sh`)
- **Created manually** for local development (in `public/` directory)
- **Loaded in HTML** before the main application script
- **Accessed via `window.configs`** in the application code

### Config File Locations

| Portal | Config Path | HTML Reference |
|--------|------------|----------------|
| Admin Portal | `admin-portal/public/config.js` | `./config.js` |
| Consent Portal | `consent-portal/public/config.js` | `./config.js` |
| Member Portal | `member-portal/public/config.js` | `./config.js` |

**Note**: Vite serves files from `public/` at the root (`/`), so `public/config.js` is accessible as `/config.js`.

## Testing

See [PORTAL_TESTING.md](./PORTAL_TESTING.md) for comprehensive testing instructions.
