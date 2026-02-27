package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
)

// InitializeEmailApp wires all Email service dependencies and returns
// a configured Gin engine, service config, logger, and a shutdown function.
func InitializeEmailApp() (*gin.Engine, *config.EmailServiceConfig, *logger.Logger, func(context.Context) error, error) {
	ctr, err := NewEmailContainer()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize email container: %w", err)
	}

	cfg := ctr.GetConfig()
	log := ctr.GetLogger()
	log.Info("Initializing MQT Email Service")

	emailHandler := ctr.GetEmailHandler()
	alertHandler := ctr.GetAlertHandler()

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST("/alerts/bucket-fill", alertHandler.HandleBucketFillAlert)
	router.POST("/send-otp", emailHandler.SendOTP)
	router.POST("/send-password-reset", emailHandler.SendPasswordReset)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	shutdown := func(ctx context.Context) error {
		return ctr.Shutdown(ctx)
	}

	return router, cfg, log, shutdown, nil
}

