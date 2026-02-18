package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

const (
	OTPExpiryMinutes       = 15
	PasswordResetExpiryMin = 60
	OTPRateLimitSeconds    = 60
	PasswordResetRateMin   = 15
)

// VerificationService handles OTP and password reset token generation and email sending
type VerificationService struct {
	tokenRepo      interfaces.VerificationTokenRepository
	emailServiceURL string
	frontendBaseURL string
	rateLimit      map[string]time.Time
	rateMu         sync.Mutex
}

// NewVerificationService creates a new verification service
func NewVerificationService(
	tokenRepo interfaces.VerificationTokenRepository,
	emailServiceURL, frontendBaseURL string,
) *VerificationService {
	return &VerificationService{
		tokenRepo:       tokenRepo,
		emailServiceURL: emailServiceURL,
		frontendBaseURL: frontendBaseURL,
		rateLimit:       make(map[string]time.Time),
	}
}

// GenerateAndStoreSignupOTP stores pending user data with OTP, sends email. User is NOT created until OTP is verified.
// If email send fails, the token is deleted (rollback) and no user is created.
func (s *VerificationService) GenerateAndStoreSignupOTP(ctx context.Context, email string, meta *auth_models.SignupMetadata) error {
	// Rate limit: 1 per email per 60 seconds
	s.rateMu.Lock()
	key := "otp:" + email
	if t, ok := s.rateLimit[key]; ok && time.Since(t) < OTPRateLimitSeconds*time.Second {
		s.rateMu.Unlock()
		return fmt.Errorf("please wait %d seconds before requesting another OTP", OTPRateLimitSeconds)
	}
	s.rateLimit[key] = time.Now()
	s.rateMu.Unlock()

	// Invalidate any existing OTP for this email
	_ = s.tokenRepo.InvalidateByEmailAndPurpose(ctx, email, auth_models.PurposeSignupOTP)

	otp, err := generateNumericOTP(6)
	if err != nil {
		return err
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	hash := sha256Hash(otp)
	expiresAt := time.Now().Add(OTPExpiryMinutes * time.Minute)
	token := auth_models.NewVerificationToken(email, hash, auth_models.PurposeSignupOTP, expiresAt, metaJSON)

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return err
	}

	if err := s.sendOTPEmail(email, otp); err != nil {
		_ = s.tokenRepo.Delete(ctx, token.ID) // Rollback: remove token since email failed
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// ValidateOTPAndGetSignupMetadata validates the OTP, marks it used, and returns the stored signup metadata.
// Returns nil metadata if OTP is invalid/expired.
func (s *VerificationService) ValidateOTPAndGetSignupMetadata(ctx context.Context, email, otp string) (*auth_models.SignupMetadata, error) {
	stored, err := s.tokenRepo.GetValidByEmailAndPurpose(ctx, email, auth_models.PurposeSignupOTP)
	if err != nil || stored == nil {
		return nil, nil
	}
	hash := sha256Hash(otp)
	if hash != stored.TokenHash {
		return nil, nil
	}
	meta, err := auth_models.ParseSignupMetadata(stored.Metadata)
	if err != nil || meta == nil {
		return nil, nil
	}
	if err := s.tokenRepo.MarkUsed(ctx, stored.ID); err != nil {
		return nil, err
	}
	return meta, nil
}

// ResendSignupOTP creates a new OTP using metadata from the latest signup token for this email.
func (s *VerificationService) ResendSignupOTP(ctx context.Context, email string) error {
	// Get metadata from the most recent signup token (even if expired/used)
	stored, err := s.tokenRepo.GetLatestSignupTokenByEmail(ctx, email)
	if err != nil || stored == nil {
		return fmt.Errorf("no pending registration found. Please register again")
	}
	meta, err := auth_models.ParseSignupMetadata(stored.Metadata)
	if err != nil || meta == nil {
		return fmt.Errorf("invalid registration data. Please register again")
	}

	// Invalidate old OTPs and create new one with same metadata
	_ = s.tokenRepo.InvalidateByEmailAndPurpose(ctx, email, auth_models.PurposeSignupOTP)
	return s.GenerateAndStoreSignupOTP(ctx, email, meta)
}


// GenerateAndStorePasswordResetToken generates a secure token, stores hash, sends email
func (s *VerificationService) GenerateAndStorePasswordResetToken(ctx context.Context, email, userName string) error {
	// Rate limit: 1 per email per 15 minutes
	s.rateMu.Lock()
	key := "pwd_reset:" + email
	if t, ok := s.rateLimit[key]; ok && time.Since(t) < PasswordResetRateMin*time.Minute {
		s.rateMu.Unlock()
		return nil // Idempotent: don't reveal if rate limited, just return success
	}
	s.rateLimit[key] = time.Now()
	s.rateMu.Unlock()

	// Invalidate any existing reset tokens for this email
	_ = s.tokenRepo.InvalidateByEmailAndPurpose(ctx, email, auth_models.PurposePasswordReset)

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	tokenStr := hex.EncodeToString(tokenBytes)
	hash := sha256Hash(tokenStr)
	expiresAt := time.Now().Add(PasswordResetExpiryMin * time.Minute)
	vt := auth_models.NewVerificationToken(email, hash, auth_models.PurposePasswordReset, expiresAt, nil)

	if err := s.tokenRepo.Create(ctx, vt); err != nil {
		return err
	}

	resetLink := s.frontendBaseURL + "/reset-password?token=" + tokenStr
	if err := s.sendPasswordResetEmail(email, resetLink, userName); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ValidateAndConsumePasswordResetToken validates the token, returns email, and marks it used
func (s *VerificationService) ValidateAndConsumePasswordResetToken(ctx context.Context, tokenStr string) (string, error) {
	hash := sha256Hash(tokenStr)
	// We need to find the token by hash - our repo doesn't support that. We need to add GetByTokenHash or similar.
	// Alternative: iterate by purpose and check. Or add GetValidByTokenHash to the repo.
	// Simpler: store the token hash, and we can only validate by having the token. We need to look up by email+ purpose for OTP.
	// For password reset, we have the token in the URL. We need to find the verification_tokens row where token_hash = hash and used_at is null and expires_at > now.
	// Add GetValidByTokenHash to the interface.
	// For now, let me add that method to the repo.
	// Actually looking at the interface - we have GetValidByEmailAndPurpose. For password reset we don't have the email from the token. We need GetValidByTokenHash.
	// Let me add that to the interface and implementation.
	stored, err := s.tokenRepo.GetValidByTokenHash(ctx, hash, auth_models.PurposePasswordReset)
	if err != nil || stored == nil {
		return "", nil
	}
	if err := s.tokenRepo.MarkUsed(ctx, stored.ID); err != nil {
		return "", err
	}
	return stored.Email, nil
}

// sendOTPEmail calls the EmailService to send OTP
func (s *VerificationService) sendOTPEmail(to, otp string) error {
	payload := map[string]string{
		"to":  to,
		"otp": otp,
		"purpose": "signup",
	}
	return s.postToEmailService("/send-otp", payload)
}

// sendPasswordResetEmail calls the EmailService to send password reset link
func (s *VerificationService) sendPasswordResetEmail(to, resetLink, userName string) error {
	payload := map[string]string{
		"to":         to,
		"reset_link": resetLink,
		"user_name":  userName,
	}
	return s.postToEmailService("/send-password-reset", payload)
}

func (s *VerificationService) postToEmailService(path string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := s.emailServiceURL + path
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("email service returned status %d", resp.StatusCode)
	}
	return nil
}

func generateNumericOTP(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[n.Int64()]
	}
	return string(result), nil
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
