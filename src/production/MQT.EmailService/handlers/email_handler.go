package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
)

// EmailHandler handles auth-related email sending (OTP, password reset)
type EmailHandler struct {
	emailService *service.EmailService
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *service.EmailService) *EmailHandler {
	return &EmailHandler{emailService: emailService}
}

// SendOTPRequest represents the request to send an OTP email
type SendOTPRequest struct {
	To      string `json:"to" binding:"required"`
	OTP     string `json:"otp" binding:"required"`
	Purpose string `json:"purpose"`
}

// SendOTP sends an OTP verification email
func (h *EmailHandler) SendOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = "signup"
	}

	subject := "Verify your email - MapleSense"
	body := fmt.Sprintf(`<html><body>
<h2>Email Verification</h2>
<p>Your verification code is:</p>
<p style="font-size: 24px; font-weight: bold; letter-spacing: 4px;">%s</p>
<p>This code expires in 15 minutes.</p>
<p>If you did not request this, please ignore this email.</p>
<p>— MapleSense</p>
</body></html>`, req.OTP)

	if err := h.emailService.SendHTML(req.To, subject, body); err != nil {
		log.Printf("SendOTP: SMTP error for %s: %v", req.To, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent", "message": "OTP email sent"})
}

// SendPasswordResetRequest represents the request to send a password reset email
type SendPasswordResetRequest struct {
	To        string `json:"to" binding:"required"`
	ResetLink string `json:"reset_link" binding:"required"`
	UserName  string `json:"user_name"`
}

// SendPasswordReset sends a password reset link email
func (h *EmailHandler) SendPasswordReset(c *gin.Context) {
	var req SendPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userName := req.UserName
	if userName == "" {
		userName = "User"
	}

	subject := "Reset your password - MapleSense"
	body := fmt.Sprintf(`<html><body>
<h2>Password Reset Request</h2>
<p>Hello %s,</p>
<p>We received a request to reset your password. Click the link below to set a new password:</p>
<p><a href="%s" style="background-color: #000; color: #fff; padding: 10px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a></p>
<p>Or copy and paste this link into your browser:</p>
<p style="word-break: break-all;">%s</p>
<p>This link expires in 1 hour.</p>
<p>If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>
<p>— MapleSense</p>
</body></html>`, userName, req.ResetLink, req.ResetLink)

	if err := h.emailService.SendHTML(req.To, subject, body); err != nil {
		log.Printf("SendPasswordReset: SMTP error for %s: %v", req.To, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent", "message": "password reset email sent"})
}
