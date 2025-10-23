package interfaces

import (
	"context"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

// PaginationResult represents a paginated result
type PaginationResult struct {
	Items    interface{} `json:"items"`
	NextPage *int        `json:"next_page,omitempty"`
	Total    int         `json:"total,omitempty"`
}

type UserRepository interface {
	// Create user
	Create(ctx context.Context, user *auth_models.User) (*auth_models.User, error)

	// Read users
	GetByID(ctx context.Context, userID string) (*auth_models.User, error)
	FindByID(ctx context.Context, userID string) (*auth_models.User, error)
	GetByUsername(ctx context.Context, username string) (*auth_models.User, error)
	GetAll(ctx context.Context) ([]*auth_models.User, error)
	List(ctx context.Context, page, pageSize int, role string) (*PaginationResult, error)
	GetUser(ctx context.Context, userID string) (*auth_models.User, error)
	GetByRole(ctx context.Context, role string) ([]*auth_models.User, error)

	// Update user
	Update(ctx context.Context, user *auth_models.User) error

	// Delete user
	Delete(ctx context.Context, userID string, hardDelete bool) error
}
