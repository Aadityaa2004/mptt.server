package auth_models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// DeviceLocationArray is a custom type for handling JSONB array in database
type DeviceLocationArray []DeviceLocation

// Value implements the driver.Valuer interface for database storage
func (a DeviceLocationArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for database retrieval
func (a *DeviceLocationArray) Scan(value interface{}) error {
	if value == nil {
		*a = DeviceLocationArray{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	if len(bytes) == 0 || string(bytes) == "[]" {
		*a = DeviceLocationArray{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// User represents a user in the system
type User struct {
	UserID         string              `json:"user_id" db:"user_id"`
	Username       string              `json:"username" db:"username"`
	Email          string              `json:"email" db:"email"`
	Password       string              `json:"-" db:"password"` // Password is not exposed in JSON
	Role           string              `json:"role" db:"role"`
	Active         bool                `json:"active" db:"active"`
	EmailVerifiedAt *time.Time         `json:"email_verified_at,omitempty" db:"email_verified_at"`
	Latitude       *float64            `json:"latitude,omitempty" db:"latitude"`
	Longitude      *float64            `json:"longitude,omitempty" db:"longitude"`
	Locations      DeviceLocationArray `json:"locations,omitempty" db:"locations"`
	CreatedAt      time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at" db:"updated_at"`
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
		Locations: DeviceLocationArray{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
