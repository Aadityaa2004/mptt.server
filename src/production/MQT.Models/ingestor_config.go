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
