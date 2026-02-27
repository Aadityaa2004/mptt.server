package jwt

import (
	"context"
	"testing"
	"time"

	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// mockUserRepo returns a user for FindByID
type mockUserRepo struct {
	user *auth_models.User
}

func (m *mockUserRepo) FindByID(ctx context.Context, userID string) (*auth_models.User, error) {
	if m.user != nil && m.user.UserID == userID {
		return m.user, nil
	}
	return nil, nil
}

// Stub other interface methods (not used in RefreshTokens test)
func (m *mockUserRepo) Create(ctx context.Context, user *auth_models.User) (*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, userID string) (*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetAll(ctx context.Context) ([]*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) List(ctx context.Context, page, pageSize int, role string) (*interfaces.PaginationResult, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUser(ctx context.Context, userID string) (*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByRole(ctx context.Context, role string) ([]*auth_models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *auth_models.User) error {
	return nil
}
func (m *mockUserRepo) Delete(ctx context.Context, userID string, hardDelete bool) error {
	return nil
}
func (m *mockUserRepo) GetUserByDeviceID(ctx context.Context, deviceID string) (*auth_models.User, error) {
	return nil, nil
}

func TestJWT_RefreshTokens(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}
	svc := NewService(cfg)

	// Generate initial tokens
	pair, err := svc.GenerateTokens("user-789", "user")
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	// Mock user repo
	userRepo := &mockUserRepo{
		user: &auth_models.User{UserID: "user-789", Username: "test", Email: "test@test.com", Role: "user"},
	}

	// Refresh tokens
	newPair, err := svc.RefreshTokens(pair.RefreshToken, userRepo)
	if err != nil {
		t.Fatalf("RefreshTokens: %v", err)
	}
	if newPair.AccessToken == pair.AccessToken {
		t.Error("expected new access token to differ from old")
	}
	if newPair.RefreshToken == pair.RefreshToken {
		t.Error("expected new refresh token to differ from old")
	}

	// Validate new access token
	claims, err := svc.ValidateAccessToken(newPair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if claims.UserID != "user-789" || claims.Role != "user" {
		t.Errorf("claims: user_id=%q role=%q", claims.UserID, claims.Role)
	}
}

func TestJWT_RefreshTokens_InvalidToken(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}
	svc := NewService(cfg)
	userRepo := &mockUserRepo{user: &auth_models.User{UserID: "u1", Role: "user"}}

	_, err := svc.RefreshTokens("invalid.token.here", userRepo)
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}
