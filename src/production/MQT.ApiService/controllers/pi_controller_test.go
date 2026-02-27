package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type mockPiRepoForPiController struct {
	pi       *hardware_models.Pi
	list     *interfaces.PaginationResult
	err      error
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockPiRepoForPiController) CreateOrUpdatePi(ctx context.Context, p hardware_models.Pi) error {
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}
func (m *mockPiRepoForPiController) GetPi(ctx context.Context, id string) (*hardware_models.Pi, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.pi != nil && m.pi.PiID == id {
		return m.pi, nil
	}
	return nil, nil
}
func (m *mockPiRepoForPiController) ListPis(ctx context.Context, userID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.list != nil {
		return m.list, nil
	}
	return &interfaces.PaginationResult{Items: []hardware_models.Pi{}, NextPage: nil}, nil
}
func (m *mockPiRepoForPiController) UpdatePi(ctx context.Context, p hardware_models.Pi) error {
	return m.updateErr
}
func (m *mockPiRepoForPiController) DeletePi(ctx context.Context, id string, cascade bool) error {
	return m.deleteErr
}
func (m *mockPiRepoForPiController) UnassignPisByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

func setupPiRouter(pc *PiController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	pc.RegisterRoutes(r)
	return r
}

func TestPiController_CreatePi(t *testing.T) {
	piRepo := &mockPiRepoForPiController{}
	userRepo := &mockUserRepoForController{byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice"}}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	body, _ := json.Marshal(map[string]string{"pi_id": "p1", "user_id": "u1"})
	r := setupPiRouter(pc)
	req := httptest.NewRequest("POST", "/pis", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("CreatePi: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPiController_CreatePi_NoUserID(t *testing.T) {
	piRepo := &mockPiRepoForPiController{}
	userRepo := &mockUserRepoForController{byID: map[string]*auth_models.User{}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	body, _ := json.Marshal(map[string]string{"pi_id": "p1"})
	r := setupPiRouter(pc)
	req := httptest.NewRequest("POST", "/pis", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("CreatePi no user: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPiController_ListPis_Admin(t *testing.T) {
	piRepo := &mockPiRepoForPiController{
		list: &interfaces.PaginationResult{Items: []hardware_models.Pi{{PiID: "p1", UserID: "u1"}}, NextPage: nil},
	}
	userRepo := &mockUserRepoForController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupPiRouter(pc)
	req := httptest.NewRequest("GET", "/pis", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ListPis: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPiController_GetPi_Admin(t *testing.T) {
	piRepo := &mockPiRepoForPiController{
		pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"},
	}
	userRepo := &mockUserRepoForController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupPiRouter(pc)
	req := httptest.NewRequest("GET", "/pis/p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetPi: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPiController_GetPi_Owner(t *testing.T) {
	piRepo := &mockPiRepoForPiController{
		pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"},
	}
	userRepo := &mockUserRepoForController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupPiRouter(pc)
	req := httptest.NewRequest("GET", "/pis/p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetPi owner: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPiController_GetPi_Forbidden(t *testing.T) {
	piRepo := &mockPiRepoForPiController{
		pi: &hardware_models.Pi{PiID: "p1", UserID: "other"},
	}
	userRepo := &mockUserRepoForController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupPiRouter(pc)
	req := httptest.NewRequest("GET", "/pis/p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestPiController_UpdatePi(t *testing.T) {
	piRepo := &mockPiRepoForPiController{
		pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"},
	}
	userRepo := &mockUserRepoForController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	uid := "u2"
	body, _ := json.Marshal(map[string]string{"user_id": uid})
	r := setupPiRouter(pc)
	req := httptest.NewRequest("PATCH", "/pis/p1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("UpdatePi: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestPiController_DeletePi(t *testing.T) {
	piRepo := &mockPiRepoForPiController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	pc := NewPiController(piRepo, userRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupPiRouter(pc)
	req := httptest.NewRequest("DELETE", "/pis/p1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("DeletePi: status=%d body=%s", w.Code, w.Body.String())
	}
}
