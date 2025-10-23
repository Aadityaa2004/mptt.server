package rbac

import (
	"errors"

	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
)

// Authorizer provides authorization operations
type Authorizer struct {
	rbacService *Service
	jwtService  *jwt.Service
}

// NewAuthorizer creates a new authorizer
func NewAuthorizer(rbacService *Service, jwtService *jwt.Service) *Authorizer {
	return &Authorizer{
		rbacService: rbacService,
		jwtService:  jwtService,
	}
}

// AuthorizeWithToken checks role using access token
func (a *Authorizer) AuthorizeWithToken(accessToken string, requiredRole string) error {
	// Validate access token
	accessClaims, err := a.jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		return err
	}

	// Check if the user has the required role
	if accessClaims.Role != requiredRole {
		return errors.New("unauthorized: insufficient role")
	}

	return nil
}

// RequireAdmin checks if user is admin
func (a *Authorizer) RequireAdmin(accessToken string) error {
	return a.AuthorizeWithToken(accessToken, "admin")
}

// IsOwner checks if user owns the resource
func (a *Authorizer) IsOwner(userID, resourceUserID string) bool {
	return userID == resourceUserID
}

// RequireOwnerOrAdmin checks if user owns resource or is admin
func (a *Authorizer) RequireOwnerOrAdmin(accessToken, resourceUserID string) error {
	// Validate access token
	accessClaims, err := a.jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		return err
	}

	// Admin can access anything
	if accessClaims.Role == "admin" {
		return nil
	}

	// User can only access their own resources
	if accessClaims.UserID == resourceUserID {
		return nil
	}

	return errors.New("unauthorized: insufficient permissions")
}
