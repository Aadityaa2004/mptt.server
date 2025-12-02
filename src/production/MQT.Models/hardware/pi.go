package hardware_models

// Pi represents a Raspberry Pi gateway
// Note: metadata fields like created_at are intentionally omitted from the schema.
type Pi struct {
	PiID   string `json:"pi_id" db:"pi_id"`
	UserID string `json:"user_id" db:"user_id"`
}
