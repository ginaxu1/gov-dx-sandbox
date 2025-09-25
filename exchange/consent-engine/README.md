# Consent Engine (CE)

Service that manages data owner consent workflows for data access requests with user JWT authentication support.

## Overview

- **Technology**: Go + In-memory storage
- **Port**: 8081
- **Purpose**: Consent management and workflow coordination
- **Authentication**: User JWT authentication with ownership checks for public endpoints, internal endpoints require no authentication
- **Test Coverage**: 34% with comprehensive unit and integration tests

## Quick Start

### 1. Environment Setup

First, set up your environment variables:

```bash
# Run the setup script to create .env.local
./setup-env.sh

# Edit .env.local with your Asgardeo credentials
nano .env.local
```

The service will automatically load environment variables from `.env.local` if it exists. You can also set them directly in your shell or use the example file:

```bash
# Copy the example file
cp env.example .env.local

# Edit with your values
nano .env.local
```

### 2. Run the Service

```bash
# Run locally
cd consent-engine && go run *.go

# Run tests
go test -v

# Run tests with coverage
go test -v -cover

# Docker
docker build -t ce . && docker run -p 8081:8081 ce
```

## API Endpoints

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/consents` | POST | Create new consent (Internal) | None |
| `/consents/{id}` | GET | Get consent information | **User JWT Auth** |
| `/consents/{id}` | PUT | Update consent status | **User JWT Auth** |
| `/consents/{id}` | POST | Update consent status (alternative) | **User JWT Auth** |
| `/consents/{id}` | DELETE | Revoke consent | **User JWT Auth** |
| `/data-info/{id}` | GET | Get data owner information | None |
| `/consent-portal` | POST | Create consent via portal | None |
| `/consent-portal` | PUT | Update consent via portal | None |
| `/consent-portal` | GET | Get consent portal info | None |
| `/consent-website` | GET | Serve consent portal website | None |
| `/data-owner/{id}` | GET | Get consents by data owner | None |
| `/consumer/{id}` | GET | Get consents by consumer | None |
| `/admin/expiry-check` | POST | Check expired consents | None |
| `/health` | GET | Health check | None |

**User JWT Auth**: All requests require user JWT with email ownership validation

## User JWT Authentication

The Consent Engine uses user JWT authentication for public endpoints with email ownership validation:

### User JWT Tokens (From Asgardeo)
- **Require JWT authentication** with valid email claim
- **Email must match** the consent owner's email for ownership validation
- **Used by** Consent Portal for user-facing operations
- **Protected endpoints**: GET, PUT, PATCH, DELETE `/consents/{id}`

### Internal Endpoints
- **POST `/consents`**: Internal endpoint, no authentication required
- **Used by** Orchestration Engine for service-to-service communication
- **Access**: Project-internal only, not exposed publicly

### Environment Variables

#### Required - User JWT Configuration (Asgardeo)
- `ASGARDEO_JWKS_URL` - JWKS endpoint URL for user token validation
- `ASGARDEO_ISSUER` - JWT issuer URL for user tokens
- `ASGARDEO_AUDIENCE` - JWT audience for user tokens
- `ASGARDEO_ORG_NAME` - Your organization name

#### Required - Service Configuration
- `CONSENT_PORTAL_URL` - Consent portal URL (default: http://localhost:5173)
- `ORCHESTRATION_ENGINE_URL` - Orchestration engine URL (default: http://localhost:4000)
- `ENVIRONMENT` - Environment (production/development)

#### Optional - Service Configuration
- `PORT` - Service port (default: 8081)
- `LOG_LEVEL` - Log level (default: info)
- `LOG_FORMAT` - Log format (default: text)
- `CORS` - Enable CORS (default: true)
- `RATE_LIMIT` - Rate limit per minute (default: 100)

#### Test Configuration
- `TEST_CONSENT_PORTAL_URL` - Test consent portal URL (default: http://localhost:5173)
- `TEST_ASGARDEO_JWKS_URL` - Test Asgardeo JWKS URL for user token validation
- `TEST_ASGARDEO_ISSUER` - Test Asgardeo issuer for user tokens
- `TEST_ASGARDEO_AUDIENCE` - Test Asgardeo audience for user tokens
- `TEST_ASGARDEO_ORG_NAME` - Test organization name

### JWT Token Format

```bash
Authorization: Bearer <jwt_token>
```

#### User JWT Tokens (Asgardeo)
The user JWT token must contain an email claim in one of these fields:
- `email`
- `sub` (subject)
- `preferred_username`

### Authorization Rules

For protected endpoints (`/consents/{id}`):
- **User JWT**: Email must match the consent owner's email for ownership validation
- **Invalid/Unknown tokens**: Rejected with 401 Unauthorized
- **Ownership mismatch**: Rejected with 403 Forbidden

For internal endpoints (`POST /consents`):
- **No authentication required** - project-internal access only

## Data Owner Information

**Endpoint:** `GET /data-info/{consentId}`

Retrieves only the owner ID and email for a specific consent record. This endpoint does not require authentication.

**Response:**
```json
{
  "owner_id": "test-owner-123",
  "owner_email": "owner@example.com"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8081/data-info/consent_122af00e
```

## Create Consent

**Endpoint:** `POST /consents`

**Request:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["person.permanentAddress", "person.photo"]
    }
  ],
  "purpose": "passport_application",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback"
}
```

**Response:**
```json
{
  "consent_id": "consent_122af00e",
  "owner_id": "199512345678",
  "app_id": "passport-app",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-15T14:51:14.395412+05:30",
  "updated_at": "2025-09-15T14:51:14.395412+05:30",
  "expires_at": "2025-10-15T14:51:14.395389+05:30",
  "grant_duration": "30d",
  "fields": ["person.permanentAddress", "person.photo"],
  "session_id": "session_123",
  "redirect_url": "http://localhost:5173/?consent_id=consent_122af00e",
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "passport-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "199512345678",
        "fields": ["person.permanentAddress"]
      }
    ],
    "purpose": "passport_application",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback"
  }'
```

## Get Consent Information

**Endpoint:** `GET /consents/{id}`  
**Authentication:** User JWT Required (email must match consent owner)

**Response (User-facing format):**
```json
{
  "consent_id": "consent_122af00e",
  "app_display_name": "Passport Application",
  "created_at": "2025-09-15T14:51:14.395412+05:30",
  "fields": ["person.permanentAddress", "person.photo"],
  "owner_name": "199512345678",
  "status": "pending",
  "type": "realtime"
}
```

**User Request (Consent Portal - JWT Required):**
```bash
curl -X GET http://localhost:8081/consents/consent_122af00e \
  -H "Authorization: Bearer <asgardeo_user_jwt_token>"
```

**Error Responses:**
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - User JWT email doesn't match consent owner
- `404 Not Found` - Consent record not found

## Update Consent Status

**Endpoint:** `PUT /consents/{id}`  
**Authentication:** User JWT Required (email must match consent owner)

**Request:**
```json
{
  "status": "approved",
  "owner_id": "199512345678",
  "message": "User approved consent"
}
```

**Response:**
```json
{
  "consent_id": "consent_122af00e",
  "status": "approved",
  "updated_at": "2025-09-15T14:55:00.000000+05:30",
  "message": "Consent status updated successfully"
}
```

**cURL Example:**
```bash
curl -X PUT http://localhost:8081/consents/consent_122af00e \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt_token>" \
  -d '{
    "status": "approved",
    "owner_id": "199512345678",
    "message": "User approved consent"
  }'
```

**Error Responses:**
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - User JWT email doesn't match consent owner
- `404 Not Found` - Consent record not found

## Revoke Consent

**Endpoint:** `DELETE /consents/{id}`  
**Authentication:** User JWT Required (email must match consent owner)

**Request:**
```json
{
  "reason": "User requested data deletion"
}
```

**Response:**
```json
{
  "consent_id": "consent_122af00e",
  "status": "revoked",
  "updated_at": "2025-09-15T14:55:00.000000+05:30",
  "message": "Consent revoked successfully"
}
```

**cURL Example:**
```bash
curl -X DELETE http://localhost:8081/consents/consent_122af00e \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt_token>" \
  -d '{"reason": "User requested data deletion"}'
```

**Error Responses:**
- `401 Unauthorized` - Invalid or missing JWT token
- `403 Forbidden` - User JWT email doesn't match consent owner
- `404 Not Found` - Consent record not found

## How to Run and Test Locally

This guide will help you set up and test the consent engine locally with PostgreSQL database.

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL client (optional, for direct database access)

### 1. Database Setup

The consent engine uses PostgreSQL for persistent storage. Start the test database:

```bash
# Start the test PostgreSQL container
make setup-test-db

# Verify the database is running
docker ps
```

This will start a PostgreSQL container on port `5433` with:
- Database: `consent_engine_test`
- Username: `test_user`
- Password: `test_password`

### 2. Running the Service Locally

Manual Environment Variables
```bash
# Set environment variables and run
CHOREO_OPENDIF_DB_HOSTNAME=localhost \
CHOREO_OPENDIF_DB_PORT=5433 \
CHOREO_OPENDIF_DB_USERNAME=test_user \
CHOREO_OPENDIF_DB_PASSWORD=test_password \
CHOREO_OPENDIF_DB_DATABASENAME=consent_engine_test \
DB_SSLMODE=disable \
go run . &
```

The service will start on `http://localhost:8081` and automatically:
- Connect to the PostgreSQL database
- Create necessary tables and indexes
- Initialize the consent engine

### 3. Testing the Service

#### Health Check
```bash
curl http://localhost:8081/health
```

Expected response:
```json
{
  "service": "consent-engine",
  "status": "healthy"
}
```

#### Complete Test Workflow

1. **Create a consent request:**
```bash
CONSENT_RESPONSE=$(curl -s -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "passport-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "199512345678",
        "fields": ["person.permanentAddress", "person.photo"]
      }
    ],
    "purpose": "passport_application",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback"
  }')

echo "Consent created: $CONSENT_RESPONSE"
```

2. **Extract consent ID and get consent information:**
```bash
CONSENT_ID=$(echo $CONSENT_RESPONSE | jq -r '.consent_id')
echo "Consent ID: $CONSENT_ID"

# Get consent information (no auth required for basic info)
curl -X GET http://localhost:8081/consents/$CONSENT_ID
```

3. **Update consent status:**
```bash
curl -X PUT http://localhost:8081/consents/$CONSENT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "owner_id": "199512345678",
    "message": "User approved consent"
  }'
```

4. **Get data owner information:**
```bash
curl -X GET http://localhost:8081/data-info/$CONSENT_ID
```

### 4. Database Access

#### Connect to the Local Database
```bash
# Connect using psql
PGPASSWORD=test_password psql -h localhost -p 5433 -U test_user -d consent_engine_test

# Or using docker exec
docker exec -it $(docker ps -q -f name=postgres-test) psql -U test_user -d consent_engine_test
```

#### Useful Database Queries
```sql
-- View all consent records
SELECT * FROM consent_records ORDER BY created_at DESC;

-- View consent records by status
SELECT consent_id, owner_id, status, created_at FROM consent_records WHERE status = 'pending';

-- View consent records by owner
SELECT * FROM consent_records WHERE owner_id = '199512345678';

-- Check database schema
\d consent_records
```

### 5. Running Tests

#### Unit Tests (In-Memory Engine)
```bash
# Run all tests with in-memory engine
make test

# Run with coverage
go test -v -cover
```

#### Integration Tests (PostgreSQL)
```bash
# Run tests with PostgreSQL database
make test-local

# Run specific test
go test -v -run TestHybridAuthMiddleware
```

#### Test Structure
- **Unit Tests**: Located in root directory (required for package access)
- **Integration Tests**: Located in `tests/` directory
- **Test Utilities**: Located in `testutils/` directory with reusable helpers

### 6. Troubleshooting

#### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check database logs
docker logs $(docker ps -q -f name=postgres-test)

# Restart database
make clean
make setup-test-db
```

#### Service Issues
```bash
# Check if service is running
curl http://localhost:8081/health

# Check service logs (if running in background)
ps aux | grep "go run"
```

#### Port Conflicts
If port 8081 is already in use:
```bash
# Kill existing process
lsof -ti:8081 | xargs kill -9

# Or use a different port
PORT=8082 go run .
```

### 7. Cleanup

```bash
# Stop and remove test database
make clean

# Stop the service (if running in background)
pkill -f "go run"
```

### Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `CHOREO_OPENDIF_DB_HOSTNAME` | `localhost` | Database host |
| `CHOREO_OPENDIF_DB_PORT` | `5432` | Database port |
| `CHOREO_OPENDIF_DB_USERNAME` | `postgres` | Database username |
| `CHOREO_OPENDIF_DB_PASSWORD` | `password` | Database password |
| `CHOREO_OPENDIF_DB_DATABASENAME` | `consent_engine` | Database name |
| `DB_SSLMODE` | `require` | SSL mode for database connection |
| `PORT` | `8081` | Service port |
| `CONSENT_PORTAL_URL` | `http://localhost:5173` | Consent portal URL |

## Consent States

- **`pending`**: Awaiting data owner decision
- **`approved`**: Data owner has approved consent
- **`rejected`**: Data owner has rejected consent
- **`expired`**: Consent has expired
- **`revoked`**: Consent has been revoked

## Health Check

```bash
curl http://localhost:8081/health
```

**Response:**
```json
{
  "service": "consent-engine",
  "status": "healthy"
}
```

## Configuration

### Environment Variables

#### Required (for JWT authentication)
- `ASGARDEO_JWKS_URL` - JWKS endpoint URL for user token validation
- `ASGARDEO_ISSUER` - JWT issuer URL for user tokens
- `ASGARDEO_AUDIENCE` - JWT audience for user tokens
- `ASGARDEO_ORG_NAME` - Your organization name

#### Optional (with defaults)
- `PORT` - Service port (default: 8081)
- `CONSENT_PORTAL_URL` - Consent portal URL (default: http://localhost:5173)

### Development Constants

The `constants.go` file contains a fallback mapping for local development and testing:

```go
// ownerIDToEmailMap - Fallback mapping for development
// TODO: Remove this file once SCIM integration is fully tested and deployed
var ownerIDToEmailMap = map[string]string{
    "199512345678": "test@opensource.lk",
    // ... more mappings
}
```

**Important**: This file is temporary and will be removed once the SCIM integration is fully tested and deployed. In production, `owner_email` is resolved via Asgardeo's SCIM API using the `owner_id` (NIC).

### Getting Asgardeo Credentials

1. Go to [Asgardeo Console](https://console.asgardeo.io/)
2. Navigate to your organization
3. Go to Applications → Your App → Settings
4. Copy the following values:
   - **Base URL**: Your organization URL
   - **Client ID**: From the application settings
   - **Client Secret**: From the application settings (if using confidential client)

## Integration

The Consent Engine integrates with:
- **Policy Decision Point**: Provides consent requirements for authorization decisions
- **Consent Portal**: Web interface for consent management (uses Asgardeo user JWT tokens)
- **Orchestration Engine**: Service-to-service communication (uses internal endpoint access)
- **Data Consumer Applications**: Receives consent requests and provides status updates

### Authentication Flow

1. **User Access (Consent Portal)**:
   - User authenticates with Asgardeo
   - Consent Portal receives user JWT token
   - Consent Portal sends requests with `Authorization: Bearer <user_jwt>`
   - Consent Engine validates user JWT and enforces ownership checks

2. **Internal Access (Orchestration Engine)**:
   - Orchestration Engine uses project-internal network access
   - Orchestration Engine calls `POST /consents` without authentication
   - Consent Engine processes internal requests without JWT validation