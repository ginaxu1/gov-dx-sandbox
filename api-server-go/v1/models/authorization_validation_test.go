package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAllPermissions verifies that GetAllPermissions returns all defined permissions
func TestGetAllPermissions(t *testing.T) {
	permissions := GetAllPermissions()
	
	// Verify we have the expected number of permissions
	// Schema: 5, Schema Submission: 6, Application: 5, Application Submission: 6, Member: 5 = 27 total
	expectedCount := 27
	assert.Equal(t, expectedCount, len(permissions), "Should have exactly %d permissions", expectedCount)

	// Verify all permission constants are included
	allExpectedPermissions := []Permission{
		// Schema permissions
		PermissionCreateSchema,
		PermissionReadSchema,
		PermissionUpdateSchema,
		PermissionDeleteSchema,
		PermissionReadAllSchemas,
		// Schema submission permissions
		PermissionCreateSchemaSubmission,
		PermissionReadSchemaSubmission,
		PermissionUpdateSchemaSubmission,
		PermissionDeleteSchemaSubmission,
		PermissionReadAllSchemaSubmissions,
		PermissionApproveSchemaSubmission,
		// Application permissions
		PermissionCreateApplication,
		PermissionReadApplication,
		PermissionUpdateApplication,
		PermissionDeleteApplication,
		PermissionReadAllApplications,
		// Application submission permissions
		PermissionCreateApplicationSubmission,
		PermissionReadApplicationSubmission,
		PermissionUpdateApplicationSubmission,
		PermissionDeleteApplicationSubmission,
		PermissionReadAllApplicationSubmissions,
		PermissionApproveApplicationSubmission,
		// Member permissions
		PermissionCreateMember,
		PermissionReadMember,
		PermissionUpdateMember,
		PermissionDeleteMember,
		PermissionReadAllMembers,
	}

	// Build a map of returned permissions for quick lookup
	permissionMap := make(map[Permission]bool)
	for _, perm := range permissions {
		permissionMap[perm] = true
	}

	// Verify all expected permissions are present
	for _, expectedPerm := range allExpectedPermissions {
		assert.True(t, permissionMap[expectedPerm], "Permission %s should be in GetAllPermissions()", expectedPerm)
	}
}

// TestValidateRolePermissions ensures all permissions assigned to roles are valid
func TestValidateRolePermissions(t *testing.T) {
	errors := ValidateRolePermissions()
	require.Empty(t, errors, "All role permissions should be valid. Errors: %v", errors)
}

// TestIsValidPermission verifies that IsValidPermission correctly identifies valid and invalid permissions
func TestIsValidPermission(t *testing.T) {
	// Test valid permissions
	validPermissions := GetAllPermissions()
	for _, perm := range validPermissions {
		assert.True(t, IsValidPermission(perm), "Permission %s should be valid", perm)
	}

	// Test invalid permissions
	invalidPermissions := []Permission{
		Permission("invalid:permission"),
		Permission("schema:invalid"),
		Permission(""),
		Permission("unknown:action"),
	}

	for _, invalidPerm := range invalidPermissions {
		assert.False(t, IsValidPermission(invalidPerm), "Permission %s should be invalid", invalidPerm)
	}
}

