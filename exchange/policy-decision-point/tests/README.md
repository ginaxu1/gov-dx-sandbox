# Tests Directory

This directory contains test files that were moved during the V1 refactoring cleanup of the policy-decision-point service.

## Status of Test Files

### 🔧 Needs Updates - Legacy Tests
The following test files are from the **legacy architecture** (before V1 refactoring) and need to be updated to work with the new V1 architecture:

#### `endpoint_test.go`
- **Issue**: Tests legacy API endpoints that no longer exist
- **Legacy endpoints**: `/policy-metadata`, `/allow-list`  
- **New V1 endpoints**: `/api/v1/policy/metadata`, `/api/v1/policy/update-allowlist`, `/api/v1/policy/decide`
- **Required changes**: 
  - Update endpoint URLs
  - Use V1 models and DTOs
  - Update test assertions for new response formats

#### `policy_consent_test.go` 
- **Issue**: Tests the legacy PolicyEvaluator with OPA (Open Policy Agent)
- **Legacy architecture**: Used OPA + Rego policies + legacy database service
- **New V1 architecture**: Uses GORM + policy metadata service
- **Required changes**:
  - Replace PolicyEvaluator tests with PolicyMetadataService tests
  - Update models to use V1 GORM models
  - Rewrite test logic for new policy decision algorithm

#### `policy_test.go`
- **Issue**: Tests legacy policy evaluation logic
- **Required changes**: Similar to policy_consent_test.go

#### `test_utils.go`
- **Issue**: Contains utility functions for legacy architecture
- **Required changes**: Update to support V1 architecture

### 🔄 Migration Guide

To update these tests for V1 architecture:

1. **Update imports**:
   ```go
   // Old
   import "github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
   
   // New
   import "github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
   ```

2. **Update API endpoints**:
   ```go
   // Old
   req, _ := http.NewRequest("POST", "/policy-metadata", body)
   
   // New  
   req, _ := http.NewRequest("POST", "/api/v1/policy/metadata", body)
   ```

3. **Update models**:
   ```go
   // Old models.PolicyDecisionRequest
   req := models.PolicyDecisionRequest{
       AppID:          "test-app",
       RequestID:      "req-1", 
       RequiredFields: []string{"person.fullName"},
   }
   
   // New V1 models.PolicyDecisionRequest
   req := models.PolicyDecisionRequest{
       ApplicationID:  "test-app",
       RequiredFields: []models.PolicyDecisionRequestRecord{
           {FieldName: "person.fullName", SchemaID: "schema1"},
       },
   }
   ```

4. **Replace PolicyEvaluator with PolicyMetadataService**:
   ```go
   // Old
   evaluator, _ := NewPolicyEvaluator(ctx)
   decision, _ := evaluator.Authorize(ctx, input)
   
   // New
   service := services.NewPolicyMetadataService(gormDB)
   response, _ := service.GetPolicyDecision(&request)
   ```

### 🧪 Shell Test Scripts

#### `test_metadata_fix.sh` and `test_system_provider.sh`
- **Status**: ✅ Should work as-is (they test external endpoints)
- **Note**: May need URL updates if testing V1 endpoints

### 🏃‍♂️ Running Tests

Currently, the Go tests will **not compile** due to the architectural changes. To run tests:

1. **Option 1**: Update the tests following the migration guide above
2. **Option 2**: Create new V1-specific tests from scratch
3. **Option 3**: Run shell tests only: `./test_metadata_fix.sh`

### 📝 Recommended Next Steps

1. **Create new V1 integration tests** that test the actual V1 endpoints
2. **Create V1 unit tests** for the PolicyMetadataService
3. **Update or replace the legacy tests** with V1-compatible versions
4. **Add database migration tests** to ensure schema changes work correctly

### 🔍 V1 Architecture Overview

The new V1 architecture uses:
- **GORM** instead of sql.DB
- **PolicyMetadataService** instead of PolicyEvaluator  
- **Database-driven metadata** instead of OPA policies
- **REST API with /api/v1/policy/** prefix
- **Structured DTOs** with validation

For examples of V1 architecture, see:
- `v1/handler.go` - V1 HTTP handlers
- `v1/services/policy_metadata_service.go` - V1 business logic
- `v1/models/` - V1 data models and DTOs