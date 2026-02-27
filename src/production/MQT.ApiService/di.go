package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/controllers"
	authService "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"
	authMiddleware "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	container "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Container"
	implementation "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Implementation"
)

// InitializeServer wires all API dependencies and returns a configured HTTP server
// together with the underlying ApiContainer.
func InitializeServer(ctx context.Context) (*http.Server, *container.ApiContainer, error) {
	// Initialize dependency container
	ctr, err := container.NewApiContainer()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	logger := ctr.GetLogger()
	logger.Info("Starting API Service")

	// Initialize database schema
	if err := ctr.InitializeDatabase(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Get database connection
	db, err := ctr.GetDatabase()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create repositories
	readingRepo := implementation.NewPostgresReadingRepository(db)
	userRepo := implementation.NewPostgresUserRepository(db)
	piRepo := implementation.NewPostgresPiRepository(db)
	deviceRepo := implementation.NewPostgresDeviceRepository(db)
	roleRepo := implementation.NewPostgresRoleRepository(db)
	verificationTokenRepo := implementation.NewPostgresVerificationTokenRepository(db)

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

	// Initialize verification service for OTP and password reset
	verificationService := authService.NewVerificationService(
		verificationTokenRepo,
		config.Alert.EmailServiceURL,
		config.FrontendBaseURL,
	)

	// Initialize auth services
	authServiceInstance := authService.NewAuthService(userRepo, roleRepo, jwtService, rbacService, verificationService)
	userServiceInstance := authService.NewUserService(userRepo, piRepo)

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
		return nil, nil, fmt.Errorf("failed to initialize roles: %w", err)
	}
	if err := roleInitializer.InitializeAdminUser(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize admin user: %w", err)
	}

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
	authController := controllers.NewAuthController(authServiceInstance, config.OpenWeatherAPIKey, config.TurnstileSecretKey)
	userController := controllers.NewUserController(userServiceInstance)
	locationController := controllers.NewLocationController(authServiceInstance, deviceRepo)
	piController := controllers.NewPiController(piRepo, userRepo, logger, authMiddlewareInstance)
	deviceController := controllers.NewDeviceController(deviceRepo, piRepo, logger, authMiddlewareInstance)
	readingController := controllers.NewReadingController(readingRepo, piRepo, deviceRepo, logger, authMiddlewareInstance)
	healthController := controllers.NewHealthController(readingRepo, piRepo, logger, authMiddlewareInstance)
	internalController := controllers.NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, config)

	// Register all routes
	authController.RegisterRoutes(router, authMiddlewareInstance)
	userController.RegisterRoutes(router, authMiddlewareInstance)
	locationController.RegisterRoutes(router, authMiddlewareInstance)
	piController.RegisterRoutes(router)
	deviceController.RegisterRoutes(router)
	readingController.RegisterRoutes(router)
	healthController.RegisterRoutes(router)
	internalController.RegisterRoutes(router)

	// HTTP server with timeouts from config
	port := config.Server.Port
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	return srv, ctr, nil
}

