package hardware_models

import "time"

// Device represents a device attached to a Raspberry Pi
type Device struct {
	PiID       string    `json:"pi_id" db:"pi_id"`
	DeviceID   int       `json:"device_id" db:"device_id"`
	DeviceType string    `json:"device_type" db:"device_type"` // temperature, humidity, light, pressure
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
