package mqtmodels

import "time"

type IngestorConfig struct {
	// MQTT
	BrokerHost  string
	BrokerPort  int
	BrokerUser  string
	BrokerPass  string
	UseTLS      bool
	CACertPath  string
	Topic       string
	ClientID    string
	SharedGroup string // e.g., "ingestors" to enable $share group consumption

	// PostgreSQL
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	// Ingestion
	BatchSize   int
	BatchWindow time.Duration
}

// NewIngestorConfig returns a new IngestorConfig with sensible defaults
func NewIngestorConfig() *IngestorConfig {
	return &IngestorConfig{
		// MQTT defaults
		BrokerPort: 8883, // Secure MQTT port
		UseTLS:     true,
		Topic:      "sensors/+/+/+", // pi_id/device_id/reading format
		ClientID:   "mqtt-ingestor",

		// PostgreSQL defaults
		PostgresPort:    5432,
		PostgresSSLMode: "require", // Secure by default in production

		// Ingestion defaults
		BatchSize:   1000,            // Batch 1000 readings at a time
		BatchWindow: 5 * time.Second, // Or flush every 5 seconds
	}
}
