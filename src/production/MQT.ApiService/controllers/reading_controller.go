package controllers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// ReadingController handles Reading management requests
type ReadingController struct {
	readingRepo    interfaces.ReadingRepository
	piRepo         interfaces.PiRepository
	deviceRepo     interfaces.DeviceRepository
	logger         *logger.Logger
	authMiddleware *middleware.AuthMiddleware
}

// NewReadingController creates a new reading controller
func NewReadingController(readingRepo interfaces.ReadingRepository, piRepo interfaces.PiRepository, deviceRepo interfaces.DeviceRepository, logger *logger.Logger, authMiddleware *middleware.AuthMiddleware) *ReadingController {
	return &ReadingController{
		readingRepo:    readingRepo,
		piRepo:         piRepo,
		deviceRepo:     deviceRepo,
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
		// Admin only: delete readings in date range
		readings.DELETE("/pis/:pi_id/devices/:device_id", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.DeleteDeviceReadingsByTimeRange)
	}
}

// calculateFillPercentage computes the fill percentage for a frustum-shaped bucket.
// height, topDiameter, bottomDiameter are bucket dimensions (cm).
// sensorDistance is the top-down distance from sensor to sap surface (cm).
func calculateFillPercentage(height, topDiameter, bottomDiameter, sensorDistance float64) float64 {
	if height <= 0 || topDiameter <= 0 || bottomDiameter <= 0 {
		return 0
	}

	sapHeight := height - sensorDistance
	if sapHeight <= 0 {
		return 0
	}
	if sapHeight > height {
		sapHeight = height
	}

	Rb := bottomDiameter / 2.0
	Rt := topDiameter / 2.0

	// Total volume of the frustum
	totalVolume := (math.Pi * height / 3.0) * (Rb*Rb + Rb*Rt + Rt*Rt)

	// Radius at the sap level (interpolated)
	Rsap := Rb + (Rt-Rb)*(sapHeight/height)

	// Volume of sap (frustum from bottom to sapHeight)
	sapVolume := (math.Pi * sapHeight / 3.0) * (Rb*Rb + Rb*Rsap + Rsap*Rsap)

	fillPct := (sapVolume / totalVolume) * 100.0

	// Clamp to 0-100
	if fillPct < 0 {
		fillPct = 0
	}
	if fillPct > 100 {
		fillPct = 100
	}

	return math.Round(fillPct*100) / 100 // round to 2 decimal places
}

// readingResponse wraps a reading with optional fill_percentage and sap_depth_cm.
// sap_depth_cm is the fill depth from bottom (cm); user-facing alternative to raw sensor distance.
type readingResponse struct {
	hardware_models.Reading
	FillPercentage *float64 `json:"fill_percentage,omitempty"`
	SapDepthCm     *float64 `json:"sap_depth_cm,omitempty"`
}

// enrichReadingsWithFillPercentage adds fill percentage to readings if device dimensions are set
func (c *ReadingController) enrichReadingsWithFillPercentage(ctx *gin.Context, readings []hardware_models.Reading) []readingResponse {
	result := make([]readingResponse, len(readings))

	// Cache device lookups by pi_id+device_id
	deviceCache := make(map[string]*hardware_models.Device)

	for i, r := range readings {
		result[i] = readingResponse{Reading: r}

		key := r.PiID + ":" + r.DeviceID
		device, ok := deviceCache[key]
		if !ok {
			d, err := c.deviceRepo.GetDevice(ctx, r.PiID, r.DeviceID)
			if err != nil {
				deviceCache[key] = nil
				continue
			}
			device = d
			deviceCache[key] = device
		}

		if device == nil {
			continue
		}

		// Check if dimensions are set and level sensor data exists
		if device.Height > 0 && device.TopDiameter > 0 && device.BottomDiameter > 0 &&
			r.Payload.Sensors.Level != nil {
			sensorDist := r.Payload.Sensors.Level.Value
			fillPct := calculateFillPercentage(device.Height, device.TopDiameter, device.BottomDiameter, sensorDist)
			result[i].FillPercentage = &fillPct
			// Sap depth = fill from bottom (cm) — user-facing metric
			sapDepth := device.Height - sensorDist
			if sapDepth < 0 {
				sapDepth = 0
			}
			if sapDepth > device.Height {
				sapDepth = device.Height
			}
			result[i].SapDepthCm = &sapDepth
		}
	}

	return result
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

	enriched := c.enrichReadingsWithFillPercentage(ctx, readings)
	ctx.JSON(http.StatusOK, gin.H{"items": enriched})
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

	enriched := c.enrichReadingsWithFillPercentage(ctx, result.Items)
	ctx.JSON(http.StatusOK, gin.H{
		"items":           enriched,
		"next_page_token": result.NextPageToken,
		"total":           result.Total,
	})
}

func (c *ReadingController) GetDeviceReadings(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceID := ctx.Param("device_id")

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

	result, err := c.readingRepo.GetReadingsByDevice(ctx, piID, deviceID, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	enriched := c.enrichReadingsWithFillPercentage(ctx, result.Items)
	ctx.JSON(http.StatusOK, gin.H{
		"items":           enriched,
		"next_page_token": result.NextPageToken,
		"total":           result.Total,
	})
}

// DeleteDeviceReadingsByTimeRange deletes readings for a device within a date range (admin only).
func (c *ReadingController) DeleteDeviceReadingsByTimeRange(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceID := ctx.Param("device_id")
	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")

	if fromStr == "" || toStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query params 'from' and 'to' (RFC3339) are required"})
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' format, use RFC3339"})
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' format, use RFC3339"})
		return
	}
	if from.After(to) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "'from' must be before or equal to 'to'"})
		return
	}

	if err := c.readingRepo.DeleteReadingsByTimeRange(ctx, piID, deviceID, from, to); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
