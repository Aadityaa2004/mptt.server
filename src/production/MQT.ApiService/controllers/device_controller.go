package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
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
	DeviceID   int    `json:"device_id" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"`
}

func (c *DeviceController) CreateDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")

	var req CreateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := hardware_models.Device{
		PiID:       piID,
		DeviceID:   req.DeviceID,
		DeviceType: req.DeviceType,
		CreatedAt:  time.Now(),
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
	deviceIDStr := ctx.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
		return
	}

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
	DeviceType *string `json:"device_type,omitempty"`
}

func (c *DeviceController) UpdateDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceIDStr := ctx.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
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

	var req UpdateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update device_type if provided
	if req.DeviceType != nil {
		existingDevice.DeviceType = *req.DeviceType
	}

	if err := c.deviceRepo.UpdateDevice(ctx, *existingDevice); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingDevice)
}

func (c *DeviceController) DeleteDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceIDStr := ctx.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
		return
	}

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
