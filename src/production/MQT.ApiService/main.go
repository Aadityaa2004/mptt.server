package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Use a bounded context for initialization
	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv, ctr, err := InitializeServer(initCtx)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize API server: %v", err))
	}
	defer ctr.Shutdown(context.Background())

	logger := ctr.GetLogger()
	config := ctr.GetConfig()
	port := config.Server.Port

	// Start HTTP server in a goroutine
	go func() {
		logger.Info("HTTP server starting on port " + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.FatalWithError(err, "Failed to start HTTP server")
		}
	}()

	logger.Info("API service running... press Ctrl+C to stop")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.ErrorWithError(err, "Server forced to shutdown")
	}
}
