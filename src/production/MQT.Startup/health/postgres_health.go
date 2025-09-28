package health

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	// Try to load .env file, but don't fail if it doesn't exist
	// Environment variables can also be set directly
	if err := godotenv.Load(); err != nil {
		// Silently ignore .env file loading errors
		// This allows the application to work with environment variables set directly
	}
}

// ConnectPostgresWithTimeout creates a PostgreSQL connection with a timeout context using environment variables
func ConnectPostgresWithTimeout(timeout time.Duration) (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return nil, fmt.Errorf("POSTGRES_USER environment variable not set")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD environment variable not set")
	}

	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "iot"
	}

	sslmode := os.Getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open PostgreSQL connection: %v", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping PostgreSQL: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// CreateTables creates the required tables if they don't exist
func CreateTables(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create users table
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			user_id     TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			role        TEXT NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			meta        JSONB NOT NULL DEFAULT '{}'::jsonb
		);
	`

	// Create pis table
	createPisTable := `
		CREATE TABLE IF NOT EXISTS pis (
			pi_id       TEXT PRIMARY KEY,
			user_id     TEXT,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
			meta        JSONB NOT NULL DEFAULT '{}'::jsonb,
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
			meta        JSONB NOT NULL DEFAULT '{}'::jsonb,
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

	// Create indexes
	createIndexes := `
		CREATE INDEX IF NOT EXISTS idx_readings_pi_device_ts_desc ON readings (pi_id, device_id, ts DESC);
		CREATE INDEX IF NOT EXISTS idx_readings_ts_desc ON readings (ts DESC);
		CREATE INDEX IF NOT EXISTS idx_readings_payload_gin ON readings USING GIN (payload);
	`

	queries := []string{
		createUsersTable,
		createPisTable,
		createDevicesTable,
		createReadingsTable,
		createIndexes,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return nil
}

// PingPostgres checks if the PostgreSQL connection is healthy
func PingPostgres(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	return DB.PingContext(ctx)
}
