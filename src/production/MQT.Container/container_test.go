package container

import (
	"context"
	"os"
	"testing"
)

func setContainerEnv(t *testing.T) {
	t.Helper()
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("INTERNAL_API_SECRET", "test-secret")
	os.Setenv("API_SERVICE_URL", "http://api:9002")
}

func unsetContainerEnv(t *testing.T) {
	t.Helper()
	os.Unsetenv("POSTGRES_USER")
	os.Unsetenv("POSTGRES_PASSWORD")
	os.Unsetenv("INTERNAL_API_SECRET")
	os.Unsetenv("API_SERVICE_URL")
}

func TestNewContainer(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	c, err := NewContainer()
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil container")
	}
}

func TestContainer_GetConfig(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	c, err := NewContainer()
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}

	cfg := c.GetConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("config.Database.User = %q", cfg.Database.User)
	}
}

func TestContainer_GetLogger(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	c, err := NewContainer()
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}

	log := c.GetLogger()
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestContainer_RegisterService_GetService(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	c, err := NewContainer()
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}

	svc := "test-service"
	c.RegisterService("myservice", svc)

	got, exists := c.GetService("myservice")
	if !exists {
		t.Fatal("expected service to exist")
	}
	if got != svc {
		t.Errorf("GetService = %v", got)
	}

	_, exists = c.GetService("nonexistent")
	if exists {
		t.Error("expected nonexistent service to not exist")
	}
}

func TestContainer_AddCleanupFunc(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	c, err := NewContainer()
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}

	called := false
	c.AddCleanupFunc(func() error {
		called = true
		return nil
	})

	ctx := context.Background()
	if err := c.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
	if !called {
		t.Error("cleanup func was not called")
	}
}

func TestContainer_Shutdown(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	c, err := NewContainer()
	if err != nil {
		t.Fatalf("NewContainer: %v", err)
	}

	ctx := context.Background()
	if err := c.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

func TestNewIngestorContainer(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	ic, err := NewIngestorContainer()
	if err != nil {
		t.Fatalf("NewIngestorContainer: %v", err)
	}
	if ic == nil {
		t.Fatal("expected non-nil container")
	}
	if ic.GetConfig() == nil {
		t.Fatal("expected non-nil config")
	}
	if ic.GetLogger() == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestIngestorContainer_Shutdown(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	ic, err := NewIngestorContainer()
	if err != nil {
		t.Fatalf("NewIngestorContainer: %v", err)
	}

	ctx := context.Background()
	if err := ic.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

func TestNewApiContainer(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	ac, err := NewApiContainer()
	if err != nil {
		t.Fatalf("NewApiContainer: %v", err)
	}
	if ac == nil {
		t.Fatal("expected non-nil container")
	}
	if ac.Container == nil {
		t.Fatal("expected non-nil base container")
	}
	if ac.GetConfig() == nil {
		t.Fatal("expected non-nil config")
	}
	if ac.GetLogger() == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestApiContainer_RegisterService_GetService(t *testing.T) {
	setContainerEnv(t)
	defer unsetContainerEnv(t)

	ac, err := NewApiContainer()
	if err != nil {
		t.Fatalf("NewApiContainer: %v", err)
	}

	ac.RegisterService("api-svc", "test")
	got, exists := ac.GetService("api-svc")
	if !exists || got != "test" {
		t.Errorf("GetService: exists=%v got=%v", exists, got)
	}
}
