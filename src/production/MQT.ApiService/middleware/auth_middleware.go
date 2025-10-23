package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"

	"github.com/gin-gonic/gin"
)

// Key types for request context
type contextKey string

const (
	// Context keys
	UserIDContextKey      contextKey = "user_id"
	UserRoleContextKey    contextKey = "user_role"
	TokenIDContextKey     contextKey = "token_id"
	AccessTokenContextKey contextKey = "access_token"
)

// AuthMiddleware provides middleware functions for authentication and authorization
type AuthMiddleware struct {
	jwtService  *jwt.Service
	rbacService *rbac.Service
	authorizer  *rbac.Authorizer
	config      Config
}

// Config holds middleware configuration
type Config struct {
	// HTTP header names for tokens
	AccessTokenHeader string

	// Cookie names for tokens (optional alternative to headers)
	AccessTokenCookie string
}

// DefaultConfig returns a default middleware configuration
func DefaultConfig() Config {
	return Config{
		AccessTokenHeader: "Authorization",
		AccessTokenCookie: "access_token",
	}
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *jwt.Service, rbacService *rbac.Service, config Config) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		rbacService: rbacService,
		authorizer:  rbac.NewAuthorizer(rbacService, jwtService),
		config:      config,
	}
}

// extractToken gets a token from either header or cookie
func extractToken(r *http.Request, headerName, cookieName string) string {
	// Try to get from header first
	token := r.Header.Get(headerName)
	if token != "" {
		// Handle Authorization: Bearer token format
		if strings.HasPrefix(token, "Bearer ") {
			return strings.TrimPrefix(token, "Bearer ")
		}
		return token
	}

	// Try to get from cookie if header is empty and cookie name is provided
	if cookieName != "" {
		cookie, err := r.Cookie(cookieName)
		if err == nil {
			return cookie.Value
		}
	}

	return ""
}

// Authenticate middleware verifies access token
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract access token
		accessToken := extractToken(c.Request, m.config.AccessTokenHeader, m.config.AccessTokenCookie)
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Validate access token
		accessClaims, err := m.jwtService.ValidateAccessToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			c.Abort()
			return
		}

		// Add user data to context
		c.Set(string(UserIDContextKey), accessClaims.UserID)
		c.Set(string(UserRoleContextKey), accessClaims.Role)
		c.Set(string(TokenIDContextKey), accessClaims.TokenID)
		c.Set(string(AccessTokenContextKey), accessToken)

		c.Next()
	}
}

// RequireAdmin ensures the user has admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract access token from context
		accessToken, ok := c.Get(string(AccessTokenContextKey))
		if !ok {
			// Try to extract from request if not in context
			accessToken = extractToken(c.Request, m.config.AccessTokenHeader, m.config.AccessTokenCookie)
			if accessToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}
		}

		// Check if user is admin
		err := m.authorizer.RequireAdmin(accessToken.(string))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnerOrAdmin ensures the user owns the resource or is admin
func (m *AuthMiddleware) RequireOwnerOrAdmin(resourceUserID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract access token from context
		accessToken, ok := c.Get(string(AccessTokenContextKey))
		if !ok {
			// Try to extract from request if not in context
			accessToken = extractToken(c.Request, m.config.AccessTokenHeader, m.config.AccessTokenCookie)
			if accessToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}
		}

		// Check if user owns resource or is admin
		err := m.authorizer.RequireOwnerOrAdmin(accessToken.(string), resourceUserID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole ensures the user has a specific role
func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get role from context first
		userRoleVal, exists := c.Get(string(UserRoleContextKey))

		// If not in context, extract from token
		if !exists {
			accessToken := extractToken(c.Request, m.config.AccessTokenHeader, m.config.AccessTokenCookie)
			if accessToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}

			// Validate access token to get role
			accessClaims, err := m.jwtService.ValidateAccessToken(accessToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
				c.Abort()
				return
			}

			userRoleVal = accessClaims.Role
		}

		userRole, ok := userRoleVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role format in context"})
			c.Abort()
			return
		}

		// Check if user has the required role
		if userRole != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserFromContext retrieves user ID from request context
func GetUserFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user not found in context")
	}
	return userID, nil
}

// GetRoleFromContext retrieves user role from request context
func GetRoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value(UserRoleContextKey).(string)
	if !ok || role == "" {
		return "", errors.New("role not found in context")
	}
	return role, nil
}

// GetUserFromGinContext retrieves user ID from Gin context
func GetUserFromGinContext(c *gin.Context) (string, error) {
	userIDVal, exists := c.Get(string(UserIDContextKey))
	if !exists {
		return "", errors.New("user not found in context")
	}

	userID, ok := userIDVal.(string)
	if !ok {
		return "", errors.New("invalid user ID format in context")
	}

	return userID, nil
}

// GetRoleFromGinContext retrieves user role from Gin context
func GetRoleFromGinContext(c *gin.Context) (string, error) {
	roleVal, exists := c.Get(string(UserRoleContextKey))
	if !exists {
		return "", errors.New("role not found in context")
	}

	role, ok := roleVal.(string)
	if !ok {
		return "", errors.New("invalid role format in context")
	}

	return role, nil
}
