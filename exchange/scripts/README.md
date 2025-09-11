# Scripts

Essential management scripts for the Exchange Services platform.

## Core Scripts

| Script | Purpose |
|--------|---------|
| `common.sh` | Common configuration and functions used by all scripts |
| `manage.sh` | Consolidated management script for all service operations |
| `test.sh` | Run basic API tests with help functionality |
| `restore-local-build.sh` | Restore to local development state |
| `prepare-docker-build.sh` | Prepare for production/Choreo deployment |

## Management Commands

```bash
# Environment Management
./scripts/manage.sh start-local     # Start local environment
./scripts/manage.sh start-prod      # Start production environment
./scripts/manage.sh stop            # Stop all services

# Monitoring & Debugging
./scripts/manage.sh status          # Check service status and health
./scripts/manage.sh logs [service]  # View logs (all or specific service)

# Help
./scripts/manage.sh help            # Show available commands
```

## Configuration

All configuration is centralized in `common.sh`:
- Service ports and URLs
- Health check endpoints
- Test data payloads
- Common utility functions

To modify service configuration, update `common.sh` and all scripts will automatically use the new values.

> **For basic script usage, see [Main README](../README.md#scripts)**
