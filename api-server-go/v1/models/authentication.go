package models

import (
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
func (u *AuthenticatedUser) GetPermissions() []Permission {
	permissionSet := make(map[Permission]bool)

	for _, role := range u.Roles {
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

// IsTokenExpired checks if the user's token is expired
func (u *AuthenticatedUser) IsTokenExpired() bool {
	return time.Now().After(u.ExpiresAt)
}

// NewAuthenticatedUser creates a new authenticated user from JWT claims
func NewAuthenticatedUser(claims *UserClaims) *AuthenticatedUser {
	// Convert string roles to Role type
	var roles []Role
	for _, roleStr := range claims.Roles.ToStringSlice() {
		role := Role(roleStr)
		if role.IsValid() {
			roles = append(roles, role)
		}
	}

	// If no valid roles found, default to member
	if len(roles) == 0 {
		roles = []Role{RoleMember}
		// TODO: If no roles are found, consider restricting access or logging a warning
	}

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
	}
}
