package implementation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create user
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user mqtmodels.User) error {
	query := `
		INSERT INTO users (user_id, name, role, created_at, meta) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) 
		DO UPDATE SET name = EXCLUDED.name, role = EXCLUDED.role, meta = EXCLUDED.meta
	`

	metaJSON, err := json.Marshal(ensureMetaNotNull(user.Meta))
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, user.UserID, user.Name, user.Role, user.CreatedAt, metaJSON)
	return err
}

// Read users
func (r *PostgresUserRepository) GetUser(ctx context.Context, userID string) (*mqtmodels.User, error) {
	query := `SELECT user_id, name, role, created_at, meta FROM users WHERE user_id = $1`

	var user mqtmodels.User
	var metaJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&user.UserID, &user.Name, &user.Role, &user.CreatedAt, &metaJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(metaJSON, &user.Meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) ListUsers(ctx context.Context, page, pageSize int, role string) (*interfaces.PaginationResult, error) {
	offset := (page - 1) * pageSize
	var query string
	var args []interface{}

	if role != "" {
		query = `SELECT user_id, name, role, created_at, meta FROM users WHERE role = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{role, pageSize, offset}
	} else {
		query = `SELECT user_id, name, role, created_at, meta FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{pageSize, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []mqtmodels.User
	for rows.Next() {
		var user mqtmodels.User
		var metaJSON []byte

		if err := rows.Scan(&user.UserID, &user.Name, &user.Role, &user.CreatedAt, &metaJSON); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metaJSON, &user.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &interfaces.PaginationResult{
		Items: users,
	}

	// Check if there are more pages
	if len(users) == pageSize {
		nextPage := page + 1
		result.NextPage = &nextPage
	}

	return result, nil
}

// Update user
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user mqtmodels.User) error {
	query := `
		UPDATE users 
		SET name = $1, role = $2, meta = $3 
		WHERE user_id = $4
	`

	metaJSON, err := json.Marshal(ensureMetaNotNull(user.Meta))
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, user.Name, user.Role, metaJSON, user.UserID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete user
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, userID string, hardDelete bool) error {
	var query string
	if hardDelete {
		query = `DELETE FROM users WHERE user_id = $1`
	} else {
		// Soft delete - you might want to add a deleted_at column
		query = `UPDATE users SET role = 'deleted' WHERE user_id = $1`
	}

	result, err := r.db.ExecContext(ctx, query, userID)
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
