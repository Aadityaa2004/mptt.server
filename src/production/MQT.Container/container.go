package container

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/health"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
)

// Container manages dependencies and their lifecycle
type Container struct {
	config *config.Config
	logger *logger.Logger
	db     *sql.DB

	// Health components
	healthChecker   *health.HealthChecker
	databaseManager *health.DatabaseManager

	// Services will be added here as they are implemented
	services map[string]interface{}

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Cleanup functions
	cleanupFuncs []func() error
}

// IngestorContainer manages dependencies for the MQTT Ingestor service
type IngestorContainer struct {
	config *config.IngestorConfig
	logger *logger.Logger
}

// ApiContainer manages dependencies for the API service
type ApiContainer struct {
	*Container
}

// NewContainer creates a new dependency injection container
func NewContainer() (*Container, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	log := logger.NewLogger(&cfg.Logging)

	container := &Container{
		config:   cfg,
		logger:   log,
		services: make(map[string]interface{}),
	}

	// Register cleanup functions
	container.registerCleanup()

	return container, nil
}

// NewIngestorContainer creates a new container for the MQTT Ingestor service
func NewIngestorContainer() (*IngestorContainer, error) {
	// Load ingestor-specific configuration
	cfg, err := config.LoadIngestorConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load ingestor configuration: %w", err)
	}

	// Initialize logger
	log := logger.NewLogger(&cfg.Logging)

	return &IngestorContainer{
		config: cfg,
		logger: log,
	}, nil
}

// NewApiContainer creates a new container for the API service
func NewApiContainer() (*ApiContainer, error) {
	// Load API-specific configuration
	cfg, err := config.LoadApiConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load API configuration: %w", err)
	}

	// Initialize logger
	log := logger.NewLogger(&cfg.Logging)

	baseContainer := &Container{
		config:   cfg,
		logger:   log,
		services: make(map[string]interface{}),
	}

	// Register cleanup functions
	baseContainer.registerCleanup()

	return &ApiContainer{Container: baseContainer}, nil
}

// GetConfig returns the configuration
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetConfig returns the ingestor configuration
func (c *IngestorContainer) GetConfig() *config.IngestorConfig {
	return c.config
}

// GetLogger returns the logger
func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

// GetLogger returns the logger
func (c *IngestorContainer) GetLogger() *logger.Logger {
	return c.logger
}

// GetDatabase returns the database connection
func (c *Container) GetDatabase() (*sql.DB, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		db, err := health.ConnectPostgresWithTimeout(c.config, 20*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		c.db = db
	}

	return c.db, nil
}

// GetHealthChecker returns the health checker
func (c *Container) GetHealthChecker() (*health.HealthChecker, error) {
	c.mu.Lock()
	if c.healthChecker != nil {
		c.mu.Unlock()
		return c.healthChecker, nil
	}
	c.mu.Unlock()

	// Get database without holding the lock to avoid deadlock
	db, err := c.GetDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database for health checker: %w", err)
	}

	// Now acquire lock again to set the health checker
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.healthChecker == nil {
		c.healthChecker = health.NewHealthChecker(db)
	}

	return c.healthChecker, nil
}

// GetDatabaseManager returns the database manager
func (c *Container) GetDatabaseManager() (*health.DatabaseManager, error) {
	c.mu.Lock()
	if c.databaseManager != nil {
		c.mu.Unlock()
		return c.databaseManager, nil
	}
	c.mu.Unlock()

	// Get database without holding the lock to avoid deadlock
	db, err := c.GetDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database for database manager: %w", err)
	}

	// Now acquire lock again to set the database manager
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.databaseManager == nil {
		c.databaseManager = health.NewDatabaseManager(db)
	}

	return c.databaseManager, nil
}

// RegisterService registers a service in the container
func (c *Container) RegisterService(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// GetService retrieves a service from the container
func (c *Container) GetService(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	service, exists := c.services[name]
	return service, exists
}

// InitializeDatabase initializes the database and creates tables
func (c *Container) InitializeDatabase(ctx context.Context) error {
	dbManager, err := c.GetDatabaseManager()
	if err != nil {
		return fmt.Errorf("failed to get database manager: %w", err)
	}

	if err := dbManager.CreateTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	c.logger.Info("Database initialized successfully")
	return nil
}

// HealthCheck performs a comprehensive health check
func (c *Container) HealthCheck(ctx context.Context) map[string]interface{} {
	healthChecker, err := c.GetHealthChecker()
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	return healthChecker.GetHealthStatus(ctx)
}

// Shutdown gracefully shuts down the container and all its dependencies
func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down container...")

	// Execute cleanup functions in reverse order
	for i := len(c.cleanupFuncs) - 1; i >= 0; i-- {
		if err := c.cleanupFuncs[i](); err != nil {
			c.logger.ErrorWithError(err, "Error during cleanup")
		}
	}

	// Close database connection
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			c.logger.ErrorWithError(err, "Error closing database connection")
		}
	}

	c.logger.Info("Container shutdown complete")
	return nil
}

// Shutdown gracefully shuts down the ingestor container
func (c *IngestorContainer) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down ingestor container...")
	c.logger.Info("Ingestor container shutdown complete")
	return nil
}

// registerCleanup registers cleanup functions
func (c *Container) registerCleanup() {
	c.cleanupFuncs = append(c.cleanupFuncs, func() error {
		if c.db != nil {
			return c.db.Close()
		}
		return nil
	})
}

// AddCleanupFunc adds a cleanup function
func (c *Container) AddCleanupFunc(fn func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanupFuncs = append(c.cleanupFuncs, fn)
}
