package mqtmodels

import "time"

// Device represents a device attached to a Raspberry Pi
type Device struct {
	PiID      string                 `json:"pi_id" db:"pi_id"`
	DeviceID  string                 `json:"device_id" db:"device_id"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	Meta      map[string]interface{} `json:"meta" db:"meta"`
}