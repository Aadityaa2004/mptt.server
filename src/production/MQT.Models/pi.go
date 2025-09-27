package mqtmodels

import "time"

// Pi represents a Raspberry Pi gateway
type Pi struct {
	PiID      string                 `json:"pi_id" db:"pi_id"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	Meta      map[string]interface{} `json:"meta" db:"meta"`
}