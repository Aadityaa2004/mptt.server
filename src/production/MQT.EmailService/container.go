package main

import (
	"context"
	"fmt"

	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/cooldown"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/handlers"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
)

// EmailContainer manages dependencies for the Email service
type EmailContainer struct {
	config            *config.EmailServiceConfig
	logger            *logger.Logger
	emailService      *service.EmailService
	cooldownTracker   *cooldown.Tracker
	emptyAlertTracker *cooldown.Tracker
	emailHandler      *handlers.EmailHandler
	alertHandler      *handlers.AlertHandler
}

// NewEmailContainer creates a new Email service container
func NewEmailContainer() (*EmailContainer, error) {
	cfg, err := config.LoadEmailConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load email config: %w", err)
	}

	log := logger.NewLogger(&cfg.Logging)

	emailSvc := service.NewEmailService(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPEncryption,
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.FromName,
		cfg.Email.FromAddress,
	)

	cooldownTracker := cooldown.NewTracker(cfg.Alert.CooldownMinutes, cfg.Alert.ResetThreshold, false)
	emptyAlertTracker := cooldown.NewTracker(cfg.Alert.EmptyCooldownMinutes, cfg.Alert.EmptyResetThreshold, true)
	emailHandler := handlers.NewEmailHandler(emailSvc, log)
	alertHandler := handlers.NewAlertHandler(emailSvc, cooldownTracker, emptyAlertTracker, log)

	return &EmailContainer{
		config:            cfg,
		logger:            log,
		emailService:      emailSvc,
		cooldownTracker:   cooldownTracker,
		emptyAlertTracker: emptyAlertTracker,
		emailHandler:      emailHandler,
		alertHandler:      alertHandler,
	}, nil
}

// GetConfig returns the configuration
func (c *EmailContainer) GetConfig() *config.EmailServiceConfig {
	return c.config
}

// GetLogger returns the logger
func (c *EmailContainer) GetLogger() *logger.Logger {
	return c.logger
}

// GetEmailHandler returns the email handler
func (c *EmailContainer) GetEmailHandler() *handlers.EmailHandler {
	return c.emailHandler
}

// GetAlertHandler returns the alert handler
func (c *EmailContainer) GetAlertHandler() *handlers.AlertHandler {
	return c.alertHandler
}

// Shutdown gracefully shuts down the container
func (c *EmailContainer) Shutdown(_ context.Context) error {
	c.logger.Info("Email container shutdown complete")
	return nil
}
