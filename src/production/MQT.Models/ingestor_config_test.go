package mqtmodels

import (
	"testing"
)

func TestNewIngestorConfig(t *testing.T) {
	cfg := NewIngestorConfig()
	if cfg == nil {
		t.Fatal("NewIngestorConfig returned nil")
	}
	if cfg.BrokerPort != 8883 || !cfg.UseTLS {
		t.Errorf("unexpected MQTT defaults: %+v", cfg)
	}
	if cfg.BatchSize != 1000 {
		t.Errorf("unexpected BatchSize: %d", cfg.BatchSize)
	}
}
