package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServiceAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	r := gin.New()
	r.POST("/internal/test", ServiceAuthMiddleware(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("POST", "/internal/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

func TestServiceAuthMiddleware_InvalidFormat(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	r := gin.New()
	r.POST("/internal/test", ServiceAuthMiddleware(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("POST", "/internal/test", nil)
	req.Header.Set("Authorization", "Basic xyz")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

func TestServiceAuthMiddleware_EmptyToken(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	r := gin.New()
	r.POST("/internal/test", ServiceAuthMiddleware(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("POST", "/internal/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

func TestServiceAuthMiddleware_WrongToken(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	r := gin.New()
	r.POST("/internal/test", ServiceAuthMiddleware(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("POST", "/internal/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", w.Code)
	}
}

func TestServiceAuthMiddleware_NoSecretConfigured(t *testing.T) {
	os.Unsetenv("INTERNAL_API_SECRET")
	r := gin.New()
	r.POST("/internal/test", ServiceAuthMiddleware(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("POST", "/internal/test", nil)
	req.Header.Set("Authorization", "Bearer any-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", w.Code)
	}
}

func TestServiceAuthMiddleware_Success(t *testing.T) {
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	defer os.Unsetenv("INTERNAL_API_SECRET")

	r := gin.New()
	r.POST("/internal/test", ServiceAuthMiddleware(), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("POST", "/internal/test", nil)
	req.Header.Set("Authorization", "Bearer test-secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}
