package service

import (
	"strings"
	"testing"

	"github.com/k1LoW/smtptest"
)

func TestNewEmailService(t *testing.T) {
	svc := NewEmailService("localhost", 587, "starttls", "user", "pass", "From", "from@test.com")
	if svc == nil {
		t.Fatal("NewEmailService returned nil")
	}
	if svc.smtpHost != "localhost" || svc.smtpPort != 587 {
		t.Errorf("unexpected host/port")
	}
}

func TestEmailService_SendHTML_PlainSMTP(t *testing.T) {
	ts, _, err := smtptest.NewServerWithAuth(smtptest.WithPlainAuth("testuser", "testpass"))
	if err != nil {
		t.Fatalf("smtptest.NewServerWithAuth: %v", err)
	}
	t.Cleanup(func() { ts.Close() })

	svc := NewEmailService(ts.Host, ts.Port, "plain", "testuser", "testpass", "Test", "sender@test.com")

	err = svc.SendHTML("recipient@test.com", "Test Subject", "<p>Hello</p>")
	if err != nil {
		t.Fatalf("SendHTML: %v", err)
	}

	msgs := ts.Messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if !strings.Contains(msgs[0].Header.Get("Subject"), "Test Subject") {
		t.Errorf("unexpected subject: %s", msgs[0].Header.Get("Subject"))
	}
}

func TestEmailService_SendHTML_STARTTLS_ConnectionError(t *testing.T) {
	// Connects to invalid host to exercise sendWithSTARTTLS error path (dial failure).
	// Full STARTTLS success path is covered by integration test (see email_service_integration_test.go).
	svc := NewEmailService("invalid.invalid", 587, "starttls", "u", "p", "T", "from@x.com")
	err := svc.SendHTML("to@x.com", "Subj", "<p>Hi</p>")
	if err == nil {
		t.Fatal("expected error for invalid host")
	}
}
