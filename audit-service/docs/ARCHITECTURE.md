# Architecture & Project Structure

Understanding the Audit Service architecture and codebase organization.

## Overview

The Audit Service follows a clean architecture with API versioning support, separating concerns into distinct layers for maintainability and future extensibility.

## Directory Structure

```
audit-service/
├── config/                  # Shared configuration
│   ├── config.go           # Environment variable loader
│   ├── config_test.go      # Configuration tests
│   ├── enums.yaml          # Enum definitions
│   └── README.md           # Configuration documentation
│
├── database/                # Database infrastructure layer
│   ├── client.go           # Connection management & config
│   └── client_test.go      # Connection tests
│
├── middleware/              # Shared HTTP middleware
│   ├── cors.go             # CORS middleware
│   └── cors_test.go        # CORS tests
│
├── v1/                      # API Version 1
│   ├── database/           # Data access layer
│   │   ├── database.go     # Repository interface
│   │   └── gorm_repository.go  # GORM implementation
│   │
│   ├── handlers/           # HTTP handlers (controllers)
│   │   ├── audit_handler.go
│   │   └── audit_handler_test.go
│   │
│   ├── models/             # Domain models & DTOs
│   │   ├── audit_log.go    # Core domain model
│   │   ├── base.go         # Common types
│   │   ├── request_dtos.go # Request DTOs
│   │   └── response_dtos.go # Response DTOs
│   │
│   ├── services/           # Business logic layer
│   │   ├── audit_service.go
│   │   ├── audit_service_test.go
│   │   └── errors.go       # Domain errors
│   │
│   ├── testutil/           # Test utilities
│   │   └── mock_repository.go
│   │
│   └── utils/              # Helper functions
│       └── validation.go
│
├── docs/                    # Documentation
│   ├── API.md              # API documentation
│   ├── ARCHITECTURE.md     # This file
│   └── DATABASE_CONFIGURATION.md
│
├── main.go                  # Service entry point
├── go.mod                   # Go module definition
├── Dockerfile              # Container image
├── docker-compose.yml      # Local deployment
├── openapi.yaml            # OpenAPI specification
├── .env.example            # Environment template
└── README.md               # Getting started guide
```

## Architecture Layers

### 1. Entry Point (`main.go`)

**Responsibilities:**

- Initialize configuration
- Establish database connection
- Set up HTTP routes
- Start HTTP server
- Handle graceful shutdown

**Key Functions:**

```go
func main() {
    // Load configuration
    // Connect to database
    // Initialize services
    // Setup HTTP routes
    // Start server
    // Handle signals
}
```

### 2. Infrastructure Layer

#### Configuration (`config/`)

**Purpose:** Centralized configuration management

**Components:**

- `config.go` - Environment variable loader with defaults
- `enums.yaml` - Enum definitions for validation
- Enum validation and loading logic

**Example:**

```go
// Load environment variable with default
port := config.GetEnvOrDefault("PORT", "3001")

// Load enum configuration
enums, err := config.LoadEnums("config/enums.yaml")
```

#### Database (`database/`)

**Purpose:** Database connection and configuration management

**Components:**

- `client.go` - Database config struct and connection logic
- Support for SQLite and PostgreSQL
- Connection pool management
- Auto-detection of database type

**Key Features:**

- In-memory SQLite when no config provided
- File-based SQLite for persistence
- PostgreSQL for production
- Automatic directory creation
- Connection pool configuration

**Example:**

```go
// Create database config from environment
config := database.NewDatabaseConfig()

// Establish GORM connection
db, err := database.ConnectGormDB(config)
```

#### Middleware (`middleware/`)

**Purpose:** HTTP request/response interceptors

**Components:**

- `cors.go` - CORS middleware for cross-origin requests

**Example:**

```go
corsMiddleware := middleware.NewCORSMiddleware()
handler := corsMiddleware(mux)
```

### 3. API Layer (`v1/`)

Implements version 1 of the API with clean separation of concerns.

#### Database Repository (`v1/database/`)

**Purpose:** Data access abstraction

**Repository Pattern:**

