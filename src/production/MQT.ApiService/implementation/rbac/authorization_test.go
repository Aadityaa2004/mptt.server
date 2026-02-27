package rbac

import (
	"testing"
	"time"

	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
)

func TestAuthorizer_AuthorizeWithToken(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	rbacSvc := NewService()
	auth := NewAuthorizer(rbacSvc, jwtSvc)

	pair, _ := jwtSvc.GenerateTokens("u1", "user")

	err := auth.AuthorizeWithToken(pair.AccessToken, "user")
	if err != nil {
		t.Errorf("AuthorizeWithToken: %v", err)
	}

	err = auth.AuthorizeWithToken(pair.AccessToken, "admin")
	if err == nil {
		t.Error("expected error for wrong role")
	}
}

func TestAuthorizer_RequireAdmin(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	auth := NewAuthorizer(NewService(), jwtSvc)

	adminPair, _ := jwtSvc.GenerateTokens("a1", "admin")
	userPair, _ := jwtSvc.GenerateTokens("u1", "user")

	if err := auth.RequireAdmin(adminPair.AccessToken); err != nil {
		t.Errorf("RequireAdmin(admin): %v", err)
	}
	if err := auth.RequireAdmin(userPair.AccessToken); err == nil {
		t.Error("expected error for non-admin")
	}
}

func TestAuthorizer_IsOwner(t *testing.T) {
	cfg := api_models.Config{SecretKey: "x", AccessTokenDuration: time.Hour, RefreshTokenDuration: 24 * time.Hour, Issuer: "i"}
	auth := NewAuthorizer(NewService(), jwt.NewService(cfg))
	if !auth.IsOwner("u1", "u1") {
		t.Error("IsOwner: same ID should be true")
	}
	if auth.IsOwner("u1", "u2") {
		t.Error("IsOwner: different ID should be false")
	}
}

func TestAuthorizer_RequireOwnerOrAdmin(t *testing.T) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	auth := NewAuthorizer(NewService(), jwtSvc)

	adminPair, _ := jwtSvc.GenerateTokens("a1", "admin")
	ownerPair, _ := jwtSvc.GenerateTokens("u1", "user")
	otherPair, _ := jwtSvc.GenerateTokens("u2", "user")

	if err := auth.RequireOwnerOrAdmin(adminPair.AccessToken, "any"); err != nil {
		t.Errorf("admin should access any: %v", err)
	}
	if err := auth.RequireOwnerOrAdmin(ownerPair.AccessToken, "u1"); err != nil {
		t.Errorf("owner should access own: %v", err)
	}
	if err := auth.RequireOwnerOrAdmin(otherPair.AccessToken, "u1"); err == nil {
		t.Error("other user should not access")
	}
}
