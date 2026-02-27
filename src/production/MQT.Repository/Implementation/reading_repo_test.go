package implementation

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

func TestPostgresReadingRepository_CreateReading(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ts := time.Now()
	payload := hardware_models.ReadingPayload{DeviceID: "d1", PiID: "p1"}
	reading := hardware_models.Reading{PiID: "p1", DeviceID: "d1", Ts: ts, Payload: payload}
	payloadJSON := []byte(`{"device_id":"d1","pi_id":"p1","timestamp":"","sensors":{},"battery_percentage":0}`)

	mock.ExpectExec("INSERT INTO readings").
		WithArgs("p1", "d1", ts, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewPostgresReadingRepository(db)
	err = repo.CreateReading(context.Background(), reading)
	if err != nil {
		t.Fatalf("CreateReading: %v", err)
	}
	// Use AnyArg for payload since json.Marshal output may vary
	_ = payloadJSON
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_CreateReadings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ts := time.Now()
	readings := []hardware_models.Reading{
		{PiID: "p1", DeviceID: "d1", Ts: ts, Payload: hardware_models.ReadingPayload{}},
	}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO readings").
		WithArgs("p1", "d1", ts, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewPostgresReadingRepository(db)
	err = repo.CreateReadings(context.Background(), readings)
	if err != nil {
		t.Fatalf("CreateReadings: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_CreateReadings_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresReadingRepository(db)
	err = repo.CreateReadings(context.Background(), nil)
	if err != nil {
		t.Fatalf("CreateReadings empty: %v", err)
	}
	// No DB calls expected
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_GetLatestReadings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ts := time.Now()
	payloadJSON := []byte(`{"device_id":"d1","pi_id":"p1","timestamp":"","sensors":{},"battery_percentage":0}`)
	rows := sqlmock.NewRows([]string{"pi_id", "device_id", "ts", "payload"}).
		AddRow("p1", "d1", ts, payloadJSON)
	mock.ExpectQuery("SELECT DISTINCT ON").
		WithArgs("p1").
		WillReturnRows(rows)

	repo := NewPostgresReadingRepository(db)
	readings, err := repo.GetLatestReadings(context.Background(), "p1")
	if err != nil {
		t.Fatalf("GetLatestReadings: %v", err)
	}
	if len(readings) != 1 || readings[0].PiID != "p1" {
		t.Errorf("unexpected readings: %+v", readings)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_GetReadings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ts := time.Now()
	payloadJSON := []byte(`{"device_id":"d1","pi_id":"p1","timestamp":"","sensors":{},"battery_percentage":0}`)
	rows := sqlmock.NewRows([]string{"pi_id", "device_id", "ts", "payload"}).
		AddRow("p1", "d1", ts, payloadJSON)
	mock.ExpectQuery("SELECT pi_id, device_id, ts, payload FROM readings").
		WithArgs(10, 0).
		WillReturnRows(rows)

	repo := NewPostgresReadingRepository(db)
	params := interfaces.ReadingQueryParams{Limit: 10, Page: 1}
	result, err := repo.GetReadings(context.Background(), params)
	if err != nil {
		t.Fatalf("GetReadings: %v", err)
	}
	if result == nil || len(result.Items) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_GetReadingsByDevice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	ts := time.Now()
	payloadJSON := []byte(`{"device_id":"d1","pi_id":"p1","timestamp":"","sensors":{},"battery_percentage":0}`)
	rows := sqlmock.NewRows([]string{"pi_id", "device_id", "ts", "payload"}).
		AddRow("p1", "d1", ts, payloadJSON)
	mock.ExpectQuery("SELECT pi_id, device_id, ts, payload FROM readings").
		WithArgs("p1", "d1", 10, 0).
		WillReturnRows(rows)

	repo := NewPostgresReadingRepository(db)
	params := interfaces.ReadingQueryParams{Limit: 10, Page: 1}
	result, err := repo.GetReadingsByDevice(context.Background(), "p1", "d1", params)
	if err != nil {
		t.Fatalf("GetReadingsByDevice: %v", err)
	}
	if result == nil || len(result.Items) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_GetSummaryStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	firstTS := time.Now().Add(-time.Hour)
	lastTS := time.Now()
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM readings").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
	mock.ExpectQuery("SELECT MIN\\(ts\\), MAX\\(ts\\) FROM readings").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"min", "max"}).AddRow(firstTS, lastTS))

	repo := NewPostgresReadingRepository(db)
	params := interfaces.ReadingQueryParams{}
	result, err := repo.GetSummaryStats(context.Background(), params)
	if err != nil {
		t.Fatalf("GetSummaryStats: %v", err)
	}
	if result == nil || result.Count != 42 {
		t.Errorf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresReadingRepository_DeleteReadingsByTimeRange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	start := time.Now().Add(-time.Hour)
	end := time.Now()
	mock.ExpectExec("DELETE FROM readings").
		WithArgs("p1", "d1", start, end).
		WillReturnResult(sqlmock.NewResult(0, 5))

	repo := NewPostgresReadingRepository(db)
	err = repo.DeleteReadingsByTimeRange(context.Background(), "p1", "d1", start, end)
	if err != nil {
		t.Fatalf("DeleteReadingsByTimeRange: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
