package implementation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type PostgresDeviceRepository struct {
	db *sql.DB
}

func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{db: db}
}

// Create device (idempotent upsert)
func (r *PostgresDeviceRepository) CreateOrUpdateDevice(ctx context.Context, device mqtmodels.Device) error {
	query := `
		INSERT INTO devices (pi_id, device_id, device_type, created_at, meta) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (pi_id, device_id) 
		DO UPDATE SET device_type = EXCLUDED.device_type, meta = EXCLUDED.meta
	`

	metaJSON, err := json.Marshal(ensureMetaNotNull(device.Meta))
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, device.PiID, device.DeviceID, device.DeviceType, device.CreatedAt, metaJSON)
	return err
}

// Read devices
func (r *PostgresDeviceRepository) GetDevice(ctx context.Context, piID string, deviceID int) (*mqtmodels.Device, error) {
	query := `SELECT pi_id, device_id, device_type, created_at, meta FROM devices WHERE pi_id = $1 AND device_id = $2`

	var device mqtmodels.Device
	var metaJSON []byte

	err := r.db.QueryRowContext(ctx, query, piID, deviceID).Scan(&device.PiID, &device.DeviceID, &device.DeviceType, &device.CreatedAt, &metaJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	if err := json.Unmarshal(metaJSON, &device.Meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}

	return &device, nil
}

func (r *PostgresDeviceRepository) ListDevicesByPi(ctx context.Context, piID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	offset := (page - 1) * pageSize
	query := `SELECT pi_id, device_id, device_type, created_at, meta FROM devices WHERE pi_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, piID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []mqtmodels.Device
	for rows.Next() {
		var device mqtmodels.Device
		var metaJSON []byte

		if err := rows.Scan(&device.PiID, &device.DeviceID, &device.DeviceType, &device.CreatedAt, &metaJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metaJSON, &device.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}

		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &interfaces.PaginationResult{
		Items: devices,
	}

	// Check if there are more pages
	if len(devices) == pageSize {
		nextPage := page + 1
		result.NextPage = &nextPage
	}

	return result, nil
}

// Update device
func (r *PostgresDeviceRepository) UpdateDevice(ctx context.Context, device mqtmodels.Device) error {
	query := `
		UPDATE devices 
		SET meta = $1 
		WHERE pi_id = $2 AND device_id = $3
	`

	metaJSON, err := json.Marshal(ensureMetaNotNull(device.Meta))
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, metaJSON, device.PiID, device.DeviceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

// Delete device
func (r *PostgresDeviceRepository) DeleteDevice(ctx context.Context, piID string, deviceID int, cascade bool) error {
	var query string
	if cascade {
		// Delete associated readings first
		// Note: This would need to be implemented with proper foreign key constraints
		query = `DELETE FROM devices WHERE pi_id = $1 AND device_id = $2`
	} else {
		query = `DELETE FROM devices WHERE pi_id = $1 AND device_id = $2`
	}

	result, err := r.db.ExecContext(ctx, query, piID, deviceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
