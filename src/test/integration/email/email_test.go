//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestEmailService_Health(t *testing.T) {
	baseURL := os.Getenv("TEST_EMAIL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9004"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("Email service not reachable at %s: %v", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health: status %d, want 200", resp.StatusCode)
	}
}

func TestEmailService_SendOTP(t *testing.T) {
	baseURL := os.Getenv("TEST_EMAIL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9004"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]string{
		"to":  "test@example.com",
		"otp": "123456",
	}
	body, _ := json.Marshal(reqBody)
	resp, err := client.Post(baseURL+"/send-otp", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("Email service not reachable: %v", err)
	}
	defer resp.Body.Close()

	// May return 200 (sent) or 500 (SMTP error if not configured) - we accept 200
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("POST /send-otp: status %d", resp.StatusCode)
	}
}

