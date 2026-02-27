package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type mockReadingRepoForController struct {
	latest        []hardware_models.Reading
	readings      *interfaces.ReadingQueryResult
	byDevice      *interfaces.ReadingQueryResult
	err           error
	deleteErr     error
}

func (m *mockReadingRepoForController) CreateReading(ctx context.Context, r hardware_models.Reading) error {
	return nil
}
func (m *mockReadingRepoForController) CreateReadings(ctx context.Context, readings []hardware_models.Reading) error {
	return nil
}
func (m *mockReadingRepoForController) GetLatestReadings(ctx context.Context, piID string) ([]hardware_models.Reading, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.latest != nil {
		return m.latest, nil
	}
	return []hardware_models.Reading{}, nil
}
func (m *mockReadingRepoForController) GetReadings(ctx context.Context, params interfaces.ReadingQueryParams) (*interfaces.ReadingQueryResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.readings != nil {
		return m.readings, nil
	}
	return &interfaces.ReadingQueryResult{Items: []hardware_models.Reading{}, NextPageToken: nil, Total: 0}, nil
}
func (m *mockReadingRepoForController) GetReadingsByDevice(ctx context.Context, piID, deviceID string, params interfaces.ReadingQueryParams) (*interfaces.ReadingQueryResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.byDevice != nil {
		return m.byDevice, nil
	}
	return &interfaces.ReadingQueryResult{Items: []hardware_models.Reading{}, NextPageToken: nil, Total: 0}, nil
}
func (m *mockReadingRepoForController) GetSummaryStats(ctx context.Context, params interfaces.ReadingQueryParams) (*interfaces.SummaryStats, error) {
	return &interfaces.SummaryStats{Count: 0}, nil
}
func (m *mockReadingRepoForController) DeleteReadingsByTimeRange(ctx context.Context, piID, deviceID string, start, end time.Time) error {
	return m.deleteErr
}

func setupReadingRouter(rc *ReadingController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rc.RegisterRoutes(r)
	return r
}

func TestCalculateFillPercentage(t *testing.T) {
	tests := []struct {
		name            string
		height          float64
		topDiameter     float64
		bottomDiameter  float64
		sensorDistance  float64
		wantZeroOrMore  bool
	}{
		{"valid", 10, 5, 3, 2, true},
		{"zero height", 0, 5, 3, 2, true},
		{"sap full", 10, 5, 3, 0, true},
		{"sap above height", 10, 5, 3, 15, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateFillPercentage(tt.height, tt.topDiameter, tt.bottomDiameter, tt.sensorDistance)
			if tt.wantZeroOrMore && got < 0 {
				t.Errorf("calculateFillPercentage() = %v, want >= 0", got)
			}
			if got > 100 {
				t.Errorf("calculateFillPercentage() = %v, want <= 100", got)
			}
		})
	}
}

func TestReadingController_GetLatestReadings(t *testing.T) {
	readingRepo := &mockReadingRepoForController{
		latest: []hardware_models.Reading{{PiID: "p1", DeviceID: "d1", Ts: time.Now()}},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	deviceRepo := &mockDeviceRepo{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	rc := NewReadingController(readingRepo, piRepo, deviceRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupReadingRouter(rc)
	req := httptest.NewRequest("GET", "/readings/latest?pi_id=p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetLatestReadings: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestReadingController_GetLatestReadings_NoPiID(t *testing.T) {
	readingRepo := &mockReadingRepoForController{}
	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	rc := NewReadingController(readingRepo, piRepo, deviceRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupReadingRouter(rc)
	req := httptest.NewRequest("GET", "/readings/latest", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestReadingController_GetReadings(t *testing.T) {
	readingRepo := &mockReadingRepoForController{
		readings: &interfaces.ReadingQueryResult{Items: []hardware_models.Reading{}, NextPageToken: nil, Total: 0},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	deviceRepo := &mockDeviceRepo{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	rc := NewReadingController(readingRepo, piRepo, deviceRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupReadingRouter(rc)
	req := httptest.NewRequest("GET", "/readings?pi_id=p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetReadings: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestReadingController_GetDeviceReadings(t *testing.T) {
	readingRepo := &mockReadingRepoForController{
		byDevice: &interfaces.ReadingQueryResult{Items: []hardware_models.Reading{}, NextPageToken: nil, Total: 0},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	deviceRepo := &mockDeviceRepo{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	rc := NewReadingController(readingRepo, piRepo, deviceRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupReadingRouter(rc)
	req := httptest.NewRequest("GET", "/readings/pis/p1/devices/d1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetDeviceReadings: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestReadingController_DeleteDeviceReadingsByTimeRange(t *testing.T) {
	readingRepo := &mockReadingRepoForController{}
	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	rc := NewReadingController(readingRepo, piRepo, deviceRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	from := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	to := time.Now().Format(time.RFC3339)
	r := setupReadingRouter(rc)
	req := httptest.NewRequest("DELETE", "/readings/pis/p1/devices/d1?from="+from+"&to="+to, nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("DeleteDeviceReadingsByTimeRange: status=%d body=%s", w.Code, w.Body.String())
	}
}
