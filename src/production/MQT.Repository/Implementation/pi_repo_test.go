package implementation

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

func TestPostgresPiRepository_CreateOrUpdatePi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO pis").
		WithArgs("p1", "u1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresPiRepository(db)
	pi := hardware_models.Pi{PiID: "p1", UserID: "u1"}
	err = repo.CreateOrUpdatePi(context.Background(), pi)
	if err != nil {
		t.Fatalf("CreateOrUpdatePi: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresPiRepository_GetPi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"pi_id", "user_id"}).
		AddRow("p1", "u1")
	mock.ExpectQuery("SELECT pi_id, user_id FROM pis").
		WithArgs("p1").
		WillReturnRows(rows)

	repo := NewPostgresPiRepository(db)
	pi, err := repo.GetPi(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetPi: %v", err)
	}
	if pi == nil || pi.PiID != "p1" || pi.UserID != "u1" {
		t.Errorf("unexpected pi: %+v", pi)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresPiRepository_GetPi_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT pi_id, user_id FROM pis").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresPiRepository(db)
	pi, err := repo.GetPi(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetPi: %v", err)
	}
	if pi != nil {
		t.Error("expected nil for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresPiRepository_ListPis(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"pi_id", "user_id"}).
		AddRow("p1", "u1")
	mock.ExpectQuery("SELECT pi_id, user_id FROM pis").
		WithArgs("u1", 10, 0).
		WillReturnRows(rows)

	repo := NewPostgresPiRepository(db)
	result, err := repo.ListPis(context.Background(), "u1", 1, 10)
	if err != nil {
		t.Fatalf("ListPis: %v", err)
	}
	pis, ok := result.Items.([]hardware_models.Pi)
	if !ok || result == nil || len(pis) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresPiRepository_UpdatePi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE pis").
		WithArgs("u2", "p1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresPiRepository(db)
	pi := hardware_models.Pi{PiID: "p1", UserID: "u2"}
	err = repo.UpdatePi(context.Background(), pi)
	if err != nil {
		t.Fatalf("UpdatePi: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresPiRepository_UnassignPisByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE pis SET user_id = NULL").
		WithArgs("u1").
		WillReturnResult(sqlmock.NewResult(0, 3))

	repo := NewPostgresPiRepository(db)
	n, err := repo.UnassignPisByUserID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("UnassignPisByUserID: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 affected, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresPiRepository_DeletePi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM pis").
		WithArgs("p1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresPiRepository(db)
	err = repo.DeletePi(context.Background(), "p1", false)
	if err != nil {
		t.Fatalf("DeletePi: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
