package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	service "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/auth"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/middleware"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
)

// mockVerificationTokenRepo for auth controller tests
type mockVerificationTokenRepo struct {
	tokens map[string]*auth_models.VerificationToken
	byHash map[string]*auth_models.VerificationToken
}

func newMockVerificationTokenRepo() *mockVerificationTokenRepo {
	return &mockVerificationTokenRepo{
		tokens: make(map[string]*auth_models.VerificationToken),
		byHash: make(map[string]*auth_models.VerificationToken),
	}
}

func (m *mockVerificationTokenRepo) Create(ctx context.Context, t *auth_models.VerificationToken) error {
	m.tokens[t.ID] = t
	m.byHash[t.TokenHash] = t
	return nil
}
func (m *mockVerificationTokenRepo) GetValidByEmailAndPurpose(ctx context.Context, email, purpose string) (*auth_models.VerificationToken, error) {
	for _, t := range m.tokens {
		if t.Email == email && t.Purpose == purpose && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, nil
}
func (m *mockVerificationTokenRepo) GetValidByTokenHash(ctx context.Context, hash, purpose string) (*auth_models.VerificationToken, error) {
	t := m.byHash[hash]
	if t != nil && t.Purpose == purpose && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
		return t, nil
	}
	return nil, nil
}
func (m *mockVerificationTokenRepo) GetLatestSignupTokenByEmail(ctx context.Context, email string) (*auth_models.VerificationToken, error) {
	var latest *auth_models.VerificationToken
	for _, t := range m.tokens {
		if t.Email == email && t.Purpose == auth_models.PurposeSignupOTP {
			if latest == nil || t.CreatedAt.After(latest.CreatedAt) {
				latest = t
			}
		}
	}
	return latest, nil
}
func (m *mockVerificationTokenRepo) MarkUsed(ctx context.Context, id string) error {
	if t, ok := m.tokens[id]; ok {
		now := time.Now()
		t.UsedAt = &now
	}
	return nil
}
func (m *mockVerificationTokenRepo) InvalidateByEmailAndPurpose(ctx context.Context, email, purpose string) error {
	for id, t := range m.tokens {
		if t.Email == email && t.Purpose == purpose {
			delete(m.tokens, id)
			delete(m.byHash, t.TokenHash)
		}
	}
	return nil
}
func (m *mockVerificationTokenRepo) Delete(ctx context.Context, id string) error {
	if t, ok := m.tokens[id]; ok {
		delete(m.byHash, t.TokenHash)
		delete(m.tokens, id)
	}
	return nil
}
func (m *mockVerificationTokenRepo) DeleteExpired(ctx context.Context, before time.Time) error { return nil }

type mockRoleRepoForAuth struct{}

func (m *mockRoleRepoForAuth) Create(ctx context.Context, r *auth_models.Role) (*auth_models.Role, error) { return nil, nil }
func (m *mockRoleRepoForAuth) FindByID(ctx context.Context, id string) (*auth_models.Role, error)       { return nil, nil }
func (m *mockRoleRepoForAuth) FindByName(ctx context.Context, name string) (*auth_models.Role, error)  { return nil, nil }
func (m *mockRoleRepoForAuth) FindAll(ctx context.Context) ([]*auth_models.Role, error)                { return nil, nil }
func (m *mockRoleRepoForAuth) Update(ctx context.Context, r *auth_models.Role) error                   { return nil }
func (m *mockRoleRepoForAuth) Delete(ctx context.Context, id string) error                             { return nil }

func newTestAuthService(t *testing.T, userRepo *mockUserRepoForController, tokenRepo *mockVerificationTokenRepo) *service.AuthService {
	emailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(emailServer.Close)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	verSvc := service.NewVerificationService(tokenRepo, emailServer.URL, "http://localhost:3000")
	return service.NewAuthService(userRepo, &mockRoleRepoForAuth{}, jwtSvc, rbac.NewService(), verSvc)
}

func setupAuthRouter(ac *AuthController, authMw *middleware.AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ac.RegisterRoutes(r, authMw)
	return r
}

