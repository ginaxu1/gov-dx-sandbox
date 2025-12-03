# Orchestration Engine

A Go-based GraphQL service that orchestrates data requests from consumers to multiple data providers, handling authorization, consent checks, argument mapping, and data aggregation.

## Overview

The Orchestration Engine (OE) is the central component that:
- Receives GraphQL queries from data consumers
- Validates authorization with the Policy Decision Point (PDP)
- Verifies consent with the Consent Engine (CE)
- Fetches data from multiple providers
- Aggregates and returns unified responses

## Features

- **GraphQL API** - Unified query interface for data consumers
- **Multi-Provider Support** - Fetch data from multiple providers in a single query
- **Authorization Integration** - Policy Decision Point (PDP) for access control
- **Consent Management** - Consent Engine (CE) integration for consent verification
- **Schema Management** - Dynamic GraphQL schema versioning and activation
- **Field-Level Routing** - Intelligent field mapping to appropriate providers
- **Data Aggregation** - Combines responses from multiple providers

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+ (for schema management)
- Access to Policy Decision Point (PDP) service
- Access to Consent Engine (CE) service

### 1. Configuration

Create a `config.json` file based on `config.example.json`:

```json
{
  "pdpUrl": "http://localhost:8082",
  "ceUrl": "http://localhost:8081",
  "providers": [
    {
      "providerKey": "primary",
      "providerUrl": "http://localhost:8080/graphql"
    }
  ]
}
```

### 2. Environment Variables

```bash
# Server Configuration
PORT=4000  # Default: 4000

# Database Configuration (for schema management)
CHOREO_DB_OE_HOSTNAME=localhost
CHOREO_DB_OE_PORT=5432
CHOREO_DB_OE_USERNAME=postgres
CHOREO_DB_OE_PASSWORD=your_password
CHOREO_DB_OE_DATABASENAME=orchestration_engine

# Or use standard DB variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=orchestration_engine
DB_SSLMODE=disable

# Audit Service (optional)
CHOREO_AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
```

### 3. Run the Service

```bash
# Install dependencies
go mod download

# Run the service
go run main.go

# Or build and run
go build -o orchestration-engine
./orchestration-engine
```

The service runs on port 4000 by default.

## API Endpoints

### GraphQL Endpoints

- **POST /public/graphql** - Public GraphQL query endpoint
- **POST /graphql** - Authenticated GraphQL query endpoint

### Schema Management

- **GET /sdl** - Get active GraphQL schema
- **POST /sdl** - Create new schema version
- **GET /sdl/versions** - List all schema versions
- **POST /sdl/versions/{version}/activate** - Activate a schema version
- **POST /sdl/validate** - Validate SDL syntax
- **POST /sdl/check-compatibility** - Check backward compatibility

### Health & Debug

- **GET /health** - Health check endpoint

## GraphQL Schema

The service uses a GraphQL schema defined in `schema.graphql`. Each field can include `@sourceInfo` directives:

```graphql
type Person {
  fullName: String @sourceInfo(providerKey: "primary", providerField: "name")
  birthDate: String @sourceInfo(providerKey: "primary", providerField: "dob")
}
```

### Schema Directives

- `@sourceInfo` - Maps GraphQL fields to provider fields
  - `providerKey` - Unique identifier for the data provider
  - `providerField` - Field name in the provider's schema

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

See [HOW_TO_TEST.md](HOW_TO_TEST.md) for detailed testing instructions.

### Test Database Setup

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=your_password
export TEST_DB_NAME=orchestration_engine_test
export TEST_DB_SSLMODE=disable
```

## Architecture

### Components

- **Federator** - Core orchestration logic and query processing
- **Provider Handler** - Manages communication with data providers
- **Schema Service** - Handles GraphQL schema versioning
- **Policy Client** - Integration with Policy Decision Point
- **Consent Client** - Integration with Consent Engine
- **Audit Middleware** - Logs all data exchange events

### Request Flow

```
Consumer → Orchestration Engine → [PDP] → [CE] → Providers → Aggregation → Consumer
```

1. Consumer Request → GraphQL query received
2. Authorization Check → PDP validates access permissions
3. Consent Verification → CE checks required consents
4. Provider Routing → Fields mapped to appropriate providers
5. Data Fetching → Parallel requests to multiple providers
6. Aggregation → Responses combined into unified result
7. Response → GraphQL response returned to consumer

## Provider Configuration

For detailed provider integration steps, see [Provider Configuration Guide](PROVIDER_CONFIGURATION.md).

### Provider Requirements

Each provider must:
- Expose a GraphQL endpoint
- Support the fields specified in the schema
- Return data in the expected format

## Docker

```bash
# Build image
docker build -t orchestration-engine .

# Run container
docker run -p 4000:4000 \
  -v $(pwd)/config.json:/app/config.json \
  orchestration-engine
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check database credentials in environment variables
   - Verify PostgreSQL is running
   - Service continues without database (schema management disabled)

2. **Provider Connection Errors**
   - Verify provider URLs in `config.json`
   - Check network connectivity
   - Ensure providers are running

3. **Authorization Failures**
   - Verify PDP service is accessible
   - Check PDP URL in configuration
   - Review authorization policies

## Related Documentation

- [Provider Configuration](PROVIDER_CONFIGURATION.md) - Provider integration guide
- [Testing Guide](HOW_TO_TEST.md) - Comprehensive testing instructions
