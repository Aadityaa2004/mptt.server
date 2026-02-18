package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// DeviceController handles Device management requests
type DeviceController struct {
	deviceRepo     interfaces.DeviceRepository
	piRepo         interfaces.PiRepository
	logger         *logger.Logger
	authMiddleware *middleware.AuthMiddleware
}

// NewDeviceController creates a new device controller
func NewDeviceController(deviceRepo interfaces.DeviceRepository, piRepo interfaces.PiRepository, logger *logger.Logger, authMiddleware *middleware.AuthMiddleware) *DeviceController {
	return &DeviceController{
		deviceRepo:     deviceRepo,
		piRepo:         piRepo,
		logger:         logger,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the device routes with Gin
func (c *DeviceController) RegisterRoutes(router *gin.Engine) {
	devices := router.Group("/pis/:pi_id/devices")
	{
		// Admin only - create/update/delete
		devices.POST("", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.CreateDevice)
		devices.PATCH("/:device_id", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.UpdateDevice)
		devices.DELETE("/:device_id", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.DeleteDevice)

		// Admin: all devices, User: devices from their PIs
		devices.GET("", c.authMiddleware.Authenticate(), c.ListDevices)
		devices.GET("/:device_id", c.authMiddleware.Authenticate(), c.GetDevice)
	}
}

type CreateDeviceRequest struct {
	DeviceID       string  `json:"device_id" binding:"required"`
	Height         float64 `json:"height,omitempty"`
	TopDiameter    float64 `json:"top_diameter,omitempty"`
	BottomDiameter float64 `json:"bottom_diameter,omitempty"`
}

func (c *DeviceController) CreateDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")

	var req CreateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := hardware_models.Device{
		PiID:           piID,
		DeviceID:       req.DeviceID,
		Height:         req.Height,
		TopDiameter:    req.TopDiameter,
		BottomDiameter: req.BottomDiameter,
	}

	if err := c.deviceRepo.CreateOrUpdateDevice(ctx, device); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, device)
}

func (c *DeviceController) ListDevices(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	// Check if user has access to this PI
	userRole, _ := middleware.GetRoleFromGinContext(ctx)
	if userRole != "admin" {
		currentUserID, _ := middleware.GetUserFromGinContext(ctx)
		pi, err := c.piRepo.GetPi(ctx, piID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "pi not found"})
			return
		}
		if pi.UserID != currentUserID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	result, err := c.deviceRepo.ListDevicesByPi(ctx, piID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *DeviceController) GetDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceID := ctx.Param("device_id")

	device, err := c.deviceRepo.GetDevice(ctx, piID, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if user has access to this PI
	userRole, _ := middleware.GetRoleFromGinContext(ctx)
	if userRole != "admin" {
		currentUserID, _ := middleware.GetUserFromGinContext(ctx)
		pi, err := c.piRepo.GetPi(ctx, piID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "pi not found"})
			return
		}
		if pi.UserID != currentUserID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	ctx.JSON(http.StatusOK, device)
}

type UpdateDeviceRequest struct {
	Height         *float64 `json:"height,omitempty"`
	TopDiameter    *float64 `json:"top_diameter,omitempty"`
	BottomDiameter *float64 `json:"bottom_diameter,omitempty"`
}

func (c *DeviceController) UpdateDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceID := ctx.Param("device_id")

	var req UpdateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing device
	existingDevice, err := c.deviceRepo.GetDevice(ctx, piID, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Height != nil {
		existingDevice.Height = *req.Height
	}
	if req.TopDiameter != nil {
		existingDevice.TopDiameter = *req.TopDiameter
	}
	if req.BottomDiameter != nil {
		existingDevice.BottomDiameter = *req.BottomDiameter
	}

	if err := c.deviceRepo.UpdateDevice(ctx, *existingDevice); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingDevice)
}

func (c *DeviceController) DeleteDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceID := ctx.Param("device_id")

	cascade := ctx.DefaultQuery("cascade", "false") == "true"

	if err := c.deviceRepo.DeleteDevice(ctx, piID, deviceID, cascade); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
