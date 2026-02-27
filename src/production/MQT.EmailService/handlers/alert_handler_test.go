package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/cooldown"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.EmailService/service"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/test/testutil"
)

func TestAlertHandler_BucketFill_Sent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{}
	tracker := cooldown.NewTracker(60, 70)
	log := testutil.NewDiscardLogger()
	handler := NewAlertHandler(mock, tracker, log)

	r := gin.New()
	r.POST("/alerts/bucket-fill", handler.HandleBucketFillAlert)

	body := `{"user_email":"u@x.com","user_name":"User","device_id":"d1","fill_percentage":85}`
	req := httptest.NewRequest("POST", "/alerts/bucket-fill", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "sent") {
		t.Errorf("body %q should contain 'sent'", w.Body.String())
	}
}

func TestAlertHandler_BucketFill_BelowThreshold(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{}
	tracker := cooldown.NewTracker(60, 70)
	log := testutil.NewDiscardLogger()
	handler := NewAlertHandler(mock, tracker, log)

	r := gin.New()
	r.POST("/alerts/bucket-fill", handler.HandleBucketFillAlert)

	body := `{"user_email":"u@x.com","user_name":"User","device_id":"d1","fill_percentage":50}`
	req := httptest.NewRequest("POST", "/alerts/bucket-fill", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "below_threshold") {
		t.Errorf("body %q should contain 'below_threshold'", w.Body.String())
	}
}

func TestAlertHandler_BucketFill_Cooldown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{}
	tracker := cooldown.NewTracker(60, 70)
	log := testutil.NewDiscardLogger()
	handler := NewAlertHandler(mock, tracker, log)

	key := "u@x.com:d1"
	tracker.Record(key)

	r := gin.New()
	r.POST("/alerts/bucket-fill", handler.HandleBucketFillAlert)

	body := `{"user_email":"u@x.com","user_name":"User","device_id":"d1","fill_percentage":85}`
	req := httptest.NewRequest("POST", "/alerts/bucket-fill", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "cooldown") {
		t.Errorf("body %q should contain 'cooldown'", w.Body.String())
	}
}

func TestAlertHandler_BucketFill_SMTPError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockEmailSender{sendErr: errors.New("SMTP failed")}
	tracker := cooldown.NewTracker(60, 70)
	log := testutil.NewDiscardLogger()
	handler := NewAlertHandler(mock, tracker, log)

	r := gin.New()
	r.POST("/alerts/bucket-fill", handler.HandleBucketFillAlert)

	body := `{"user_email":"u@x.com","user_name":"User","device_id":"d1","fill_percentage":85}`
	req := httptest.NewRequest("POST", "/alerts/bucket-fill", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", w.Code)
	}
}

var _ service.EmailSender = (*mockEmailSender)(nil)
