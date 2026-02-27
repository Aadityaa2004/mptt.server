package main

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitializeEmailApp_BasicWiring(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router, cfg, log, shutdown, err := InitializeEmailApp()
	if err != nil {
		t.Fatalf("InitializeEmailApp returned error: %v", err)
	}
	if router == nil {
		t.Fatal("router is nil")
	}
	if cfg == nil {
		t.Fatal("config is nil")
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

