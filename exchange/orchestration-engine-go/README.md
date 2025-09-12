# Orchestration Engine

The Orchestration Engine (OE) is the central coordinator in the data exchange platform, managing the flow between data consumers, policy decisions, and consent management.

## Overview

The Orchestration Engine acts as the central hub that:
- Receives data requests from consumers
- Coordinates with the Policy Decision Point (PDP) for authorization
- Manages consent workflows through the Consent Engine (CE)
- Routes approved requests to appropriate data providers

## Architecture

```
Data Consumer → Orchestration Engine → Policy Decision Point (PDP)
                     ↓
              Consent Engine (CE) ← (if consent required)
                     ↓
              Data Provider
```

## Service Dependencies

- **Policy Decision Point (PDP)** - Port 8082: Authorization decisions
- **Consent Engine (CE)** - Port 8081: Data owner consent management
- **Data Providers** - Various ports: Actual data sources

## API Integration

### 1. Policy Decision Point Integration

**Endpoint:** `POST http://localhost:8082/decide`

**Request Format:**
```json
{
  "consumer_id": "passport-app",
  "app_id": "passport-app", 
  "request_id": "req_789",
  "required_fields": [
    "person.fullName",
    "person.nic",
    "person.birthDate",
    "person.permanentAddress",
    "person.photo"
  ]
}
```

**Response Format:**
```json
{
  "allow": true,
  "consent_required": true,
  "consent_required_fields": [
    "person.permanentAddress",
    "person.photo"
  ]
}
```

### 2. Consent Engine Integration

**Create Consent Request:**
- **Endpoint:** `POST http://localhost:8081/consent`
- **Purpose:** Create a new consent record for data owner approval

**Request Format:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": [
        "person.permanentAddress",
        "person.photo"
      ]
    }
  ],
  "purpose": "passport_application",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback"
}
```

**Response Format:**
```json
{
  "id": "consent_abc123",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:20:00Z",
  "expires_at": "2025-10-10T10:20:00Z",
  "data_consumer": "passport-app",
  "data_owner": "199512345678",
  "fields": [
    "person.permanentAddress",
    "person.photo"
  ],
  "consent_portal_url": "/consent-portal/xyz789",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "metadata": {
    "purpose": "passport_application",
    "request_id": "req_789"
  }
}
```

**Check Consent Status:**
- **Endpoint:** `GET http://localhost:8081/consent/{consent_id}`
- **Purpose:** Monitor consent approval status

**Consent Status Values:**
- `pending` - Waiting for data owner decision
- `approved` - Consent approved, proceed with data access
- `denied` - Consent denied, return error to consumer
- `expired` - Consent expired, may need renewal
- `revoked` - Consent revoked by data owner

## Workflow Implementation

### 1. Data Request Processing

```go
type DataRequest struct {
    ConsumerID    string   `json:"consumer_id"`
    DataOwnerID   string   `json:"data_owner_id"`
    RequiredFields []string `json:"required_fields"`
    Purpose       string   `json:"purpose"`
    RequestID     string   `json:"request_id"`
}
```

### 2. Authorization Flow

```go
// Step 1: Call Policy Decision Point
pdpRequest := PDPRequest{
    ConsumerID:    req.ConsumerID,
    AppID:        req.ConsumerID,
    RequestID:    req.RequestID,
    RequiredFields: req.RequiredFields,
}

pdpResponse, err := callPDP(pdpRequest)
if err != nil {
    return handleError("PDP call failed", err)
}

if !pdpResponse.Allow {
    return handleError("Access denied by PDP", nil)
}
```

### 3. Consent Management

