package main

import (
	"context"
	"os"
	"testing"
)

func TestInitializeIngestorApp_BasicWiring(t *testing.T) {
	// Minimal required env for IngestorConfig
	if err := os.Setenv("INTERNAL_API_SECRET", "test-secret"); err != nil {
		t.Fatalf("failed to set INTERNAL_API_SECRET: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("INTERNAL_API_SECRET") })

	ing, cfg, apiClient, log, shutdown, err := InitializeIngestorApp()
	if err != nil {
		t.Fatalf("InitializeIngestorApp returned error: %v", err)
	}
	if ing == nil {
		t.Fatal("ingestor is nil")
	}
	if cfg == nil {
		t.Fatal("config is nil")
	}
	if apiClient == nil {
		t.Fatal("apiClient is nil")
	}
	if log == nil {
		t.Fatal("logger is nil")
	}
	if shutdown == nil {
		t.Fatal("shutdown func is nil")
	}

	// Ensure shutdown is callable
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}
}

