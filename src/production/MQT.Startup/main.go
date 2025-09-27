package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	mqtingestor "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Ingestor"
	implementation "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Implementation"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Startup/health"
)

func main() {
	// Load configuration
	cfg := mqtingestor.Load()

	// Connect to MongoDB using the standalone health package
	mc, err := health.ConnectDBWithTimeout(20 * time.Second)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	defer mc.Disconnect(context.Background())

	// Set the global client for health checks
	health.Client = mc

	// Get MongoDB collection and create repository using standalone function
	coll := health.GetCollection(mc)
	r := implementation.NewMongoReadingRepository(coll)

	// Create and start MQTT ingestor
	ing := mqtingestor.New(cfg, r)
	if err := ing.Start(context.Background()); err != nil {
		log.Fatal("Error starting MQTT ingestor:", err)
	}
	defer ing.Stop()

	// Initialize Gin router for health endpoints
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("PORT environment variable not set, using default 8080")
	}

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "mqtt-ingestor",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// Detailed health check
	router.GET("/health/detailed", func(c *gin.Context) {
		// Check MongoDB connection
		mongoStatus := "ok"
		if err := health.Client.Ping(context.Background(), nil); err != nil {
			mongoStatus = "error"
		}

		// Check MQTT connection
		mqttStatus := "ok"
		if !ing.IsConnected() {
			mqttStatus = "disconnected"
		}

		status := "ok"
		if mongoStatus != "ok" || mqttStatus != "ok" {
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":    status,
			"service":   "mqtt-ingestor",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
			"checks": gin.H{
				"mongodb": gin.H{
					"status": mongoStatus,
				},
				"mqtt": gin.H{
					"status": mqttStatus,
				},
			},
		})
	})

	// Readiness check
	router.GET("/ready", func(c *gin.Context) {
		// Check if all dependencies are ready
		if err := health.Client.Ping(context.Background(), nil); err != nil {
			c.JSON(503, gin.H{
				"status":  "not ready",
				"reason":  "mongodb connection failed",
				"details": err.Error(),
			})
			return
		}

		if !ing.IsConnected() {
			c.JSON(503, gin.H{
				"status": "not ready",
				"reason": "mqtt connection failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ready",
		})
	})

	// Start HTTP server in a goroutine
	go func() {
		fmt.Printf("HTTP server starting on port %s...\n", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatal("Error starting HTTP server:", err)
		}
	}()

	log.Println("MQTT ingestor running... press Ctrl+C to stop")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
}
