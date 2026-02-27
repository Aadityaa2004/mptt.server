//go:build integration

package service

import (
	"os"
	"testing"
)

// TestEmailService_SendHTML_STARTTLS_Integration exercises the full STARTTLS flow
// against a real SMTP server. Requires a running SMTP server with STARTTLS support
// (e.g. Mailpit via docker-compose.test.yml).
//
// Run with: go test -tags=integration ./...
// Or skip in CI if no SMTP server is available.
func TestEmailService_SendHTML_STARTTLS_Integration(t *testing.T) {
	host := os.Getenv("SMTP_TEST_HOST")
	port := 1025 // Mailpit default
	if host == "" {
		t.Skip("SMTP_TEST_HOST not set - skip STARTTLS integration test")
	}

	svc := NewEmailService(host, port, "starttls", "", "", "Test", "test@test.com")
	err := svc.SendHTML("recipient@test.com", "Integration Test", "<p>Hello</p>")
	if err != nil {
		t.Fatalf("SendHTML with STARTTLS: %v", err)
	}
}
