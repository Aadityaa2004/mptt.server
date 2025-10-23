package implementation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create user
func (r *PostgresUserRepository) Create(ctx context.Context, user *auth_models.User) (*auth_models.User, error) {
	if user.UserID == "" {
		user.UserID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (user_id, username, email, password, role, active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) 
		DO UPDATE SET username = EXCLUDED.username, email = EXCLUDED.email, password = EXCLUDED.password, 
		              role = EXCLUDED.role, active = EXCLUDED.active, 
		              updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query, user.UserID, user.Username, user.Email,
		user.Password, user.Role, user.Active, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Read users
func (r *PostgresUserRepository) GetByID(ctx context.Context, userID string) (*auth_models.User, error) {
	query := `SELECT user_id, username, email, password, role, active, created_at, updated_at FROM users WHERE user_id = $1`

	var user auth_models.User

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&user.UserID, &user.Username, &user.Email,
		&user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// FindByID is an alias for GetByID for backward compatibility
func (r *PostgresUserRepository) FindByID(ctx context.Context, userID string) (*auth_models.User, error) {
	return r.GetByID(ctx, userID)
}

// GetUser is an alias for GetByID for backward compatibility
func (r *PostgresUserRepository) GetUser(ctx context.Context, userID string) (*auth_models.User, error) {
	return r.GetByID(ctx, userID)
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*auth_models.User, error) {
	query := `SELECT user_id, username, email, password, role, active, created_at, updated_at FROM users WHERE username = $1`

	var user auth_models.User

	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.UserID, &user.Username, &user.Email,
		&user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetAll(ctx context.Context) ([]*auth_models.User, error) {
	query := `SELECT user_id, username, email, password, role, active, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*auth_models.User
	for rows.Next() {
		var user auth_models.User

		if err := rows.Scan(&user.UserID, &user.Username, &user.Email,
			&user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *PostgresUserRepository) List(ctx context.Context, page, pageSize int, role string) (*interfaces.PaginationResult, error) {
	offset := (page - 1) * pageSize
	var query string
	var args []interface{}

	if role != "" {
		query = `SELECT user_id, username, email, password, role, active, created_at, updated_at FROM users WHERE role = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{role, pageSize, offset}
	} else {
		query = `SELECT user_id, username, email, password, role, active, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{pageSize, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []auth_models.User
	for rows.Next() {
		var user auth_models.User

		if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Password,
			&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
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
func (r *PostgresUserRepository) Update(ctx context.Context, user *auth_models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET username = $1, email = $2, password = $3, role = $4, active = $5, updated_at = $6 
		WHERE user_id = $7
	`

	result, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Password,
		user.Role, user.Active, user.UpdatedAt, user.UserID)
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

// GetByRole retrieves users by role
func (r *PostgresUserRepository) GetByRole(ctx context.Context, role string) ([]*auth_models.User, error) {
	query := `SELECT user_id, username, email, password, role, active, created_at, updated_at FROM users WHERE role = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*auth_models.User
	for rows.Next() {
		var user auth_models.User

		if err := rows.Scan(&user.UserID, &user.Username, &user.Email,
			&user.Password, &user.Role, &user.Active,
			&user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Delete user
func (r *PostgresUserRepository) Delete(ctx context.Context, userID string, hardDelete bool) error {
	var query string
	if hardDelete {
		query = `DELETE FROM users WHERE user_id = $1`
	} else {
		// Soft delete - set active to false
		query = `UPDATE users SET active = false, updated_at = now() WHERE user_id = $1`
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
