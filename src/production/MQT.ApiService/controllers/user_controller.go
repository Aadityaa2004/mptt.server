package controllers

import (
	"net/http"

	service "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"

	"github.com/gin-gonic/gin"
)

// UserController handles user management requests
type UserController struct {
	userService *service.UserService
}

// NewUserController creates a new user controller
func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// RegisterRoutes registers the user routes with Gin
func (h *UserController) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes
	users := router.Group("/api/users", authMiddleware.Authenticate())
	{
		// Get all users - requires admin role
		users.GET("",
			authMiddleware.RequireAdmin(),
			h.GetAllUsers)

		// Get user by ID - requires admin role or own user
		users.GET("/:id",
			h.GetUserByID)

		// Update user - requires admin role
		users.PUT("/:id",
			authMiddleware.RequireAdmin(),
			h.UpdateUser)

		// Delete user - requires admin role
		users.DELETE("/:id",
			authMiddleware.RequireAdmin(),
			h.DeleteUser)

		// Update user role - requires admin role
		users.PUT("/:id/role",
			authMiddleware.RequireAdmin(),
			h.UpdateUserRole)
	}
}

// GetAllUsers retrieves all users
func (h *UserController) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserByID retrieves a user by ID
func (h *UserController) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Check ownership if not admin
	userRole, err := middleware.GetRoleFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user role"})
		return
	}

	if userRole != "admin" {
		currentUserID, err := middleware.GetUserFromGinContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current user"})
			return
		}

		if user.UserID != currentUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser updates a user
func (h *UserController) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing user
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
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
		hashedPassword, err := h.userService.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}

	// Update user in database
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser deletes a user (hard delete)
func (h *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Check if user exists
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Delete user
	err = h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// UpdateUserRole updates a user's role
func (h *UserController) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUserRole(c.Request.Context(), userID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
