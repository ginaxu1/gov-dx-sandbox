## Summary
Implements comprehensive audit logging for Policy Decision Point (PDP) and Consent Engine (CE) services, and completes the full audit flow in Orchestration Engine with provider fetch events. This PR adds audit logging from the service perspective (PDP and CE) and implements `PROVIDER_FETCH_REQUEST`/`PROVIDER_FETCH_RESPONSE` events to replace the generic `DATA_EXCHANGE` event.

**Key Changes:**
- **PDP Audit Logging**: Added audit logging to Policy Decision Point that logs `POLICY_CHECK` events (one record per API call, status reflects result)
- **CE Audit Logging**: Added audit logging to Consent Engine that logs `CONSENT_CHECK` events (one record per API call, status reflects result)
- **Provider Fetch Events**: Implemented `PROVIDER_FETCH` event in Orchestration Engine (one record per API call)
- **Removed DATA_EXCHANGE**: Removed redundant `DATA_EXCHANGE` event type in favor of more specific provider fetch events
- **Removed REQUEST/RESPONSE Variants**: Removed `POLICY_CHECK_REQUEST`, `POLICY_CHECK_RESPONSE`, `CONSENT_CHECK_REQUEST`, `CONSENT_CHECK_RESPONSE` - only the receiving service logs one audit record per API call
- **Complete Traceability**: Full audit trail from OE → PDP/CE → Providers, with all events linked by trace ID
- **Reusable Audit Service**: Moved OpenDIF-specific event types to sample `enums.yaml` config, keeping `DefaultEnums` generic

## Example Audit Flow

A single OE request now generates a complete sequence of audit events linked by one `traceID`:

1. **`ORCHESTRATION_REQUEST_RECEIVED`** (traceID: `abc-123`)
   - Logged by Orchestration Engine when GraphQL request is received
   - Includes: application ID, GraphQL query
   - Status: `SUCCESS`

2. **`POLICY_CHECK`** (traceID: `abc-123`)
   - Logged by Policy Decision Point when responding to authorization request
   - One audit record per API call
   - Includes: application ID, required fields, authorization status, consent requirements, unauthorized/expired fields
   - Status: `SUCCESS` if API call succeeds, `FAILURE` if API call fails or unauthorized/expired
   - Target: `policy-decision-point`

3. **`CONSENT_CHECK`** (traceID: `abc-123`)
   - Logged by Consent Engine when responding to consent check request
   - One audit record per API call
   - Includes: application ID, owner ID/email, consent ID, consent status, consent portal URL, fields count
   - Status: `SUCCESS` if API call succeeds, `FAILURE` if API call fails
   - Target: `consent-engine`

4. **`PROVIDER_FETCH`** (traceID: `abc-123`, target: `provider-1`, `provider-2`, etc.)
   - Logged by Orchestration Engine after receiving response from each provider
   - One audit record per API call
   - Includes: application ID, schema ID, service key, requested fields, GraphQL query, response status, errors, data keys
   - Status: `SUCCESS` or `FAILURE` based on provider response
   - Target: Provider service key

**Example**: For a request that requires authorization and consent, and fetches data from 2 providers, the complete flow generates 5 audit events:
- 1 orchestration request received
- 1 policy check event (logged by PDP)
- 1 consent check event (logged by CE)
- 2 provider fetch events (one per provider)

All events can be retrieved using: `GET /api/audit-logs?traceId=abc-123`

## Changes

### Policy Decision Point (`exchange/policy-decision-point/`)

- **New Audit Middleware**: Created `v1/middleware/audit.go` with trace ID extraction and audit client integration
- **Handler Updates**: Added audit logging to `GetPolicyDecision` handler:
  - Logs `POLICY_CHECK` event when responding to authorization request (one record per API call)
  - Status reflects API call result: `SUCCESS` if call succeeds, `FAILURE` if call fails or unauthorized/expired
  - Includes: application ID, required fields, authorization status, consent requirements, unauthorized/expired fields
