# Data Exchange Portal Backend

A secure, scalable Go-based REST Portal Backend for managing data exchange workflows, including member management, schema submissions, and application processing with comprehensive authentication and authorization.

## 🚀 Features

- **JWT Authentication** with Asgardeo integration
- **Role-Based Access Control (RBAC)** with granular permissions
- **PostgreSQL Database** with automatic schema management
- **Thread-Safe Caching** for optimal performance
- **OpenAPI/Swagger Documentation**
- **Comprehensive Health Monitoring**
- **Docker Support** for containerized deployment
- **Audit Logging** for compliance and debugging

## 📋 API Endpoints

### Core Resources

- **Members** - `/api/v1/members` - User profile and membership management
- **Schemas** - `/api/v1/schemas` - Data schema definitions and management
- **Schema Submissions** - `/api/v1/schema-submissions` - Schema submission workflow
- **Applications** - `/api/v1/applications` - Application definitions
- **Application Submissions** - `/api/v1/application-submissions` - Application submission workflow

### System Endpoints

- **Health Check** - `/health` - System health and database status
- **API Documentation** - `/openapi.yaml` - OpenAPI specification

## 🔐 Authentication & Authorization

### Supported Roles

- **OpenDIF_Admin** - Full system access and management capabilities
- **OpenDIF_Member** - Standard user access to own resources
- **OpenDIF_System** - System-level operations with read access

### Permission System

The API implements fine-grained permissions including:

- Resource creation, reading, updating permissions
- Ownership-based access control
- Admin override capabilities
- Cached permission evaluation for performance

### JWT Token Requirements

- **Issuer**: Asgardeo identity provider
- **Audience**: Configured client IDs (member-portal, admin-portal)
- **Claims**: Must include valid roles and user information
- **Validation**: JWKS-based signature verification with key rotation support

## 🛠️ Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Docker (optional)

### 1. Environment Setup

Create a `.env` file:

```bash
# Database Configuration
CHOREO_DB_portal_backend_HOSTNAME=localhost
CHOREO_DB_portal_backend_PORT=5432
CHOREO_DB_portal_backend_USERNAME=postgres
CHOREO_DB_portal_backend_PASSWORD=your_password
CHOREO_DB_portal_backend_DATABASENAME=portal_backend

# JWT Authentication (Required)
ASGARDEO_BASE_URL=https://api.asgardeo.io/t/your-org
ASGARDEO_MEMBER_CLIENT_ID=your_member_client_id
ASGARDEO_ADMIN_CLIENT_ID=your_admin_client_id

# Policy Decision Point
CHOREO_PDP_CONNECTION_SERVICEURL=http://localhost:9000
CHOREO_PDP_CONNECTION_CHOREOAPIKEY=your_pdp_key

# Optional: Asgardeo Management (for member creation)
ASGARDEO_CLIENT_ID=management_client_id
ASGARDEO_CLIENT_SECRET=management_client_secret
ASGARDEO_SCOPES="internal_user_mgt_create internal_user_mgt_list"
```

### 2. Database Setup

```bash
# Start PostgreSQL with Docker
make setup-test-db

# Or use your own PostgreSQL instance
# Database tables will be auto-created on startup
```

### 3. Run the Application

```bash
# Install dependencies
go mod download

# Run the server
make run

# Or run directly
go run main.go
```

### 4. Test the API

```bash
# Health check
curl http://localhost:3000/health

# API documentation
curl http://localhost:3000/openapi.yaml
```

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

```bash
cd ../exchange
docker-compose up postgres portal-backend
```

### Standalone Docker

```bash
# Build image
docker build -t portal-backend .

# Run container
docker run -p 3000:3000 \
  -e CHOREO_DB_portal_backend_HOSTNAME=host.docker.internal \
  -e CHOREO_DB_portal_backend_PASSWORD=password \
  --env-file .env \
  portal-backend
```

## 🧪 Testing

### Run All Tests

```bash
make test-all
```

### Test Categories

```bash
# Unit tests only
go test ./...

# Integration tests with PostgreSQL
make test-postgres

# Tests with race detection
go test -race ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Database Setup

For integration tests, set:

```bash
export TEST_DB_PASSWORD=test_password
make test-local
```

## 📊 Database Schema

### Core Tables

- **`members`** - User profiles and membership information
- **`schemas`** - Data schema definitions with versioning
- **`schema_submissions`** - Schema submission workflow and status
- **`applications`** - Application templates and definitions
- **`application_submissions`** - Application submission workflow

### Features

- **Auto-migration** on startup
- **Connection pooling** with configurable limits
- **Health monitoring** with metrics
- **Transaction support** with timeouts
- **Query optimization** with prepared statements

## 🔧 Configuration

### Database Optimization

```bash
DB_MAX_OPEN_CONNS=25              # Maximum open connections
DB_MAX_IDLE_CONNS=5               # Maximum idle connections
DB_CONN_MAX_LIFETIME=1h           # Connection maximum lifetime
DB_QUERY_TIMEOUT=30s              # Query timeout duration
DB_ENABLE_MONITORING=true         # Enable connection pool monitoring
```

### JWT Security

```bash
JWT_VALIDATION_STRICT=true        # Strict JWT validation mode
JWT_CACHE_DURATION=15m            # JWKS cache duration
JWT_TIMEOUT=10s                   # Token validation timeout
```

### Server Configuration

```bash
PORT=3000                         # Server port (default: 3000)
LOG_LEVEL=info                    # Logging level (debug, info, warn, error)
CORS_ALLOWED_ORIGINS=*            # CORS allowed origins
```

## 📈 Monitoring & Health

### Health Endpoint

`GET /health` returns comprehensive system status:

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "database": {
    "status": "connected",
    "open_connections": 5,
    "max_open_connections": 25,
    "idle_connections": 3
  },
  "version": "1.0.0"
}
```

### Logging

- **Structured JSON logging** with configurable levels
- **Request tracing** with correlation IDs
- **Performance metrics** for database operations
- **Security events** for authentication failures

## 🏗️ Architecture

### Project Structure

```
portal-backend/
├── main.go                 # Application entry point
├── v1/                     # API version 1
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # Authentication & authorization
│   ├── models/            # Data models and DTOs
│   ├── services/          # Business logic layer
│   └── utils/             # Utility functions
├── shared/                # Shared utilities
├── idp/                   # Identity provider integrations
└── middleware/            # Global middleware
```

### Security Architecture

```
Request → CORS → JWT Validation → Authorization → Resource Access
    ↓        ↓           ↓              ↓             ↓
 Origin   Token      Role Check    Permission    Ownership
 Check    Verify     & Claims      Validation    Validation
```

## 🤝 Contributing

### Development Setup

1. Clone the repository
2. Set up environment variables
3. Start PostgreSQL database
4. Run tests to verify setup
5. Start the development server

### Code Standards

- **Go formatting** with `gofmt`
- **Linting** with `golangci-lint`
- **Testing** with minimum 80% coverage
- **Documentation** for public APIs
- **Security** best practices enforced

### Submitting Changes

1. Create feature branch
2. Write tests for new functionality
3. Ensure all tests pass
4. Update documentation as needed
5. Submit pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🆘 Support

For issues, questions, or contributions:

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check OpenAPI spec at `/openapi.yaml`
- **Logs**: Check application logs for debugging information

---

**Built with ❤️ using Go, PostgreSQL, and modern security practices.**
