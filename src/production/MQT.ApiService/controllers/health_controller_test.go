package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type mockReadingRepo struct {
	summaryStats *interfaces.SummaryStats
	err          error
}

func (m *mockReadingRepo) CreateReading(ctx context.Context, r hardware_models.Reading) error      { return nil }
func (m *mockReadingRepo) CreateReadings(ctx context.Context, readings []hardware_models.Reading) error { return nil }
func (m *mockReadingRepo) GetLatestReadings(ctx context.Context, piID string) ([]hardware_models.Reading, error) {
	return nil, nil
}
func (m *mockReadingRepo) GetReadings(ctx context.Context, params interfaces.ReadingQueryParams) (*interfaces.ReadingQueryResult, error) {
	return nil, nil
}
func (m *mockReadingRepo) GetReadingsByDevice(ctx context.Context, piID, deviceID string, params interfaces.ReadingQueryParams) (*interfaces.ReadingQueryResult, error) {
	return nil, nil
}
func (m *mockReadingRepo) GetSummaryStats(ctx context.Context, params interfaces.ReadingQueryParams) (*interfaces.SummaryStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.summaryStats != nil {
		return m.summaryStats, nil
	}
	return &interfaces.SummaryStats{Count: 0}, nil
}
func (m *mockReadingRepo) DeleteReadingsByTimeRange(ctx context.Context, piID, deviceID string, start, end time.Time) error {
	return nil
}

type mockPiRepo struct {
	pi  *hardware_models.Pi
	err error
}

func (m *mockPiRepo) CreateOrUpdatePi(ctx context.Context, pi hardware_models.Pi) error { return nil }
func (m *mockPiRepo) GetPi(ctx context.Context, piID string) (*hardware_models.Pi, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pi, nil
}
func (m *mockPiRepo) ListPis(ctx context.Context, userID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	return nil, nil
}
func (m *mockPiRepo) UpdatePi(ctx context.Context, pi hardware_models.Pi) error         { return nil }
func (m *mockPiRepo) DeletePi(ctx context.Context, piID string, cascade bool) error     { return nil }
func (m *mockPiRepo) UnassignPisByUserID(ctx context.Context, userID string) (int64, error) { return 0, nil }

func setupHealthRouter(t *testing.T, readingRepo interfaces.ReadingRepository, piRepo interfaces.PiRepository, authMw *middleware.AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	hc := NewHealthController(readingRepo, piRepo, log, authMw)
	r := gin.New()
	hc.RegisterRoutes(r)
	return r
}

func newAuthMiddleware(t *testing.T) *middleware.AuthMiddleware {
	apiCfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(apiCfg)
	rbacSvc := rbac.NewService()
	return middleware.NewAuthMiddleware(jwtSvc, rbacSvc, middleware.DefaultConfig())
}

func TestHealthController_HealthLive(t *testing.T) {
	r := setupHealthRouter(t, &mockReadingRepo{}, &mockPiRepo{}, newAuthMiddleware(t))
	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("HealthLive: status=%d", w.Code)
	}
}

func TestHealthController_HealthReady(t *testing.T) {
	r := setupHealthRouter(t, &mockReadingRepo{}, &mockPiRepo{}, newAuthMiddleware(t))
	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("HealthReady: status=%d", w.Code)
	}
}

func TestHealthController_Metrics(t *testing.T) {
	r := setupHealthRouter(t, &mockReadingRepo{}, &mockPiRepo{}, newAuthMiddleware(t))
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Metrics: status=%d", w.Code)
	}
}

func TestHealthController_GetSummaryStats_Admin(t *testing.T) {
	authMw := newAuthMiddleware(t)
	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "admin")

	readingRepo := &mockReadingRepo{summaryStats: &interfaces.SummaryStats{Count: 42}}
	r := setupHealthRouter(t, readingRepo, &mockPiRepo{}, authMw)

	req := httptest.NewRequest("GET", "/stats/summary", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetSummaryStats admin: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHealthController_GetSummaryStats_NonAdmin_NoPiID(t *testing.T) {
	authMw := newAuthMiddleware(t)
	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupHealthRouter(t, &mockReadingRepo{}, &mockPiRepo{}, authMw)

	req := httptest.NewRequest("GET", "/stats/summary", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-admin without pi_id, got %d", w.Code)
	}
}

func TestHealthController_GetSummaryStats_NonAdmin_Forbidden(t *testing.T) {
	authMw := newAuthMiddleware(t)
	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	piRepo := &mockPiRepo{pi: &hardware_models.Pi{PiID: "p1", UserID: "other-user"}}
	r := setupHealthRouter(t, &mockReadingRepo{}, piRepo, authMw)

	req := httptest.NewRequest("GET", "/stats/summary?pi_id=p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 when accessing another user's pi, got %d", w.Code)
	}
}

func TestHealthController_GetSummaryStats_NonAdmin_OK(t *testing.T) {
	authMw := newAuthMiddleware(t)
	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	piRepo := &mockPiRepo{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	readingRepo := &mockReadingRepo{summaryStats: &interfaces.SummaryStats{Count: 10}}
	r := setupHealthRouter(t, readingRepo, piRepo, authMw)

	req := httptest.NewRequest("GET", "/stats/summary?pi_id=p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetSummaryStats user: status=%d body=%s", w.Code, w.Body.String())
	}
}

var _ interfaces.ReadingRepository = (*mockReadingRepo)(nil)
var _ interfaces.PiRepository = (*mockPiRepo)(nil)
