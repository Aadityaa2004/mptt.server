package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/cooldown"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
)

type AlertHandler struct {
	emailSender          service.EmailSender
	cooldownTracker      *cooldown.Tracker
	emptyAlertTracker    *cooldown.Tracker
	logger               *logger.Logger
}

func NewAlertHandler(emailSender service.EmailSender, cooldownTracker *cooldown.Tracker, emptyAlertTracker *cooldown.Tracker, log *logger.Logger) *AlertHandler {
	return &AlertHandler{
		emailSender:       emailSender,
		cooldownTracker:   cooldownTracker,
		emptyAlertTracker: emptyAlertTracker,
		logger:            log,
	}
}

type BucketFillAlertRequest struct {
	UserEmail      string  `json:"user_email" binding:"required"`
	UserName       string  `json:"user_name" binding:"required"`
	DeviceID       string  `json:"device_id" binding:"required"`
	FillPercentage float64 `json:"fill_percentage" binding:"required"`
	Timestamp      string  `json:"timestamp"`
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

	if err := h.emailSender.SendHTML(req.UserEmail, subject, body); err != nil {
		h.logger.ErrorWithError(err, "HandleBucketFillAlert: failed to send email for "+req.UserEmail)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	// Record the alert time for cooldown
	h.cooldownTracker.Record(key)

	c.JSON(http.StatusOK, gin.H{"status": "sent", "message": "alert email sent"})
}

func (h *AlertHandler) HandleBucketEmptyAlert(c *gin.Context) {
	var req BucketFillAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reset cooldown once the bucket refills above the reset threshold
	key := req.UserEmail + ":" + req.DeviceID
	if h.emptyAlertTracker.ShouldReset(req.FillPercentage) {
		h.emptyAlertTracker.Reset(key)
		c.JSON(http.StatusOK, gin.H{"status": "above_reset_threshold", "message": "fill above reset threshold, empty-alert cooldown reset"})
		return
	}

	if h.emptyAlertTracker.IsInCooldown(key) {
		c.JSON(http.StatusOK, gin.H{"status": "cooldown", "message": "alert suppressed due to cooldown"})
		return
	}

	subject := "MapleSense Notification: Your Bucket Has Been Emptied"
	body := fmt.Sprintf(`<html><body>
<p>Your MapleSense bucket is now reading below 15%% full. This means that your bucket was picked up and emptied!</p>
<p><strong>If you did not empty this bucket, please CHECK ON IT IMMEDIATELY.</strong> A sudden drop could indicate the bucket was moved, tipped over, leaking, stolen, or a sensor issue.</p>
<br>
<p>Device ID: %s</p>
<p>Current Fill: %.2f%%</p>
<p>Timestamp: %s</p>
<p>— MapleSense Alerts</p>
</body></html>`, req.DeviceID, req.FillPercentage, req.Timestamp)

	if err := h.emailSender.SendHTML(req.UserEmail, subject, body); err != nil {
		h.logger.ErrorWithError(err, "HandleBucketEmptyAlert: failed to send email for "+req.UserEmail)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	h.emptyAlertTracker.Record(key)

	c.JSON(http.StatusOK, gin.H{"status": "sent", "message": "empty alert email sent"})
}
