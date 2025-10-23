package interfaces

import (
	"context"

	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

type DeviceRepository interface {
	// Create device (idempotent upsert)
	CreateOrUpdateDevice(ctx context.Context, device hardware_models.Device) error

	// Read devices
	GetDevice(ctx context.Context, piID string, deviceID int) (*hardware_models.Device, error)
	ListDevicesByPi(ctx context.Context, piID string, page, pageSize int) (*PaginationResult, error)

	// Update device
	UpdateDevice(ctx context.Context, device hardware_models.Device) error

	// Delete device
	DeleteDevice(ctx context.Context, piID string, deviceID int, cascade bool) error
}
