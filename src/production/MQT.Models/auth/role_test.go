package auth_models

import (
	"testing"
)

func TestNewRole(t *testing.T) {
	r := NewRole("admin", "Administrator")
	if r == nil {
		t.Fatal("NewRole returned nil")
	}
	if r.Name != "admin" || r.Description != "Administrator" {
		t.Errorf("unexpected role: %+v", r)
	}
}
