package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/utils"
)

// TestUser represents different test user personas
type TestUser struct {
	IdpUserID string
	Email     string
	Roles     []models.Role
	User      *models.AuthenticatedUser
}

// Predefined test users with different roles
var (
	// AdminUser has full admin privileges
	AdminUser = TestUser{
		IdpUserID: "admin-test-user-123",
		Email:     "admin@test.com",
		Roles:     []models.Role{models.RoleAdmin},
	}

	// MemberUser has standard member privileges
	MemberUser = TestUser{
		IdpUserID: "member-test-user-456",
		Email:     "member@test.com",
		Roles:     []models.Role{models.RoleMember},
	}

	// SystemUser has system-level read access
	SystemUser = TestUser{
		IdpUserID: "system-test-user-789",
		Email:     "system@test.com",
		Roles:     []models.Role{models.RoleSystem},
	}

	// UnauthorizedUser has no roles (should fail most operations)
	UnauthorizedUser = TestUser{
		IdpUserID: "unauthorized-test-user-000",
		Email:     "unauthorized@test.com",
		Roles:     []models.Role{},
	}
)

// init initializes the test users with their AuthenticatedUser instances
func init() {
	AdminUser.User = createTestUser(AdminUser.IdpUserID, AdminUser.Email, AdminUser.Roles)
	MemberUser.User = createTestUser(MemberUser.IdpUserID, MemberUser.Email, MemberUser.Roles)
	SystemUser.User = createTestUser(SystemUser.IdpUserID, SystemUser.Email, SystemUser.Roles)
	UnauthorizedUser.User = createTestUser(UnauthorizedUser.IdpUserID, UnauthorizedUser.Email, UnauthorizedUser.Roles)
}

// createTestUser creates an AuthenticatedUser from test data
func createTestUser(idpUserID, email string, roles []models.Role) *models.AuthenticatedUser {
	// Convert roles to string slice for UserClaims
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}

	claims := &models.UserClaims{
		IdpUserID: idpUserID,
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Roles:     models.FlexibleStringSlice(roleStrings),
	}

	return models.NewAuthenticatedUser(claims)
}

// WithAuth creates a new HTTP request with the specified user authentication context
func WithAuth(req *http.Request, testUser TestUser) *http.Request {
	ctx := utils.SetAuthenticatedUser(req.Context(), testUser.User)
	return req.WithContext(ctx)
}

// WithAdminAuth is a convenience function to add admin authentication to a request
func WithAdminAuth(req *http.Request) *http.Request {
	return WithAuth(req, AdminUser)
}

// WithMemberAuth is a convenience function to add member authentication to a request
func WithMemberAuth(req *http.Request) *http.Request {
	return WithAuth(req, MemberUser)
}

// WithSystemAuth is a convenience function to add system authentication to a request
func WithSystemAuth(req *http.Request) *http.Request {
	return WithAuth(req, SystemUser)
}

// WithUnauthorizedAuth is a convenience function to add unauthorized user context to a request
func WithUnauthorizedAuth(req *http.Request) *http.Request {
	return WithAuth(req, UnauthorizedUser)
}

// NewAuthenticatedRequest creates a new HTTP request with authentication context
func NewAuthenticatedRequest(method, url string, body io.Reader, testUser TestUser) *http.Request {
	req := httptest.NewRequest(method, url, body)
	return WithAuth(req, testUser)
}

// NewAdminRequest creates a new HTTP request with admin authentication
func NewAdminRequest(method, url string, body io.Reader) *http.Request {
	return NewAuthenticatedRequest(method, url, body, AdminUser)
}

// NewMemberRequest creates a new HTTP request with member authentication
func NewMemberRequest(method, url string, body io.Reader) *http.Request {
	return NewAuthenticatedRequest(method, url, body, MemberUser)
}

// NewSystemRequest creates a new HTTP request with system authentication
func NewSystemRequest(method, url string, body io.Reader) *http.Request {
	return NewAuthenticatedRequest(method, url, body, SystemUser)
}

// NewUnauthenticatedRequest creates a new HTTP request without authentication (should fail)
func NewUnauthenticatedRequest(method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	// Intentionally don't add authentication context
	return req
}

// CreateCustomTestUser creates a test user with custom roles and permissions
func CreateCustomTestUser(idpUserID, email string, roles []models.Role) TestUser {
	user := createTestUser(idpUserID, email, roles)
	return TestUser{
		IdpUserID: idpUserID,
		Email:     email,
		Roles:     roles,
		User:      user,
	}
}
