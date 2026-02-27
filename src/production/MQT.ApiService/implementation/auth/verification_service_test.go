package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type mockTokenRepo struct {
	tokens map[string]*auth_models.VerificationToken
	byHash map[string]*auth_models.VerificationToken
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		tokens: make(map[string]*auth_models.VerificationToken),
		byHash: make(map[string]*auth_models.VerificationToken),
	}
}

func (m *mockTokenRepo) Create(ctx context.Context, t *auth_models.VerificationToken) error {
	m.tokens[t.ID] = t
	m.byHash[t.TokenHash] = t
	return nil
}
func (m *mockTokenRepo) GetValidByEmailAndPurpose(ctx context.Context, email, purpose string) (*auth_models.VerificationToken, error) {
	for _, t := range m.tokens {
		if t.Email == email && t.Purpose == purpose && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, nil
}
func (m *mockTokenRepo) GetValidByTokenHash(ctx context.Context, hash, purpose string) (*auth_models.VerificationToken, error) {
	t := m.byHash[hash]
	if t != nil && t.Purpose == purpose && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
		return t, nil
	}
	return nil, nil
}
func (m *mockTokenRepo) GetLatestSignupTokenByEmail(ctx context.Context, email string) (*auth_models.VerificationToken, error) {
	var latest *auth_models.VerificationToken
	for _, t := range m.tokens {
		if t.Email == email && t.Purpose == auth_models.PurposeSignupOTP {
			if latest == nil || t.CreatedAt.After(latest.CreatedAt) {
				latest = t
			}
		}
	}
	return latest, nil
}
func (m *mockTokenRepo) MarkUsed(ctx context.Context, id string) error {
	if t, ok := m.tokens[id]; ok {
		now := time.Now()
		t.UsedAt = &now
	}
	return nil
}
func (m *mockTokenRepo) InvalidateByEmailAndPurpose(ctx context.Context, email, purpose string) error {
	for id, t := range m.tokens {
		if t.Email == email && t.Purpose == purpose {
			delete(m.tokens, id)
			delete(m.byHash, t.TokenHash)
		}
	}
	return nil
}
func (m *mockTokenRepo) Delete(ctx context.Context, id string) error {
	if t, ok := m.tokens[id]; ok {
		delete(m.byHash, t.TokenHash)
		delete(m.tokens, id)
	}
	return nil
}
func (m *mockTokenRepo) DeleteExpired(ctx context.Context, before time.Time) error { return nil }

func TestVerificationService_ValidateOTPAndGetSignupMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tokenRepo := newMockTokenRepo()
	meta := &auth_models.SignupMetadata{Username: "u1", PasswordHash: "hash", Role: "user"}
	svc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	ctx := context.Background()

	// First create a token via GenerateAndStoreSignupOTP to exercise happy path
	if err := svc.GenerateAndStoreSignupOTP(ctx, "test@x.com", meta); err != nil {
		t.Fatalf("GenerateAndStoreSignupOTP: %v", err)
	}

	// Now insert a token with a known hash so we can test success path of ValidateOTPAndGetSignupMetadata
	otp := "654321"
	hash := sha256Hash(otp)
	meta2 := &auth_models.SignupMetadata{Username: "u2", PasswordHash: "hash2", Role: "admin"}
	metaJSON, err := json.Marshal(meta2)
	if err != nil {
		t.Fatalf("Marshal meta2: %v", err)
	}
	token := auth_models.NewVerificationToken("test2@x.com", hash, auth_models.PurposeSignupOTP, time.Now().Add(10*time.Minute), metaJSON)
	if err := tokenRepo.Create(ctx, token); err != nil {
		t.Fatalf("Create token: %v", err)
	}

	storedBefore, _ := tokenRepo.GetValidByEmailAndPurpose(ctx, "test2@x.com", auth_models.PurposeSignupOTP)
	if storedBefore == nil {
		t.Fatal("expected stored token before validation")
	}

	gotMeta, err := svc.ValidateOTPAndGetSignupMetadata(ctx, "test2@x.com", otp)
	if err != nil {
		t.Fatalf("ValidateOTPAndGetSignupMetadata success: %v", err)
	}
	if gotMeta == nil || gotMeta.Username != "u2" || gotMeta.Role != "admin" {
		t.Errorf("unexpected metadata: %+v", gotMeta)
	}

	// After validation, token should be marked used
	storedAfter, _ := tokenRepo.GetValidByEmailAndPurpose(ctx, "test2@x.com", auth_models.PurposeSignupOTP)
	if storedAfter != nil {
		t.Error("expected token to be marked used and not returned as valid")
	}
}

