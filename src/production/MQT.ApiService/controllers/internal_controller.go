package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// InternalController handles internal API endpoints for service-to-service communication
type InternalController struct {
	piRepo      interfaces.PiRepository
	deviceRepo  interfaces.DeviceRepository
	readingRepo interfaces.ReadingRepository
}

// NewInternalController creates a new internal controller
func NewInternalController(piRepo interfaces.PiRepository, deviceRepo interfaces.DeviceRepository, readingRepo interfaces.ReadingRepository) *InternalController {
	return &InternalController{
		piRepo:      piRepo,
		deviceRepo:  deviceRepo,
		readingRepo: readingRepo,
	}
}

// ValidatePiRequest represents the request to validate a Pi
type ValidatePiRequest struct {
	PiID string `json:"pi_id" binding:"required"`
}

// ValidatePiResponse represents the response from Pi validation
type ValidatePiResponse struct {
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

// ValidateDeviceRequest represents the request to validate a Device
type ValidateDeviceRequest struct {
	PiID     string `json:"pi_id" binding:"required"`
	DeviceID int    `json:"device_id" binding:"required"`
}

// ValidateDeviceResponse represents the response from Device validation
type ValidateDeviceResponse struct {
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

// CreateReadingRequest represents the request to create a reading
type CreateReadingRequest struct {
	PiID     string                 `json:"pi_id" binding:"required"`
	DeviceID int                    `json:"device_id" binding:"required"`
	Ts       string                 `json:"ts" binding:"required"`
	Payload  map[string]interface{} `json:"payload" binding:"required"`
}

// CreateReadingResponse represents the response from reading creation
type CreateReadingResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ValidatePi checks if a Pi exists
func (c *InternalController) ValidatePi(ctx *gin.Context) {
	var req ValidatePiRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ValidatePiResponse{
			Exists: false,
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	// Check if Pi exists
	_, err := c.piRepo.GetPi(ctx, req.PiID)
	if err != nil {
		ctx.JSON(http.StatusOK, ValidatePiResponse{
			Exists: false,
			Error:  "",
		})
		return
	}

	ctx.JSON(http.StatusOK, ValidatePiResponse{
		Exists: true,
		Error:  "",
	})
}

// ValidateDevice checks if a Device exists for a given Pi
func (c *InternalController) ValidateDevice(ctx *gin.Context) {
	var req ValidateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ValidateDeviceResponse{
			Exists: false,
			Error:  "Invalid request: " + err.Error(),
		})
		return
	}

	// Check if Device exists
	_, err := c.deviceRepo.GetDevice(ctx, req.PiID, req.DeviceID)
	if err != nil {
		ctx.JSON(http.StatusOK, ValidateDeviceResponse{
			Exists: false,
			Error:  "",
		})
		return
	}

	ctx.JSON(http.StatusOK, ValidateDeviceResponse{
		Exists: true,
		Error:  "",
	})
}

// CreateReading creates a reading
func (c *InternalController) CreateReading(ctx *gin.Context) {
	var req CreateReadingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, CreateReadingResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Parse timestamp
	ts, err := parseTimeString(req.Ts)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, CreateReadingResponse{
			Success: false,
			Error:   "Invalid timestamp format: " + err.Error(),
		})
		return
	}

	// Create reading
	reading := hardware_models.Reading{
		PiID:     req.PiID,
		DeviceID: req.DeviceID,
		Ts:       ts,
		Payload:  req.Payload,
	}

	if err := c.readingRepo.CreateReading(ctx, reading); err != nil {
		ctx.JSON(http.StatusInternalServerError, CreateReadingResponse{
			Success: false,
			Error:   "Failed to create reading: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, CreateReadingResponse{
		Success: true,
		Error:   "",
	})
}

// RegisterRoutes registers the internal API routes
func (c *InternalController) RegisterRoutes(router *gin.Engine) {
	// Internal API group with service-to-service authentication
	internal := router.Group("/internal")
	internal.Use(middleware.ServiceAuthMiddleware())

	// Pi validation endpoint
	internal.POST("/pis/validate", c.ValidatePi)

	// Device validation endpoint
	internal.POST("/devices/validate", c.ValidateDevice)

	// Reading creation endpoint
	internal.POST("/readings", c.CreateReading)
}

// parseTimeString parses a time string in RFC3339 format
func parseTimeString(timeStr string) (time.Time, error) {
	// Try RFC3339 format first
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// Try other common formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}
