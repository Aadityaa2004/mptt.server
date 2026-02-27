package implementation

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

func TestPostgresDeviceRepository_CreateOrUpdateDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO devices").
		WithArgs("p1", "d1", 10.0, 5.0, 4.0, true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresDeviceRepository(db)
	device := hardware_models.Device{PiID: "p1", DeviceID: "d1", Height: 10, TopDiameter: 5, BottomDiameter: 4, CollectionEnabled: true}
	err = repo.CreateOrUpdateDevice(context.Background(), device)
	if err != nil {
		t.Fatalf("CreateOrUpdateDevice: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresDeviceRepository_GetDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"pi_id", "device_id", "height", "top_diameter", "bottom_diameter", "collection_enabled"}).
		AddRow("p1", "d1", 10.0, 5.0, 4.0, true)
	mock.ExpectQuery("SELECT .+ FROM devices").
		WithArgs("p1", "d1").
		WillReturnRows(rows)

	repo := NewPostgresDeviceRepository(db)
	device, err := repo.GetDevice(context.Background(), "p1", "d1")
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if device == nil || device.PiID != "p1" || device.DeviceID != "d1" {
		t.Errorf("unexpected device: %+v", device)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresDeviceRepository_GetDevice_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT .+ FROM devices").
		WithArgs("p1", "nonexistent").
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresDeviceRepository(db)
	device, err := repo.GetDevice(context.Background(), "p1", "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
	if device != nil {
		t.Error("expected nil for not found")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresDeviceRepository_ListDevicesByPi(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"pi_id", "device_id", "height", "top_diameter", "bottom_diameter", "collection_enabled"}).
		AddRow("p1", "d1", 10.0, 5.0, 4.0, true)
	mock.ExpectQuery("SELECT .+ FROM devices").
		WithArgs("p1", 10, 0).
		WillReturnRows(rows)

	repo := NewPostgresDeviceRepository(db)
	result, err := repo.ListDevicesByPi(context.Background(), "p1", 1, 10)
	if err != nil {
		t.Fatalf("ListDevicesByPi: %v", err)
	}
	devs, ok := result.Items.([]hardware_models.Device)
	if !ok || result == nil || len(devs) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresDeviceRepository_UpdateDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE devices").
		WithArgs(12.0, 6.0, 5.0, false, "p1", "d1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresDeviceRepository(db)
	device := hardware_models.Device{PiID: "p1", DeviceID: "d1", Height: 12, TopDiameter: 6, BottomDiameter: 5, CollectionEnabled: false}
	err = repo.UpdateDevice(context.Background(), device)
	if err != nil {
		t.Fatalf("UpdateDevice: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresDeviceRepository_DeleteDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM devices").
		WithArgs("p1", "d1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresDeviceRepository(db)
	err = repo.DeleteDevice(context.Background(), "p1", "d1", false)
	if err != nil {
		t.Fatalf("DeleteDevice: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
