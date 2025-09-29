package implementation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type PostgresPiRepository struct {
	db *sql.DB
}

func NewPostgresPiRepository(db *sql.DB) *PostgresPiRepository {
	return &PostgresPiRepository{db: db}
}

// Create pi (idempotent upsert)
func (r *PostgresPiRepository) CreateOrUpdatePi(ctx context.Context, pi mqtmodels.Pi) error {
	query := `
		INSERT INTO pis (pi_id, user_id, created_at, meta) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (pi_id) 
		DO UPDATE SET user_id = EXCLUDED.user_id, meta = EXCLUDED.meta
	`

	metaJSON, err := json.Marshal(ensureMetaNotNull(pi.Meta))
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, pi.PiID, pi.UserID, pi.CreatedAt, metaJSON)
	return err
}

// Read pis
func (r *PostgresPiRepository) GetPi(ctx context.Context, piID string) (*mqtmodels.Pi, error) {
	query := `SELECT pi_id, user_id, created_at, meta FROM pis WHERE pi_id = $1`

	var pi mqtmodels.Pi
	var metaJSON []byte

	err := r.db.QueryRowContext(ctx, query, piID).Scan(&pi.PiID, &pi.UserID, &pi.CreatedAt, &metaJSON)
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

func (r *PostgresPiRepository) ListPis(ctx context.Context, userID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	offset := (page - 1) * pageSize
	var query string
	var args []interface{}

	if userID != "" {
		query = `SELECT pi_id, user_id, created_at, meta FROM pis WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{userID, pageSize, offset}
	} else {
		query = `SELECT pi_id, user_id, created_at, meta FROM pis ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{pageSize, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pis []mqtmodels.Pi
	for rows.Next() {
		var pi mqtmodels.Pi
		var metaJSON []byte

		if err := rows.Scan(&pi.PiID, &pi.UserID, &pi.CreatedAt, &metaJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metaJSON, &pi.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}

		pis = append(pis, pi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &interfaces.PaginationResult{
		Items: pis,
	}

	// Check if there are more pages
	if len(pis) == pageSize {
		nextPage := page + 1
		result.NextPage = &nextPage
	}

	return result, nil
}

// Update pi
func (r *PostgresPiRepository) UpdatePi(ctx context.Context, pi mqtmodels.Pi) error {
	query := `
		UPDATE pis 
		SET user_id = $1, meta = $2 
		WHERE pi_id = $3
	`

	metaJSON, err := json.Marshal(ensureMetaNotNull(pi.Meta))
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, pi.UserID, metaJSON, pi.PiID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pi not found")
	}

	return nil
}

// Delete pi
func (r *PostgresPiRepository) DeletePi(ctx context.Context, piID string, cascade bool) error {
	var query string
	if cascade {
		// Delete associated devices and readings first
		// Note: This would need to be implemented with proper foreign key constraints
		query = `DELETE FROM pis WHERE pi_id = $1`
	} else {
		query = `DELETE FROM pis WHERE pi_id = $1`
	}

	result, err := r.db.ExecContext(ctx, query, piID)
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
