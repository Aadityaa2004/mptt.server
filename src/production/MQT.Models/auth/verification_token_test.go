package auth_models

import (
	"testing"
	"time"
)

func TestNewVerificationToken(t *testing.T) {
	exp := time.Now().Add(15 * time.Minute)
	meta := []byte(`{"username":"u1","password_hash":"h","role":"user"}`)
	token := NewVerificationToken("a@b.com", "hash123", PurposeSignupOTP, exp, meta)
	if token == nil {
		t.Fatal("NewVerificationToken returned nil")
	}
	if token.ID == "" {
		t.Error("expected non-empty ID")
	}
	if token.Email != "a@b.com" || token.TokenHash != "hash123" || token.Purpose != PurposeSignupOTP {
		t.Errorf("unexpected token fields: %+v", token)
	}
	if len(token.Metadata) != len(meta) {
		t.Error("metadata not preserved")
	}
}

func TestParseSignupMetadata_Empty(t *testing.T) {
	m, err := ParseSignupMetadata(nil)
	if err != nil {
		t.Fatalf("ParseSignupMetadata nil: %v", err)
	}
	if m != nil {
		t.Errorf("expected nil for empty, got %+v", m)
	}

	m2, err := ParseSignupMetadata([]byte{})
	if err != nil {
		t.Fatalf("ParseSignupMetadata empty: %v", err)
	}
	if m2 != nil {
		t.Errorf("expected nil for empty bytes, got %+v", m2)
	}
}

func TestParseSignupMetadata_ValidJSON(t *testing.T) {
	json := `{"username":"alice","password_hash":"h123","role":"user"}`
	m, err := ParseSignupMetadata([]byte(json))
	if err != nil {
		t.Fatalf("ParseSignupMetadata: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil metadata")
	}
	if m.Username != "alice" || m.PasswordHash != "h123" || m.Role != "user" {
		t.Errorf("unexpected metadata: %+v", m)
	}
}

func TestParseSignupMetadata_InvalidJSON(t *testing.T) {
	m, err := ParseSignupMetadata([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if m != nil {
		t.Errorf("expected nil on error, got %+v", m)
	}
}
