package api_models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds JWT configuration
type Config struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// AccessClaims represents the JWT claims for user access
type AccessClaims struct {
	jwt.RegisteredClaims
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
	TokenID string `json:"token_id"`
}

// RefreshClaims represents the JWT claims for refresh tokens
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID  string `json:"user_id"`
	TokenID string `json:"token_id"`
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenID      string `json:"token_id"`
	ExpiresAt    int64  `json:"expires_at"`
}
