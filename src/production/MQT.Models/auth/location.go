package auth_models

// DeviceLocation represents a device location on the map
type DeviceLocation struct {
	DeviceID  string  `json:"device_id" binding:"required"` // MAC Address
	PiID      string  `json:"pi_id" binding:"required"`     // Pi ID
	Latitude  float64 `json:"latitude" binding:"required"`  // Latitude coordinate
	Longitude float64 `json:"longitude" binding:"required"` // Longitude coordinate
}
