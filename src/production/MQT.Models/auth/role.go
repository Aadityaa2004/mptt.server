package auth_models

import (
	"time"
)

// Role represents a role in the system
type Role struct {
	RoleID      string    `json:"role_id" db:"role_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NewRole creates a new Role instance
func NewRole(name, description string) *Role {
	now := time.Now()
	return &Role{
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
