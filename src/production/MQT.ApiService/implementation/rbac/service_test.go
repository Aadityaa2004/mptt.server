package rbac

import (
	"testing"
)

func TestService_IsValidRole(t *testing.T) {
	svc := NewService()

	if !svc.IsValidRole("admin") {
		t.Error("admin should be valid")
	}
	if !svc.IsValidRole("user") {
		t.Error("user should be valid")
	}
	if svc.IsValidRole("invalid") {
		t.Error("invalid should not be valid")
	}
}

func TestService_IsAdmin(t *testing.T) {
	svc := NewService()

	if !svc.IsAdmin("admin") {
		t.Error("admin should be admin")
	}
	if svc.IsAdmin("user") {
		t.Error("user should not be admin")
	}
}

func TestService_IsUser(t *testing.T) {
	svc := NewService()

	if !svc.IsUser("user") {
		t.Error("user should be user")
	}
	if svc.IsUser("admin") {
		t.Error("admin should not be user")
	}
}

func TestService_AddRole(t *testing.T) {
	svc := NewService()

	if svc.IsValidRole("custom") {
		t.Error("custom should not be valid initially")
	}
	svc.AddRole("custom")
	if !svc.IsValidRole("custom") {
		t.Error("custom should be valid after AddRole")
	}
}

func TestService_GetValidRoles(t *testing.T) {
	svc := NewService()
	roles := svc.GetValidRoles()

	if len(roles) < 2 {
		t.Errorf("expected at least 2 roles, got %d", len(roles))
	}
	foundAdmin := false
	foundUser := false
	for _, r := range roles {
		if r == "admin" {
			foundAdmin = true
		}
		if r == "user" {
			foundUser = true
		}
	}
	if !foundAdmin || !foundUser {
		t.Errorf("expected admin and user in roles, got %v", roles)
	}
}
