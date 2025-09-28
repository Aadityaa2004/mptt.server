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
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Startup/controllers"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Startup/health"
)

func main() {
	// Load configuration
	cfg := mqtingestor.Load()

	// Connect to PostgreSQL using the standalone health package
	db, err := health.ConnectPostgresWithTimeout(20 * time.Second)
	if err != nil {
		log.Fatal("Error connecting to PostgreSQL:", err)
	}
	defer db.Close()

	// Set the global client for health checks
	health.DB = db

	// Create tables if they don't exist
	if err := health.CreateTables(db); err != nil {
		log.Fatal("Error creating tables:", err)
	}

	// Create repositories
	readingRepo := implementation.NewPostgresReadingRepository(db)
	userRepo := implementation.NewPostgresUserRepository(db)
	piRepo := implementation.NewPostgresPiRepository(db)
	deviceRepo := implementation.NewPostgresDeviceRepository(db)

	// Create and start MQTT ingestor
	ing := mqtingestor.New(cfg, readingRepo, piRepo, deviceRepo)
	if err := ing.Start(context.Background()); err != nil {
		log.Fatal("Error starting MQTT ingestor:", err)
	}
	defer ing.Stop()

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Create controllers and setup routes
	controllersInstance := controllers.NewControllers(userRepo, piRepo, deviceRepo, readingRepo)
	controllers.SetupRepoRoutes(router, controllersInstance)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("PORT environment variable not set, using default 8080")
	}

	// Legacy health endpoints (for backward compatibility)
	router.GET("/health", func(c *gin.Context) {
		// Check PostgreSQL connection
		postgresStatus := "ok"
		if err := health.PingPostgres(context.Background()); err != nil {
			postgresStatus = "error"
		}

		// Check MQTT connection
		mqttStatus := "ok"
		if !ing.IsConnected() {
			mqttStatus = "disconnected"
		}

		status := "ok"
		if postgresStatus != "ok" || mqttStatus != "ok" {
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":    status,
			"service":   "mqtt-ingestor",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
			"checks": gin.H{
				"postgres": gin.H{
					"status": postgresStatus,
				},
				"mqtt": gin.H{
					"status": mqttStatus,
				},
			},
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check if all dependencies are ready
		if err := health.PingPostgres(context.Background()); err != nil {
			c.JSON(503, gin.H{
				"status":  "not ready",
				"reason":  "postgresql connection failed",
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
