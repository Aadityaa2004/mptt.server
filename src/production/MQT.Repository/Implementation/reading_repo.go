package implementation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

type PostgresReadingRepository struct {
	db *sql.DB
}

func NewPostgresReadingRepository(db *sql.DB) *PostgresReadingRepository {
	return &PostgresReadingRepository{db: db}
}

// Pi operations
func (r *PostgresReadingRepository) UpsertPi(ctx context.Context, pi mqtmodels.Pi) error {
	query := `
		INSERT INTO pis (pi_id, created_at, meta) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (pi_id) 
		DO UPDATE SET meta = EXCLUDED.meta
	`

	metaJSON, err := json.Marshal(pi.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, pi.PiID, pi.CreatedAt, metaJSON)
	return err
}

func (r *PostgresReadingRepository) GetPi(ctx context.Context, piID string) (*mqtmodels.Pi, error) {
	query := `SELECT pi_id, created_at, meta FROM pis WHERE pi_id = $1`

	var pi mqtmodels.Pi
	var metaJSON []byte

	err := r.db.QueryRowContext(ctx, query, piID).Scan(&pi.PiID, &pi.CreatedAt, &metaJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(metaJSON, &pi.Meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}

	return &pi, nil
}

func (r *PostgresReadingRepository) ListPis(ctx context.Context) ([]mqtmodels.Pi, error) {
	query := `SELECT pi_id, created_at, meta FROM pis ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pis []mqtmodels.Pi
	for rows.Next() {
		var pi mqtmodels.Pi
		var metaJSON []byte

		if err := rows.Scan(&pi.PiID, &pi.CreatedAt, &metaJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metaJSON, &pi.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}

		pis = append(pis, pi)
	}

	return pis, rows.Err()
}

// Device operations
func (r *PostgresReadingRepository) UpsertDevice(ctx context.Context, device mqtmodels.Device) error {
	query := `
		INSERT INTO devices (pi_id, device_id, created_at, meta) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (pi_id, device_id) 
		DO UPDATE SET meta = EXCLUDED.meta
	`

	metaJSON, err := json.Marshal(device.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, device.PiID, device.DeviceID, device.CreatedAt, metaJSON)
	return err
}

func (r *PostgresReadingRepository) GetDevice(ctx context.Context, piID, deviceID string) (*mqtmodels.Device, error) {
	query := `SELECT pi_id, device_id, created_at, meta FROM devices WHERE pi_id = $1 AND device_id = $2`

	var device mqtmodels.Device
	var metaJSON []byte

	err := r.db.QueryRowContext(ctx, query, piID, deviceID).Scan(&device.PiID, &device.DeviceID, &device.CreatedAt, &metaJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(metaJSON, &device.Meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}

	return &device, nil
}

func (r *PostgresReadingRepository) ListDevicesByPi(ctx context.Context, piID string) ([]mqtmodels.Device, error) {
	query := `SELECT pi_id, device_id, created_at, meta FROM devices WHERE pi_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, piID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []mqtmodels.Device
	for rows.Next() {
		var device mqtmodels.Device
		var metaJSON []byte

		if err := rows.Scan(&device.PiID, &device.DeviceID, &device.CreatedAt, &metaJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metaJSON, &device.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

func (r *PostgresReadingRepository) ListAllDevices(ctx context.Context) ([]mqtmodels.Device, error) {
	query := `SELECT pi_id, device_id, created_at, meta FROM devices ORDER BY pi_id, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []mqtmodels.Device
	for rows.Next() {
		var device mqtmodels.Device
		var metaJSON []byte

		if err := rows.Scan(&device.PiID, &device.DeviceID, &device.CreatedAt, &metaJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metaJSON, &device.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}

		devices = append(devices, device)
	}

	return devices, rows.Err()
}

// Reading operations
func (r *PostgresReadingRepository) InsertReading(ctx context.Context, reading mqtmodels.Reading) error {
	query := `
		INSERT INTO readings (pi_id, device_id, ts, payload) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (pi_id, device_id, ts) 
		DO UPDATE SET payload = EXCLUDED.payload
	`

	payloadJSON, err := json.Marshal(reading.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, reading.PiID, reading.DeviceID, reading.Ts, payloadJSON)
	return err
}

func (r *PostgresReadingRepository) InsertReadings(ctx context.Context, readings []mqtmodels.Reading) error {
	if len(readings) == 0 {
		return nil
	}

	// Prepare the COPY statement for bulk insert
	txn, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	stmt, err := txn.Prepare(pq.CopyIn("readings", "pi_id", "device_id", "ts", "payload"))
	if err != nil {
		return err
	}

	for _, reading := range readings {
		payloadJSON, err := json.Marshal(reading.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		_, err = stmt.Exec(reading.PiID, reading.DeviceID, reading.Ts, payloadJSON)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (r *PostgresReadingRepository) GetReadingsByPi(ctx context.Context, piID string, limit, offset int) ([]mqtmodels.Reading, error) {
	query := `
		SELECT pi_id, device_id, ts, payload 
		FROM readings 
		WHERE pi_id = $1 
		ORDER BY ts DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, piID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReadings(rows)
}

func (r *PostgresReadingRepository) GetReadingsByDevice(ctx context.Context, piID, deviceID string, limit, offset int) ([]mqtmodels.Reading, error) {
	query := `
		SELECT pi_id, device_id, ts, payload 
		FROM readings 
		WHERE pi_id = $1 AND device_id = $2 
		ORDER BY ts DESC 
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, piID, deviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReadings(rows)
}

func (r *PostgresReadingRepository) GetReadingsByTimeRange(ctx context.Context, piID, deviceID string, start, end time.Time, limit, offset int) ([]mqtmodels.Reading, error) {
	query := `
		SELECT pi_id, device_id, ts, payload 
		FROM readings 
		WHERE pi_id = $1 AND device_id = $2 AND ts BETWEEN $3 AND $4 
		ORDER BY ts DESC 
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.QueryContext(ctx, query, piID, deviceID, start, end, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReadings(rows)
}

func (r *PostgresReadingRepository) GetLatestReadings(ctx context.Context, piID string, limit int) ([]mqtmodels.Reading, error) {
	query := `
		SELECT DISTINCT ON (device_id) pi_id, device_id, ts, payload 
		FROM readings 
		WHERE pi_id = $1 
		ORDER BY device_id, ts DESC 
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, piID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReadings(rows)
}

func (r *PostgresReadingRepository) scanReadings(rows *sql.Rows) ([]mqtmodels.Reading, error) {
	var readings []mqtmodels.Reading

	for rows.Next() {
		var reading mqtmodels.Reading
		var payloadJSON []byte

		if err := rows.Scan(&reading.PiID, &reading.DeviceID, &reading.Ts, &payloadJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(payloadJSON, &reading.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		readings = append(readings, reading)
	}

	return readings, rows.Err()
}
