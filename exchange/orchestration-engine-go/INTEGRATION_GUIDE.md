# Orchestration Engine Integration Guide

This is a guide for implementing the orchestration engine's communication with the PDP and CE services.

## Overview

The Orchestration Engine acts as the central coordinator in the data exchange platform, managing the flow between data consumers, policy decisions, and consent management. It communicates with two key services:

- **Policy Decision Point (PDP)** - Port 8082: Handles authorization decisions
- **Consent Engine (CE)** - Port 8081: Manages data owner consent workflows

## Service Communication Flow

```
Data Consumer → Orchestration Engine → Policy Decision Point
                     ↓
              Consent Engine (if consent required)
                     ↓
              Data Provider
```

## 1. Policy Decision Point Integration

### 1.1 Authorization Request

When a data consumer requests data, the OE must first check authorization with the PDP.

**Endpoint:** `POST http://localhost:8082/decide`

**Request Payload:**
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

### 1.2 Response Handling

The PDP response determines the next steps:

- **`allow: false`** → Return error to consumer (access denied)
- **`allow: true, consent_required: false`** → Proceed directly to data provider
- **`allow: true, consent_required: true`** → Proceed to consent workflow

### 1.3 Error Handling

Handle common PDP errors:
- **400 Bad Request** - Invalid request format
- **500 Internal Server Error** - PDP processing error
- **Timeout** - PDP service unavailable

## 2. Consent Engine Integration

### 2.1 Consent Request

When consent is required, the OE must create a consent record with the CE.

**Endpoint:** `POST http://localhost:8081/consent`

**Request Payload:**
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

### 2.2 Consent Status Monitoring

Monitor consent status to determine when to proceed.

**Endpoint:** `GET http://localhost:8081/consent/{consent_id}`

**Response Format:**
```json
{
  "id": "consent_abc123",
  "status": "approved",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:25:00Z",
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

### 2.3 Consent Status Values

- **`pending`** - Waiting for data owner decision
- **`approved`** - Consent granted, proceed with data access
- **`denied`** - Consent denied, return error to consumer
- **`expired`** - Consent expired, may need renewal
- **`revoked`** - Consent revoked by data owner

## 3. Complete Integration Workflow

### 3.1 Step-by-Step Process

1. **Receive Data Request**
   ```go
   type DataRequest struct {
       ConsumerID    string   `json:"consumer_id"`
       DataOwnerID   string   `json:"data_owner_id"`
       RequiredFields []string `json:"required_fields"`
       Purpose       string   `json:"purpose"`
       RequestID     string   `json:"request_id"`
   }
   ```

2. **Call Policy Decision Point**
   ```go
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

3. **Handle Consent Requirements**
   ```go
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

4. **Proceed with Data Access**
   ```go
   // Once consent is approved (or not required)
   dataResponse, err := callDataProvider(req)
   if err != nil {
       return handleError("Data provider call failed", err)
   }
   
   return dataResponse
   ```

### 3.2 Consent Status Polling

```go
func waitForConsent(consentID string) (*DataResponse, error) {
    maxAttempts := 30 // 5 minutes with 10-second intervals
    attempt := 0
    
    for attempt < maxAttempts {
        status, err := getConsentStatus(consentID)
        if err != nil {
            return nil, err
        }
        
        switch status.Status {
        case "approved":
            return proceedWithDataAccess()
        case "denied":
            return nil, errors.New("consent denied by data owner")
        case "expired":
            return nil, errors.New("consent expired")
        case "revoked":
            return nil, errors.New("consent revoked")
        case "pending":
            time.Sleep(10 * time.Second)
            attempt++
        }
    }
    
    return nil, errors.New("consent timeout")
}
```

## 4. Error Handling

### 4.1 Common Error Scenarios

| Scenario | Action | Response to Consumer |
|----------|--------|---------------------|
| PDP unavailable | Retry with backoff | 503 Service Unavailable |
| PDP denies access | Return immediately | 403 Forbidden |
| CE unavailable | Retry with backoff | 503 Service Unavailable |
| Consent denied | Return immediately | 403 Forbidden |
| Consent timeout | Return immediately | 408 Request Timeout |
| Data provider error | Return error | 502 Bad Gateway |

### 4.2 Retry Logic

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

func isRetryableError(err error) bool {
    // Check for network errors, timeouts, 5xx status codes
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection refused") ||
           strings.Contains(err.Error(), "503") ||
           strings.Contains(err.Error(), "502")
}
```

## 5. Configuration

### 5.1 Service Endpoints

```go
type Config struct {
    PDPEndpoint string `json:"pdp_endpoint"` // http://localhost:8082
    CEEndpoint  string `json:"ce_endpoint"`  // http://localhost:8081
    Timeout     int    `json:"timeout"`      // seconds
    MaxRetries  int    `json:"max_retries"`
}
```

### 5.2 Environment Variables

```bash
PDP_ENDPOINT=http://localhost:8082
CE_ENDPOINT=http://localhost:8081
REQUEST_TIMEOUT=30
MAX_RETRIES=3
```

## 6. Monitoring and Logging

### 6.1 Key Metrics to Track

- PDP response times
- CE response times
- Consent approval rates
- Error rates by service
- Request throughput

### 6.2 Logging Format

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

log.Info("Consent status changed",
    "consent_id", consentID,
    "old_status", oldStatus,
    "new_status", newStatus)
```

## 7. Performance Optimization

### 7.1 Caching

- Cache PDP responses for similar requests
- Cache consent status to avoid repeated polling
- Implement cache invalidation strategies

### 7.2 Async Processing

- Use async patterns for consent polling
- Implement webhooks for consent status changes
- Queue long-running operations