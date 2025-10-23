package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServiceAuthMiddleware validates service-to-service authentication
func ServiceAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Empty token",
			})
			c.Abort()
			return
		}

		// Get the expected secret from environment
		expectedSecret := os.Getenv("INTERNAL_API_SECRET")
		if expectedSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal API secret not configured",
			})
			c.Abort()
			return
		}

		// Validate the token
		if token != expectedSecret {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid service token",
			})
			c.Abort()
			return
		}

		// Add service context to the request
		c.Set("service_auth", true)
		c.Set("service_name", "mqtt-ingestor")

		// Continue to the next handler
		c.Next()
	}
}
