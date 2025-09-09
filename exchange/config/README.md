# Configuration Management

Centralized configuration management for all Go services in the exchange system

## Overview

This package provides a unified configuration system that supports:
- **Environment-specific settings** (local, production)
- **Service-specific overrides** (consent-engine, policy-decision-point)
- **Smart defaults** with minimal configuration required
- **WSO2 Choreo deployment** compatibility

## Architecture

```
config/
├── config.go              # Core configuration structs and defaults
├── loader.go              # Environment file loading and validation
├── consent-engine-local.env      # Consent Engine local overrides
├── consent-engine-production.env # Consent Engine production overrides
├── policy-decision-point-local.env      # Policy Decision Point local overrides
└── policy-decision-point-production.env # Policy Decision Point production overrides
```

## Usage

### Basic Configuration Loading

```go
import "github.com/gov-dx-sandbox/exchange/config"

// Load configuration for a service
cfg, err := config.LoadConfigForEnvironment("consent-engine", "local")
if err != nil {
    log.Fatal(err)
}

// Access configuration values
fmt.Printf("Service: %s on port %s\n", cfg.Service.Name, cfg.Service.Port)
fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
```

### Service-Specific Configuration

Each service automatically loads its own environment file:

- **Consent Engine**: `consent-engine-{environment}.env`
- **Policy Decision Point**: `policy-decision-point-{environment}.env`

## Configuration Structure

### Service Configuration
- **Name**: Service identifier
- **Port**: Service port (8081 for consent-engine, 8082 for policy-decision-point)
- **Host**: Bind address (default: 0.0.0.0)
- **Timeouts**: Read/Write/Idle timeouts

### Database Configuration
- **Host**: Database server address
- **Port**: Database port (default: 5432)
- **User/Password**: Authentication credentials
- **Name**: Database name (auto-generated: `{service}-{environment}`)
- **SSL Mode**: Security mode (disable for local, require for production)
- **Max Connections**: Connection pool size

### Logging Configuration
- **Level**: Log level (debug for local, warn for production)
- **Format**: Output format (text for local, json for production)

### Security Configuration
- **JWT Secret**: Authentication secret
- **CORS**: Cross-origin resource sharing settings
- **Rate Limiting**: Requests per minute limits

## Environment Files

### Local Development
```bash
# consent-engine-local.env
ENVIRONMENT=local
PORT=8081
ENABLE_CORS=true
RATE_LIMIT_PER_MINUTE=1000

# policy-decision-point-local.env
ENVIRONMENT=local
PORT=8082
LOG_LEVEL=debug
LOG_FORMAT=text
ENABLE_CORS=true
RATE_LIMIT_PER_MINUTE=1000
```

### Production
```bash
# consent-engine-production.env
ENVIRONMENT=production
PORT=8081
ENABLE_CORS=false
RATE_LIMIT_PER_MINUTE=100

# policy-decision-point-production.env
ENVIRONMENT=production
PORT=8082
LOG_LEVEL=warn
LOG_FORMAT=json
ENABLE_CORS=false
RATE_LIMIT_PER_MINUTE=100
```

## Default Values

The system provides intelligent defaults based on environment:

| Setting | Local | Production |
|---------|-------|------------|
| Log Level | debug | warn |
| Log Format | text | json |
| CORS | enabled | disabled |
| Rate Limit | 1000/min | 100/min |
| DB SSL | disable | require |
| Max Connections | 5 | 50 |

## Service Ports

- **Consent Engine**: 8081
- **Policy Decision Point**: 8082

## WSO2 Choreo Integration

This configuration system should work seamlessly with WSO2 Choreo:

- Environment variables are automatically loaded from Choreo's configuration
- Service-specific settings are applied based on the component
- Database connections use Choreo's connection references
- Security settings integrate with Choreo's secret management

## Development

### Adding New Services

1. Create service-specific environment files:
   ```bash
   # new-service-local.env
   ENVIRONMENT=local
   PORT=8083
   ```

2. Add port mapping in `config.go`:
   ```go
   func getDefaultPort(serviceName string) string {
       ports := map[string]string{
           "consent-engine":        "8081",
           "policy-decision-point": "8082",
           "new-service":           "8083", // Add here
       }
       // ...
   }
   ```

### Adding New Configuration Options

1. Add field to appropriate config struct in `config.go`
2. Add default value function
3. Update environment files as needed
4. Update this README

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure each service uses a unique port
2. **Missing environment files**: Check that service-specific files exist
3. **Configuration validation**: Verify all required fields are set
4. **WSO2 Choreo deployment**: Ensure component.yaml files reference correct environment variables

### Debug Configuration

```go
// Print current configuration
cfg, _ := config.LoadConfigForEnvironment("consent-engine", "local")
fmt.Printf("%+v\n", cfg)
```