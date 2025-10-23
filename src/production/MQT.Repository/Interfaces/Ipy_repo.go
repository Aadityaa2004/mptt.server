package interfaces

import (
	"context"

	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

type PiRepository interface {
	// Create pi (idempotent upsert)
	CreateOrUpdatePi(ctx context.Context, pi hardware_models.Pi) error

	// Read pis
	GetPi(ctx context.Context, piID string) (*hardware_models.Pi, error)
	ListPis(ctx context.Context, userID string, page, pageSize int) (*PaginationResult, error)

	// Update pi
	UpdatePi(ctx context.Context, pi hardware_models.Pi) error

	// Delete pi
	DeletePi(ctx context.Context, piID string, cascade bool) error
}
