package logger

import (
	"testing"

	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
)

func TestNewLogger(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:        "info",
		Format:       "text",
		Output:       "stdout",
		EnableCaller: false,
	}
	log := NewLogger(cfg)
	if log == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestNewLogger_JSON(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:        "debug",
		Format:       "json",
		Output:       "stdout",
		EnableCaller: false,
	}
	log := NewLogger(cfg)
	if log == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestLogger_WithField(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log2 := log.WithField("key", "value")
	if log2 == nil {
		t.Fatal("WithField returned nil")
	}
}

func TestLogger_WithFields(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log2 := log.WithFields(map[string]interface{}{"a": 1, "b": "x"})
	if log2 == nil {
		t.Fatal("WithFields returned nil")
	}
}

func TestLogger_WithError(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log2 := log.WithError(nil)
	if log2 == nil {
		t.Fatal("WithError returned nil")
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log2 := log.WithRequestID("req-1")
	if log2 == nil {
		t.Fatal("WithRequestID returned nil")
	}
}

func TestLogger_WithService(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log2 := log.WithService("test-svc")
	if log2 == nil {
		t.Fatal("WithService returned nil")
	}
}

func TestLogger_WithComponent(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log2 := log.WithComponent("component")
	if log2 == nil {
		t.Fatal("WithComponent returned nil")
	}
}

func TestLogger_LogMethods(t *testing.T) {
	cfg := &config.LoggingConfig{Level: "debug", Format: "text", Output: "stdout"}
	log := NewLogger(cfg)
	log.Debug("debug msg")
	log.Info("info msg")
	log.Warn("warn msg")
	log.Error("error msg")
}

func TestGetGlobalLogger(t *testing.T) {
	log := GetGlobalLogger()
	if log == nil {
		t.Fatal("GetGlobalLogger returned nil")
	}
}
