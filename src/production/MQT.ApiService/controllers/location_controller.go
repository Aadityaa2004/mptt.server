package controllers

import (
	"log"
	"net/http"

	service "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"

	"github.com/gin-gonic/gin"
)

// LocationController handles device location management requests
type LocationController struct {
	authService *service.AuthService
	deviceRepo  interfaces.DeviceRepository
}

// NewLocationController creates a new location controller
func NewLocationController(authService *service.AuthService, deviceRepo interfaces.DeviceRepository) *LocationController {
	return &LocationController{
		authService: authService,
		deviceRepo:  deviceRepo,
	}
}

// RegisterRoutes registers the location routes with Gin
func (h *LocationController) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes - users can only manage their own locations
	protected := router.Group("/api/locations", authMiddleware.Authenticate())
	{
		// Get all locations for the authenticated user
		protected.GET("", h.GetLocations)

		// Add a new location
		protected.POST("", h.AddLocation)

		// Update a location by device_id
		protected.PUT("/:device_id", h.UpdateLocation)

		// Delete a location by device_id
		protected.DELETE("/:device_id", h.DeleteLocation)
	}
}

// GetLocations retrieves all locations for the authenticated user
func (h *LocationController) GetLocations(c *gin.Context) {
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

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"locations": user.Locations})
}

// AddLocation adds a new location to the authenticated user's locations array
func (h *LocationController) AddLocation(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var location auth_models.DeviceLocation
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Check if device_id already exists
	for i, loc := range user.Locations {
		if loc.DeviceID == location.DeviceID {
			c.JSON(http.StatusConflict, gin.H{"error": "device location already exists. Use PUT to update instead."})
			return
		}
		// Also check if we're updating an existing one (in case device_id changed but we want to update by index)
		_ = i // Suppress unused variable warning
	}

	// Add new location to the array
	user.Locations = append(user.Locations, location)

	// Update user in database
	updatedUser, err := h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sync bucket dimensions to the devices table
	if location.Height > 0 || location.TopDiameter > 0 || location.BottomDiameter > 0 {
		h.syncDeviceDimensions(c, location)
	}

	// Find and return the newly added location
	var addedLocation *auth_models.DeviceLocation
	for i := range updatedUser.Locations {
		if updatedUser.Locations[i].DeviceID == location.DeviceID {
			addedLocation = &updatedUser.Locations[i]
			break
		}
	}

	if addedLocation == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve added location"})
		return
	}

	c.JSON(http.StatusCreated, addedLocation)
}

// UpdateLocation updates an existing location by device_id
func (h *LocationController) UpdateLocation(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	var location auth_models.DeviceLocation
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the device_id in the body matches the path parameter
	location.DeviceID = deviceID

	// Get current user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Find and update the location
	found := false
	for i := range user.Locations {
		if user.Locations[i].DeviceID == deviceID {
			user.Locations[i] = location
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found for device_id: " + deviceID})
		return
	}

	// Update user in database
	updatedUser, err := h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sync bucket dimensions to the devices table
	if location.Height > 0 || location.TopDiameter > 0 || location.BottomDiameter > 0 {
		h.syncDeviceDimensions(c, location)
	}

	// Find and return the updated location
	var updatedLocation *auth_models.DeviceLocation
	for i := range updatedUser.Locations {
		if updatedUser.Locations[i].DeviceID == deviceID {
			updatedLocation = &updatedUser.Locations[i]
			break
		}
	}

	if updatedLocation == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated location"})
		return
	}

	c.JSON(http.StatusOK, updatedLocation)
}

// DeleteLocation deletes a location by device_id
func (h *LocationController) DeleteLocation(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	// Get current user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Find and remove the location
	found := false
	newLocations := make(auth_models.DeviceLocationArray, 0, len(user.Locations))
	for _, loc := range user.Locations {
		if loc.DeviceID == deviceID {
			found = true
			continue // Skip this location
		}
		newLocations = append(newLocations, loc)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found for device_id: " + deviceID})
		return
	}

	// Update user with new locations array
	user.Locations = newLocations

	// Update user in database
	_, err = h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "location deleted successfully"})
}

// syncDeviceDimensions updates the devices table with bucket dimensions from a location
func (h *LocationController) syncDeviceDimensions(c *gin.Context, location auth_models.DeviceLocation) {
	device, err := h.deviceRepo.GetDevice(c.Request.Context(), location.PiID, location.DeviceID)
	if err != nil {
		// Device may not exist in devices table yet - log and continue
		log.Printf("Could not find device %s/%s to sync dimensions: %v", location.PiID, location.DeviceID, err)
		return
	}

	device.Height = location.Height
	device.TopDiameter = location.TopDiameter
	device.BottomDiameter = location.BottomDiameter

	if err := h.deviceRepo.UpdateDevice(c.Request.Context(), *device); err != nil {
		log.Printf("Failed to sync dimensions to devices table for %s/%s: %v", location.PiID, location.DeviceID, err)
	}
}

