package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/controllers"
	container "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Container"
	implementation "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Implementation"

	// Auth imports
	authService "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"
	authMiddleware "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
)

func main() {
	// Initialize dependency injection container
	ctr, err := container.NewApiContainer()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize container: %v", err))
	}
	defer ctr.Shutdown(context.Background())

	logger := ctr.GetLogger()
	logger.Info("Starting API Service")

	// Initialize database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ctr.InitializeDatabase(ctx); err != nil {
		logger.FatalWithError(err, "Failed to initialize database")
	}

	// Get database connection
	db, err := ctr.GetDatabase()
	if err != nil {
		logger.FatalWithError(err, "Failed to get database connection")
	}

	// Create repositories
	readingRepo := implementation.NewPostgresReadingRepository(db)
	userRepo := implementation.NewPostgresUserRepository(db)
	piRepo := implementation.NewPostgresPiRepository(db)
	deviceRepo := implementation.NewPostgresDeviceRepository(db)
	roleRepo := implementation.NewPostgresRoleRepository(db)

	// Get configuration
	config := ctr.GetConfig()

	// Initialize JWT service for token validation
	jwtConfig := api_models.Config{
		SecretKey:            config.Auth.JWTSecretKey,
		AccessTokenDuration:  config.Auth.AccessTokenDuration,
		RefreshTokenDuration: config.Auth.RefreshTokenDuration,
		Issuer:               config.Auth.JWTIssuer,
	}
	jwtService := jwt.NewService(jwtConfig)

	// Initialize RBAC service
	rbacService := rbac.NewService()

	// Create auth middleware
	middlewareConfig := authMiddleware.Config{
		AccessTokenHeader: "Authorization",
		AccessTokenCookie: "access_token",
	}
	authMiddlewareInstance := authMiddleware.NewAuthMiddleware(jwtService, rbacService, middlewareConfig)

	// Initialize auth services
	authServiceInstance := authService.NewAuthService(userRepo, roleRepo, jwtService, rbacService)
	userServiceInstance := authService.NewUserService(userRepo)

	// Initialize role initializer
	roleInitializer := authService.NewRoleInitializerService(
		roleRepo,
		userRepo,
		rbacService,
		logger,
		authService.AdminConfig{
			Username: config.Auth.Admin.Username,
			Email:    config.Auth.Admin.Email,
			Password: config.Auth.Admin.Password,
		},
	)

	// Initialize roles and admin user
	if err := roleInitializer.InitializeRoles(ctx); err != nil {
		logger.FatalWithError(err, "Failed to initialize roles")
	}
	if err := roleInitializer.InitializeAdminUser(ctx); err != nil {
		logger.FatalWithError(err, "Failed to initialize admin user")
	}

	// Note: MQTT ingestor is now a separate service

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configure CORS from config
	corsConfig := cors.Config{
		AllowOrigins:     config.CORS.AllowedOrigins,
		AllowMethods:     config.CORS.AllowedMethods,
		AllowHeaders:     config.CORS.AllowedHeaders,
		ExposeHeaders:    config.CORS.ExposedHeaders,
		AllowCredentials: config.CORS.AllowCredentials,
		MaxAge:           time.Duration(config.CORS.MaxAge) * time.Second,
	}
	router.Use(cors.New(corsConfig))

	// Create controllers and register routes
	authController := controllers.NewAuthController(authServiceInstance)
	userController := controllers.NewUserController(userServiceInstance)
	piController := controllers.NewPiController(piRepo, userRepo, logger, authMiddlewareInstance)
	deviceController := controllers.NewDeviceController(deviceRepo, piRepo, logger, authMiddlewareInstance)
	readingController := controllers.NewReadingController(readingRepo, piRepo, logger, authMiddlewareInstance)
	healthController := controllers.NewHealthController(readingRepo, piRepo, logger, authMiddlewareInstance)
	internalController := controllers.NewInternalController(piRepo, deviceRepo, readingRepo)

	// Register all routes
	authController.RegisterRoutes(router, authMiddlewareInstance)
	userController.RegisterRoutes(router, authMiddlewareInstance)
	piController.RegisterRoutes(router)
	deviceController.RegisterRoutes(router)
	readingController.RegisterRoutes(router)
	healthController.RegisterRoutes(router)
	internalController.RegisterRoutes(router)

	// Get port from configuration
	port := config.Server.Port

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

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
