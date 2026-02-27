package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/test/testutil"
)

type mockEmailSender struct {
	sendErr error
}

func (m *mockEmailSender) SendHTML(to, subject, htmlBody string) error {
	return m.sendErr
}

func TestEmailHandler_SendOTP_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{}
	log := testutil.NewDiscardLogger()
	handler := NewEmailHandler(mock, log)

	r := gin.New()
	r.POST("/send-otp", handler.SendOTP)

	body := `{"to":"test@example.com","otp":"123456"}`
	req := httptest.NewRequest("POST", "/send-otp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

func TestEmailHandler_SendOTP_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{}
	log := testutil.NewDiscardLogger()
	handler := NewEmailHandler(mock, log)

	r := gin.New()
	r.POST("/send-otp", handler.SendOTP)

	req := httptest.NewRequest("POST", "/send-otp", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", w.Code)
	}
}

func TestEmailHandler_SendOTP_SMTPError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{sendErr: errors.New("SMTP failed")}
	log := testutil.NewDiscardLogger()
	handler := NewEmailHandler(mock, log)

	r := gin.New()
	r.POST("/send-otp", handler.SendOTP)

	body := `{"to":"test@example.com","otp":"123456"}`
	req := httptest.NewRequest("POST", "/send-otp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", w.Code)
	}
}

func TestEmailHandler_SendPasswordReset_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{}
	log := testutil.NewDiscardLogger()
	handler := NewEmailHandler(mock, log)

	r := gin.New()
	r.POST("/send-password-reset", handler.SendPasswordReset)

	body := `{"to":"test@example.com","reset_link":"https://example.com/reset?token=abc"}`
	req := httptest.NewRequest("POST", "/send-password-reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
}

// Ensure EmailService implements EmailSender
var _ service.EmailSender = (*service.EmailService)(nil)
