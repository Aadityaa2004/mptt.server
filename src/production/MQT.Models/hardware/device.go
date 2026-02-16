package hardware_models

// Device represents a device attached to a Raspberry Pi.
// Note: metadata fields like created_at and the device_type classification
// are intentionally omitted from the schema.
type Device struct {
	PiID           string  `json:"pi_id" db:"pi_id"`
	DeviceID       string  `json:"device_id" db:"device_id"`
	Height         float64 `json:"height,omitempty" db:"height"`
	TopDiameter    float64 `json:"top_diameter,omitempty" db:"top_diameter"`
	BottomDiameter float64 `json:"bottom_diameter,omitempty" db:"bottom_diameter"`
}
