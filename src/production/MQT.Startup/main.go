package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	// "github.com/sirupsen/logrus"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Startup/health"
)

func main() {

	err := health.ConnectDB()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT environment variable not set")
	}

	// Simple health endpoint and debug run-once route
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Start the server
	fmt.Printf("Server is starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
