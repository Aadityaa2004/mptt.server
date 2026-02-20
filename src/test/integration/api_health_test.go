//go:build integration

package integration

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestAPI_HealthLive(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9002"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health/live")
	if err != nil {
		t.Skipf("API not reachable at %s (is the stack up?): %v", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health/live: status %d, want 200", resp.StatusCode)
	}
}

func TestAPI_HealthReady(t *testing.T) {
	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9002"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health/ready")
	if err != nil {
		t.Skipf("API not reachable at %s (is the stack up?): %v", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health/ready: status %d, want 200", resp.StatusCode)
	}
}
