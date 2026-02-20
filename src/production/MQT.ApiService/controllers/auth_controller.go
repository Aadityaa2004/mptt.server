package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	service "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	weather_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/weather"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication requests
type AuthController struct {
	authService         *service.AuthService
	openWeatherAPIKey   string
	turnstileSecretKey  string
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *service.AuthService, openWeatherAPIKey string, turnstileSecretKey string) *AuthController {
	return &AuthController{
		authService:        authService,
		openWeatherAPIKey:  openWeatherAPIKey,
		turnstileSecretKey: turnstileSecretKey,
	}
}

// verifyTurnstile checks the token with Cloudflare's siteverify API.
// When turnstileSecretKey is empty, verification is skipped (returns nil).
func (h *AuthController) verifyTurnstile(token string) error {
	if h.turnstileSecretKey == "" {
		return nil
	}
	if token == "" {
		return fmt.Errorf("Turnstile verification required but no token provided")
	}
	form := url.Values{}
	form.Set("secret", h.turnstileSecretKey)
	form.Set("response", token)
	req, err := http.NewRequest("POST", "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool     `json:"success"`
		Errors  []string `json:"error-codes,omitempty"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if !result.Success {
		if len(result.Errors) > 0 {
			return fmt.Errorf("Turnstile verification failed: %v", result.Errors)
		}
		return fmt.Errorf("Turnstile verification failed")
	}
	return nil
}

// Register handles user registration
func (h *AuthController) Register(c *gin.Context) {
	var req struct {
		service.RegisterRequest
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.verifyTurnstile(req.TurnstileToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force user role for regular registration - no admin role allowed
	req.Role = "user"

	resp, err := h.authService.Register(c.Request.Context(), req.RegisterRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := gin.H{
		"requires_verification": resp.RequiresVerification,
		"message":               resp.Message,
		"email":                 req.Email,
		"username":              req.Username,
	}
	if resp.User != nil {
		payload["id"] = resp.User.UserID
		payload["role"] = resp.User.Role
	} else {
		payload["role"] = "user"
	}
	c.JSON(http.StatusCreated, payload)
}

// RegisterAdmin handles admin user registration (also requires email verification)
func (h *AuthController) RegisterAdmin(c *gin.Context) {
	var req struct {
		service.RegisterRequest
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.verifyTurnstile(req.TurnstileToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force admin role
	req.Role = "admin"

	resp, err := h.authService.Register(c.Request.Context(), req.RegisterRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := gin.H{
		"requires_verification": resp.RequiresVerification,
		"message":               resp.Message,
		"email":                 req.Email,
		"username":              req.Username,
	}
	if resp.User != nil {
		payload["id"] = resp.User.UserID
		payload["role"] = resp.User.Role
	} else {
		payload["role"] = req.Role
	}
	c.JSON(http.StatusCreated, payload)
}

// Login handles user login
func (h *AuthController) Login(c *gin.Context) {
	var req struct {
		service.LoginRequest
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.verifyTurnstile(req.TurnstileToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, tokenPair, err := h.authService.Login(c.Request.Context(), req.LoginRequest)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set refresh token as HTTP-only cookie
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(time.Unix(tokenPair.ExpiresAt, 0)).Seconds()),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HTTP only
	)

	// Return access token in the response body
	c.JSON(http.StatusOK, response)
}

// RefreshTokens handles token refresh
func (h *AuthController) RefreshTokens(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	response, tokenPair, err := h.authService.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new refresh token as HTTP-only cookie
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(time.Unix(tokenPair.ExpiresAt, 0)).Seconds()),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HTTP only
	)

	// Return new access token in the response body
	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthController) Logout(c *gin.Context) {
	// Clear the refresh token cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HTTP only
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Profile retrieves the authenticated user's profile
func (h *AuthController) Profile(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user profile
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile handles updating the authenticated user's profile
func (h *AuthController) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Username  string   `json:"username,omitempty"`
		Email     string   `json:"email,omitempty"`
		Password  string   `json:"password,omitempty"`
		Latitude  *float64 `json:"latitude,omitempty"`
		Longitude *float64 `json:"longitude,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		// Hash the new password
		hashedPassword, err := h.authService.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}
	if req.Latitude != nil {
		user.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		user.Longitude = req.Longitude
	}

	// Update user in database
	updatedUser, err := h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// UpdateLocation handles updating the authenticated user's location (latitude and longitude)
func (h *AuthController) UpdateLocation(c *gin.Context) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update location fields
	user.Latitude = &req.Latitude
	user.Longitude = &req.Longitude

	// Update user in database
	updatedUser, err := h.authService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// getCoordinates retrieves coordinates from the authenticated user's profile
func (h *AuthController) getCoordinates(c *gin.Context) (float64, float64, error) {
	// Get user ID from context
	userID, err := middleware.GetUserFromGinContext(c)
	if err != nil {
		return 0, 0, fmt.Errorf("unauthorized")
	}

	// Get user profile
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user: %v", err)
	}

	// Check if user has location set
	if user.Latitude == nil || user.Longitude == nil {
		return 0, 0, fmt.Errorf("location not set. Please set your location using PUT /api/auth/location")
	}

	return *user.Latitude, *user.Longitude, nil
}

