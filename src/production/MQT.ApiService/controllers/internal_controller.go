package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// InternalController handles internal API endpoints for service-to-service communication
type InternalController struct {
	piRepo      interfaces.PiRepository
	deviceRepo  interfaces.DeviceRepository
	readingRepo interfaces.ReadingRepository
	userRepo    interfaces.UserRepository
	config      *config.Config
}

// NewInternalController creates a new internal controller
func NewInternalController(piRepo interfaces.PiRepository, deviceRepo interfaces.DeviceRepository, readingRepo interfaces.ReadingRepository, userRepo interfaces.UserRepository, cfg *config.Config) *InternalController {
	return &InternalController{
		piRepo:      piRepo,
		deviceRepo:  deviceRepo,
		readingRepo: readingRepo,
		userRepo:    userRepo,
		config:      cfg,
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
	DeviceID string `json:"device_id" binding:"required"`
}

// ValidateDeviceResponse represents the response from Device validation
type ValidateDeviceResponse struct {
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

// CreateReadingRequest represents the request to create a reading
type CreateReadingRequest struct {
	PiID     string                         `json:"pi_id" binding:"required"`
	DeviceID string                         `json:"device_id" binding:"required"`
	Ts       string                         `json:"ts" binding:"required"`
	Payload  hardware_models.ReadingPayload `json:"payload" binding:"required"`
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

	// Check if device has collection enabled; if not, accept but do not store
	device, err := c.deviceRepo.GetDevice(ctx, req.PiID, req.DeviceID)
	if err == nil && device != nil && !device.CollectionEnabled {
		ctx.JSON(http.StatusOK, CreateReadingResponse{
			Success: true,
			Error:   "",
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

	// Check if we should send a bucket fill alert (fire-and-forget)
	if req.Payload.Sensors.Level != nil {
		go c.checkAndSendAlert(req.PiID, req.DeviceID, req.Payload.Sensors.Level.Value)
	}

	ctx.JSON(http.StatusCreated, CreateReadingResponse{
		Success: true,
		Error:   "",
	})
}

// checkAndSendAlert calculates fill percentage and sends alert to email service if threshold exceeded
func (c *InternalController) checkAndSendAlert(piID, deviceID string, sensorDistance float64) {
	device, err := c.deviceRepo.GetDevice(context.Background(), piID, deviceID)
	if err != nil || device == nil {
		return
	}

	// Skip if dimensions not set
	if device.Height <= 0 || device.TopDiameter <= 0 || device.BottomDiameter <= 0 {
		return
	}

	fillPct := calculateFillPercentage(device.Height, device.TopDiameter, device.BottomDiameter, sensorDistance)

	// Look up the user who owns this device
	user, err := c.userRepo.GetUserByDeviceID(context.Background(), deviceID)
	if err != nil || user == nil {
		return
	}

	threshold := float64(c.config.Alert.BucketThreshold)
	if user.SapAlertThresholdPct != nil && *user.SapAlertThresholdPct >= 0 && *user.SapAlertThresholdPct <= 100 {
		threshold = float64(*user.SapAlertThresholdPct)
	}
	if fillPct < threshold {
		return
	}

	// Send alert to email service
	alertPayload := map[string]interface{}{
		"user_email":      user.Email,
		"user_name":       user.Username,
		"device_id":       deviceID,
		"fill_percentage": fillPct,
	}

	body, err := json.Marshal(alertPayload)
	if err != nil {
		log.Printf("Failed to marshal alert payload: %v", err)
		return
	}

	emailURL := c.config.Alert.EmailServiceURL + "/alerts/bucket-fill"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(emailURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("Failed to send alert to email service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("Email service returned status %d for device %s", resp.StatusCode, deviceID)
	}
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
	// Try RFC3339 with nanoseconds first (covers microsecond precision timestamps)
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		return t, nil
	}

	// Then try standard RFC3339
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
