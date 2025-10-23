package hardware_models

import "time"

// Pi represents a Raspberry Pi gateway
type Pi struct {
	PiID      string    `json:"pi_id" db:"pi_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
