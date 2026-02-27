package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/client"
	mqtingestor "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/ingestor"
)

func main() {
	// Use a bounded context for ingestor lifetime
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ing, cfg, apiClient, logger, shutdown, err := InitializeIngestorApp()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize MQTT Ingestor Service: %v", err))
	}
	defer shutdown(context.Background())

	// Create and start MQTT ingestor
	if err := ing.Start(ctx); err != nil {
		logger.FatalWithError(err, "Failed to start MQTT ingestor")
	}

	// Start health check server
	go startHealthServer(cfg.Server.Port, logger, ing, apiClient)

	logger.Info("MQTT ingestor running... press Ctrl+C to stop")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down MQTT Ingestor Service...")
}

// startHealthServer starts a simple HTTP server for health checks
func startHealthServer(port string, logger mqtingestorLogger, ing *mqtingestor.Ingestor, apiClient client.APIReadingsClient) {
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

	logger.Info("Health server starting on port " + port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.FatalWithError(err, "Failed to start health server")
	}
}

// mqtingestorLogger is the subset of logger.Logger used by startHealthServer.
type mqtingestorLogger interface {
	Info(msg string)
	FatalWithError(err error, msg string)
}

