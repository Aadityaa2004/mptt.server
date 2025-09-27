package mqtmodels

import "time"

type IngestorConfig struct {
	// MQTT
	BrokerHost   string
	BrokerPort   int
	BrokerUser   string
	BrokerPass   string
	UseTLS       bool
	CACertPath   string
	Topic        string
	ClientID     string
	SharedGroup  string // e.g., "ingestors" to enable $share group consumption

	// Mongo
	MongoURI     string
	DBName       string
	CollName     string

	// Ingestion
	BatchSize    int
	BatchWindow  time.Duration
}

