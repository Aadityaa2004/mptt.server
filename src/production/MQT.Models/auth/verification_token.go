package auth_models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	PurposeSignupOTP     = "signup_otp"
	PurposePasswordReset = "password_reset"
)

// SignupMetadata holds pending user data stored with signup OTP until verification
type SignupMetadata struct {
	Username     string   `json:"username"`
	PasswordHash string   `json:"password_hash"`
	Role         string   `json:"role"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

// VerificationToken represents an OTP or password reset token stored for verification
type VerificationToken struct {
	ID        string     `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	TokenHash string     `json:"token_hash" db:"token_hash"`
	Purpose   string     `json:"purpose" db:"purpose"`
	Metadata  []byte     `json:"metadata,omitempty" db:"metadata"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// NewVerificationToken creates a new VerificationToken with generated ID
func NewVerificationToken(email, tokenHash, purpose string, expiresAt time.Time, metadata []byte) *VerificationToken {
	return &VerificationToken{
		ID:        uuid.New().String(),
		Email:     email,
		TokenHash: tokenHash,
		Purpose:   purpose,
		Metadata:  metadata,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

// ParseSignupMetadata parses Metadata bytes into SignupMetadata
func ParseSignupMetadata(data []byte) (*SignupMetadata, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var m SignupMetadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
