package health

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealthChecker_PingPostgres_NilDB(t *testing.T) {
	hc := NewHealthChecker(nil)
	ctx := context.Background()
	err := hc.PingPostgres(ctx)
	if err == nil {
		t.Fatal("expected error for nil db")
	}
	if err.Error() != "database connection is nil" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHealthChecker_GetHealthStatus_NilDB(t *testing.T) {
	hc := NewHealthChecker(nil)
	ctx := context.Background()
	status := hc.GetHealthStatus(ctx)
	if status["status"] != "degraded" {
		t.Errorf("expected degraded, got %v", status["status"])
	}
	checks := status["checks"].(map[string]interface{})
	pg := checks["postgres"].(map[string]interface{})
	if pg["status"] != "error" {
		t.Errorf("expected postgres status error, got %v", pg["status"])
	}
}

func TestHealthChecker_PingPostgres_OK(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()

	hc := NewHealthChecker(db)
	ctx := context.Background()
	err = hc.PingPostgres(ctx)
	if err != nil {
		t.Fatalf("PingPostgres: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestHealthChecker_CheckDatabaseHealth_OK(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()
	rows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery("SELECT 1").WillReturnRows(rows)

	hc := NewHealthChecker(db)
	ctx := context.Background()
	err = hc.CheckDatabaseHealth(ctx)
	if err != nil {
		t.Fatalf("CheckDatabaseHealth: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestHealthChecker_GetHealthStatus_OK(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectPing()
	rows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery("SELECT 1").WillReturnRows(rows)

	hc := NewHealthChecker(db)
	ctx := context.Background()
	status := hc.GetHealthStatus(ctx)
	if status["status"] != "ok" {
		t.Errorf("expected ok, got %v", status["status"])
	}
	checks := status["checks"].(map[string]interface{})
	pg := checks["postgres"].(map[string]interface{})
	if pg["status"] != "ok" {
		t.Errorf("expected postgres status ok, got %v", pg["status"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDatabaseManager_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.ExpectClose()
	dm := NewDatabaseManager(db)
	err = dm.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestDatabaseManager_Close_NilDB(t *testing.T) {
	dm := &DatabaseManager{db: nil}
	err := dm.Close()
	if err != nil {
		t.Fatalf("Close with nil db should not error: %v", err)
	}
}

