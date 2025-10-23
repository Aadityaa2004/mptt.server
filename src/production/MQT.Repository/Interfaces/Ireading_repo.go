package interfaces

import (
	"context"
	"time"

	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

// ReadingQueryParams represents parameters for reading queries
type ReadingQueryParams struct {
	PiID     string
	DeviceID string
	From     *time.Time
	To       *time.Time
	Limit    int
	Page     int
}

// ReadingQueryResult represents the result of a reading query with pagination
type ReadingQueryResult struct {
	Items         []hardware_models.Reading `json:"items"`
	NextPageToken *string             `json:"next_page_token,omitempty"`
	Total         int                 `json:"total,omitempty"`
}

// SummaryStats represents aggregate statistics
type SummaryStats struct {
	Count    int64         `json:"count"`
	FirstTS  *time.Time    `json:"first_ts,omitempty"`
	LastTS   *time.Time    `json:"last_ts,omitempty"`
	ByDevice []DeviceStats `json:"by_device,omitempty"`
}

// DeviceStats represents stats for a specific device
type DeviceStats struct {
	PiID     string     `json:"pi_id"`
	DeviceID string     `json:"device_id"`
	Count    int64      `json:"count"`
	FirstTS  *time.Time `json:"first_ts,omitempty"`
	LastTS   *time.Time `json:"last_ts,omitempty"`
}

type ReadingRepository interface {
	// Reading operations
	CreateReading(ctx context.Context, reading hardware_models.Reading) error
	CreateReadings(ctx context.Context, readings []hardware_models.Reading) error

	// Query operations with pagination
	GetLatestReadings(ctx context.Context, piID string) ([]hardware_models.Reading, error)
	GetReadings(ctx context.Context, params ReadingQueryParams) (*ReadingQueryResult, error)
	GetReadingsByDevice(ctx context.Context, piID string, deviceID int, params ReadingQueryParams) (*ReadingQueryResult, error)

	// Statistics
	GetSummaryStats(ctx context.Context, params ReadingQueryParams) (*SummaryStats, error)

	// Delete operations
	DeleteReadingsByTimeRange(ctx context.Context, piID string, deviceID int, start, end time.Time) error
}
