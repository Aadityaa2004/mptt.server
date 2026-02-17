package auth_models

// DeviceLocation represents a device location on the map
type DeviceLocation struct {
	DeviceID       string  `json:"device_id" binding:"required"` // MAC Address
	PiID           string  `json:"pi_id" binding:"required"`     // Pi ID
	Latitude       float64 `json:"latitude" binding:"required"`  // Latitude coordinate
	Longitude      float64 `json:"longitude" binding:"required"` // Longitude coordinate
	Color          string  `json:"color,omitempty"`              // Marker color (e.g., hex code, color name)
	Height         float64 `json:"height,omitempty"`             // Bucket height in cm
	TopDiameter    float64 `json:"top_diameter,omitempty"`       // Bucket top diameter in cm
	BottomDiameter float64 `json:"bottom_diameter,omitempty"`    // Bucket bottom diameter in cm
}
