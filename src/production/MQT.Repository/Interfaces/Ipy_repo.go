package interfaces

import (
	"context"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

type PiRepository interface {
	// Create pi (idempotent upsert)
	CreateOrUpdatePi(ctx context.Context, pi mqtmodels.Pi) error

	// Read pis
	GetPi(ctx context.Context, piID string) (*mqtmodels.Pi, error)
	ListPis(ctx context.Context, userID string, page, pageSize int) (*PaginationResult, error)

	// Update pi
	UpdatePi(ctx context.Context, pi mqtmodels.Pi) error

	// Delete pi
	DeletePi(ctx context.Context, piID string, cascade bool) error
}
