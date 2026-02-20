package implementation

import (
	"context"
	"database/sql"

	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type PostgresDeviceRepository struct {
	db *sql.DB
}

func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{db: db}
}

// Create device (idempotent upsert)
func (r *PostgresDeviceRepository) CreateOrUpdateDevice(ctx context.Context, device hardware_models.Device) error {
	query := `
		INSERT INTO devices (pi_id, device_id, height, top_diameter, bottom_diameter, collection_enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (pi_id, device_id)
		DO UPDATE SET height = EXCLUDED.height, top_diameter = EXCLUDED.top_diameter, bottom_diameter = EXCLUDED.bottom_diameter, collection_enabled = EXCLUDED.collection_enabled
	`
	_, err := r.db.ExecContext(ctx, query, device.PiID, device.DeviceID, device.Height, device.TopDiameter, device.BottomDiameter, device.CollectionEnabled)
	return err
}

// Read devices
func (r *PostgresDeviceRepository) GetDevice(ctx context.Context, piID string, deviceID string) (*hardware_models.Device, error) {
	query := `SELECT pi_id, device_id, COALESCE(height, 0), COALESCE(top_diameter, 0), COALESCE(bottom_diameter, 0), COALESCE(collection_enabled, true) FROM devices WHERE pi_id = $1 AND device_id = $2`

	var device hardware_models.Device

	err := r.db.QueryRowContext(ctx, query, piID, deviceID).Scan(&device.PiID, &device.DeviceID, &device.Height, &device.TopDiameter, &device.BottomDiameter, &device.CollectionEnabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &device, nil
}

func (r *PostgresDeviceRepository) ListDevicesByPi(ctx context.Context, piID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	offset := (page - 1) * pageSize
	query := `SELECT pi_id, device_id, COALESCE(height, 0), COALESCE(top_diameter, 0), COALESCE(bottom_diameter, 0), COALESCE(collection_enabled, true) FROM devices WHERE pi_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, piID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []hardware_models.Device
	for rows.Next() {
		var device hardware_models.Device

		if err := rows.Scan(&device.PiID, &device.DeviceID, &device.Height, &device.TopDiameter, &device.BottomDiameter, &device.CollectionEnabled); err != nil {
			return nil, err
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
func (r *PostgresDeviceRepository) UpdateDevice(ctx context.Context, device hardware_models.Device) error {
	query := `UPDATE devices SET height = $1, top_diameter = $2, bottom_diameter = $3, collection_enabled = $4 WHERE pi_id = $5 AND device_id = $6`

	result, err := r.db.ExecContext(ctx, query, device.Height, device.TopDiameter, device.BottomDiameter, device.CollectionEnabled, device.PiID, device.DeviceID)
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

// Delete device
func (r *PostgresDeviceRepository) DeleteDevice(ctx context.Context, piID string, deviceID string, cascade bool) error {
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
