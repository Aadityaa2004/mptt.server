package mqtingestor

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	// Clear relevant env vars to test defaults
	orig := make(map[string]string)
	for _, k := range []string{"BROKER_HOST", "BROKER_PORT", "MQTT_TOPIC", "MQTT_CLIENT_ID", "BATCH_SIZE", "BATCH_WINDOW"} {
		orig[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	defer func() {
		for k, v := range orig {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	cfg := LoadFromEnv()

	if cfg.Topic != "sensors/#" {
		t.Errorf("Topic = %q, want sensors/#", cfg.Topic)
	}
	if cfg.ClientID != "mqtt-ingestor-1" {
		t.Errorf("ClientID = %q, want mqtt-ingestor-1", cfg.ClientID)
	}
	if cfg.BrokerPort != 1883 {
		t.Errorf("BrokerPort = %d, want 1883", cfg.BrokerPort)
	}
	if cfg.BatchSize != 200 {
		t.Errorf("BatchSize = %d, want 200", cfg.BatchSize)
	}
	if cfg.BatchWindow != 1*time.Second {
		t.Errorf("BatchWindow = %v, want 1s", cfg.BatchWindow)
	}
}

func TestLoadFromEnv_Override(t *testing.T) {
	os.Setenv("MQTT_TOPIC", "custom/topic")
	os.Setenv("BATCH_SIZE", "50")
	defer os.Unsetenv("MQTT_TOPIC")
	defer os.Unsetenv("BATCH_SIZE")

	cfg := LoadFromEnv()

	if cfg.Topic != "custom/topic" {
		t.Errorf("Topic = %q, want custom/topic", cfg.Topic)
	}
	if cfg.BatchSize != 50 {
		t.Errorf("BatchSize = %d, want 50", cfg.BatchSize)
	}
}
