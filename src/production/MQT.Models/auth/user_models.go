package auth_models

import (
	"time"
)

// User represents a user in the system
type User struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Password is not exposed in JSON
	Role      string    `json:"role" db:"role"`
	Active    bool      `json:"active" db:"active"`
	Latitude  *float64  `json:"latitude,omitempty" db:"latitude"`
	Longitude *float64  `json:"longitude,omitempty" db:"longitude"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser creates a new User instance
func NewUser(username, email, password, role string, latitude, longitude *float64) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Email:     email,
		Password:  password, // Note: This should be hashed before saving
		Role:      role,
		Active:    true,
		Latitude:  latitude,
		Longitude: longitude,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