// TestGetPermissionsByResourceType verifies that permissions are correctly grouped by resource type
func TestGetPermissionsByResourceType(t *testing.T) {
	permissionsByType := GetPermissionsByResourceType()

	// Verify all expected resource types are present
	expectedTypes := []string{"schema", "schema_submission", "application", "application_submission", "member"}
	for _, resourceType := range expectedTypes {
		assert.Contains(t, permissionsByType, resourceType, "Resource type %s should be present", resourceType)
	}

	// Verify schema permissions
	schemaPerms := permissionsByType["schema"]
	assert.Contains(t, schemaPerms, PermissionCreateSchema)
	assert.Contains(t, schemaPerms, PermissionReadSchema)
	assert.Contains(t, schemaPerms, PermissionUpdateSchema)
	assert.Contains(t, schemaPerms, PermissionDeleteSchema)
	assert.Contains(t, schemaPerms, PermissionReadAllSchemas)
	assert.Equal(t, 5, len(schemaPerms), "Schema should have 5 permissions")

	// Verify schema_submission permissions
	submissionPerms := permissionsByType["schema_submission"]
	assert.Contains(t, submissionPerms, PermissionCreateSchemaSubmission)
	assert.Contains(t, submissionPerms, PermissionApproveSchemaSubmission)
	assert.Equal(t, 6, len(submissionPerms), "Schema submission should have 6 permissions")

	// Verify application permissions
	appPerms := permissionsByType["application"]
	assert.Contains(t, appPerms, PermissionCreateApplication)
	assert.Contains(t, appPerms, PermissionReadApplication)
	assert.Equal(t, 5, len(appPerms), "Application should have 5 permissions")

	// Verify member permissions
	memberPerms := permissionsByType["member"]
	assert.Contains(t, memberPerms, PermissionCreateMember)
	assert.Contains(t, memberPerms, PermissionReadMember)
	assert.Equal(t, 5, len(memberPerms), "Member should have 5 permissions")
}

// TestGetPermissionsByRole verifies that GetPermissionsByRole returns correct permissions for each role
func TestGetPermissionsByRole(t *testing.T) {
	// Test Admin role - should have all permissions
	adminPerms := GetPermissionsByRole(RoleAdmin)
	assert.Greater(t, len(adminPerms), 20, "Admin should have many permissions")

	// Test Member role - should have limited permissions
	memberPerms := GetPermissionsByRole(RoleMember)
	assert.Greater(t, len(memberPerms), 0, "Member should have some permissions")
	assert.Less(t, len(memberPerms), len(adminPerms), "Member should have fewer permissions than Admin")

	// Test System role - should have read permissions
	systemPerms := GetPermissionsByRole(RoleSystem)
	assert.Greater(t, len(systemPerms), 0, "System should have some permissions")

	// Test invalid role
	invalidPerms := GetPermissionsByRole(Role("invalid_role"))
	assert.Empty(t, invalidPerms, "Invalid role should return empty permissions")
}

// TestPermissionEnumerationCompleteness ensures that all permissions defined as constants
// are included in GetAllPermissions() and that no permissions are missing
func TestPermissionEnumerationCompleteness(t *testing.T) {
	// Get all permissions from the enumeration function
	enumeratedPerms := GetAllPermissions()
	enumeratedMap := make(map[Permission]bool)
	for _, perm := range enumeratedPerms {
		enumeratedMap[perm] = true
	}

	// Manually list all permission constants to ensure completeness
	allPermissionConstants := []Permission{
		PermissionCreateSchema,
		PermissionReadSchema,
		PermissionUpdateSchema,
		PermissionDeleteSchema,
		PermissionReadAllSchemas,
		PermissionCreateSchemaSubmission,
		PermissionReadSchemaSubmission,
		PermissionUpdateSchemaSubmission,
		PermissionDeleteSchemaSubmission,
		PermissionReadAllSchemaSubmissions,
		PermissionApproveSchemaSubmission,
		PermissionCreateApplication,
		PermissionReadApplication,
		PermissionUpdateApplication,
		PermissionDeleteApplication,
		PermissionReadAllApplications,
		PermissionCreateApplicationSubmission,
		PermissionReadApplicationSubmission,
		PermissionUpdateApplicationSubmission,
		PermissionDeleteApplicationSubmission,
		PermissionReadAllApplicationSubmissions,
		PermissionApproveApplicationSubmission,
		PermissionCreateMember,
		PermissionReadMember,
		PermissionUpdateMember,
		PermissionDeleteMember,
		PermissionReadAllMembers,
	}

	// Verify every constant is in the enumeration
	for _, perm := range allPermissionConstants {
		assert.True(t, enumeratedMap[perm], "Permission constant %s must be included in GetAllPermissions()", perm)
	}

	// Verify no extra permissions in enumeration (should match exactly)
	assert.Equal(t, len(allPermissionConstants), len(enumeratedPerms),
		"GetAllPermissions() should return exactly the same number of permissions as defined constants")
}

