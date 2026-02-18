package hardware_models

// Payload models for MQTT sensor readings

// TemperatureSensorPayload represents a temperature sensor reading.
type TemperatureSensorPayload struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// LevelSensorPayload represents a level sensor reading.
type LevelSensorPayload struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// SensorsPayload groups all sensor readings in the payload.
type SensorsPayload struct {
	Temperature *TemperatureSensorPayload `json:"temperature,omitempty"`
	Level       *LevelSensorPayload       `json:"level,omitempty"`
}

// ReadingPayload represents the inner "payload" object of the MQTT envelope.
//
// NOTE: The timestamp is kept as a string here and is parsed into a time.Time
// where needed (e.g. when persisting to the database).
type ReadingPayload struct {
	DeviceID          string         `json:"device_id"`
	PiID              string         `json:"pi_id"`
	Timestamp         string         `json:"timestamp"`
	Sensors           SensorsPayload `json:"sensors"`
	BatteryPercentage float64        `json:"battery_percentage"`
}

// MqttEnvelope represents the "mqtt_envelope" object sent by devices.
type MqttEnvelope struct {
	Topic     string         `json:"topic"`
	Payload   ReadingPayload `json:"payload"`
	QoS       int            `json:"qos"`
	Retain    bool           `json:"retain"`
	MessageID int64          `json:"message_id"`
	Duplicate bool           `json:"duplicate"`
}

// IngestedMqttMessage is the top-level structure received by the ingestor.
type IngestedMqttMessage struct {
	MqttEnvelope MqttEnvelope `json:"mqtt_envelope"`
}


