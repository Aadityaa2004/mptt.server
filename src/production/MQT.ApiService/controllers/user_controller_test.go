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
	service "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type mockUserRepoForController struct {
	byID       map[string]*auth_models.User
	byUsername map[string]*auth_models.User
	byEmail    map[string]*auth_models.User
	byDevice   map[string]*auth_models.User
	getAllErr  error
	getErr     error
	updateErr  error
	deleteErr  error
}

func (m *mockUserRepoForController) Create(ctx context.Context, u *auth_models.User) (*auth_models.User, error) {
	if m.byID == nil {
		m.byID = make(map[string]*auth_models.User)
	}
	if m.byUsername == nil {
		m.byUsername = make(map[string]*auth_models.User)
	}
	if m.byEmail == nil {
		m.byEmail = make(map[string]*auth_models.User)
	}
	m.byID[u.UserID] = u
	m.byUsername[u.Username] = u
	m.byEmail[u.Email] = u
	return u, nil
}
func (m *mockUserRepoForController) GetByID(ctx context.Context, id string) (*auth_models.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.byID[id], nil
}
func (m *mockUserRepoForController) FindByID(ctx context.Context, id string) (*auth_models.User, error) {
	return m.byID[id], nil
}
func (m *mockUserRepoForController) GetByUsername(ctx context.Context, s string) (*auth_models.User, error) {
	if m.byUsername != nil {
		return m.byUsername[s], nil
	}
	return nil, nil
}
func (m *mockUserRepoForController) GetByEmail(ctx context.Context, s string) (*auth_models.User, error) {
	if m.byEmail != nil {
		return m.byEmail[s], nil
	}
	return nil, nil
}
func (m *mockUserRepoForController) GetAll(ctx context.Context) ([]*auth_models.User, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	var out []*auth_models.User
	for _, u := range m.byID {
		out = append(out, u)
	}
	return out, nil
}
func (m *mockUserRepoForController) List(ctx context.Context, page, pageSize int, role string) (*interfaces.PaginationResult, error) {
	return &interfaces.PaginationResult{Items: []auth_models.User{}, NextPage: nil}, nil
}
func (m *mockUserRepoForController) GetUser(ctx context.Context, id string) (*auth_models.User, error) {
	return m.byID[id], nil
}
func (m *mockUserRepoForController) GetByRole(ctx context.Context, s string) ([]*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForController) Update(ctx context.Context, u *auth_models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.byID == nil {
		m.byID = make(map[string]*auth_models.User)
	}
	if m.byUsername == nil {
		m.byUsername = make(map[string]*auth_models.User)
	}
	if m.byEmail == nil {
		m.byEmail = make(map[string]*auth_models.User)
	}
	m.byID[u.UserID] = u
	m.byUsername[u.Username] = u
	m.byEmail[u.Email] = u
	return nil
}
func (m *mockUserRepoForController) Delete(ctx context.Context, id string, hard bool) error {
	return m.deleteErr
}
func (m *mockUserRepoForController) GetUserByDeviceID(ctx context.Context, devID string) (*auth_models.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.byDevice[devID], nil
}

type mockPiRepoForController struct{}

func (m *mockPiRepoForController) CreateOrUpdatePi(ctx context.Context, p hardware_models.Pi) error {
	return nil
}
func (m *mockPiRepoForController) GetPi(ctx context.Context, id string) (*hardware_models.Pi, error) {
	return nil, nil
}
func (m *mockPiRepoForController) ListPis(ctx context.Context, userID string, page, pageSize int) (*interfaces.PaginationResult, error) {
	return nil, nil
}
func (m *mockPiRepoForController) UpdatePi(ctx context.Context, p hardware_models.Pi) error {
	return nil
}
func (m *mockPiRepoForController) DeletePi(ctx context.Context, id string, cascade bool) error {
	return nil
}
func (m *mockPiRepoForController) UnassignPisByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

func setupUserRouter(uc *UserController, authMw *middleware.AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uc.RegisterRoutes(r, authMw)
	return r
}

func TestUserController_GetAllUsers_Admin(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetAllUsers: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_GetAllUsers_NonAdmin(t *testing.T) {
	userRepo := &mockUserRepoForController{byID: map[string]*auth_models.User{}}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-admin, got %d", w.Code)
	}
}

func TestUserController_GetUserByID_Admin(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users/u1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetUserByID: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_GetUserByID_OwnUser(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users/u1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetUserByID own: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_GetUserByID_Forbidden(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"other": {UserID: "other", Username: "bob"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users/other", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 when accessing other user, got %d", w.Code)
	}
}

func TestUserController_GetUserByID_NotFound(t *testing.T) {
	userRepo := &mockUserRepoForController{byID: map[string]*auth_models.User{}}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUserController_UpdateUser(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice", Email: "a@b.com"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	body, _ := json.Marshal(map[string]string{"username": "alice2", "email": "a2@b.com"})
	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("PUT", "/api/users/u1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("UpdateUser: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_DeleteUser(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("DELETE", "/api/users/u1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("DeleteUser: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_UpdateUserRole(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": {UserID: "u1", Username: "alice", Role: "user"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("admin", "admin")

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("PUT", "/api/users/u1/role", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("UpdateUserRole: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_GetUserByDeviceID(t *testing.T) {
	userRepo := &mockUserRepoForController{
		byDevice: map[string]*auth_models.User{"dev1": {UserID: "u1", Username: "alice"}},
	}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users/by-device/dev1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GetUserByDeviceID: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserController_GetUserByDeviceID_NotFound(t *testing.T) {
	userRepo := &mockUserRepoForController{byDevice: map[string]*auth_models.User{}}
	userSvc := service.NewUserService(userRepo, &mockPiRepoForController{})
	uc := NewUserController(userSvc)
	authMw := newAuthMiddleware(t)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	r := setupUserRouter(uc, authMw)
	req := httptest.NewRequest("GET", "/api/users/by-device/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
