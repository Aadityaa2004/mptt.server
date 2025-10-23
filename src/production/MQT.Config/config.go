package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// MQTT configuration
	MQTT MQTTConfig `json:"mqtt"`

	// Auth configuration
	Auth AuthConfig `json:"auth"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`

	// CORS configuration
	CORS CORSConfig `json:"cors"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
	MaxConns int    `json:"max_conns"`
	MinConns int    `json:"min_conns"`
}

// MQTTConfig holds MQTT-related configuration
type MQTTConfig struct {
	BrokerHost  string        `json:"broker_host"`
	BrokerPort  int           `json:"broker_port"`
	BrokerUser  string        `json:"broker_user"`
	BrokerPass  string        `json:"broker_pass"`
	UseTLS      bool          `json:"use_tls"`
	CACertPath  string        `json:"ca_cert_path"`
	Topic       string        `json:"topic"`
	ClientID    string        `json:"client_id"`
	SharedGroup string        `json:"shared_group"`
	KeepAlive   time.Duration `json:"keep_alive"`
	PingTimeout time.Duration `json:"ping_timeout"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecretKey               string        `json:"jwt_secret_key"`
	JWTIssuer                  string        `json:"jwt_issuer"`
	AccessTokenDuration        time.Duration `json:"access_token_duration"`
	RefreshTokenDuration       time.Duration `json:"refresh_token_duration"`
	PasswordMinLength          int           `json:"password_min_length"`
	PasswordRequireSpecialChar bool          `json:"password_require_special_char"`
	Admin                      AdminConfig   `json:"admin"`
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level        string `json:"level"`
	Format       string `json:"format"` // json or text
	Output       string `json:"output"` // stdout, stderr, or file path
	EnableCaller bool   `json:"enable_caller"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// BatchConfig holds batch processing configuration
type BatchConfig struct {
	Size   int           `json:"size"`
	Window time.Duration `json:"window"`
}

// IngestorConfig holds configuration for the MQTT Ingestor service
type IngestorConfig struct {
	Server            ServerConfig  `json:"server"`
	MQTT              MQTTConfig    `json:"mqtt"`
	Logging           LoggingConfig `json:"logging"`
	ApiServiceURL     string        `json:"api_service_url"`
	InternalAPISecret string        `json:"internal_api_secret"`
}

// LoadIngestorConfig loads configuration for the MQTT Ingestor service
func LoadIngestorConfig() (*IngestorConfig, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		// Silently ignore .env file loading errors
		// This allows the application to work with environment variables set directly
	}

	config := &IngestorConfig{
		Server: ServerConfig{
			Port:         getEnv("INGESTOR_PORT", "9003"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 120*time.Second),
		},
		MQTT: MQTTConfig{
			BrokerHost:  getEnv("BROKER_HOST", "localhost"),
			BrokerPort:  getInt("BROKER_PORT", 1883),
			BrokerUser:  getEnv("BROKER_USER", ""),
			BrokerPass:  getEnv("BROKER_PASS", ""),
			UseTLS:      getBool("BROKER_TLS", false),
			CACertPath:  getEnv("BROKER_CA_FILE", ""),
			Topic:       getEnv("MQTT_TOPIC", "sensors/#"),
			ClientID:    getEnv("MQTT_CLIENT_ID", "mqtt-ingestor"),
			SharedGroup: getEnv("MQTT_SHARED_GROUP", ""),
			KeepAlive:   getDuration("MQTT_KEEP_ALIVE", 30*time.Second),
			PingTimeout: getDuration("MQTT_PING_TIMEOUT", 10*time.Second),
		},
		Logging: LoggingConfig{
			Level:        getEnv("LOG_LEVEL", "info"),
			Format:       getEnv("LOG_FORMAT", "text"),
			Output:       getEnv("LOG_OUTPUT", "stdout"),
			EnableCaller: getBool("LOG_ENABLE_CALLER", false),
		},
		ApiServiceURL:     getEnv("API_SERVICE_URL", "http://api-service:9002"),
		InternalAPISecret: getRequiredEnv("INTERNAL_API_SECRET"),
	}

	// Validate configuration
	if config.ApiServiceURL == "" {
		return nil, fmt.Errorf("API_SERVICE_URL is required")
	}
	if config.InternalAPISecret == "" {
		return nil, fmt.Errorf("INTERNAL_API_SECRET is required")
	}

	return config, nil
}

// LoadApiConfig loads configuration for the API service
func LoadApiConfig() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		// Silently ignore .env file loading errors
		// This allows the application to work with environment variables set directly
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "9002"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getInt("POSTGRES_PORT", 5432),
			User:     getRequiredEnv("POSTGRES_USER"),
			Password: getRequiredEnv("POSTGRES_PASSWORD"),
			DBName:   getEnv("POSTGRES_DB", "iot"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: getInt("POSTGRES_MAX_CONNS", 25),
			MinConns: getInt("POSTGRES_MIN_CONNS", 5),
		},
		// Minimal MQTT config for API (not used but required for struct)
		MQTT: MQTTConfig{
			BrokerHost:  getEnv("BROKER_HOST", "localhost"),
			BrokerPort:  getInt("BROKER_PORT", 1883),
			BrokerUser:  getEnv("BROKER_USER", ""),
			BrokerPass:  getEnv("BROKER_PASS", ""),
			UseTLS:      getBool("BROKER_TLS", false),
			CACertPath:  getEnv("BROKER_CA_FILE", ""),
			Topic:       getEnv("MQTT_TOPIC", "sensors/#"),
			ClientID:    getEnv("MQTT_CLIENT_ID", "api-service"),
			SharedGroup: getEnv("MQTT_SHARED_GROUP", ""),
			KeepAlive:   getDuration("MQTT_KEEP_ALIVE", 30*time.Second),
			PingTimeout: getDuration("MQTT_PING_TIMEOUT", 10*time.Second),
		},
		Auth: AuthConfig{
			JWTSecretKey:               getEnv("JWT_SECRET_KEY", "change-this-secret-in-production"),
			JWTIssuer:                  getEnv("JWT_ISSUER", "mpt-api-service"),
			AccessTokenDuration:        getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration:       getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			PasswordMinLength:          getInt("PASSWORD_MIN_LENGTH", 8),
			PasswordRequireSpecialChar: getBool("PASSWORD_REQUIRE_SPECIAL_CHAR", true),
			Admin: AdminConfig{
				Username: getEnv("ADMIN_USERNAME", "admin"),
				Email:    getEnv("ADMIN_EMAIL", "admin@example.com"),
				Password: getEnv("ADMIN_PASSWORD", "adminpassword123"),
			},
		},
		Logging: LoggingConfig{
			Level:        getEnv("LOG_LEVEL", "info"),
			Format:       getEnv("LOG_FORMAT", "text"),
			Output:       getEnv("LOG_OUTPUT", "stdout"),
			EnableCaller: getBool("LOG_ENABLE_CALLER", false),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods:   getStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getStringSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "token"}),
			ExposedHeaders:   getStringSlice("CORS_EXPOSED_HEADERS", []string{"Content-Length"}),
			AllowCredentials: getBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getInt("CORS_MAX_AGE", 43200), // 12 hours
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Load loads configuration from environment variables with fallback defaults
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		// Silently ignore .env file loading errors
		// This allows the application to work with environment variables set directly
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getInt("POSTGRES_PORT", 5432),
			User:     getRequiredEnv("POSTGRES_USER"),
			Password: getRequiredEnv("POSTGRES_PASSWORD"),
			DBName:   getEnv("POSTGRES_DB", "iot"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: getInt("POSTGRES_MAX_CONNS", 25),
			MinConns: getInt("POSTGRES_MIN_CONNS", 5),
		},
		MQTT: MQTTConfig{
			BrokerHost:  getEnv("BROKER_HOST", "localhost"),
			BrokerPort:  getInt("BROKER_PORT", 1883),
			BrokerUser:  getEnv("BROKER_USER", ""),
			BrokerPass:  getEnv("BROKER_PASS", ""),
			UseTLS:      getBool("BROKER_TLS", false),
			CACertPath:  getEnv("BROKER_CA_FILE", ""),
			Topic:       getEnv("MQTT_TOPIC", "sensors/#"),
			ClientID:    getEnv("MQTT_CLIENT_ID", "mqtt-ingestor"),
			SharedGroup: getEnv("MQTT_SHARED_GROUP", ""),
			KeepAlive:   getDuration("MQTT_KEEP_ALIVE", 30*time.Second),
			PingTimeout: getDuration("MQTT_PING_TIMEOUT", 10*time.Second),
		},
		Auth: AuthConfig{
			JWTSecretKey:               getEnv("JWT_SECRET_KEY", "change-this-secret-in-production"),
			JWTIssuer:                  getEnv("JWT_ISSUER", "mpt-auth-service"),
			AccessTokenDuration:        getDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration:       getDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			PasswordMinLength:          getInt("PASSWORD_MIN_LENGTH", 8),
			PasswordRequireSpecialChar: getBool("PASSWORD_REQUIRE_SPECIAL_CHAR", true),
			Admin: AdminConfig{
				Username: getEnv("ADMIN_USERNAME", "admin"),
				Email:    getEnv("ADMIN_EMAIL", "admin@example.com"),
				Password: getEnv("ADMIN_PASSWORD", "adminpassword123"),
			},
		},
		Logging: LoggingConfig{
			Level:        getEnv("LOG_LEVEL", "info"),
			Format:       getEnv("LOG_FORMAT", "text"),
			Output:       getEnv("LOG_OUTPUT", "stdout"),
			EnableCaller: getBool("LOG_ENABLE_CALLER", false),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods:   getStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getStringSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "token"}),
			ExposedHeaders:   getStringSlice("CORS_EXPOSED_HEADERS", []string{"Content-Length"}),
			AllowCredentials: getBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getInt("CORS_MAX_AGE", 43200), // 12 hours
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.User == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}
	if c.Auth.JWTSecretKey == "change-this-secret-in-production" {
		log.Println("WARNING: Using default JWT secret key. Change JWT_SECRET_KEY in production!")
	}
	if c.Auth.PasswordMinLength < 6 {
		return fmt.Errorf("password minimum length must be at least 6")
	}
	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.DBName, c.Database.SSLMode)
}

// GetMQTTBrokerURL returns the MQTT broker URL
func (c *Config) GetMQTTBrokerURL() string {
	scheme := "tcp"
	if c.MQTT.UseTLS {
		scheme = "tcps"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, c.MQTT.BrokerHost, c.MQTT.BrokerPort)
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return value
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("invalid %s: %v", key, err)
	}
	return intValue
}

func getBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if value == "1" || value == "true" || value == "TRUE" {
		return true
	}
	if value == "0" || value == "false" || value == "FALSE" {
		return false
	}
	log.Fatalf("invalid %s: %q (expected true/false or 1/0)", key, value)
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("invalid %s: %v", key, err)
	}
	return duration
}

func getStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// Simple comma-separated parsing
	// For more complex parsing, consider using a proper CSV library
	parts := make([]string, 0)
	for _, part := range splitString(value, ",") {
		if trimmed := trimString(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// Simple string splitting and trimming helpers
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
