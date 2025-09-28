package interfaces

import (
	"context"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

// PaginationResult represents a paginated result
type PaginationResult struct {
	Items    interface{} `json:"items"`
	NextPage *int        `json:"next_page,omitempty"`
	Total    int         `json:"total,omitempty"`
}

type UserRepository interface {
	// Create user
	CreateUser(ctx context.Context, user mqtmodels.User) error

	// Read users
	GetUser(ctx context.Context, userID string) (*mqtmodels.User, error)
	ListUsers(ctx context.Context, page, pageSize int, role string) (*PaginationResult, error)

	// Update user
	UpdateUser(ctx context.Context, user mqtmodels.User) error

	// Delete user
	DeleteUser(ctx context.Context, userID string, hardDelete bool) error
}