func TestVerificationService_ResendSignupOTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tokenRepo := newMockTokenRepo()
	svc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	ctx := context.Background()

	// Manually create a signup token so GetLatestSignupTokenByEmail finds it (avoids rate limit from GenerateAndStoreSignupOTP)
	exp := time.Now().Add(15 * time.Minute)
	metaJSON := []byte(`{"username":"u1","password_hash":"hash","role":"user"}`)
	token := auth_models.NewVerificationToken("resend2@x.com", "hash", auth_models.PurposeSignupOTP, exp, metaJSON)
	tokenRepo.Create(ctx, token)

	err := svc.ResendSignupOTP(ctx, "resend2@x.com")
	if err != nil {
		t.Fatalf("ResendSignupOTP: %v", err)
	}
}

func TestVerificationService_ResendSignupOTP_NoPending(t *testing.T) {
	svc := NewVerificationService(newMockTokenRepo(), "http://localhost:9999", "http://localhost:3000")
	ctx := context.Background()
	err := svc.ResendSignupOTP(ctx, "nonexistent@x.com")
	if err == nil {
		t.Error("expected error when no pending registration")
	}
}

func TestVerificationService_GenerateAndStorePasswordResetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewVerificationService(newMockTokenRepo(), server.URL, "http://localhost:3000")
	ctx := context.Background()
	err := svc.GenerateAndStorePasswordResetToken(ctx, "user@x.com", "User")
	if err != nil {
		t.Fatalf("GenerateAndStorePasswordResetToken: %v", err)
	}
}

func TestVerificationService_ValidateAndConsumePasswordResetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tokenRepo := newMockTokenRepo()
	svc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	ctx := context.Background()

	// Create a reset token manually with a known token string/hash
	tokenStr := "known-reset-token"
	hash := sha256Hash(tokenStr)
	vt := auth_models.NewVerificationToken("reset@x.com", hash, auth_models.PurposePasswordReset, time.Now().Add(time.Hour), nil)
	if err := tokenRepo.Create(ctx, vt); err != nil {
		t.Fatalf("Create reset token: %v", err)
	}

	email, err := svc.ValidateAndConsumePasswordResetToken(ctx, tokenStr)
	if err != nil {
		t.Fatalf("ValidateAndConsumePasswordResetToken success: %v", err)
	}
	if email != "reset@x.com" {
		t.Errorf("expected email 'reset@x.com', got %q", email)
	}

	// Token should now be marked used; validating again should return empty email
	email2, err := svc.ValidateAndConsumePasswordResetToken(ctx, tokenStr)
	if err != nil {
		t.Fatalf("ValidateAndConsumePasswordResetToken second time: %v", err)
	}
	if email2 != "" {
		t.Errorf("expected empty email after token consumed, got %q", email2)
	}
}

func TestVerificationService_GenerateAndStoreSignupOTP_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tokenRepo := newMockTokenRepo()
	svc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	ctx := context.Background()

	meta := &auth_models.SignupMetadata{Username: "u1", PasswordHash: "hash", Role: "user"}
	if err := svc.GenerateAndStoreSignupOTP(ctx, "rate@x.com", meta); err != nil {
		t.Fatalf("first GenerateAndStoreSignupOTP: %v", err)
	}
	// Second call within OTPRateLimitSeconds should be rate limited
	if err := svc.GenerateAndStoreSignupOTP(ctx, "rate@x.com", meta); err == nil {
		t.Error("expected rate limit error on second OTP request")
	}
}

func TestVerificationService_GenerateAndStorePasswordResetToken_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewVerificationService(newMockTokenRepo(), server.URL, "http://localhost:3000")
	ctx := context.Background()

	if err := svc.GenerateAndStorePasswordResetToken(ctx, "rl@x.com", "User"); err != nil {
		t.Fatalf("first GenerateAndStorePasswordResetToken: %v", err)
	}
	// Second call within PasswordResetRateMin minutes should be silently allowed (returns nil) due to rate limit
	if err := svc.GenerateAndStorePasswordResetToken(ctx, "rl@x.com", "User"); err != nil {
		t.Fatalf("expected nil error on rate-limited password reset request, got: %v", err)
	}
}

func TestVerificationService_EmailServiceErrors(t *testing.T) {
	// Email service returns 500 to trigger error paths for signup and password reset emails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tokenRepo := newMockTokenRepo()
	svc := NewVerificationService(tokenRepo, server.URL, "http://localhost:3000")
	ctx := context.Background()

	meta := &auth_models.SignupMetadata{Username: "uerr", PasswordHash: "hash", Role: "user"}
	if err := svc.GenerateAndStoreSignupOTP(ctx, "emailerr@x.com", meta); err == nil {
		t.Error("expected error when OTP email service returns non-2xx")
	}

	if err := svc.GenerateAndStorePasswordResetToken(ctx, "emailerr@x.com", "User"); err == nil {
		t.Error("expected error when password reset email service returns non-2xx")
	}
}

// Ensure mock implements interface
var _ interfaces.VerificationTokenRepository = (*mockTokenRepo)(nil)
