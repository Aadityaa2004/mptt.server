package mqtmodels

import (
	"time"
)

// Reading represents a time-series reading from a device
type Reading struct {
	PiID     string                 `json:"pi_id" db:"pi_id"`
	DeviceID int                    `json:"device_id" db:"device_id"`
	Ts       time.Time              `json:"ts" db:"ts"`
	Payload  map[string]interface{} `json:"payload" db:"payload"`
}

// ReadingWithTopic represents a reading with topic information for MQTT processing
type ReadingWithTopic struct {
	PiID       string                 `json:"pi_id"`
	DeviceID   string                 `json:"device_id"`
	Topic      string                 `json:"topic"`
	Payload    map[string]interface{} `json:"payload"`
	ReceivedAt time.Time              `json:"received_at"`
}
