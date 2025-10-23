package controllers

import (
	"net/http"
	"time"

	service "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication requests
type AuthController struct {
	authService *service.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthController) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force user role for regular registration - no admin role allowed
	req.Role = "user"

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// RegisterAdmin handles admin user registration
func (h *AuthController) RegisterAdmin(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force admin role
	req.Role = "admin"

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

// Login handles user login
func (h *AuthController) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, tokenPair, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set refresh token as HTTP-only cookie
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(time.Unix(tokenPair.ExpiresAt, 0)).Seconds()),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HTTP only
	)

	// Return access token in the response body
	c.JSON(http.StatusOK, response)
}

// RefreshTokens handles token refresh
func (h *AuthController) RefreshTokens(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	response, tokenPair, err := h.authService.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new refresh token as HTTP-only cookie
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(time.Unix(tokenPair.ExpiresAt, 0)).Seconds()),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HTTP only
	)

	// Return new access token in the response body
	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthController) Logout(c *gin.Context) {
	// Clear the refresh token cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HTTP only
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Profile retrieves the authenticated user's profile
func (h *AuthController) Profile(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user profile
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile handles updating the authenticated user's profile
func (h *AuthController) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		// Hash the new password
		hashedPassword, err := h.authService.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}

	// Update user in database
	updatedUser, err := h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// RegisterRoutes registers the auth routes with Gin
func (h *AuthController) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Public routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshTokens)
		auth.POST("/logout", h.Logout)
	}

	// Protected routes
	protected := auth.Group("", authMiddleware.Authenticate())
	{
		protected.GET("/profile", h.Profile)
		protected.PATCH("/profile", h.UpdateProfile)
	}

	// Admin-only routes (requires authentication + admin role)
	adminOnly := auth.Group("", authMiddleware.Authenticate(), authMiddleware.RequireAdmin())
	{
		adminOnly.POST("/register/admin", h.RegisterAdmin)
	}
}
