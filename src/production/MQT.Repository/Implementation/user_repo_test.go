package implementation

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	user := auth_models.NewUser("alice", "alice@test.com", "hash", "user", nil, nil)
	user.UserID = "u1"
	mock.ExpectExec("INSERT INTO users").
		WithArgs("u1", "alice", "alice@test.com", "hash", "user", true, nil, nil, nil, []byte("[]"), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresUserRepository(db)
	_, err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "alice@test.com", "hash", "user", true, nil, nil, nil, []byte("[]"), nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users WHERE user_id").
		WithArgs("u1").
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	user, err := repo.GetByID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if user == nil || user.Username != "alice" {
		t.Errorf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .+ FROM users WHERE user_id").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresUserRepository(db)
	user, err := repo.GetByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if user != nil {
		t.Error("expected nil for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "a@b.com", "h", "user", true, nil, nil, nil, []byte("[]"), nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users WHERE username").
		WithArgs("alice").
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	user, err := repo.GetByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if user == nil || user.Username != "alice" {
		t.Errorf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "a@b.com", "h", "user", true, nil, nil, nil, []byte("[]"), nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users WHERE email").
		WithArgs("a@b.com").
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	user, err := repo.GetByEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if user == nil || user.Email != "a@b.com" {
		t.Errorf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "a@b.com", "h", "user", true, nil, nil, nil, []byte("[]"), nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users ORDER BY").
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	users, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "a@b.com", "h", "user", true, nil, nil, nil, []byte("[]"), nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users").
		WithArgs(10, 0).
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	result, err := repo.List(context.Background(), 1, 10, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	user := auth_models.NewUser("alice", "a@b.com", "h", "user", nil, nil)
	user.UserID = "u1"
	mock.ExpectExec("UPDATE users").
		WithArgs("alice", "a@b.com", "h", "user", true, nil, nil, nil, []byte("[]"), nil, sqlmock.AnyArg(), "u1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresUserRepository(db)
	err = repo.Update(context.Background(), user)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetByRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "a@b.com", "h", "admin", true, nil, nil, nil, []byte("[]"), nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users WHERE role").
		WithArgs("admin").
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	users, err := repo.GetByRole(context.Background(), "admin")
	if err != nil {
		t.Fatalf("GetByRole: %v", err)
	}
	if len(users) != 1 || users[0].Role != "admin" {
		t.Errorf("unexpected users: %+v", users)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_GetUserByDeviceID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Now()
	locJSON := []byte(`[{"device_id":"d1","pi_id":"p1","latitude":1,"longitude":2}]`)
	rows := sqlmock.NewRows([]string{"user_id", "username", "email", "password", "role", "active", "email_verified_at", "latitude", "longitude", "locations", "sap_alert_threshold_percent", "created_at", "updated_at"}).
		AddRow("u1", "alice", "a@b.com", "h", "user", true, nil, nil, nil, locJSON, nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM users WHERE locations").
		WithArgs(`[{"device_id": "d1"}]`).
		WillReturnRows(rows)

	repo := NewPostgresUserRepository(db)
	user, err := repo.GetUserByDeviceID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("GetUserByDeviceID: %v", err)
	}
	if user == nil || user.UserID != "u1" {
		t.Errorf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_Delete_Hard(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM users").
		WithArgs("u1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresUserRepository(db)
	err = repo.Delete(context.Background(), "u1", true)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresUserRepository_Delete_Soft(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE users SET active").
		WithArgs("u1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresUserRepository(db)
	err = repo.Delete(context.Background(), "u1", false)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
