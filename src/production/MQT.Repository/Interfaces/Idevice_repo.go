package interfaces

import (
	"context"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

type DeviceRepository interface {
	// Create device (idempotent upsert)
	CreateOrUpdateDevice(ctx context.Context, device mqtmodels.Device) error

	// Read devices
	GetDevice(ctx context.Context, piID string, deviceID int) (*mqtmodels.Device, error)
	ListDevicesByPi(ctx context.Context, piID string, page, pageSize int) (*PaginationResult, error)

	// Update device
	UpdateDevice(ctx context.Context, device mqtmodels.Device) error

	// Delete device
	DeleteDevice(ctx context.Context, piID string, deviceID int, cascade bool) error
}
