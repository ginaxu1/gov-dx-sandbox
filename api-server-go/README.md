# API Server (Go)

A RESTful API server for government data exchange portal management. Built with Go and runs on port 3000.

## Overview

The API server provides RESTful endpoints for managing:
- Consumer applications
- Provider submissions and profiles
- Provider schemas
- Admin functions

## Architecture

```
api-server-go/
├── main.go              # Server entry point
├── handlers/
│   └── server.go        # Generic HTTP handlers (unified)
├── models/              # Data structures
│   ├── consumer.go      # Consumer types
│   └── provider.go      # Provider types
├── services/            # Business logic
│   ├── consumer.go      # Consumer operations
│   ├── provider.go      # Provider operations
│   └── admin.go         # Admin dashboard
├── tests/               # Unit tests
└── go.mod              # Dependencies
```

**Key Features:**
- Generic handler pattern reduces code duplication
- Shared utils package for common operations
- In-memory storage with thread-safe operations
- Comprehensive input validation

## API Endpoints

### Consumer Management
- `GET /consumers` - List applications
- `POST /consumers` - Create application
- `GET /consumers/{id}` - Get application
- `PUT /consumers/{id}` - Update application
- `DELETE /consumers/{id}` - Delete application

### Provider Management
- `GET /provider-submissions` - List submissions
- `POST /provider-submissions` - Create submission
- `GET /provider-submissions/{id}` - Get submission
- `PUT /provider-submissions/{id}` - Update submission
- `GET /provider-profiles` - List profiles
- `GET /provider-profiles/{id}` - Get profile
- `GET /provider-schemas` - List schemas
- `POST /provider-schemas` - Create schema
- `GET /provider-schemas/{id}` - Get schema
- `PUT /provider-schemas/{id}` - Update schema

### Admin
- `GET /admin/dashboard` - Dashboard with metrics and recent activity

## Development

### Running
```bash
go run main.go
# Server starts on http://localhost:3000
```

### Building
```bash
go build -o api-server
./api-server
```

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./tests

# Run specific test
go test -v ./tests -run TestConsumerService
```

### Dependencies
- Uses shared utils from `github.com/gov-dx-sandbox/exchange/utils`
- Standard Go modules for HTTP, JSON, and logging

## Implementation Details

**Handler Pattern:** Generic `handleCollection` and `handleItem` functions reduce code duplication across all endpoints.

**Data Storage:** In-memory maps with `sync.RWMutex` for thread-safe concurrent access.

**Validation:** Server-side validation for required fields with detailed error messages.

**Response Format:** Consistent JSON responses with proper HTTP status codes and error handling.

**Admin Dashboard:** Dynamic generation of metrics, recent activity, and aggregated data from all services.