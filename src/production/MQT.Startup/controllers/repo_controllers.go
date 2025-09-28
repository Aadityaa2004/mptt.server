package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// Controllers holds all repository dependencies
type Controllers struct {
	userRepo    interfaces.UserRepository
	piRepo      interfaces.PiRepository
	deviceRepo  interfaces.DeviceRepository
	readingRepo interfaces.ReadingRepository
}

// NewControllers creates a new Controllers instance
func NewControllers(userRepo interfaces.UserRepository, piRepo interfaces.PiRepository, deviceRepo interfaces.DeviceRepository, readingRepo interfaces.ReadingRepository) *Controllers {
	return &Controllers{
		userRepo:    userRepo,
		piRepo:      piRepo,
		deviceRepo:  deviceRepo,
		readingRepo: readingRepo,
	}
}

// SetupRepoRoutes wires all the REST API endpoints
func SetupRepoRoutes(router *gin.Engine, controllers *Controllers) {
	// Allow CORS to match app defaults.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Core endpoints
	router.GET("/health/live", controllers.HealthLive)
	router.GET("/health/ready", controllers.HealthReady)
	router.GET("/metrics", controllers.Metrics)

	// User endpoints
	users := router.Group("/users")
	{
		users.POST("", controllers.CreateUser)
		users.GET("", controllers.ListUsers)
		users.GET("/:user_id", controllers.GetUser)
		users.PATCH("/:user_id", controllers.UpdateUser)
		users.DELETE("/:user_id", controllers.DeleteUser)
	}

	// Pi endpoints
	pis := router.Group("/pis")
	{
		pis.POST("", controllers.CreatePi)
		pis.GET("", controllers.ListPis)
		pis.GET("/:pi_id", controllers.GetPi)
		pis.PATCH("/:pi_id", controllers.UpdatePi)
		pis.DELETE("/:pi_id", controllers.DeletePi)
	}

	// Device endpoints
	devices := router.Group("/pis/:pi_id/devices")
	{
		devices.POST("", controllers.CreateDevice)
		devices.GET("", controllers.ListDevices)
		devices.GET("/:device_id", controllers.GetDevice)
		devices.PATCH("/:device_id", controllers.UpdateDevice)
		devices.DELETE("/:device_id", controllers.DeleteDevice)
	}

	// Reading endpoints
	readings := router.Group("/readings")
	{
		readings.GET("/latest", controllers.GetLatestReadings)
		readings.GET("", controllers.GetReadings)
		readings.GET("/pis/:pi_id/devices/:device_id", controllers.GetDeviceReadings)
	}

	// Stats endpoint
	router.GET("/stats/summary", controllers.GetSummaryStats)
}

// Core Endpoints

func (c *Controllers) HealthLive(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (c *Controllers) HealthReady(ctx *gin.Context) {
	// This would typically check database connectivity
	// For now, we'll assume it's ready
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"db":     true,
		"mqtt":   true,
	})
}

func (c *Controllers) Metrics(ctx *gin.Context) {
	// Basic metrics endpoint - can be enhanced with Prometheus metrics
	ctx.String(http.StatusOK, "# HELP mqtt_ingestor_health Health status of MQTT ingestor\n# TYPE mqtt_ingestor_health gauge\nmqtt_ingestor_health 1\n")
}

// User Endpoints

type CreateUserRequest struct {
	UserID string                 `json:"user_id" binding:"required"`
	Name   string                 `json:"name" binding:"required"`
	Role   string                 `json:"role" binding:"required"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

func (c *Controllers) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := mqtmodels.User{
		UserID:    req.UserID,
		Name:      req.Name,
		Role:      req.Role,
		CreatedAt: time.Now(),
		Meta:      req.Meta,
	}

	if err := c.userRepo.CreateUser(ctx, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (c *Controllers) ListUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	role := ctx.Query("role")

	result, err := c.userRepo.ListUsers(ctx, page, pageSize, role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *Controllers) GetUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	user, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type UpdateUserRequest struct {
	Name *string                `json:"name,omitempty"`
	Role *string                `json:"role,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

func (c *Controllers) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")

	// Get existing user
	existingUser, err := c.userRepo.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existingUser.Name = *req.Name
	}
	if req.Role != nil {
		existingUser.Role = *req.Role
	}
	if req.Meta != nil {
		existingUser.Meta = req.Meta
	}

	if err := c.userRepo.UpdateUser(ctx, *existingUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingUser)
}

