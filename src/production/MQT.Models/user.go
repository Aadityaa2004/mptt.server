package mqtmodels

import "time"

// User represents a user in the system
type User struct {
	UserID    string                 `json:"user_id" db:"user_id"`
	Name      string                 `json:"name" db:"name"`
	Role      string                 `json:"role" db:"role"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	Meta      map[string]interface{} `json:"meta" db:"meta"`
}
