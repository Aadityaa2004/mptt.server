package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	container "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Container"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/client"
	mqtingestor "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/ingestor"
)

func main() {
	// Initialize dependency injection container
	ctr, err := container.NewIngestorContainer()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize container: %v", err))
	}
	defer ctr.Shutdown(context.Background())

	logger := ctr.GetLogger()
	logger.Info("Starting MQTT Ingestor Service")

	// Get configuration
	config := ctr.GetConfig()

	// Create API client
	apiClient := client.NewAPIClient(config.ApiServiceURL, config.InternalAPISecret)

	// Create MQTT ingestor configuration from environment
	cfg := mqtingestor.LoadFromEnv()

	// Create and start MQTT ingestor
	ing := mqtingestor.New(cfg, apiClient, logger)
	if err := ing.Start(context.Background()); err != nil {
		logger.FatalWithError(err, "Failed to start MQTT ingestor")
	}
	defer ing.Stop()

	// Start health check server
	go startHealthServer(ctr, ing, apiClient)

	logger.Info("MQTT ingestor running... press Ctrl+C to stop")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")
}

// startHealthServer starts a simple HTTP server for health checks
func startHealthServer(ctr *container.IngestorContainer, ing *mqtingestor.Ingestor, apiClient *client.APIClient) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Check MQTT connection
		mqttStatus := "disconnected"
		if ing.IsConnected() {
			mqttStatus = "connected"
		}

		// Check API service connection
		apiStatus := "disconnected"
		if err := apiClient.Health(ctx); err == nil {
			apiStatus = "connected"
		}

		// Return health status
		status := "healthy"
		if mqttStatus != "connected" || apiStatus != "connected" {
			status = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		if status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Get circuit breaker status
		circuitBreakerStatus := apiClient.GetCircuitBreakerStatus()

		fmt.Fprintf(w, `{
			"status": "%s",
			"timestamp": "%s",
			"services": {
				"mqtt": "%s",
				"api_service": "%s"
			},
			"circuit_breaker": {
				"state": "%s",
				"failure_count": %d
			}
		}`, status, time.Now().UTC().Format(time.RFC3339), mqttStatus, apiStatus,
			circuitBreakerStatus["state"], circuitBreakerStatus["failure_count"])
	})

	port := ctr.GetConfig().Server.Port
	logger := ctr.GetLogger()
	logger.Info("Health server starting on port " + port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.FatalWithError(err, "Failed to start health server")
	}
}
