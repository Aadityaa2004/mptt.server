package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/cooldown"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/handlers"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
)

func main() {
	port := getEnv("PORT", "9004")

	// SMTP config
	smtpHost := getEnv("SMTP_HOST", "smtp-relay.brevo.com")
	smtpPort := getIntEnv("SMTP_PORT", 587)
	smtpEncryption := getEnv("SMTP_ENCRYPTION", "starttls")
	smtpUsername := getEnv("SMTP_USERNAME", "")
	smtpPassword := getEnv("SMTP_PASSWORD", "")
	fromName := getEnv("EMAIL_FROM_NAME", "MapleSense Alerts")
	fromAddress := getEnv("EMAIL_FROM_ADDRESS", "")

	// Cooldown config
	cooldownMinutes := getIntEnv("ALERT_COOLDOWN_MINUTES", 720)
	resetThreshold := getIntEnv("ALERT_RESET_THRESHOLD", 70)

	emailSvc := service.NewEmailService(smtpHost, smtpPort, smtpEncryption, smtpUsername, smtpPassword, fromName, fromAddress)
	cooldownTracker := cooldown.NewTracker(cooldownMinutes, resetThreshold)
	alertHandler := handlers.NewAlertHandler(emailSvc, cooldownTracker)

	router := gin.Default()
	router.POST("/alerts/bucket-fill", alertHandler.HandleBucketFillAlert)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("MQT.EmailService starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return i
}
