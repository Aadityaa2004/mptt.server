package mqtingestor

import (
	"context"
	"testing"
	"time"

	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/client"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
)

type mockAPIReadingsClient struct {
	validatePiFn     func(ctx context.Context, piID string) (bool, error)
	validateDeviceFn func(ctx context.Context, piID, deviceID string) (bool, error)
	createReadingFn  func(ctx context.Context, reading hardware_models.Reading) error
}

func (m *mockAPIReadingsClient) ValidatePi(ctx context.Context, piID string) (bool, error) {
	if m.validatePiFn != nil {
		return m.validatePiFn(ctx, piID)
	}
	return true, nil
}
func (m *mockAPIReadingsClient) ValidateDevice(ctx context.Context, piID, deviceID string) (bool, error) {
	if m.validateDeviceFn != nil {
		return m.validateDeviceFn(ctx, piID, deviceID)
	}
	return true, nil
}
func (m *mockAPIReadingsClient) CreateReading(ctx context.Context, reading hardware_models.Reading) error {
	if m.createReadingFn != nil {
		return m.createReadingFn(ctx, reading)
	}
	return nil
}
func (m *mockAPIReadingsClient) Health(ctx context.Context) error { return nil }
func (m *mockAPIReadingsClient) GetCircuitBreakerStatus() map[string]interface{} {
	return map[string]interface{}{"state": "closed"}
}

var _ client.APIReadingsClient = (*mockAPIReadingsClient)(nil)

func TestParseMQTTMessage_ValidPayload(t *testing.T) {
	payload := []byte(`{"mqtt_envelope":{"topic":"sensors/p1/d1/temp","payload":{"device_id":"d1","pi_id":"p1","timestamp":"","sensors":{},"battery_percentage":0}}}`)
	rwt, err := ParseMQTTMessage("sensors/p1/d1/temp", payload)
	if err != nil {
		t.Fatalf("ParseMQTTMessage: %v", err)
	}
	if rwt == nil || rwt.PiID != "p1" || rwt.DeviceID != "d1" {
		t.Errorf("unexpected result: %+v", rwt)
	}
}

func TestParseMQTTMessage_TopicFallback(t *testing.T) {
	payload := []byte(`{"mqtt_envelope":{"topic":"sensors/p1/d1/temp","payload":{"device_id":"","pi_id":"","timestamp":"","sensors":{},"battery_percentage":0}}}`)
	rwt, err := ParseMQTTMessage("sensors/p1/d1/temp", payload)
	if err != nil {
		t.Fatalf("ParseMQTTMessage: %v", err)
	}
	if rwt == nil || rwt.PiID != "p1" || rwt.DeviceID != "d1" {
		t.Errorf("unexpected result: %+v", rwt)
	}
}

func TestParseMQTTMessage_InvalidTopicFormat(t *testing.T) {
	payload := []byte(`{"mqtt_envelope":{"topic":"short","payload":{"device_id":"","pi_id":"","timestamp":"","sensors":{},"battery_percentage":0}}}`)
	_, err := ParseMQTTMessage("short", payload)
	if err == nil {
		t.Fatal("expected error for invalid topic")
	}
}

func TestParseMQTTMessage_InvalidJSON(t *testing.T) {
	payload := []byte(`{invalid`)
	rwt, err := ParseMQTTMessage("sensors/p1/d1/temp", payload)
	if err != nil {
		t.Fatalf("invalid JSON should fallback to empty payload: %v", err)
	}
	if rwt == nil || rwt.PiID != "p1" || rwt.DeviceID != "d1" {
		t.Errorf("unexpected result: %+v", rwt)
	}
}

func TestIngestor_brokerURL(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{BrokerHost: "broker", BrokerPort: 1883, UseTLS: false}
	ing := New(cfg, nil, nil)
	url := ing.brokerURL()
	if url != "tcp://broker:1883" {
		t.Errorf("got %s", url)
	}
	cfg.UseTLS = true
	ing2 := New(cfg, nil, nil)
	url2 := ing2.brokerURL()
	if url2 != "tcps://broker:1883" {
		t.Errorf("got %s", url2)
	}
}