```go
// Step 2: Handle consent if required
if pdpResponse.ConsentRequired {
    consentRequest := ConsentRequest{
        AppID: req.ConsumerID,
        DataFields: []DataField{
            {
                OwnerType: "citizen",
                OwnerID:   req.DataOwnerID,
                Fields:    pdpResponse.ConsentRequiredFields,
            },
        },
        Purpose:     req.Purpose,
        SessionID:   generateSessionID(),
        RedirectURL: buildRedirectURL(req.ConsumerID),
    }
    
    consentResponse, err := callConsentEngine(consentRequest)
    if err != nil {
        return handleError("Consent creation failed", err)
    }
    
    // Wait for consent approval
    return waitForConsent(consentResponse.ID)
}
```

### 4. Data Access

```go
// Step 3: Proceed with data access once authorized
dataResponse, err := callDataProvider(req)
if err != nil {
    return handleError("Data provider call failed", err)
}

return dataResponse
```

## Error Handling

### Common Error Scenarios

| Scenario | Action | Response to Consumer |
|----------|--------|---------------------|
| PDP unavailable | Retry with backoff | 503 Service Unavailable |
| PDP denies access | Return immediately | 403 Forbidden |
| CE unavailable | Retry with backoff | 503 Service Unavailable |
| Consent denied | Return immediately | 403 Forbidden |
| Consent timeout | Return immediately | 408 Request Timeout |
| Data provider error | Return error | 502 Bad Gateway |

### Retry Logic

```go
func callServiceWithRetry(serviceCall func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := serviceCall()
        if err == nil {
            return nil
        }
        
        if isRetryableError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        
        return err
    }
    return errors.New("max retries exceeded")
}
```

## Configuration

### Service Endpoints

```go
type Config struct {
    PDPEndpoint string `json:"pdp_endpoint"` // http://localhost:8082
    CEEndpoint  string `json:"ce_endpoint"`  // http://localhost:8081
    Timeout     int    `json:"timeout"`      // seconds
    MaxRetries  int    `json:"max_retries"`
}
```

### Environment Variables

```bash
PDP_ENDPOINT=http://localhost:8082
CE_ENDPOINT=http://localhost:8081
REQUEST_TIMEOUT=30
MAX_RETRIES=3
```

## Monitoring and Logging

### Key Metrics

- PDP response times
- CE response times
- Consent approval rates
- Error rates by service
- Request throughput

### Logging Format

```go
log.Info("PDP request", 
    "consumer_id", req.ConsumerID,
    "request_id", req.RequestID,
    "fields", req.RequiredFields,
    "duration_ms", duration)

log.Info("Consent created",
    "consent_id", consentResponse.ID,
    "data_owner", consentResponse.DataOwner,
    "fields", consentResponse.Fields)
```

## Security Considerations

### Authentication
- Use API keys or JWT tokens for service-to-service communication
- Validate all incoming requests
- Log all access attempts

### Data Privacy
- Don't log sensitive data fields
- Use request IDs for correlation without exposing data
- Implement proper data retention policies

## Performance Optimization

### Caching
- Cache PDP responses for similar requests
- Cache consent status to avoid repeated polling
- Implement cache invalidation strategies

### Async Processing
- Use async patterns for consent polling
- Implement webhooks for consent status changes
- Queue long-running operations

## Testing

### Unit Tests
Test each integration point independently:
- PDP integration tests
- Consent Engine integration tests
- Error handling tests

### Integration Tests
Test the complete flow:
- Consumer request → PDP → CE → Data provider
- Consumer request → PDP → Data provider (no consent)
- Error scenarios and edge cases

## Development

### Prerequisites
- Go 1.21 or later
- Access to PDP service (port 8082)
- Access to CE service (port 8081)

### Running the Service

```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go
```

### Building

```bash
# Build binary
go build -o orchestration-engine main.go

# Run binary
./orchestration-engine
```

## Related Documentation

- [Integration Guide](INTEGRATION_GUIDE.md) - Detailed integration examples
- [Policy Decision Point README](../policy-decision-point/README.md)
- [Consent Engine README](../consent-engine/README.md)
- [Exchange Platform README](../README.md)
