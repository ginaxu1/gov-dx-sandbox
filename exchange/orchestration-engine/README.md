# Orchestration Engine

A Go-based GraphQL service that orchestrates data requests from consumers to multiple data providers, handling JWT authentication, authorization, consent checks, argument mapping, and data aggregation with built-in security features.

## Overview

The Orchestration Engine (OE) is a Go-based service that orchestrates data requests from consumers to multiple data
providers. It handles JWT token validation, authorization checks, consent management, argument mapping, and data aggregation.

## Features

- **JWT Authentication**: Validates consumer tokens with JWKS auto-refresh and signature verification
- **SSRF Protection**: Environment-based protection against Server-Side Request Forgery attacks
- **GraphQL API**: Exposes a GraphQL API for consumers to request data
- **Multiple Data Providers**: Fetches data from multiple providers based on consumer requests
- **Authorization Checks**: Integrates with Policy Decision Point (PDP) for field-level authorization
- **Consent Management**: Verifies consumer consent via Consent Engine (CE) before data access
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals for clean service termination
- **Security Hardened**: Generic error messages to clients, detailed logging for operators

## Security Features

### JWT Token Validation

- **JWKS Auto-Refresh**: Automatic key rotation handling with hourly background refresh
- **Signature Verification**: RSA/ECDSA signature validation using JWKS
- **Claim Validation**: Validates `exp`, `nbf`, `iat`, `iss`, `aud`, `client_id`, and `sub`/`azp` claims
- **Development Bypass**: Optional bypass for local development (⚠️ never use in production)

### SSRF Protection

- **Production Mode**: Blocks private IPs (10.x, 192.168.x, 127.x) and cloud metadata endpoints (169.254.169.254)
- **Development Mode**: Allows localhost for local testing
- **Environment-Based**: Controlled via `environment` field in config.json

### Error Handling

- **Generic Client Errors**: Returns safe messages like "Unauthorized: invalid or expired token"
- **Detailed Logging**: Full error details logged internally for debugging
- **Information Disclosure Prevention**: No database errors, validation details, or internal paths exposed

## Quick Start

### Prerequisites

To set up the development environment for the Orchestration Engine, follow these steps:

1. **Install Go**: Ensure you have Go installed on your machine. You can download it from the
   official [Go website](https://golang.org/dl/).
2. **GraphQL Specification**: The Orchestration Engine uses GraphQL for its API. Familiarize yourself with the GraphQL
   specification by visiting the [GraphQL official site](https://graphql.org/).
3. **`schema.graphql` Schema File**: The GraphQL schema file is currently located in the `schemas` directory. These
   files define the structure of the API and the types of data that can be queried.
   We have placed the sample schema in it.
   - It should include `@sourceInfo` the directives in each of its leaf fields along with the following fields.
     - `providerKey` - A unique identifier for the data provider.
     - `providerField` - The field name in the provider's schema that corresponds to this field.
4. **`config.json` File**: Refer to the sample `config.example.json` file
   and create your own `config.json` file based on it. This file lists out the following information.

   - `environment` - Environment name: `development`, `staging`, or `production` (controls SSRF protection)
   - `trustUpstream` - Whether to trust upstream JWT validation (set to `false` for signature verification)
   - `pdpUrl` - The URL of the Policy Decision Point which handles authorization.
   - `ceUrl` - The URL of the Consent Engine which handles consent management.
   - `providers` - An array of data providers, each with a `providerKey` and `providerUrl`.
     For detailed provider integration steps, see the [Provider Onboarding Guide](PROVIDER_CONFIGURATION.md).
   - `jwt` - JWT configuration object:
     - `expectedIssuer` - Expected token issuer (e.g., `https://idp.example.com/oauth2/token`)
     - `validAudiences` - Array of valid audience values
     - `jwksUrl` - JWKS endpoint URL for public key retrieval (e.g., `https://idp.example.com/oauth2/jwks`)

5. **Run the Server**: You can run the Orchestration Engine server using the following command:
   ```bash
   go run main.go
   ```
   The server will start and listen for incoming requests on port 4000 (configurable via `PORT` environment variable).

## Configuration

### Environment Variables

- `PORT` - Server port (default: 4000)
- `CONFIG_PATH` - Path to config.json (default: ./config.json)
- Database configuration (see [Schema Management Guide](HOW_TO_TEST.md))

### JWT Configuration Example

```json
{
  "environment": "production",
  "trustUpstream": false,
  "jwt": {
    "expectedIssuer": "https://idp.example.com/oauth2/token",
    "validAudiences": ["your-client-id", "another-client-id"],
    "jwksUrl": "https://idp.example.com/oauth2/jwks"
  }
}
```

### SSRF Protection Behavior

| Environment   | Localhost Allowed? | Private IPs Blocked? | Cloud Metadata Blocked? |
| ------------- | ------------------ | -------------------- | ----------------------- |
| `development` | ✅ Yes             | ❌ No                | ❌ No                   |
| `staging`     | ❌ No              | ✅ Yes               | ✅ Yes                  |
| `production`  | ❌ No              | ✅ Yes               | ✅ Yes                  |

## Development Mode

For local development, set `environment: "development"` in config.json to:

- Bypass JWT token validation (⚠️ **WARNING**: Development bypass should be removed before production)
- Allow localhost/private IP JWKS URLs for testing
- Enable detailed error messages in logs

**Important**: Never deploy with `environment: "development"` to production environments.
The server will start and listen for incoming requests.
