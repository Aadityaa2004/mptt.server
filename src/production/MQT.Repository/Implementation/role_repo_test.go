package implementation

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

func TestPostgresRoleRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO roles").
		WithArgs(sqlmock.AnyArg(), "admin", "Administrator", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresRoleRepository(db)
	role := auth_models.NewRole("admin", "Administrator")
	_, err = repo.Create(context.Background(), role)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresRoleRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"role_id", "name", "description", "created_at", "updated_at"}).
		AddRow("r1", "admin", "Administrator", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .+ FROM roles WHERE role_id").
		WithArgs("r1").
		WillReturnRows(rows)

	repo := NewPostgresRoleRepository(db)
	role, err := repo.FindByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if role == nil || role.Name != "admin" {
		t.Errorf("unexpected role: %+v", role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresRoleRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .+ FROM roles WHERE role_id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresRoleRepository(db)
	role, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if role != nil {
		t.Error("expected nil for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresRoleRepository_FindByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"role_id", "name", "description", "created_at", "updated_at"}).
		AddRow("r1", "user", "User", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .+ FROM roles WHERE name").
		WithArgs("user").
		WillReturnRows(rows)

	repo := NewPostgresRoleRepository(db)
	role, err := repo.FindByName(context.Background(), "user")
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}
	if role == nil || role.Name != "user" {
		t.Errorf("unexpected role: %+v", role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresRoleRepository_FindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"role_id", "name", "description", "created_at", "updated_at"}).
		AddRow("r1", "admin", "Admin", time.Now(), time.Now()).
		AddRow("r2", "user", "User", time.Now(), time.Now())
	mock.ExpectQuery("SELECT .+ FROM roles ORDER BY").
		WillReturnRows(rows)

	repo := NewPostgresRoleRepository(db)
	roles, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresRoleRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE roles").
		WithArgs("admin", "Updated desc", sqlmock.AnyArg(), "r1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresRoleRepository(db)
	role := &auth_models.Role{RoleID: "r1", Name: "admin", Description: "Updated desc"}
	err = repo.Update(context.Background(), role)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresRoleRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM roles").
		WithArgs("r1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresRoleRepository(db)
	err = repo.Delete(context.Background(), "r1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
