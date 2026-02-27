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

func TestAPI_Internal_ValidatePi_NotExists(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9002"
	}
	secret := os.Getenv("TEST_INTERNAL_API_SECRET")
	if secret == "" {
		secret = "test-internal-secret"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]string{"pi_id": "nonexistent-pi"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", baseURL+"/internal/pis/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("API not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST /internal/pis/validate: status %d, want 200", resp.StatusCode)
	}

	var result struct {
		Exists bool   `json:"exists"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Exists {
		t.Error("expected exists=false for nonexistent pi")
	}
}

func TestAPI_Internal_Unauthorized(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9002"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := map[string]string{"pi_id": "pi-1"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", baseURL+"/internal/pis/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong-secret")
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("API not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("POST /internal/pis/validate with wrong secret: status %d, want 401", resp.StatusCode)
	}
}

