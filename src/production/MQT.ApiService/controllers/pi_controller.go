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

// PiController handles Pi management requests
type PiController struct {
	piRepo         interfaces.PiRepository
	userRepo       interfaces.UserRepository
	logger         *logger.Logger
	authMiddleware *middleware.AuthMiddleware
}

// NewPiController creates a new pi controller
func NewPiController(piRepo interfaces.PiRepository, userRepo interfaces.UserRepository, logger *logger.Logger, authMiddleware *middleware.AuthMiddleware) *PiController {
	return &PiController{
		piRepo:         piRepo,
		userRepo:       userRepo,
		logger:         logger,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes registers the pi routes with Gin
func (c *PiController) RegisterRoutes(router *gin.Engine) {
	pis := router.Group("/pis")
	{
		// Admin only - create/update/delete
		pis.POST("", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.CreatePi)
		pis.PATCH("/:pi_id", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.UpdatePi)
		pis.DELETE("/:pi_id", c.authMiddleware.Authenticate(), c.authMiddleware.RequireAdmin(), c.DeletePi)

		// Admin: all PIs, User: only their assigned PIs
		pis.GET("", c.authMiddleware.Authenticate(), c.ListPis)
		pis.GET("/:pi_id", c.authMiddleware.Authenticate(), c.GetPi)
	}
}

type CreatePiRequest struct {
	PiID   string `json:"pi_id" binding:"required"`
	UserID string `json:"user_id,omitempty"`
}

func (c *PiController) CreatePi(ctx *gin.Context) {
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

	pi := hardware_models.Pi{
		PiID:      req.PiID,
		UserID:    req.UserID,
		CreatedAt: time.Now(),
	}

	if err := c.piRepo.CreateOrUpdatePi(ctx, pi); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, pi)
}

func (c *PiController) ListPis(ctx *gin.Context) {
	userRole, _ := middleware.GetRoleFromGinContext(ctx)
	currentUserID, _ := middleware.GetUserFromGinContext(ctx)

	// If user role, filter by their user_id
	filterUserID := ctx.Query("user_id")
	if userRole != "admin" {
		filterUserID = currentUserID
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	result, err := c.piRepo.ListPis(ctx, filterUserID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *PiController) GetPi(ctx *gin.Context) {
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

	// Check ownership if not admin
	userRole, _ := middleware.GetRoleFromGinContext(ctx)
	if userRole != "admin" {
		currentUserID, _ := middleware.GetUserFromGinContext(ctx)
		if pi.UserID != currentUserID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	ctx.JSON(http.StatusOK, pi)
}

type UpdatePiRequest struct {
	UserID *string `json:"user_id,omitempty"`
}

func (c *PiController) UpdatePi(ctx *gin.Context) {
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

	if err := c.piRepo.UpdatePi(ctx, *existingPi); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingPi)
}

func (c *PiController) DeletePi(ctx *gin.Context) {
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
