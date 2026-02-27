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
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type mockDeviceRepo struct {
	device   *hardware_models.Device
	list     *interfaces.PaginationResult
	err      error
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockDeviceRepo) CreateOrUpdateDevice(ctx context.Context, d hardware_models.Device) error {
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}
func (m *mockDeviceRepo) GetDevice(ctx context.Context, piID, deviceID string) (*hardware_models.Device, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.device != nil && m.device.PiID == piID && m.device.DeviceID == deviceID {
		return m.device, nil
	}
	return nil, nil
}
func (m *mockDeviceRepo) ListDevicesByPi(ctx context.Context, piID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.list != nil {
		return m.list, nil
	}
	return &interfaces.PaginationResult{Items: []hardware_models.Device{}, NextPage: nil}, nil
}
func (m *mockDeviceRepo) UpdateDevice(ctx context.Context, d hardware_models.Device) error {
	return m.updateErr
}
func (m *mockDeviceRepo) DeleteDevice(ctx context.Context, piID, deviceID string, cascade bool) error {
	return m.deleteErr
}

func setupDeviceRouter(dc *DeviceController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	dc.RegisterRoutes(r)
	return r
}

func TestDeviceController_CreateDevice(t *testing.T) {
	deviceRepo := &mockDeviceRepo{}
	piRepo := &mockPiRepoForPiController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	body, _ := json.Marshal(map[string]string{"device_id": "d1"})
	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("POST", "/pis/p1/devices", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("CreateDevice: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDeviceController_ListDevices_Admin(t *testing.T) {
	deviceRepo := &mockDeviceRepo{
		list: &interfaces.PaginationResult{Items: []hardware_models.Device{{PiID: "p1", DeviceID: "d1"}}, NextPage: nil},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("GET", "/pis/p1/devices", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ListDevices: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDeviceController_ListDevices_Owner(t *testing.T) {
	deviceRepo := &mockDeviceRepo{
		list: &interfaces.PaginationResult{Items: []hardware_models.Device{}, NextPage: nil},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("GET", "/pis/p1/devices", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ListDevices owner: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDeviceController_ListDevices_Forbidden(t *testing.T) {
	deviceRepo := &mockDeviceRepo{}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "other"}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("GET", "/pis/p1/devices", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestDeviceController_GetDevice(t *testing.T) {
	deviceRepo := &mockDeviceRepo{
		device: &hardware_models.Device{PiID: "p1", DeviceID: "d1"},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("GET", "/pis/p1/devices/d1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetDevice: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDeviceController_UpdateDevice(t *testing.T) {
	deviceRepo := &mockDeviceRepo{
		device: &hardware_models.Device{PiID: "p1", DeviceID: "d1", Height: 10},
	}
	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	h := 20.0
	body, _ := json.Marshal(map[string]float64{"height": h})
	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("PATCH", "/pis/p1/devices/d1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("UpdateDevice: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDeviceController_DeleteDevice(t *testing.T) {
	deviceRepo := &mockDeviceRepo{}
	piRepo := &mockPiRepoForPiController{}
	cfg := &config.LoggingConfig{Level: "info", Format: "text", Output: "stdout", EnableCaller: false}
	log := logger.NewLogger(cfg)
	dc := NewDeviceController(deviceRepo, piRepo, log, newAuthMiddleware(t))

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupDeviceRouter(dc)
	req := httptest.NewRequest("DELETE", "/pis/p1/devices/d1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("DeleteDevice: status=%d body=%s", w.Code, w.Body.String())
	}
}
