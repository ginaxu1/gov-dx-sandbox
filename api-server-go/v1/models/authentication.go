package models

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents the JWT claims for a user
type UserClaims struct {
	Email       string              `json:"email"`
	FirstName   string              `json:"given_name"`
	LastName    string              `json:"family_name"`
	PhoneNumber string              `json:"phone_number"`
	Roles       FlexibleStringSlice `json:"roles"`
	Groups      []string            `json:"groups"`
	OrgName     string              `json:"org_name"`
	IdpUserID   string              `json:"sub"` // Subject is typically the user ID from IdP
	// Standard JWT claims
	Issuer    string    `json:"iss"`
	Audience  []string  `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf"`
}

// GetExpirationTime implements jwt.Claims interface
func (c *UserClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.ExpiresAt.IsZero() {
		return nil, nil
	}
	return jwt.NewNumericDate(c.ExpiresAt), nil
}

// GetIssuedAt implements jwt.Claims interface
func (c *UserClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.IssuedAt.IsZero() {
		return nil, nil
	}
	return jwt.NewNumericDate(c.IssuedAt), nil
}

// GetNotBefore implements jwt.Claims interface
func (c *UserClaims) GetNotBefore() (*jwt.NumericDate, error) {
	if c.NotBefore.IsZero() {
		return nil, nil
	}
	return jwt.NewNumericDate(c.NotBefore), nil
}

// GetIssuer implements jwt.Claims interface
func (c *UserClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

// GetSubject implements jwt.Claims interface
func (c *UserClaims) GetSubject() (string, error) {
	return c.IdpUserID, nil
}

// GetAudience implements jwt.Claims interface
func (c *UserClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(c.Audience), nil
}

// AuthenticatedUser represents the authenticated user context
type AuthenticatedUser struct {
	IdpUserID   string    `json:"idpUserId"`
	Email       string    `json:"email"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	PhoneNumber string    `json:"phoneNumber"`
	Roles       []Role    `json:"roles"`
	Groups      []string  `json:"groups"`
	OrgName     string    `json:"orgName"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`

	// Cached permissions - computed once during user creation for performance
	permissions []Permission `json:"-"` // Don't expose in JSON, use GetPermissions() method

	// Cached member ID - populated on first access to avoid repeated database queries
	memberID      string `json:"-"` // Don't expose in JSON
	memberIDError error  `json:"-"` // Cache the error state as well
}

// AuthContext represents the authentication context in HTTP requests
type AuthContext struct {
	User        *AuthenticatedUser `json:"user"`
	Token       string             `json:"-"` // Don't expose in JSON
	IssuedBy    string             `json:"issuedBy"`
	Audience    []string           `json:"audience"`
	Permissions []Permission       `json:"permissions"`
}

// HasRole checks if the user has a specific role
func (u *AuthenticatedUser) HasRole(role Role) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *AuthenticatedUser) HasAnyRole(roles ...Role) bool {
	for _, requiredRole := range roles {
		for _, userRole := range u.Roles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission based on their roles
func (u *AuthenticatedUser) HasPermission(permission Permission) bool {
	for _, role := range u.Roles {
		if role.HasPermission(permission) {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin role
func (u *AuthenticatedUser) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// IsMember checks if the user has member role
func (u *AuthenticatedUser) IsMember() bool {
	return u.HasRole(RoleMember)
}

// IsSystem checks if the user has system role
func (u *AuthenticatedUser) IsSystem() bool {
	return u.HasRole(RoleSystem)
}

// GetPrimaryRole returns the highest priority role (Admin > System > Member)
func (u *AuthenticatedUser) GetPrimaryRole() Role {
	if u.HasRole(RoleAdmin) {
		return RoleAdmin
	}
	if u.HasRole(RoleSystem) {
		return RoleSystem
	}
	if u.HasRole(RoleMember) {
		return RoleMember
	}
	return RoleMember // Default to member if no roles found
}

// GetPermissions returns all permissions the user has based on their roles
// Uses cached permissions computed during user creation for optimal performance
func (u *AuthenticatedUser) GetPermissions() []Permission {
	// Return a copy of the cached permissions to prevent external modification
	result := make([]Permission, len(u.permissions))
	copy(result, u.permissions)
	return result
}

// IsTokenExpired checks if the user's token is expired
func (u *AuthenticatedUser) IsTokenExpired() bool {
	return time.Now().After(u.ExpiresAt)
}

// GetCachedMemberID returns the cached member ID if available
func (u *AuthenticatedUser) GetCachedMemberID() (string, bool) {
	return u.memberID, u.memberID != ""
}

// SetCachedMemberID sets the cached member ID and error state
func (u *AuthenticatedUser) SetCachedMemberID(memberID string, err error) {
	u.memberID = memberID
	u.memberIDError = err
}

// GetCachedMemberIDError returns the cached error from member ID lookup
func (u *AuthenticatedUser) GetCachedMemberIDError() error {
	return u.memberIDError
}

// computePermissions calculates all permissions for the given roles
func computePermissions(roles []Role) []Permission {
	permissionSet := make(map[Permission]bool)

	for _, role := range roles {
		if permissions, exists := RolePermissions[role]; exists {
			for _, permission := range permissions {
				permissionSet[permission] = true
			}
		}
	}

	var permissions []Permission
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions
}

// NewAuthenticatedUser creates a new authenticated user from JWT claims
// Returns an error if no valid roles are found in the claims
func NewAuthenticatedUser(claims *UserClaims) (*AuthenticatedUser, error) {
	// Convert string roles to Role type
	var roles []Role
	for _, roleStr := range claims.Roles.ToStringSlice() {
		role := Role(roleStr)
		if role.IsValid() {
			roles = append(roles, role)
		}
	}

	// If no valid roles found, deny access for security
	if len(roles) == 0 {
		return nil, fmt.Errorf("access denied: no valid roles found in JWT claims for user %s", claims.IdpUserID)
	}

	// Compute permissions once during user creation for optimal performance
	permissions := computePermissions(roles)

	return &AuthenticatedUser{
		IdpUserID:   claims.IdpUserID,
		Email:       claims.Email,
		FirstName:   claims.FirstName,
		LastName:    claims.LastName,
		PhoneNumber: claims.PhoneNumber,
		Roles:       roles,
		Groups:      claims.Groups,
		OrgName:     claims.OrgName,
		IssuedAt:    claims.IssuedAt,
		ExpiresAt:   claims.ExpiresAt,
		permissions: permissions,
	}, nil
}
