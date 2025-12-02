# OpenDIF

A comprehensive data exchange platform consisting of multiple microservices and portals for secure data sharing and consent management.

## Architecture

### Backend Services (Go)

- **Orchestration Engine** - Data exchange workflow orchestration
- **Policy Decision Point** - Policy enforcement
- **Consent Engine** - User consent management and validation
- **Audit Service** - Event logging and audit trail management
- **Portal Backend** - Backend service for the `Admin Portal` and the `Member Portal`

### Frontend Portals (React/TypeScript)

- **Member Portal** - Management of `Data sources` or `Applications` by `OpenDIF Members`
- **Admin Portal** - Administrative dashboard for the `OpenDIF Admins`
- **Consent Portal** - Citizen-facing interface for data consent

## Quick Start

### Initial Setup

```bash
make setup-all
```

This command will:

1. **Install Git Hooks** - Sets up pre-commit hooks that automatically run quality checks, build validation, and tests for services with staged changes
2. **Setup Go Services** - Installs dependencies (`go mod tidy` and `go mod download`) for:

   - orchestration-engine
   - policy-decision-point
   - consent-engine
   - audit-service
   - portal-backend

3. **Setup Frontend Services** - Installs npm dependencies (`npm ci`) for:
   - member-portal
   - admin-portal
   - consent-portal

### Build and Run

```bash
# Build all services
make validate-build-all

# Run a specific service
make run <service-name>
```

## Available Commands

```bash
make help                    # Show all available commands
make setup <service>         # Setup a specific service
make validate-build <service> # Build and validate a service
make validate-test <service>  # Run tests for a service
make quality-check <service>  # Run code quality checks
```

For detailed documentation, see `docs/` directory.
