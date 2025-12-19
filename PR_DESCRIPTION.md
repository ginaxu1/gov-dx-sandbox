## Summary

This PR addresses critical security and observability issues identified during code review:

1. **Security Regression Fix**: Resolves Nginx security header inheritance issue across all portal configurations
2. **Observability Improvements**: Fixes histogram bucket configuration and prevents metric cardinality explosion in the monitoring package

**Note**: This PR focuses on security and observability fixes. Audit service improvements and OE integration changes are handled in separate PRs.

## Key Changes

### 1. Security: Nginx Header Inheritance Fix

**Problem**: In Nginx, `add_header` directives from a parent block (like `server`) are not inherited by child blocks (like `location`) if the child block defines its own `add_header` directives. This caused a security regression where location blocks with custom headers (e.g., `Cache-Control`, `Content-Type`) were missing critical security headers (`X-Frame-Options`, `X-Content-Type-Options`, `Content-Security-Policy`).

**Solution**: Explicitly added security headers to all location blocks that define their own `add_header` directives in all three portals:

#### Files Modified:
- `portals/consent-portal/nginx.conf`
- `portals/member-portal/nginx.conf`
- `portals/admin-portal/nginx.conf`

#### Changes:
- Added security headers (`X-Frame-Options`, `X-Content-Type-Options`, `Content-Security-Policy`) to all location blocks:
  - `/config.js` location block
  - Hashed static assets location block (`^.+\\.[a-f0-9]{6,}\\.(jpg|jpeg|png|...)`)
  - Non-hashed static assets location block (`\\.(jpg|jpeg|png|...)`)
  - `/health` location block
- Added `always` flag to all `add_header` directives in `admin-portal/nginx.conf` for consistency
- Fixed Content-Security-Policy in `consent-portal` to include `connect-src 'self' https: http:`

**Security Impact**: All HTTP responses from portal services now include security headers, preventing vulnerabilities like clickjacking and MIME-type sniffing attacks.

### 2. Observability: Histogram Bucket Configuration Fix

**Problem**: Only `http_request_duration_seconds` had custom histogram buckets configured. The `external_call_duration_seconds` metric was using default OpenTelemetry buckets, leading to inconsistent metric configurations.

**Solution**: Added explicit histogram bucket configuration for `external_call_duration_seconds` to match `http_request_duration_seconds`.

#### File Modified:
- `exchange/shared/monitoring/otel_metrics.go`

#### Changes:
- Added `sdkmetric.WithView` configuration for `external_call_duration_seconds` metric
- Both duration metrics now use consistent custom buckets: `[.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10]` seconds

**Impact**: Consistent histogram bucket configuration across all duration metrics enables accurate performance analysis and alerting.

### 3. Observability: Metric Cardinality Explosion Prevention

**Problem**: The `looksLikeID` function in route normalization was too broad, incorrectly classifying static path segments like `data-owner` or `list-all` as dynamic IDs because they contained hyphens. This created excessive unique metric time series and could overload monitoring systems.

**Solution**: Improved ID detection logic to be more specific and prevent false positives.

#### File Modified:
- `exchange/shared/monitoring/metrics.go`

#### Changes:
- **UUID Detection**: Added specific check for UUID format (`len(s) == 36 && strings.Count(s, "-") == 4`)
- **Separator + Number Detection**: Updated logic to require both separators (`_` or `-`) AND numeric characters: `(strings.Contains(s, "_") || strings.Contains(s, "-")) && strings.ContainsAny(s, "0123456789")`
- **Prevents False Positives**: Static paths like `data-owner`, `list-all`, `check-status` are no longer incorrectly normalized
- **Correctly Detects**: Actual IDs like `consent_abc123`, `app-456`, UUIDs are still properly detected

**Impact**: Prevents metric cardinality explosion while maintaining accurate route normalization for actual dynamic IDs.

### 4. Testing: Comprehensive Test Coverage

Added comprehensive unit tests to verify all improvements:

#### File Modified:
- `exchange/shared/monitoring/metrics_test.go`

#### New Tests Added:
1. `TestLooksLikeIDImprovedLogic` - Verifies improved ID detection prevents false positives
2. `TestRouteNormalizationWithStaticPaths` - Ensures static paths with hyphens are not incorrectly normalized
3. `TestHistogramBucketsConfiguration` - Verifies both histogram metrics use custom buckets
4. `TestIsInitialized` - Tests initialization state functions
5. `TestMultipleInitializations` - Verifies thread-safety of multiple initialization calls
6. `TestHTTPMetricsMiddlewareWithDifferentStatusCodes` - Tests different HTTP status code recording
7. `TestNormalizeRouteWith404` - Ensures 404s are normalized to "unknown"

**Impact**: Ensures all improvements work correctly and prevents regressions.

## Files Changed

### Security Fixes
- `portals/consent-portal/nginx.conf`
- `portals/member-portal/nginx.conf`
- `portals/admin-portal/nginx.conf`

### Observability Fixes
- `exchange/shared/monitoring/otel_metrics.go`
- `exchange/shared/monitoring/metrics.go`
- `exchange/shared/monitoring/metrics_test.go`

### Documentation
- `exchange/shared/monitoring/TEST_SUMMARY.md` (new)
- `exchange/shared/monitoring/VERIFICATION.md` (new)

## Testing

- [x] **Unit Tests**: All new and existing tests pass
- [x] **Build Verification**: Code compiles successfully with no linter errors
- [x] **Security Verification**: All location blocks include security headers
- [x] **Observability Verification**: Histogram buckets configured correctly, ID detection prevents false positives
- [x] **Backward Compatibility**: API contract remains the same, only internal improvements

## Verification

### Security Headers Verification
All location blocks that define their own `add_header` directives now include:
- ✅ `X-Frame-Options: SAMEORIGIN`
- ✅ `X-Content-Type-Options: nosniff`
- ✅ `Content-Security-Policy: ...` (portal-specific)

All security headers use the `always` flag to ensure they're sent even on error responses.

### Observability Verification
- ✅ Both `http_request_duration_seconds` and `external_call_duration_seconds` use custom histogram buckets
- ✅ Static paths with hyphens (e.g., `data-owner`, `list-all`) are NOT incorrectly normalized
- ✅ Actual IDs (UUIDs, numeric IDs, IDs with separators + numbers) are correctly detected
- ✅ Route normalization prevents metric cardinality explosion

## Impact

### Security
- **Critical**: All portal responses now include security headers, preventing clickjacking and MIME-type sniffing attacks
- **Compliance**: Meets security best practices for web application headers

### Observability
- **Performance**: Consistent histogram buckets enable accurate performance analysis
- **Reliability**: Prevents metric cardinality explosion that could overload monitoring systems
- **Maintainability**: Comprehensive test coverage ensures improvements work correctly

## Related Issues

- Addresses security regression in Nginx configuration
- Fixes inconsistent histogram bucket configuration
- Prevents metric cardinality explosion in route normalization