- **Main Integration**: Initialized audit middleware in `main.go` with graceful degradation support
- **Dependencies**: Added `audit-service` and `shared/audit` dependencies with local replace directives

### Consent Engine (`exchange/consent-engine/`)

- **New Audit Middleware**: Created `v1/middleware/audit.go` with trace ID extraction and audit client integration
- **Handler Updates**: Added audit logging to `GetConsent` handler:
  - Logs `CONSENT_CHECK` event when responding to consent check request (one record per API call)
  - Status reflects API call result: `SUCCESS` if call succeeds, `FAILURE` if call fails
  - Includes: application ID, owner ID/email, consent ID, consent status, consent portal URL, fields count
- **Main Integration**: Initialized audit middleware in `main.go` with graceful degradation support
- **Dependencies**: Added `audit-service` and `shared/audit` dependencies with local replace directives

### Orchestration Engine (`exchange/orchestration-engine/federator/`)

- **Provider Fetch Events**: Implemented `logProviderFetch()` function
  - `PROVIDER_FETCH`: Logged after receiving response from provider (one record per API call)
  - Includes comprehensive metadata: requested fields, query, response status, errors, data keys
  - Status reflects API call result: `SUCCESS` if call succeeds, `FAILURE` if call fails or has errors
- **Removed DATA_EXCHANGE**: Removed `logAuditEvent()` function and `DATA_EXCHANGE` event type
- **Removed Policy/Consent Logging**: Removed `logPolicyCheckRequest()`, `logPolicyCheckResponse()`, `logConsentCheckRequest()`, and `logConsentCheckResponse()` functions - these are now logged by PDP and CE respectively
- **Removed REQUEST/RESPONSE Variants**: Removed `logProviderFetchRequest()` and `logProviderFetchResponse()` functions - replaced with single `logProviderFetch()` function
- **Integration**: Added provider fetch audit logging to `performFederation()` function

### Audit Service Configuration (`audit-service/config/`)

- **Event Types**: Added OpenDIF-specific event types to `enums.yaml` (sample config):
  - `POLICY_CHECK` (logged by Policy Decision Point, one record per API call)
  - `CONSENT_CHECK` (logged by Consent Engine, one record per API call)
  - `PROVIDER_FETCH` (logged by Orchestration Engine, one record per API call)
- **Removed Event Types**: Removed `POLICY_CHECK_REQUEST`, `POLICY_CHECK_RESPONSE`, `CONSENT_CHECK_REQUEST`, `CONSENT_CHECK_RESPONSE`, `PROVIDER_FETCH_REQUEST`, `PROVIDER_FETCH_RESPONSE`, and `DATA_EXCHANGE` from configuration files
- **Reusable Defaults**: Removed OpenDIF-specific event types from `DefaultEnums` in `config.go`, keeping only generic types (`MANAGEMENT_EVENT`, `USER_MANAGEMENT`, `DATA_FETCH`)

## Event Types

The following event types are now fully implemented:

| Event Type | Logged By | Description | Status Values | Target Service |
|------------|-----------|-------------|---------------|----------------|
| `ORCHESTRATION_REQUEST_RECEIVED` | Orchestration Engine | GraphQL request received | `SUCCESS` | - |
| `POLICY_CHECK` | Policy Decision Point | Authorization decision result (one record per API call) | `SUCCESS`, `FAILURE` | `policy-decision-point` |
| `CONSENT_CHECK` | Consent Engine | Consent check result (one record per API call) | `SUCCESS`, `FAILURE` | `consent-engine` |
| `PROVIDER_FETCH` | Orchestration Engine | Provider fetch result (one record per API call) | `SUCCESS`, `FAILURE` | Provider service key |

**Key Features:**
- **Single Record Per API Call**: Each API call generates exactly one audit record, logged by the receiving service (PDP or CE)
- **Status Reflects API Result**: Status is set to `SUCCESS` if the API call succeeds, `FAILURE` if the API call fails or returns an error/unauthorized result
- **Trace ID Propagation**: All events use the same `traceID` extracted from `X-Trace-ID` HTTP header
- **Comprehensive Metadata**: Each event includes relevant context (application ID, fields, status, errors)
- **Error Handling**: Failed operations are logged with `FAILURE` status and error details

