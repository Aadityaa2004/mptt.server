package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	router, cfg, logger, shutdown, err := InitializeEmailApp()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Email service: %v", err))
	}
	defer shutdown(context.Background())

	port := cfg.Server.Port

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("MQT Email Service starting on port " + port)
		if err := router.Run(":" + port); err != nil {
			logger.FatalWithError(err, "Failed to start Email HTTP server")
		}
	}()

	logger.Info("MQT Email Service running... press Ctrl+C to stop")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down Email service...")
}