```go
// Interface (database.go)
type AuditRepository interface {
    CreateAuditLog(ctx, *AuditLog) (*AuditLog, error)
    GetAuditLogsByTraceID(ctx, string) ([]AuditLog, error)
    GetAuditLogs(ctx, *AuditLogFilters) ([]AuditLog, int64, error)
}

// GORM Implementation (gorm_repository.go)
type GormRepository struct {
    db *gorm.DB
}
```

**Benefits:**

- Database-agnostic interface
- Easy to mock for testing
- Supports multiple implementations
- Clear data access boundaries

#### Models (`v1/models/`)

**Purpose:** Domain models and data transfer objects

**Components:**

- `audit_log.go` - Core domain model with GORM tags
- `request_dtos.go` - Request validation models
- `response_dtos.go` - API response structures
- `base.go` - Common types and enums

**Example:**

```go
// Domain model
type AuditLog struct {
    ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
    TraceID      *uuid.UUID `gorm:"type:uuid;index"`
    Timestamp    time.Time `gorm:"index"`
    EventType    string
    // ...
}

// Request DTO
type CreateAuditLogRequest struct {
    TraceID   *string `json:"traceId"`
    Timestamp string  `json:"timestamp"`
    // ...
}
```

#### Services (`v1/services/`)

**Purpose:** Business logic and domain operations

**Components:**

- `audit_service.go` - Core business logic
- `errors.go` - Domain-specific errors

**Responsibilities:**

- Input validation
- Business rule enforcement
- Coordinate repository operations
- Error handling and wrapping

**Example:**

```go
type AuditService struct {
    repo database.AuditRepository
}

func (s *AuditService) CreateAuditLog(ctx context.Context, req *CreateAuditLogRequest) (*AuditLog, error) {
    // Validate input
    // Apply business rules
    // Call repository
    // Return result
}
```

#### Handlers (`v1/handlers/`)

**Purpose:** HTTP request handling (controllers)

**Components:**

- `audit_handler.go` - HTTP handlers

**Responsibilities:**

- Parse HTTP requests
- Call service layer
- Format HTTP responses
- Handle HTTP-specific errors

**Example:**

```go
type AuditHandler struct {
    service *services.AuditService
}

func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Call service
    // Write response
}
```

#### Test Utilities (`v1/testutil/`)

**Purpose:** Testing support

**Components:**

- `mock_repository.go` - Mock implementation of repository

**Example:**

```go
type MockRepository struct {
    CreateAuditLogFunc func(context.Context, *models.AuditLog) (*models.AuditLog, error)
}
```

## Design Patterns

### 1. Repository Pattern

**Purpose:** Abstract data access layer

**Benefits:**

- Decouples business logic from database
- Easy to test with mocks
- Supports multiple database backends

```go
// Interface defines contract
type AuditRepository interface {
    CreateAuditLog(ctx, *AuditLog) (*AuditLog, error)
}

// Implementation for GORM
type GormRepository struct { db *gorm.DB }

// Can add more implementations:
// type MongoRepository struct { client *mongo.Client }
```

### 2. Dependency Injection

**Purpose:** Loose coupling between components

**Implementation:**

```go
// main.go
repo := v1database.NewGormRepository(db)
service := v1services.NewAuditService(repo)
handler := v1handlers.NewAuditHandler(service)
```

**Benefits:**

- Easier testing
- Flexible component replacement
- Clear dependencies

### 3. API Versioning

**Purpose:** Support multiple API versions simultaneously

**Structure:**

```
v1/           # API Version 1
  database/
  handlers/
  models/
  services/

v2/           # API Version 2 (future)
  database/
  handlers/
  models/
  services/
```

**Benefits:**

- Backward compatibility
- Gradual migration
- Clean separation of versions

### 4. DTO Pattern

**Purpose:** Separate API contracts from domain models

**Example:**

```go
// Request DTO - what API accepts
type CreateAuditLogRequest struct {
    Timestamp string `json:"timestamp"` // String for validation
}

// Domain Model - internal representation
type AuditLog struct {
    Timestamp time.Time // Parsed time
}
```

**Benefits:**

- API stability
- Flexible internal changes
- Clear validation boundaries

## Data Flow

### Create Audit Log Flow

