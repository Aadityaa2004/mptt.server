package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
)

// ReadingController handles Reading management requests
type ReadingController struct {
	readingRepo    interfaces.ReadingRepository
	piRepo         interfaces.PiRepository
	logger         *logger.Logger
	authMiddleware *middleware.AuthMiddleware
}

// NewReadingController creates a new reading controller
func NewReadingController(readingRepo interfaces.ReadingRepository, piRepo interfaces.PiRepository, logger *logger.Logger, authMiddleware *middleware.AuthMiddleware) *ReadingController {
	return &ReadingController{
		readingRepo:    readingRepo,
		piRepo:         piRepo,
		logger:         logger,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the reading routes with Gin
func (c *ReadingController) RegisterRoutes(router *gin.Engine) {
	readings := router.Group("/readings")
	{
		// Admin: all readings, User: readings from their devices
		readings.GET("/latest", c.authMiddleware.Authenticate(), c.GetLatestReadings)
		readings.GET("", c.authMiddleware.Authenticate(), c.GetReadings)
		readings.GET("/pis/:pi_id/devices/:device_id", c.authMiddleware.Authenticate(), c.GetDeviceReadings)
	}
}

func (c *ReadingController) GetLatestReadings(ctx *gin.Context) {
	piID := ctx.Query("pi_id")
	if piID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "pi_id is required"})
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

	readings, err := c.readingRepo.GetLatestReadings(ctx, piID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"items": readings})
}

func (c *ReadingController) GetReadings(ctx *gin.Context) {
	piID := ctx.Query("pi_id")
	if piID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "pi_id is required"})
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

	deviceID := ctx.Query("device_id")
	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))

	params := interfaces.ReadingQueryParams{
		PiID:     piID,
		DeviceID: deviceID,
		Limit:    limit,
		Page:     page,
	}

	if fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			params.From = &from
		}
	}

	if toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			params.To = &to
		}
	}

	result, err := c.readingRepo.GetReadings(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *ReadingController) GetDeviceReadings(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceIDStr := ctx.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
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

	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))

	params := interfaces.ReadingQueryParams{
		PiID:     piID,
		DeviceID: deviceIDStr,
		Limit:    limit,
		Page:     page,
	}

	if fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			params.From = &from
		}
	}

	if toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			params.To = &to
		}
	}

	result, err := c.readingRepo.GetReadingsByDevice(ctx, piID, deviceID, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
