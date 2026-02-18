package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/cooldown"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
)

type AlertHandler struct {
	emailService    *service.EmailService
	cooldownTracker *cooldown.Tracker
}

func NewAlertHandler(emailService *service.EmailService, cooldownTracker *cooldown.Tracker) *AlertHandler {
	return &AlertHandler{
		emailService:    emailService,
		cooldownTracker: cooldownTracker,
	}
}

type BucketFillAlertRequest struct {
	UserEmail      string  `json:"user_email" binding:"required"`
	UserName       string  `json:"user_name" binding:"required"`
	DeviceID       string  `json:"device_id" binding:"required"`
	FillPercentage float64 `json:"fill_percentage" binding:"required"`
}

func (h *AlertHandler) HandleBucketFillAlert(c *gin.Context) {
	var req BucketFillAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check cooldown - if fill is below reset threshold, reset cooldown
	key := req.UserEmail + ":" + req.DeviceID
	if h.cooldownTracker.ShouldReset(req.FillPercentage) {
		h.cooldownTracker.Reset(key)
		c.JSON(http.StatusOK, gin.H{"status": "below_threshold", "message": "fill below reset threshold, cooldown reset"})
		return
	}

	// Check if we're in cooldown period
	if h.cooldownTracker.IsInCooldown(key) {
		c.JSON(http.StatusOK, gin.H{"status": "cooldown", "message": "alert suppressed due to cooldown"})
		return
	}

	// Send the email
	subject := fmt.Sprintf("Bucket Alert: %s is at %.1f%% capacity", req.DeviceID, req.FillPercentage)
	body := fmt.Sprintf(`<html><body>
<h2>Bucket Fill Alert</h2>
<p>Hello %s,</p>
<p>Your bucket <strong>%s</strong> is at <strong>%.1f%%</strong> capacity.</p>
<p>Please check your bucket and collect the sap soon.</p>
<p>— MapleSense Alerts</p>
</body></html>`, req.UserName, req.DeviceID, req.FillPercentage)

	if err := h.emailService.SendHTML(req.UserEmail, subject, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	// Record the alert time for cooldown
	h.cooldownTracker.Record(key)

	c.JSON(http.StatusOK, gin.H{"status": "sent", "message": "alert email sent"})
}
