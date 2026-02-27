package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

func TestAPIClient_ValidatePi_Exists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exists":true,"error":""}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-secret")
	ctx := context.Background()

	exists, err := client.ValidatePi(ctx, "pi-1")
	if err != nil {
		t.Fatalf("ValidatePi: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
}

func TestAPIClient_ValidatePi_NotExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exists":false,"error":""}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-secret")
	ctx := context.Background()

	exists, err := client.ValidatePi(ctx, "pi-nonexistent")
	if err != nil {
		t.Fatalf("ValidatePi: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}

func TestAPIClient_CreateReading_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success":true,"error":""}`))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-secret")
	ctx := context.Background()

	reading := hardware_models.Reading{
		PiID:     "pi-1",
		DeviceID: "dev-1",
		Ts:       time.Now(),
		Payload:  hardware_models.ReadingPayload{},
	}

	err := client.CreateReading(ctx, reading)
	if err != nil {
		t.Fatalf("CreateReading: %v", err)
	}
}

func TestAPIClient_GetCircuitBreakerStatus(t *testing.T) {
	client := NewAPIClient("http://localhost:9999", "secret")
	status := client.GetCircuitBreakerStatus()

	if status["state"] != "closed" {
		t.Errorf("initial state = %v, want closed", status["state"])
	}
	if fc, ok := status["failure_count"].(int); !ok || fc != 0 {
		t.Errorf("initial failure_count = %v", status["failure_count"])
	}
}
