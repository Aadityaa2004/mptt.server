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

func TestAPI_Login_Admin(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9002"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]string{
		"username": "admin",
		"password": "adminpassword123",
	}
	body, _ := json.Marshal(reqBody)
	resp, err := client.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("API not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST /api/auth/login: status %d, want 200", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		UserID      string `json:"user_id"`
		Username    string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected access_token in response")
	}
	if result.Username != "admin" {
		t.Errorf("username = %q, want admin", result.Username)
	}
}

func TestAPI_Login_InvalidCredentials(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9002"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]string{
		"username": "admin",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)
	resp, err := client.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("API not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("POST /api/auth/login with wrong password: status %d, want 401", resp.StatusCode)
	}
}