## Configuration

### Environment Variables

Both PDP and CE support the same audit configuration:

- **`CHOREO_AUDIT_CONNECTION_SERVICEURL`**: Audit service base URL (e.g., `http://localhost:3000`)
  - If not set or empty, audit logging is disabled
  - Default: empty (audit disabled)
- **`ENABLE_AUDIT`**: Explicitly enable/disable audit logging
  - Values: `true`, `1`, `yes` (case-insensitive) to enable
  - Default: enabled if `CHOREO_AUDIT_CONNECTION_SERVICEURL` is set

### Graceful Degradation

- If audit service URL is not configured, audit logging is skipped (no errors)
- Services continue to function normally without audit service
- All audit operations are asynchronous (fire-and-forget) to avoid blocking request flow
- Trace ID extraction works even when audit is disabled (for header propagation)

## Dependencies

### New Dependencies (PDP and CE)

- `github.com/gov-dx-sandbox/audit-service` - Audit service v1 models
- `github.com/gov-dx-sandbox/shared/audit` - Shared audit client package

### Local Replace Directives

All services use local replace directives for development:
- `replace github.com/gov-dx-sandbox/audit-service => ../../audit-service`
- `replace github.com/gov-dx-sandbox/shared/audit => ../../shared/audit`

### Version Simplification

- Simplified `go.mod` version strings from pseudo-versions (`v0.0.0-00010101000000-000000000000`) to `v0.0.0` for local dependencies

## Testing

- ✅ **Build Verification**: All services (PDP, CE, OE) build successfully
- ✅ **Test Execution**: All existing tests pass
- ✅ **Integration Ready**: Audit logging integrated at all interaction points
- ✅ **Trace ID Propagation**: Trace ID properly extracted from HTTP headers and propagated via context
- ✅ **Error Handling**: Failed operations properly logged with failure status

## Migration Notes

### Breaking Changes
- **Removed `DATA_EXCHANGE` event type**: Replaced with `PROVIDER_FETCH`
  - Any queries filtering by `eventType=DATA_EXCHANGE` should be updated to use `PROVIDER_FETCH`
  - Historical `DATA_EXCHANGE` events remain in the database but new events will use the new type
- **Removed REQUEST/RESPONSE event types**: Removed `POLICY_CHECK_REQUEST`, `POLICY_CHECK_RESPONSE`, `CONSENT_CHECK_REQUEST`, `CONSENT_CHECK_RESPONSE`, `PROVIDER_FETCH_REQUEST`, `PROVIDER_FETCH_RESPONSE`
  - Only `POLICY_CHECK` (logged by PDP), `CONSENT_CHECK` (logged by CE), and `PROVIDER_FETCH` (logged by OE) are used now
  - One audit record per API call, logged by the receiving service (or OE for provider fetches)
  - Historical REQUEST/RESPONSE events remain in the database but new events will use the simplified types

### Upgrade Path
1. Deploy updated audit-service with new event types in `enums.yaml`:
   - `POLICY_CHECK` (one record per API call, logged by PDP)
   - `CONSENT_CHECK` (one record per API call, logged by CE)
   - `PROVIDER_FETCH` (one record per API call, logged by OE)
2. Deploy updated Orchestration Engine (removed policy/consent logging, kept provider fetch logging)
3. Deploy updated Policy Decision Point and Consent Engine with audit logging
4. Set `CHOREO_AUDIT_CONNECTION_SERVICEURL` environment variable in all services
5. Verify audit events are being logged correctly using `GET /api/audit-logs?traceId=<trace-id>`

## Related PRs
- #384 - Refactor audit-service to be reusable, general with trace_id
- #XXX - Add shared audit client and integrate with OE and Portal Backend (340-audit-update-oe)
