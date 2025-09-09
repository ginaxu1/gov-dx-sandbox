# Configuration Management

Simplified configuration management for all Go services in the exchange system using command-line flags and environment variables.

## Overview

This package provides a unified, simple configuration system that supports:
- **Command-line flags** for local development and testing
- **Environment variables** for Docker and WSO2 Choreo deployment
- **Smart defaults** based on environment detection
- **Single configuration file** - no multiple .env files needed

## Architecture

```
config/
├── config.go              # Single configuration file with flags and defaults
└── README.md              # This documentation
```

## Usage

### Basic Configuration Loading

```go
import "github.com/gov-dx-sandbox/exchange/config"

// Load configuration using flags
cfg := config.LoadConfig("consent-engine")

// Access configuration values
fmt.Printf("Service: %s on port %s\n", cfg.Service.Name, cfg.Service.Port)
fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
```

### Command-Line Flags

```bash
# Default local settings
go run ./consent-engine

# Custom settings
go run ./consent-engine -port=8083 -log-level=info -cors=false

# Production mode
go run ./consent-engine -env=production
```

### Environment Variables

```bash
# Override with environment variables
export ENVIRONMENT=production
export PORT=8081
export LOG_LEVEL=warn
go run ./consent-engine
```

## Available Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-env` | local | Environment (local/production) |
| `-port` | 8081/8082 | Service port (auto-detected) |
| `-host` | 0.0.0.0 | Host address |
| `-timeout` | 10s | Request timeout |
| `-log-level` | debug/warn | Log level |
| `-log-format` | text/json | Log format |
| `-jwt-secret` | auto | JWT secret |
| `-cors` | true/false | Enable CORS |
| `-rate-limit` | 1000/100 | Rate limit per minute |

## Service Ports

- **Consent Engine**: 8081
- **Policy Decision Point**: 8082

## Default Values

The system provides intelligent defaults based on environment:

| Setting | Local | Production |
|---------|-------|------------|
| Log Level | debug | warn |
| Log Format | text | json |
| CORS | enabled | disabled |
| Rate Limit | 1000/min | 100/min |
| JWT Secret | local-secret-key | (must be set) |

## Deployment Examples

### Local Development
```bash
# Default local settings
go run ./consent-engine

# Custom settings
go run ./consent-engine -port=8083 -log-level=info
```

### Docker Deployment
```bash
# Build and run
docker build --build-arg SERVICE_PATH=consent-engine -t consent-engine .
docker run -p 8081:8081 consent-engine -env=production

# With environment variables
docker run -p 8081:8081 -e ENVIRONMENT=production consent-engine
```

### WSO2 Choreo Deployment
```yaml
# component.yaml
configurations:
  env:
    - name: ENVIRONMENT
      valueFrom:
        configForm:
          type: string
    - name: PORT
      valueFrom:
        configForm:
          type: string
```

## Benefits

1. **Single Configuration File** - No multiple .env files to manage
2. **Command-Line Friendly** - Easy to override settings during development
3. **Environment Variable Support** - Works with Docker and WSO2 Choreo
4. **Smart Defaults** - Automatically chooses appropriate settings
5. **Simple** - Much easier to understand and maintain
6. **Universal** - Works with all deployment methods

## Migration from Complex System

### Before (Complex)
```bash
# Multiple .env files
config/consent-engine-local.env
config/consent-engine-production.env
config/policy-decision-point-local.env
config/policy-decision-point-production.env

# Complex loading
cfg, err := config.LoadConfigForEnvironment("consent-engine", "local")
```

### After (Simple)
```bash
# Single configuration file
config/config.go

# Simple loading
cfg := config.LoadConfig("consent-engine")

# Override with flags
go run ./consent-engine -env=production -port=8081
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure each service uses a unique port
2. **Flag parsing**: Use `-help` to see available flags
3. **Environment variables**: Check that variables are properly set
4. **WSO2 Choreo deployment**: Ensure component.yaml files reference correct environment variables

### Debug Configuration

```go
// Print current configuration
cfg := config.LoadConfig("consent-engine")
fmt.Printf("%+v\n", cfg)
```

### Help

```bash
# See all available flags
go run ./consent-engine -help
```