func (c *Controllers) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	hardDelete := ctx.DefaultQuery("hard", "false") == "true"

	if err := c.userRepo.DeleteUser(ctx, userID, hardDelete); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}

// Pi Endpoints

type CreatePiRequest struct {
	PiID   string                 `json:"pi_id" binding:"required"`
	UserID string                 `json:"user_id,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

func (c *Controllers) CreatePi(ctx *gin.Context) {
	var req CreatePiRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that the user exists if user_id is provided
	if req.UserID != "" {
		user, err := c.userRepo.GetUser(ctx, req.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if user == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
			return
		}
	}

	pi := mqtmodels.Pi{
		PiID:      req.PiID,
		UserID:    req.UserID,
		CreatedAt: time.Now(),
		Meta:      req.Meta,
	}

	if err := c.piRepo.CreateOrUpdatePi(ctx, pi); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, pi)
}

func (c *Controllers) ListPis(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	result, err := c.piRepo.ListPis(ctx, userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *Controllers) GetPi(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	pi, err := c.piRepo.GetPi(ctx, piID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "pi not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, pi)
}

type UpdatePiRequest struct {
	UserID *string                `json:"user_id,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

func (c *Controllers) UpdatePi(ctx *gin.Context) {
	piID := ctx.Param("pi_id")

	// Get existing pi
	existingPi, err := c.piRepo.GetPi(ctx, piID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "pi not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req UpdatePiRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.UserID != nil {
		existingPi.UserID = *req.UserID
	}
	if req.Meta != nil {
		existingPi.Meta = req.Meta
	}

	if err := c.piRepo.UpdatePi(ctx, *existingPi); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingPi)
}

func (c *Controllers) DeletePi(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	cascade := ctx.DefaultQuery("cascade", "false") == "true"

	if err := c.piRepo.DeletePi(ctx, piID, cascade); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "pi not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}

// Device Endpoints

type CreateDeviceRequest struct {
	DeviceID   int                    `json:"device_id" binding:"required"`
	DeviceType string                 `json:"device_type" binding:"required"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}

func (c *Controllers) CreateDevice(ctx *gin.Context) {
	piID := ctx.Param("pi_id")

	var req CreateDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := mqtmodels.Device{
		PiID:       piID,
		DeviceID:   req.DeviceID,
		DeviceType: req.DeviceType,
		CreatedAt:  time.Now(),
		Meta:       req.Meta,
	}

	if err := c.deviceRepo.CreateOrUpdateDevice(ctx, device); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, device)
}

func (c *Controllers) ListDevices(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	result, err := c.deviceRepo.ListDevicesByPi(ctx, piID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *Controllers) GetDevice(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, device)
}

type UpdateDeviceRequest struct {
	Meta map[string]interface{} `json:"meta,omitempty"`
}

func (c *Controllers) UpdateDevice(ctx *gin.Context) {
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

	// Update meta if provided
	if req.Meta != nil {
		existingDevice.Meta = req.Meta
	}

	if err := c.deviceRepo.UpdateDevice(ctx, *existingDevice); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingDevice)
}

func (c *Controllers) DeleteDevice(ctx *gin.Context) {
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

// Reading Endpoints

func (c *Controllers) GetLatestReadings(ctx *gin.Context) {
	piID := ctx.Query("pi_id")
	if piID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "pi_id is required"})
		return
	}

	readings, err := c.readingRepo.GetLatestReadings(ctx, piID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"items": readings})
}

func (c *Controllers) GetReadings(ctx *gin.Context) {
	piID := ctx.Query("pi_id")
	if piID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "pi_id is required"})
		return
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

func (c *Controllers) GetDeviceReadings(ctx *gin.Context) {
	piID := ctx.Param("pi_id")
	deviceIDStr := ctx.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
		return
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

// Stats Endpoint

func (c *Controllers) GetSummaryStats(ctx *gin.Context) {
	piID := ctx.Query("pi_id")
	deviceID := ctx.Query("device_id")
	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")

	params := interfaces.ReadingQueryParams{
		PiID:     piID,
		DeviceID: deviceID,
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

	result, err := c.readingRepo.GetSummaryStats(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