```
HTTP Request
    ↓
Handler (audit_handler.go)
    ├─ Parse JSON request
    ├─ Validate request format
    ↓
Service (audit_service.go)
    ├─ Validate business rules
    ├─ Convert DTO to domain model
    ├─ Generate IDs
    ↓
Repository (gorm_repository.go)
    ├─ Execute database operation
    ├─ Handle database errors
    ↓
Database (SQLite/PostgreSQL)
    ├─ Persist data
    ↓
← Response propagates back up
```

### Get Audit Logs Flow

```
HTTP Request
    ↓
Handler (audit_handler.go)
    ├─ Parse query parameters
    ├─ Validate filters
    ↓
Service (audit_service.go)
    ├─ Build filter criteria
    ├─ Apply pagination
    ↓
Repository (gorm_repository.go)
    ├─ Build GORM query
    ├─ Execute query
    ├─ Count total
    ↓
Database (SQLite/PostgreSQL)
    ├─ Retrieve data
    ↓
← Results propagate back up
```

## Testing Strategy

### Unit Tests

**Location:** `*_test.go` files alongside source

**Focus:**

- Individual function behavior
- Business logic validation
- Error handling

**Example:**

```go
// services/audit_service_test.go
func TestCreateAuditLog_ValidInput(t *testing.T) {
    // Arrange
    mockRepo := &testutil.MockRepository{}
    service := NewAuditService(mockRepo)

    // Act
    result, err := service.CreateAuditLog(ctx, request)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests

**Approach:** Use real database (SQLite in-memory)

**Benefits:**

- Test full stack
- Verify database operations
- Catch integration issues

### Test Database Strategy

- Use `:memory:` SQLite for tests
- Automatic cleanup
- Fast execution
- No external dependencies

## Adding a New Endpoint

### Step-by-Step Guide

1. **Add to models** (`v1/models/`)

```go
// request_dtos.go
type NewFeatureRequest struct {
    Field string `json:"field"`
}

// response_dtos.go
type NewFeatureResponse struct {
    Result string `json:"result"`
}
```

2. **Add to repository interface** (`v1/database/database.go`)

```go
type AuditRepository interface {
    NewFeature(ctx context.Context, req *Request) (*Response, error)
}
```

3. **Implement in repository** (`v1/database/gorm_repository.go`)

```go
func (r *GormRepository) NewFeature(ctx context.Context, req *Request) (*Response, error) {
    // Implementation
}
```

4. **Add service method** (`v1/services/audit_service.go`)

```go
func (s *AuditService) NewFeature(ctx context.Context, req *Request) (*Response, error) {
    // Business logic
    return s.repo.NewFeature(ctx, req)
}
```

5. **Add handler** (`v1/handlers/audit_handler.go`)

```go
func (h *AuditHandler) NewFeature(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Call service
    // Write response
}
```

6. **Register route** (`main.go`)

```go
mux.HandleFunc("/api/new-feature", handler.NewFeature)
```

## Future Enhancements

### API Version 2

When adding v2:

```
v2/
  database/      # New repository interfaces
  handlers/      # New handlers
  models/        # New models/DTOs
  services/      # New business logic
```

Keep v1 unchanged for backward compatibility.

### Additional Database Backends

Add new implementations:

```go
// v1/database/mongo_repository.go
type MongoRepository struct {
    client *mongo.Client
}

func (r *MongoRepository) CreateAuditLog(...) { }
```

### Metrics & Monitoring

Potential additions:

- Prometheus metrics
- Request tracing
- Performance monitoring
- Query statistics

## Best Practices

1. **Keep layers separate** - Don't mix concerns
2. **Use interfaces** - Enable testing and flexibility
3. **Version your API** - Keep changes under v1/, v2/, etc.
4. **Write tests** - Unit and integration tests
5. **Document changes** - Update relevant docs
6. **Handle errors gracefully** - Wrap and provide context
7. **Use context** - Pass context through all layers
8. **Validate early** - Validate at handler/service boundary

## References

- [API Documentation](API.md)
- [Database Configuration](DATABASE_CONFIGURATION.md)
- [OpenAPI Specification](../openapi.yaml)
- [Configuration Guide](../config/README.md)
