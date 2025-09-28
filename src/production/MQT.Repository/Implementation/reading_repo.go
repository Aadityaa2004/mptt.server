package implementation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type PostgresReadingRepository struct {
	db *sql.DB
}

func NewPostgresReadingRepository(db *sql.DB) *PostgresReadingRepository {
	return &PostgresReadingRepository{db: db}
}

// Reading operations
func (r *PostgresReadingRepository) CreateReading(ctx context.Context, reading mqtmodels.Reading) error {
	query := `
		INSERT INTO readings (pi_id, device_id, ts, payload) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (pi_id, device_id, ts) 
		DO UPDATE SET payload = EXCLUDED.payload
	`

	payloadJSON, err := json.Marshal(ensureMetaNotNull(reading.Payload))
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, reading.PiID, reading.DeviceID, reading.Ts, payloadJSON)
	return err
}

func (r *PostgresReadingRepository) CreateReadings(ctx context.Context, readings []mqtmodels.Reading) error {
	if len(readings) == 0 {
		return nil
	}

	// Use batched VALUES upsert for conflict handling
	txn, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	// Build batched INSERT with ON CONFLICT
	valueStrings := make([]string, len(readings))
	args := make([]interface{}, 0, len(readings)*4)

	for i, reading := range readings {
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)",
			i*4+1, i*4+2, i*4+3, i*4+4)

		payloadJSON, err := json.Marshal(ensureMetaNotNull(reading.Payload))
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		args = append(args, reading.PiID, reading.DeviceID, reading.Ts, payloadJSON)
	}

	valuesClause := strings.Join(valueStrings, ",")

	query := fmt.Sprintf(`
		INSERT INTO readings (pi_id, device_id, ts, payload) 
		VALUES %s
		ON CONFLICT (pi_id, device_id, ts) 
		DO UPDATE SET payload = EXCLUDED.payload
	`, valuesClause)

	_, err = txn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return txn.Commit()
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

func (r *PostgresReadingRepository) DeleteReadingsByTimeRange(ctx context.Context, piID string, deviceID int, start, end time.Time) error {
	query := `DELETE FROM readings WHERE pi_id = $1 AND device_id = $2 AND ts BETWEEN $3 AND $4`

	_, err := r.db.ExecContext(ctx, query, piID, deviceID, start, end)
	return err
}

// Enhanced methods for new interface