func TestAuthController_Logout(t *testing.T) {
	userRepo := &mockUserRepoForController{}
	authSvc := newTestAuthService(t, userRepo, newMockVerificationTokenRepo())
	ac := NewAuthController(authSvc, "", "")
	r := setupAuthRouter(ac, newAuthMiddleware(t))

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Logout: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthController_Register(t *testing.T) {
	userRepo := &mockUserRepoForController{byID: map[string]*auth_models.User{}, byUsername: map[string]*auth_models.User{}, byEmail: map[string]*auth_models.User{}}
	authSvc := newTestAuthService(t, userRepo, newMockVerificationTokenRepo())
	ac := NewAuthController(authSvc, "", "") // empty turnstile = skip
	r := setupAuthRouter(ac, newAuthMiddleware(t))

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"email":    "new@test.com",
		"password": "pass123",
	})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Register: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthController_Login(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	now := time.Now()
	user := &auth_models.User{
		UserID:          "u1",
		Username:        "loginuser",
		Email:           "login@test.com",
		Password:        string(hashed),
		Role:            "user",
		Active:          true,
		EmailVerifiedAt: &now,
	}
	userRepo := &mockUserRepoForController{
		byID:       map[string]*auth_models.User{"u1": user},
		byUsername: map[string]*auth_models.User{"loginuser": user},
		byEmail:    map[string]*auth_models.User{"login@test.com": user},
	}
	authSvc := newTestAuthService(t, userRepo, newMockVerificationTokenRepo())
	ac := NewAuthController(authSvc, "", "")
	r := setupAuthRouter(ac, newAuthMiddleware(t))

	body, _ := json.Marshal(map[string]string{
		"username": "loginuser",
		"password": "pass123",
	})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Login: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	userRepo := &mockUserRepoForController{byUsername: map[string]*auth_models.User{}}
	authSvc := newTestAuthService(t, userRepo, newMockVerificationTokenRepo())
	ac := NewAuthController(authSvc, "", "")
	r := setupAuthRouter(ac, newAuthMiddleware(t))

	body, _ := json.Marshal(map[string]string{
		"username": "nonexistent",
		"password": "wrong",
	})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthController_Profile(t *testing.T) {
	user := &auth_models.User{UserID: "u1", Username: "alice", Email: "a@b.com", Role: "user"}
	userRepo := &mockUserRepoForController{
		byID: map[string]*auth_models.User{"u1": user},
	}
	authSvc := newTestAuthService(t, userRepo, newMockVerificationTokenRepo())
	ac := NewAuthController(authSvc, "", "")
	authMw := newAuthMiddleware(t)
	r := setupAuthRouter(ac, authMw)

	jwtSvc := jwt.NewService(api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	})
	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	req := httptest.NewRequest("GET", "/api/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Profile: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthController_ForgotPassword(t *testing.T) {
	userRepo := &mockUserRepoForController{byEmail: map[string]*auth_models.User{}}
	authSvc := newTestAuthService(t, userRepo, newMockVerificationTokenRepo())
	ac := NewAuthController(authSvc, "", "")
	r := setupAuthRouter(ac, newAuthMiddleware(t))

	body, _ := json.Marshal(map[string]string{"email": "any@test.com"})
	req := httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ForgotPassword: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthController_ResendOTP(t *testing.T) {
	userRepo := &mockUserRepoForController{}
	tokenRepo := newMockVerificationTokenRepo()
	// Seed a pending signup token directly (avoid Register which would set rate limit)
	meta := &auth_models.SignupMetadata{Username: "resenduser", PasswordHash: "hash", Role: "user"}
	metaJSON, _ := json.Marshal(meta)
	token := auth_models.NewVerificationToken("resend@test.com", "hash", auth_models.PurposeSignupOTP, time.Now().Add(time.Hour), metaJSON)
	_ = tokenRepo.Create(context.Background(), token)

	authSvc := newTestAuthService(t, userRepo, tokenRepo)
	ac := NewAuthController(authSvc, "", "")
	r := setupAuthRouter(ac, newAuthMiddleware(t))

	body, _ := json.Marshal(map[string]string{"email": "resend@test.com"})
	req := httptest.NewRequest("POST", "/api/auth/resend-otp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ResendOTP: status=%d body=%s", w.Code, w.Body.String())
	}
}
