# Exchange Services Scripts

## Core Scripts
- **`common.sh`** - Common configuration and functions used by all scripts
- **`manage.sh`** - Consolidated management script for all service operations
- **`test.sh`** - Run basic API tests with help functionality

## Consolidated Management (`manage.sh`)

The `manage.sh` script consolidates all service management operations:

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

## Usage Examples

```bash
# Start services
./scripts/manage.sh start-local
make start

# Check status
./scripts/manage.sh status
make status

# View logs
./scripts/manage.sh logs                    # All services
./scripts/manage.sh logs policy-decision-point  # Specific service
make logs

# Run tests
./scripts/test.sh
make test

# Get help
./scripts/manage.sh help
make help
```

## Configuration

All configuration is centralized in `common.sh`:
- Service ports and URLs
- Health check endpoints
- Test data payloads
- Common utility functions

To modify service configuration, update `common.sh` and all scripts will automatically use the new values.