func TestIngestor_tlsConfig_Empty(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{}
	ing := New(cfg, nil, nil)
	tlsCfg, err := ing.tlsConfig("")
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}
	if tlsCfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestIngestor_tlsConfig_InvalidFile(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{}
	ing := New(cfg, nil, nil)
	_, err := ing.tlsConfig("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIngestor_processBatch_Success(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{BatchSize: 10, BatchWindow: time.Second}
	logCfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(logCfg)

	createCalled := false
	mockAPI := &mockAPIReadingsClient{
		createReadingFn: func(ctx context.Context, r hardware_models.Reading) error {
			createCalled = true
			return nil
		},
	}

	ing := New(cfg, mockAPI, log)
	batch := []hardware_models.ReadingWithTopic{
		{PiID: "p1", DeviceID: "d1", Topic: "sensors/p1/d1/temp", Payload: hardware_models.ReadingPayload{}, ReceivedAt: time.Now()},
	}
	ing.processBatch(context.Background(), batch)
	if !createCalled {
		t.Error("CreateReading was not called")
	}
}

func TestIngestor_processBatch_PiNotFound(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{BatchSize: 10, BatchWindow: time.Second}
	logCfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(logCfg)

	mockAPI := &mockAPIReadingsClient{
		validatePiFn: func(ctx context.Context, piID string) (bool, error) {
			return false, nil
		},
	}

	ing := New(cfg, mockAPI, log)
	batch := []hardware_models.ReadingWithTopic{
		{PiID: "invalid", DeviceID: "d1", Topic: "sensors/invalid/d1/temp", Payload: hardware_models.ReadingPayload{}, ReceivedAt: time.Now()},
	}
	ing.processBatch(context.Background(), batch)
	// publishError no-ops when mqttClient is nil
}

func TestIngestor_processBatch_DeviceNotFound(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{BatchSize: 10, BatchWindow: time.Second}
	logCfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(logCfg)

	mockAPI := &mockAPIReadingsClient{
		validatePiFn: func(ctx context.Context, piID string) (bool, error) { return true, nil },
		validateDeviceFn: func(ctx context.Context, piID, deviceID string) (bool, error) {
			return false, nil
		},
	}

	ing := New(cfg, mockAPI, log)
	batch := []hardware_models.ReadingWithTopic{
		{PiID: "p1", DeviceID: "invalid", Topic: "sensors/p1/invalid/temp", Payload: hardware_models.ReadingPayload{}, ReceivedAt: time.Now()},
	}
	ing.processBatch(context.Background(), batch)
}

func TestIngestor_processBatch_Empty(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{}
	logCfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(logCfg)
	ing := New(cfg, &mockAPIReadingsClient{}, log)
	ing.processBatch(context.Background(), nil)
	ing.processBatch(context.Background(), []hardware_models.ReadingWithTopic{})
}

func TestIngestor_processBatch_TimestampFromPayload(t *testing.T) {
	cfg := mqtmodels.IngestorConfig{BatchSize: 10, BatchWindow: time.Second}
	logCfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(logCfg)

	var createdTs time.Time
	mockAPI := &mockAPIReadingsClient{
		createReadingFn: func(ctx context.Context, r hardware_models.Reading) error {
			createdTs = r.Ts
			return nil
		},
	}

	ing := New(cfg, mockAPI, log)
	tsStr := "2024-01-15T10:30:00.000Z"
	batch := []hardware_models.ReadingWithTopic{
		{PiID: "p1", DeviceID: "d1", Topic: "sensors/p1/d1/temp", Payload: hardware_models.ReadingPayload{Timestamp: tsStr}, ReceivedAt: time.Now()},
	}
	ing.processBatch(context.Background(), batch)
	expected, _ := time.Parse(time.RFC3339Nano, tsStr)
	if !createdTs.Equal(expected) {
		t.Errorf("expected ts %v, got %v", expected, createdTs)
	}
}
