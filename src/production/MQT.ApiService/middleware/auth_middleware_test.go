package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
)

func setupAuthMiddleware(t *testing.T) (*AuthMiddleware, *gin.Engine) {
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	rbacSvc := rbac.NewService()
	mw := NewAuthMiddleware(jwtSvc, rbacSvc, DefaultConfig())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	return mw, r
}

func TestAuthMiddleware_Authenticate_MissingToken(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	r.GET("/protected", mw.Authenticate(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	r.GET("/protected", mw.Authenticate(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_Authenticate_ValidToken(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-1", "user")

	r.GET("/protected", mw.Authenticate(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_Authenticate_TokenFromCookie(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-1", "user")

	r.GET("/protected", mw.Authenticate(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: pair.AccessToken})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_RequireAdmin_Forbidden(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-1", "user")

	r.GET("/admin", mw.Authenticate(), mw.RequireAdmin(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403", w.Code)
	}
}

func TestAuthMiddleware_RequireAdmin_Success(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("admin-1", "admin")

	r.GET("/admin", mw.Authenticate(), mw.RequireAdmin(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-1", "user")

	r.GET("/user-only", mw.Authenticate(), mw.RequireRole("user"), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/user-only", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_RequireRole_Insufficient(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-1", "user")

	r.GET("/admin-only", mw.Authenticate(), mw.RequireRole("admin"), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403", w.Code)
	}
}

func TestAuthMiddleware_RequireOwnerOrAdmin_Owner(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-1", "user")

	r.GET("/resource/:id", mw.Authenticate(), mw.RequireOwnerOrAdmin("user-1"), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/resource/user-1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_RequireOwnerOrAdmin_Forbidden(t *testing.T) {
	mw, r := setupAuthMiddleware(t)
	cfg := api_models.Config{
		SecretKey:            "test-secret-key-at-least-32-chars",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test",
	}
	jwtSvc := jwt.NewService(cfg)
	pair, _ := jwtSvc.GenerateTokens("user-2", "user")

	r.GET("/resource/:id", mw.Authenticate(), mw.RequireOwnerOrAdmin("user-1"), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/resource/user-1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403", w.Code)
	}
}

func TestGetUserFromContext(t *testing.T) {
	_, err := GetUserFromContext(context.Background())
	if err == nil {
		t.Error("expected error when user not in context")
	}
}

func TestGetRoleFromContext(t *testing.T) {
	_, err := GetRoleFromContext(context.Background())
	if err == nil {
		t.Error("expected error when role not in context")
	}
}

func TestGetUserFromGinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	_, err := GetUserFromGinContext(c)
	if err == nil {
		t.Error("expected error when user not in context")
	}
}

func TestGetRoleFromGinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	_, err := GetRoleFromGinContext(c)
	if err == nil {
		t.Error("expected error when role not in context")
	}
}
