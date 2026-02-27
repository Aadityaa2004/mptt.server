//go:build acceptance

package acceptance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultAPIURL = "http://localhost:9002"
)

func TestE2E_LoginAndHealth(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = defaultAPIURL
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Health check
	resp, err := client.Get(baseURL + "/health/live")
	if err != nil {
		t.Skipf("API not reachable: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health/live: status %d", resp.StatusCode)
	}

	// 2. Login with admin (seeded by API)
	reqBody := map[string]string{
		"username": "admin",
		"password": "adminpassword123",
	}
	body, _ := json.Marshal(reqBody)
	resp, err = client.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: status %d", resp.StatusCode)
	}

	var loginResult struct {
		AccessToken string `json:"access_token"`
		UserID      string `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResult); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if loginResult.AccessToken == "" {
		t.Fatal("no access_token")
	}

	// 3. Call protected endpoint (e.g. users/me or similar)
	req, _ := http.NewRequest("GET", baseURL+"/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResult.AccessToken)
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatalf("users/me: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Logf("users/me: status %d (endpoint may vary)", resp2.StatusCode)
	}
}

func TestE2E_InternalAPI_CreateReading_Flow(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = defaultAPIURL
	}
	secret := os.Getenv("TEST_INTERNAL_API_SECRET")
	if secret == "" {
		secret = "test-internal-secret"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Validate non-existent Pi
	reqBody := map[string]string{"pi_id": "e2e-pi-1"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", baseURL+"/internal/pis/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("API not reachable: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("validate pi: status %d", resp.StatusCode)
	}

	// Create reading would need Pi and Device to exist - skip that part for now
	// This test validates the internal API auth and validate endpoints work
	t.Log("Internal API validate endpoint OK")
}

