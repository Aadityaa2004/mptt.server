package mqtingestor

import (
	"log"
	"os"
	"strconv"
	"time"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

func mustInt(env string, def int) int {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid %s: %v", env, err)
	}
	return i
}

func mustBool(env string, def bool) bool {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	if v == "1" || v == "true" || v == "TRUE" {
		return true
	}
	if v == "0" || v == "false" || v == "FALSE" {
		return false
	}
	log.Fatalf("invalid %s: %q", env, v)
	return def
}

func mustDur(env string, def time.Duration) time.Duration {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("invalid %s: %v", env, err)
	}
	return d
}

func Load() mqtmodels.IngestorConfig {
	return mqtmodels.IngestorConfig{
		BrokerHost:  os.Getenv("BROKER_HOST"),
		BrokerPort:  mustInt("BROKER_PORT", 1883),
		BrokerUser:  os.Getenv("BROKER_USER"),
		BrokerPass:  os.Getenv("BROKER_PASS"),
		UseTLS:      mustBool("BROKER_TLS", false),
		CACertPath:  os.Getenv("BROKER_CA_FILE"),
		Topic:       defaultStr("MQTT_TOPIC", "sensors/#"),
		ClientID:    defaultStr("MQTT_CLIENT_ID", "go-ingestor-1"),
		SharedGroup: os.Getenv("MQTT_SHARED_GROUP"),

		PostgresHost:     defaultStr("POSTGRES_HOST", "localhost"),
		PostgresPort:     mustInt("POSTGRES_PORT", 5432),
		PostgresUser:     required("POSTGRES_USER"),
		PostgresPassword: required("POSTGRES_PASSWORD"),
		PostgresDB:       defaultStr("POSTGRES_DB", "iot"),
		PostgresSSLMode:  defaultStr("POSTGRES_SSLMODE", "disable"),

		BatchSize:   mustInt("BATCH_SIZE", 200),
		BatchWindow: mustDur("BATCH_WINDOW", 1*time.Second),
	}
}

// LoadFromEnv loads configuration from environment variables for the new microservice architecture
func LoadFromEnv() mqtmodels.IngestorConfig {
	return mqtmodels.IngestorConfig{
		BrokerHost:  os.Getenv("BROKER_HOST"),
		BrokerPort:  mustInt("BROKER_PORT", 1883),
		BrokerUser:  os.Getenv("BROKER_USER"),
		BrokerPass:  os.Getenv("BROKER_PASS"),
		UseTLS:      mustBool("BROKER_TLS", false),
		CACertPath:  os.Getenv("BROKER_CA_FILE"),
		Topic:       defaultStr("MQTT_TOPIC", "sensors/#"),
		ClientID:    defaultStr("MQTT_CLIENT_ID", "mqtt-ingestor-1"),
		SharedGroup: os.Getenv("MQTT_SHARED_GROUP"),

		// No database configuration needed for microservice architecture
		BatchSize:   mustInt("BATCH_SIZE", 200),
		BatchWindow: mustDur("BATCH_WINDOW", 1*time.Second),
	}
}

func required(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing required env var %s", k)
	}
	return v
}
func defaultStr(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
