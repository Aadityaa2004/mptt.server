package interfaces

import (
	"context"
	"time"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

type ReadingRepository interface {
	// Pi operations
	UpsertPi(ctx context.Context, pi mqtmodels.Pi) error
	GetPi(ctx context.Context, piID string) (*mqtmodels.Pi, error)
	ListPis(ctx context.Context) ([]mqtmodels.Pi, error)

	// Device operations
	UpsertDevice(ctx context.Context, device mqtmodels.Device) error
	GetDevice(ctx context.Context, piID, deviceID string) (*mqtmodels.Device, error)
	ListDevicesByPi(ctx context.Context, piID string) ([]mqtmodels.Device, error)
	ListAllDevices(ctx context.Context) ([]mqtmodels.Device, error)

	// Reading operations
	InsertReading(ctx context.Context, r mqtmodels.Reading) error
	InsertReadings(ctx context.Context, rs []mqtmodels.Reading) error
	GetReadingsByPi(ctx context.Context, piID string, limit, offset int) ([]mqtmodels.Reading, error)
	GetReadingsByDevice(ctx context.Context, piID, deviceID string, limit, offset int) ([]mqtmodels.Reading, error)
	GetReadingsByTimeRange(ctx context.Context, piID, deviceID string, start, end time.Time, limit, offset int) ([]mqtmodels.Reading, error)
	GetLatestReadings(ctx context.Context, piID string, limit int) ([]mqtmodels.Reading, error)
}
