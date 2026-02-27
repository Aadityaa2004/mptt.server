package api_models

import (
	"testing"
)

func TestGetPredefinedRoles(t *testing.T) {
	roles := GetPredefinedRoles()
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0].Name != "admin" || roles[1].Name != "user" {
		t.Errorf("unexpected roles: %+v", roles)
	}
}
