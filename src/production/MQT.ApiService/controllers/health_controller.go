package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
)

// HealthController handles health and stats requests
type HealthController struct {
	readingRepo    interfaces.ReadingRepository
	piRepo         interfaces.PiRepository
	logger         *logger.Logger
	authMiddleware *middleware.AuthMiddleware
}

// NewHealthController creates a new health controller
func NewHealthController(readingRepo interfaces.ReadingRepository, piRepo interfaces.PiRepository, logger *logger.Logger, authMiddleware *middleware.AuthMiddleware) *HealthController {
	return &HealthController{
		readingRepo:    readingRepo,
		piRepo:         piRepo,
		logger:         logger,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the health routes with Gin
func (c *HealthController) RegisterRoutes(router *gin.Engine) {
	// Public health endpoints
	router.GET("/health/live", c.HealthLive)
	router.GET("/health/ready", c.HealthReady)
	router.GET("/metrics", c.Metrics)

	// Stats endpoint with RBAC
	router.GET("/stats/summary", c.authMiddleware.Authenticate(), c.GetSummaryStats)
}

func (c *HealthController) HealthLive(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (c *HealthController) HealthReady(ctx *gin.Context) {
	// This would typically check database connectivity
	// For now, we'll assume it's ready
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"db":     true,
		"mqtt":   true,
	})
}

func (c *HealthController) Metrics(ctx *gin.Context) {
	// Basic metrics endpoint - can be enhanced with Prometheus metrics
	ctx.String(http.StatusOK, "# HELP mqtt_ingestor_health Health status of MQTT ingestor\n# TYPE mqtt_ingestor_health gauge\nmqtt_ingestor_health 1\n")
}

func (c *HealthController) GetSummaryStats(ctx *gin.Context) {
	piID := ctx.Query("pi_id")
	deviceID := ctx.Query("device_id")
	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")

	// Check user role and filter by user's PIs if not admin
	userRole, _ := middleware.GetRoleFromGinContext(ctx)
	if userRole != "admin" {
		// For users, we need to filter by their assigned PIs
		currentUserID, _ := middleware.GetUserFromGinContext(ctx)

		// If pi_id is specified, check if user has access to it
		if piID != "" {
			pi, err := c.piRepo.GetPi(ctx, piID)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "pi not found"})
				return
			}
			if pi.UserID != currentUserID {
				ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
		} else {
			// If no pi_id specified, we need to get all PIs for this user
			// and modify the query to only include readings from those PIs
			// For now, we'll require pi_id to be specified for non-admin users
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "pi_id is required for non-admin users"})
			return
		}
	}

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
