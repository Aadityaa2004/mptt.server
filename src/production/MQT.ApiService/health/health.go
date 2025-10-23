package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db *sql.DB
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// PingPostgres checks if the PostgreSQL connection is healthy
func (h *HealthChecker) PingPostgres(ctx context.Context) error {
	if h.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return h.db.PingContext(ctx)
}

// CheckDatabaseHealth performs a comprehensive database health check
func (h *HealthChecker) CheckDatabaseHealth(ctx context.Context) error {
	if err := h.PingPostgres(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	err := h.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

// GetHealthStatus returns the current health status
func (h *HealthChecker) GetHealthStatus(ctx context.Context) map[string]interface{} {
	status := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"checks":    make(map[string]interface{}),
	}

	// Check database
	dbStatus := "ok"
	if err := h.CheckDatabaseHealth(ctx); err != nil {
		dbStatus = "error"
		status["checks"].(map[string]interface{})["postgres"] = map[string]interface{}{
			"status": dbStatus,
			"error":  err.Error(),
		}
	} else {
		status["checks"].(map[string]interface{})["postgres"] = map[string]interface{}{
			"status": dbStatus,
		}
	}

	// Overall status
	overallStatus := "ok"
	if dbStatus != "ok" {
		overallStatus = "degraded"
	}
	status["status"] = overallStatus

	return status
}

// DatabaseManager handles database operations
type DatabaseManager struct {
	db *sql.DB
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(db *sql.DB) *DatabaseManager {
	return &DatabaseManager{db: db}
}

// ConnectPostgresWithTimeout creates a PostgreSQL connection with a timeout context
func ConnectPostgresWithTimeout(cfg *config.Config, timeout time.Duration) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := sql.Open("postgres", cfg.GetDatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("unable to open PostgreSQL connection: %w", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping PostgreSQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MinConns)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// CreateTables creates the required tables if they don't exist
func (dm *DatabaseManager) CreateTables(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create users table
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			user_id     TEXT PRIMARY KEY,
			username    TEXT NOT NULL UNIQUE,
			email       TEXT NOT NULL UNIQUE,
			password    TEXT NOT NULL,
			role        TEXT NOT NULL,
			active      BOOLEAN NOT NULL DEFAULT true,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`

	// Create pis table
	createPisTable := `
		CREATE TABLE IF NOT EXISTS pis (
			pi_id       TEXT PRIMARY KEY,
			user_id     TEXT,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		);
	`

	// Create devices table
	createDevicesTable := `
		CREATE TABLE IF NOT EXISTS devices (
			pi_id       TEXT NOT NULL,
			device_id   INTEGER NOT NULL,
			device_type TEXT,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (pi_id, device_id),
			FOREIGN KEY (pi_id) REFERENCES pis(pi_id) ON DELETE CASCADE
		);
	`

	// Create readings table
	createReadingsTable := `
		CREATE TABLE IF NOT EXISTS readings (
			pi_id       TEXT NOT NULL,
			device_id   INTEGER NOT NULL,
			ts          TIMESTAMPTZ NOT NULL,
			payload     JSONB NOT NULL,
			PRIMARY KEY (pi_id, device_id, ts),
			FOREIGN KEY (pi_id, device_id) REFERENCES devices(pi_id, device_id) ON DELETE CASCADE
		);
	`

	// Create roles table
	createRolesTable := `
		CREATE TABLE IF NOT EXISTS roles (
			role_id     TEXT PRIMARY KEY,
			name        TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`

	// Create indexes
	createIndexes := `
		CREATE INDEX IF NOT EXISTS idx_readings_pi_device_ts_desc ON readings (pi_id, device_id, ts DESC);
		CREATE INDEX IF NOT EXISTS idx_readings_ts_desc ON readings (ts DESC);
		CREATE INDEX IF NOT EXISTS idx_readings_payload_gin ON readings USING GIN (payload);
		CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);
	`

	queries := []string{
		createUsersTable,
		createPisTable,
		createDevicesTable,
		createReadingsTable,
		createRolesTable,
		createIndexes,
	}

	for _, query := range queries {
		if _, err := dm.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}