func (r *PostgresReadingRepository) GetLatestReadings(ctx context.Context, piID string) ([]mqtmodels.Reading, error) {
	query := `
		SELECT DISTINCT ON (device_id) pi_id, device_id, ts, payload 
		FROM readings 
		WHERE pi_id = $1 
		ORDER BY device_id, ts DESC
	`

	rows, err := r.db.QueryContext(ctx, query, piID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReadings(rows)
}

func (r *PostgresReadingRepository) GetReadings(ctx context.Context, params interfaces.ReadingQueryParams) (*interfaces.ReadingQueryResult, error) {
	offset := (params.Page - 1) * params.Limit

	query := `SELECT pi_id, device_id, ts, payload FROM readings WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if params.PiID != "" {
		query += fmt.Sprintf(" AND pi_id = $%d", argIndex)
		args = append(args, params.PiID)
		argIndex++
	}

	if params.DeviceID != "" {
		deviceIDInt, err := strconv.Atoi(params.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("invalid device_id: %w", err)
		}
		query += fmt.Sprintf(" AND device_id = $%d", argIndex)
		args = append(args, deviceIDInt)
		argIndex++
	}

	if params.From != nil {
		query += fmt.Sprintf(" AND ts >= $%d", argIndex)
		args = append(args, *params.From)
		argIndex++
	}

	if params.To != nil {
		query += fmt.Sprintf(" AND ts <= $%d", argIndex)
		args = append(args, *params.To)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY ts DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings, err := r.scanReadings(rows)
	if err != nil {
		return nil, err
	}

	result := &interfaces.ReadingQueryResult{
		Items: readings,
	}

	// Check if there are more pages
	if len(readings) == params.Limit {
		nextPageToken := strconv.Itoa(params.Page + 1)
		result.NextPageToken = &nextPageToken
	}

	return result, nil
}

func (r *PostgresReadingRepository) GetReadingsByDevice(ctx context.Context, piID string, deviceID int, params interfaces.ReadingQueryParams) (*interfaces.ReadingQueryResult, error) {
	offset := (params.Page - 1) * params.Limit

	query := `SELECT pi_id, device_id, ts, payload FROM readings WHERE pi_id = $1 AND device_id = $2`
	args := []interface{}{piID, deviceID}
	argIndex := 3

	if params.From != nil {
		query += fmt.Sprintf(" AND ts >= $%d", argIndex)
		args = append(args, *params.From)
		argIndex++
	}

	if params.To != nil {
		query += fmt.Sprintf(" AND ts <= $%d", argIndex)
		args = append(args, *params.To)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY ts DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings, err := r.scanReadings(rows)
	if err != nil {
		return nil, err
	}

	result := &interfaces.ReadingQueryResult{
		Items: readings,
	}

	// Check if there are more pages
	if len(readings) == params.Limit {
		nextPageToken := strconv.Itoa(params.Page + 1)
		result.NextPageToken = &nextPageToken
	}

	return result, nil
}

func (r *PostgresReadingRepository) GetSummaryStats(ctx context.Context, params interfaces.ReadingQueryParams) (*interfaces.SummaryStats, error) {
	query := `SELECT COUNT(*) FROM readings WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if params.PiID != "" {
		query += fmt.Sprintf(" AND pi_id = $%d", argIndex)
		args = append(args, params.PiID)
		argIndex++
	}

	if params.DeviceID != "" {
		deviceIDInt, err := strconv.Atoi(params.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("invalid device_id: %w", err)
		}
		query += fmt.Sprintf(" AND device_id = $%d", argIndex)
		args = append(args, deviceIDInt)
		argIndex++
	}

	if params.From != nil {
		query += fmt.Sprintf(" AND ts >= $%d", argIndex)
		args = append(args, *params.From)
		argIndex++
	}

	if params.To != nil {
		query += fmt.Sprintf(" AND ts <= $%d", argIndex)
		args = append(args, *params.To)
		argIndex++
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	stats := &interfaces.SummaryStats{
		Count: count,
	}

	// Get first and last timestamps
	if count > 0 {
		timeQuery := strings.Replace(query, "COUNT(*)", "MIN(ts), MAX(ts)", 1)
		var firstTS, lastTS time.Time
		err := r.db.QueryRowContext(ctx, timeQuery, args...).Scan(&firstTS, &lastTS)
		if err == nil {
			stats.FirstTS = &firstTS
			stats.LastTS = &lastTS
		}
	}

	// Get stats by device if requested
	if params.PiID != "" {
		deviceStatsQuery := `
			SELECT device_id, COUNT(*), MIN(ts), MAX(ts) 
			FROM readings 
			WHERE pi_id = $1
		`
		deviceArgs := []interface{}{params.PiID}

		if params.From != nil {
			deviceStatsQuery += " AND ts >= $2"
			deviceArgs = append(deviceArgs, *params.From)
		}

		if params.To != nil {
			deviceStatsQuery += " AND ts <= $" + strconv.Itoa(len(deviceArgs)+1)
			deviceArgs = append(deviceArgs, *params.To)
		}

		deviceStatsQuery += " GROUP BY device_id ORDER BY device_id"

		rows, err := r.db.QueryContext(ctx, deviceStatsQuery, deviceArgs...)
		if err == nil {
			defer rows.Close()

			for rows.Next() {
				var deviceStat interfaces.DeviceStats
				var firstTS, lastTS time.Time

				if err := rows.Scan(&deviceStat.DeviceID, &deviceStat.Count, &firstTS, &lastTS); err == nil {
					deviceStat.PiID = params.PiID
					deviceStat.FirstTS = &firstTS
					deviceStat.LastTS = &lastTS
					stats.ByDevice = append(stats.ByDevice, deviceStat)
				}
			}
		}
	}

	return stats, nil
}