// GetCurrentWeather retrieves current weather information
func (h *AuthController) GetCurrentWeather(c *gin.Context) {
	// Get coordinates
	lat, lon, err := h.getCoordinates(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API key
	if h.openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OPENWEATHER_API_KEY is not configured"})
		return
	}

	// Build API URL
	weatherURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%.4f&lon=%.4f&appid=%s&units=metric", lat, lon, h.openWeatherAPIKey)

	// Make API request
	resp, err := http.Get(weatherURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to connect to weather API: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("weather API returned status code %d", resp.StatusCode)})
		return
	}

	// Decode response
	var weatherData weather_models.CurrentWeather
	err = json.NewDecoder(resp.Body).Decode(&weatherData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to process weather data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, weatherData)
}

// GetWeatherForecast retrieves weather forecast information
func (h *AuthController) GetWeatherForecast(c *gin.Context) {
	// Get coordinates
	lat, lon, err := h.getCoordinates(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API key
	if h.openWeatherAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OPENWEATHER_API_KEY is not configured"})
		return
	}

	// Build API URL
	forecastURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%.4f&lon=%.4f&appid=%s&units=metric", lat, lon, h.openWeatherAPIKey)

	// Make API request
	resp, err := http.Get(forecastURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to connect to forecast API: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("forecast API returned status code %d", resp.StatusCode)})
		return
	}

	// Decode response
	var forecastData weather_models.Forecast
	err = json.NewDecoder(resp.Body).Decode(&forecastData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to process forecast data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, forecastData)
}

// VerifyEmail handles OTP verification and activates the account
func (h *AuthController) VerifyEmail(c *gin.Context) {
	var req service.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, tokenPair, err := h.authService.VerifyEmail(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set refresh token as HTTP-only cookie
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(time.Unix(tokenPair.ExpiresAt, 0)).Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, response)
}

// ResendOTP sends a new OTP to the user's email
func (h *AuthController) ResendOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResendOTP(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to your email"})
}

// ForgotPassword initiates password reset flow
func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req struct {
		Email           string `json:"email" binding:"required"`
		TurnstileToken  string `json:"turnstile_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.verifyTurnstile(req.TurnstileToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Idempotent - always return success to prevent email enumeration
	_ = h.authService.RequestPasswordReset(c.Request.Context(), req.Email)

	c.JSON(http.StatusOK, gin.H{"message": "If an account exists with that email, you will receive a password reset link."})
}

// ResetPassword resets password using the token from the email link
func (h *AuthController) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully. You can now sign in."})
}

// RegisterRoutes registers the auth routes with Gin
func (h *AuthController) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Public routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshTokens)
		auth.POST("/logout", h.Logout)
		auth.POST("/verify-email", h.VerifyEmail)
		auth.POST("/resend-otp", h.ResendOTP)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
	}

	// Protected routes
	protected := auth.Group("", authMiddleware.Authenticate())
	{
		protected.GET("/profile", h.Profile)
		protected.PATCH("/profile", h.UpdateProfile)
		protected.PUT("/location", h.UpdateLocation)
		protected.GET("/weather", h.GetCurrentWeather)
		protected.GET("/weather/forecast", h.GetWeatherForecast)
	}

	// Admin-only routes (requires authentication + admin role)
	adminOnly := auth.Group("", authMiddleware.Authenticate(), authMiddleware.RequireAdmin())
	{
		adminOnly.POST("/register/admin", h.RegisterAdmin)
	}
}
