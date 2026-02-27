package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	byUsername map[string]*auth_models.User
	byEmail    map[string]*auth_models.User
	byID       map[string]*auth_models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byUsername: make(map[string]*auth_models.User),
		byEmail:    make(map[string]*auth_models.User),
		byID:       make(map[string]*auth_models.User),
	}
}

func (m *mockUserRepo) add(u *auth_models.User) {
	m.byUsername[u.Username] = u
	m.byEmail[u.Email] = u
	m.byID[u.UserID] = u
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*auth_models.User, error) {
	return m.byUsername[username], nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*auth_models.User, error) {
	return m.byEmail[email], nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, userID string) (*auth_models.User, error) {
	return m.byID[userID], nil
}
func (m *mockUserRepo) FindByID(ctx context.Context, userID string) (*auth_models.User, error) {
	return m.byID[userID], nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *auth_models.User) (*auth_models.User, error) {
	m.add(user)
	return user, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *auth_models.User) error {
	m.add(user)
	return nil
}
func (m *mockUserRepo) GetAll(ctx context.Context) ([]*auth_models.User, error)           { return nil, nil }
func (m *mockUserRepo) List(ctx context.Context, page, pageSize int, role string) (*interfaces.PaginationResult, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUser(ctx context.Context, userID string) (*auth_models.User, error) {
	return m.byID[userID], nil
}
func (m *mockUserRepo) GetByRole(ctx context.Context, role string) ([]*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Delete(ctx context.Context, userID string, hardDelete bool) error { return nil }
func (m *mockUserRepo) GetUserByDeviceID(ctx context.Context, deviceID string) (*auth_models.User, error) {
	return nil, nil
}

type mockRoleRepo struct{}

func (m *mockRoleRepo) Create(ctx context.Context, role *auth_models.Role) (*auth_models.Role, error) {
	return nil, nil
}
func (m *mockRoleRepo) FindByID(ctx context.Context, id string) (*auth_models.Role, error)   { return nil, nil }
func (m *mockRoleRepo) FindByName(ctx context.Context, name string) (*auth_models.Role, error) { return nil, nil }
func (m *mockRoleRepo) FindAll(ctx context.Context) ([]*auth_models.Role, error)            { return nil, nil }
func (m *mockRoleRepo) Update(ctx context.Context, role *auth_models.Role) error            { return nil }
func (m *mockRoleRepo) Delete(ctx context.Context, id string) error                         { return nil }

func TestAuthService_Login(t *testing.T) {
	jwtCfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}
	jwtSvc := jwt.NewService(jwtCfg)
	rbacSvc := rbac.NewService()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	now := time.Now()
	user := &auth_models.User{
		UserID:         "u1",
		Username:       "testuser",
		Email:          "test@test.com",
		Password:       string(hashed),
		Role:           "user",
		Active:         true,
		EmailVerifiedAt: &now,
	}

	userRepo := newMockUserRepo()
	userRepo.add(user)

	verSvc := NewVerificationService(nil, "http://localhost:9004", "http://localhost:3000")
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbacSvc, verSvc)

	ctx := context.Background()

	// Success
	resp, tokenPair, err := authSvc.Login(ctx, LoginRequest{Username: "testuser", Password: "pass123"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.UserID != "u1" || resp.Username != "testuser" || resp.Role != "user" {
		t.Errorf("resp: %+v", resp)
	}
	if tokenPair.AccessToken == "" {
		t.Error("expected access token")
	}

	// Wrong password
	_, _, err = authSvc.Login(ctx, LoginRequest{Username: "testuser", Password: "wrong"})
	if err == nil {
		t.Error("expected error for wrong password")
	}

	// Inactive user
	user.Active = false
	userRepo.add(user)
	_, _, err = authSvc.Login(ctx, LoginRequest{Username: "testuser", Password: "pass123"})
	if err == nil {
		t.Error("expected error for inactive user")
	}

	// Unknown user
	user.Active = true
	userRepo.add(user)
	_, _, err = authSvc.Login(ctx, LoginRequest{Username: "nonexistent", Password: "pass123"})
	if err == nil {
		t.Error("expected error for unknown user")
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u2", Username: "u2", Email: "u2@x.com", Role: "user"}
	userRepo.add(user)

	jwtSvc := jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"})
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbac.NewService(), nil)

	ctx := context.Background()
	got, err := authSvc.GetUserByID(ctx, "u2")
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.UserID != "u2" {
		t.Errorf("got UserID %q", got.UserID)
	}
}

func TestAuthService_Register(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	jwtSvc := jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"})
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbac.NewService(), verSvc)
	ctx := context.Background()

	resp, err := authSvc.Register(ctx, RegisterRequest{
		Username: "newuser",
		Email:    "new@x.com",
		Password: "pass123",
		Role:     "user",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if !resp.RequiresVerification {
		t.Error("expected RequiresVerification=true")
	}

	// Duplicate username
	userRepo.add(&auth_models.User{UserID: "u3", Username: "taken", Email: "other@x.com"})
	_, err = authSvc.Register(ctx, RegisterRequest{Username: "taken", Email: "new2@x.com", Password: "pass"})
	if err == nil {
		t.Error("expected error for duplicate username")
	}

	// Duplicate email
	_, err = authSvc.Register(ctx, RegisterRequest{Username: "newuser2", Email: "new@x.com", Password: "pass"})
	if err == nil {
		t.Error("expected error for duplicate email")
	}
}

func TestAuthService_VerifyEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	jwtSvc := jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"})
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbac.NewService(), verSvc)
	ctx := context.Background()

	// Register first
	_, err := authSvc.Register(ctx, RegisterRequest{Username: "verifyuser", Email: "verify@x.com", Password: "pass123"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Test invalid OTP path (we don't have the real OTP from email)
	_, _, err = authSvc.VerifyEmail(ctx, VerifyEmailRequest{Email: "verify@x.com", OTP: "000000"})
	if err == nil {
		t.Error("expected error for invalid OTP")
	}
}

func TestAuthService_RequestPasswordReset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	userRepo := newMockUserRepo()
	userRepo.add(&auth_models.User{UserID: "u1", Username: "u1", Email: "reset@x.com"})
	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"}), rbac.NewService(), verSvc)
	ctx := context.Background()

	err := authSvc.RequestPasswordReset(ctx, "reset@x.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}

	// Nonexistent user - idempotent, returns nil
	err = authSvc.RequestPasswordReset(ctx, "nonexistent@x.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset for nonexistent: %v", err)
	}
}

func TestAuthService_UpdateUser(t *testing.T) {
	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u1", Username: "u1", Role: "user"}
	userRepo.add(user)
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"}), rbac.NewService(), nil)
	ctx := context.Background()

	user.Username = "updated"
	got, err := authSvc.UpdateUser(ctx, user)
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if got.Username != "updated" {
		t.Errorf("Username = %q", got.Username)
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	authSvc := NewAuthService(nil, nil, nil, nil, nil)
	hash, err := authSvc.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "secret" {
		t.Error("expected hashed password")
	}
}

func TestAuthService_RefreshTokens(t *testing.T) {
	jwtCfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}
	jwtSvc := jwt.NewService(jwtCfg)

	userRepo := newMockUserRepo()
	user := &auth_models.User{UserID: "u1", Username: "u1", Email: "u1@x.com", Role: "admin"}
	userRepo.add(user)

	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbac.NewService(), nil)
	ctx := context.Background()

	// Generate a valid refresh token
	pair, err := jwtSvc.GenerateTokens("u1", "admin")
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	resp, newPair, err := authSvc.RefreshTokens(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if resp.AccessToken == "" || newPair.AccessToken == "" {
		t.Error("expected non-empty access tokens")
	}

	// Invalid refresh token should return error
	_, _, err = authSvc.RefreshTokens(ctx, "invalid-token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

func TestAuthService_VerifyEmail_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	jwtSvc := jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"})
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbac.NewService(), verSvc)
	ctx := context.Background()

	// Manually insert a valid signup token so we can control the OTP value
	otp := "123456"
	meta := &auth_models.SignupMetadata{Username: "verifyuser2", PasswordHash: "hash", Role: "user"}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Marshal meta: %v", err)
	}
	hash := sha256Hash(otp)
	token := auth_models.NewVerificationToken("verify2@x.com", hash, auth_models.PurposeSignupOTP, time.Now().Add(time.Hour), metaJSON)
	if err := tokenRepo.Create(ctx, token); err != nil {
		t.Fatalf("Create token: %v", err)
	}

	resp, pair, err := authSvc.VerifyEmail(ctx, VerifyEmailRequest{Email: "verify2@x.com", OTP: otp})
	if err != nil {
		t.Fatalf("VerifyEmail success: %v", err)
	}
	if resp.AccessToken == "" || pair.AccessToken == "" {
		t.Errorf("expected populated tokens, got resp=%+v pair=%+v", resp, pair)
	}
	if resp.Email != "verify2@x.com" || resp.Username != "verifyuser2" {
		t.Errorf("unexpected user in response: %+v", resp)
	}

	// User should now exist in repo (even if UserID is empty, since DB normally sets it)
	created, err := userRepo.GetByEmail(ctx, "verify2@x.com")
	if err != nil || created == nil {
		t.Fatalf("expected created user, got err=%v user=%v", err, created)
	}
}

func TestAuthService_ResendOTP_Error(t *testing.T) {
	// Using a fresh VerificationService with empty token repo so ResendSignupOTP returns an error
	verSvc := NewVerificationService(newMockTokenRepo(), "http://localhost:9999", "http://localhost:3000")
	authSvc := NewAuthService(nil, nil, nil, nil, verSvc)
	ctx := context.Background()

	err := authSvc.ResendOTP(ctx, "nonexistent@x.com")
	if err == nil {
		t.Error("expected error when no pending registration")
	}
}

func TestAuthService_RequestPasswordReset_ErrorFromVerificationService(t *testing.T) {
	// Configure email service to return 500 so GenerateAndStorePasswordResetToken fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	userRepo := newMockUserRepo()
	userRepo.add(&auth_models.User{UserID: "u1", Username: "u1", Email: "reset2@x.com"})
	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"}), rbac.NewService(), verSvc)
	ctx := context.Background()

	err := authSvc.RequestPasswordReset(ctx, "reset2@x.com")
	if err == nil {
		t.Error("expected error when email service fails")
	}
}

func TestAuthService_Register_EmailError(t *testing.T) {
	// Email service returns 500 so GenerateAndStoreSignupOTP fails and Register wraps the error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	jwtSvc := jwt.NewService(api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"})
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, jwtSvc, rbac.NewService(), verSvc)
	ctx := context.Background()

	_, err := authSvc.Register(ctx, RegisterRequest{
		Username: "uemailerror",
		Email:    "uemailerror@x.com",
		Password: "pass123",
	})
	if err == nil {
		t.Error("expected error when verification email fails")
	}
}

func TestAuthService_ResetPassword_Flows(t *testing.T) {
	// Shared setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tokenRepo := newMockTokenRepo()
	verSvc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	userRepo := newMockUserRepo()
	authSvc := NewAuthService(userRepo, &mockRoleRepo{}, nil, nil, verSvc)
	ctx := context.Background()

	// Success: valid token + existing user
	tokenStr := "reset-token"
	hash := sha256Hash(tokenStr)
	exp := time.Now().Add(time.Hour)
	resetToken := auth_models.NewVerificationToken("reset-success@x.com", hash, auth_models.PurposePasswordReset, exp, nil)
	if err := tokenRepo.Create(ctx, resetToken); err != nil {
		t.Fatalf("Create reset token: %v", err)
	}
	userRepo.add(&auth_models.User{UserID: "u1", Username: "u1", Email: "reset-success@x.com", Password: "old"})

	err := authSvc.ResetPassword(ctx, ResetPasswordRequest{Token: tokenStr, NewPassword: "newpass"})
	if err != nil {
		t.Fatalf("ResetPassword success: %v", err)
	}
	updated, _ := userRepo.GetByEmail(ctx, "reset-success@x.com")
	if updated == nil || updated.Password == "old" {
		t.Error("expected password to be updated")
	}

	// Invalid token: token not found/expired
	err = authSvc.ResetPassword(ctx, ResetPasswordRequest{Token: "invalid-token", NewPassword: "pass"})
	if err == nil {
		t.Error("expected error for invalid or expired reset link")
	}

	// User not found for valid token
	tokenStr2 := "reset-token-no-user"
	hash2 := sha256Hash(tokenStr2)
	resetToken2 := auth_models.NewVerificationToken("missing-user@x.com", hash2, auth_models.PurposePasswordReset, time.Now().Add(time.Hour), nil)
	if err := tokenRepo.Create(ctx, resetToken2); err != nil {
		t.Fatalf("Create reset token2: %v", err)
	}

	err = authSvc.ResetPassword(ctx, ResetPasswordRequest{Token: tokenStr2, NewPassword: "pass"})
	if err == nil {
		t.Error("expected error when user not found for reset token")
	}
}

// Ensure mockUserRepo satisfies interface
var _ interfaces.UserRepository = (*mockUserRepo)(nil)
