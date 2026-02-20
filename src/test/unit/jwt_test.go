package unit

import (
	"testing"
	"time"

	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
)

func TestJWT_GenerateAndValidateAccessToken(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}
	svc := jwt.NewService(cfg)

	// Generate tokens
	pair, err := svc.GenerateTokens("user-123", "user")
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("expected non-empty access and refresh tokens")
	}
	if pair.TokenID == "" {
		t.Error("expected non-empty token_id")
	}

	// Validate access token
	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if claims.UserID != "user-123" || claims.Role != "user" {
		t.Errorf("claims: user_id=%q role=%q", claims.UserID, claims.Role)
	}
	if claims.TokenID != pair.TokenID {
		t.Errorf("token_id mismatch: %q vs %q", claims.TokenID, pair.TokenID)
	}
}

func TestJWT_ValidateAccessToken_Invalid(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}
	svc := jwt.NewService(cfg)

	_, err := svc.ValidateAccessToken("invalid.jwt.token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}
