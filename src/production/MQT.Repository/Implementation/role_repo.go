package implementation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

type PostgresRoleRepository struct {
	db *sql.DB
}

func NewPostgresRoleRepository(db *sql.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

// Create adds a new role to the database
func (r *PostgresRoleRepository) Create(ctx context.Context, role *auth_models.Role) (*auth_models.Role, error) {
	if role.RoleID == "" {
		role.RoleID = uuid.New().String()
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	query := `
		INSERT INTO roles (role_id, name, description, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (role_id) 
		DO UPDATE SET name = EXCLUDED.name, 
		              description = EXCLUDED.description, updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query, role.RoleID, role.Name,
		role.Description, role.CreatedAt, role.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// FindByID finds a role by ID
func (r *PostgresRoleRepository) FindByID(ctx context.Context, id string) (*auth_models.Role, error) {
	query := `SELECT role_id, name, description, created_at, updated_at FROM roles WHERE role_id = $1`

	var role auth_models.Role

	err := r.db.QueryRowContext(ctx, query, id).Scan(&role.RoleID, &role.Name,
		&role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &role, nil
}

// FindByName finds a role by name
func (r *PostgresRoleRepository) FindByName(ctx context.Context, name string) (*auth_models.Role, error) {
	query := `SELECT role_id, name, description, created_at, updated_at FROM roles WHERE name = $1`

	var role auth_models.Role

	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.RoleID, &role.Name,
		&role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &role, nil
}

// FindAll retrieves all roles
func (r *PostgresRoleRepository) FindAll(ctx context.Context) ([]*auth_models.Role, error) {
	query := `SELECT role_id, name, description, created_at, updated_at FROM roles ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*auth_models.Role
	for rows.Next() {
		var role auth_models.Role

		if err := rows.Scan(&role.RoleID, &role.Name,
			&role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}

		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// Update updates a role
func (r *PostgresRoleRepository) Update(ctx context.Context, role *auth_models.Role) error {
	role.UpdatedAt = time.Now()

	query := `
		UPDATE roles 
		SET name = $1, description = $2, updated_at = $3 
		WHERE role_id = $4
	`

	result, err := r.db.ExecContext(ctx, query, role.Name,
		role.Description, role.UpdatedAt, role.RoleID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// Delete deletes a role
func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM roles WHERE role_id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
