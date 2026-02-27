package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

var errNotFound = errors.New("not found")

func TestInternalController_ValidatePi_Exists(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	piRepo := &mockPiRepoForPiController{pi: &hardware_models.Pi{PiID: "p1", UserID: "u1"}}
	deviceRepo := &mockDeviceRepo{}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{BucketThreshold: 75, EmailServiceURL: "http://test"}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	body, _ := json.Marshal(map[string]string{"pi_id": "p1"})
	req := httptest.NewRequest("POST", "/internal/pis/validate", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ValidatePi exists: status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Exists {
		t.Error("expected exists=true")
	}
}

func TestInternalController_ValidatePi_NotExists(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	// When pi is nil and no match, GetPi returns (nil, nil). The controller treats err==nil as "exists".
	// So we need the mock to return an error for "not found". Use a fresh mock that returns err.
	piRepo := &mockPiRepoForPiController{err: errNotFound}
	deviceRepo := &mockDeviceRepo{}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{BucketThreshold: 75, EmailServiceURL: "http://test"}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	body, _ := json.Marshal(map[string]string{"pi_id": "nonexistent"})
	req := httptest.NewRequest("POST", "/internal/pis/validate", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ValidatePi not exists: status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Exists {
		t.Error("expected exists=false")
	}
}

func TestInternalController_ValidatePi_Unauthorized(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	body, _ := json.Marshal(map[string]string{"pi_id": "p1"})
	req := httptest.NewRequest("POST", "/internal/pis/validate", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer wrong-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestInternalController_ValidateDevice_Exists(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{device: &hardware_models.Device{PiID: "p1", DeviceID: "d1"}}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	body, _ := json.Marshal(map[string]string{"pi_id": "p1", "device_id": "d1"})
	req := httptest.NewRequest("POST", "/internal/devices/validate", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ValidateDevice exists: status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Exists {
		t.Error("expected exists=true")
	}
}

func TestInternalController_ValidateDevice_NotExists(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{err: errNotFound}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	body, _ := json.Marshal(map[string]string{"pi_id": "p1", "device_id": "nonexistent"})
	req := httptest.NewRequest("POST", "/internal/devices/validate", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ValidateDevice not exists: status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Exists bool `json:"exists"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Exists {
		t.Error("expected exists=false")
	}
}

func TestInternalController_CreateReading(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{device: &hardware_models.Device{PiID: "p1", DeviceID: "d1", CollectionEnabled: true}}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	payload := map[string]interface{}{
		"sensors": map[string]interface{}{},
	}
	body, _ := json.Marshal(map[string]interface{}{
		"pi_id":     "p1",
		"device_id": "d1",
		"ts":        time.Now().UTC().Format(time.RFC3339),
		"payload":   payload,
	})
	req := httptest.NewRequest("POST", "/internal/readings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("CreateReading: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestInternalController_CreateReading_CollectionDisabled(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-internal-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	piRepo := &mockPiRepoForPiController{}
	deviceRepo := &mockDeviceRepo{device: &hardware_models.Device{PiID: "p1", DeviceID: "d1", CollectionEnabled: false}}
	readingRepo := &mockReadingRepoForController{}
	userRepo := &mockUserRepoForController{}
	cfg := &config.Config{Alert: config.AlertConfig{}}

	ic := NewInternalController(piRepo, deviceRepo, readingRepo, userRepo, cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ic.RegisterRoutes(r)

	payload := map[string]interface{}{
		"sensors": map[string]interface{}{},
	}
	body, _ := json.Marshal(map[string]interface{}{
		"pi_id":     "p1",
		"device_id": "d1",
		"ts":        time.Now().UTC().Format(time.RFC3339),
		"payload":   payload,
	})
	req := httptest.NewRequest("POST", "/internal/readings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer test-internal-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("CreateReading collection disabled: status=%d body=%s", w.Code, w.Body.String())
	}
}

