package interfaces

import (
	"context"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

type RoleRepository interface {
	// Create role
	Create(ctx context.Context, role *auth_models.Role) (*auth_models.Role, error)

	// Read roles
	FindByID(ctx context.Context, id string) (*auth_models.Role, error)
	FindByName(ctx context.Context, name string) (*auth_models.Role, error)
	FindAll(ctx context.Context) ([]*auth_models.Role, error)

	// Update role
	Update(ctx context.Context, role *auth_models.Role) error

	// Delete role
	Delete(ctx context.Context, id string) error
}
