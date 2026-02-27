package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"

	"golang.org/x/crypto/bcrypt"
)

// AuthService aggregates auth operations
type AuthService struct {
	userRepo           interfaces.UserRepository
	roleRepo           interfaces.RoleRepository
	jwtService         *jwt.Service
	rbacService        *rbac.Service
	verificationService *VerificationService
}

type RegisterRequest struct {
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Role      string   `json:"role"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenID     string `json:"token_id"`
	ExpiresAt   int64  `json:"expires_at"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenID     string `json:"token_id"`
	ExpiresAt   int64  `json:"expires_at"`
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	jwtService *jwt.Service,
	rbacService *rbac.Service,
	verificationService *VerificationService,
) *AuthService {
	return &AuthService{
		userRepo:            userRepo,
		roleRepo:            roleRepo,
		jwtService:          jwtService,
		rbacService:         rbacService,
		verificationService: verificationService,
	}
}

// RegisterResponse is returned when registration requires email verification
type RegisterResponse struct {
	User                 *auth_models.User `json:"user,omitempty"`
	RequiresVerification bool              `json:"requires_verification"`
	Message              string            `json:"message"`
}

// Register stores pending user data and sends OTP. User is NOT created until OTP verification succeeds.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// If role is not provided, use default "user" role
	if req.Role == "" {
		req.Role = "user"
	}

	// Store OTP with pending user data and send email. NO user created yet.
	// If email fails, nothing is persisted - user can retry.
	meta := &auth_models.SignupMetadata{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
	}
	if err := s.verificationService.GenerateAndStoreSignupOTP(ctx, req.Email, meta); err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return &RegisterResponse{
		User:                 nil, // No user created yet
		RequiresVerification: true,
		Message:              "OTP sent to your email. Please verify to activate your account.",
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, *api_models.TokenPair, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil || user == nil {
		return nil, nil, errors.New("invalid credentials")
	}
	if !user.Active {
		return nil, nil, errors.New("account not verified. Please verify your email first.")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Generate tokens
	tokenPair, err := s.jwtService.GenerateTokens(user.UserID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AuthResponse{
		AccessToken: tokenPair.AccessToken,
		TokenID:     tokenPair.TokenID,
		ExpiresAt:   tokenPair.ExpiresAt,
		UserID:      user.UserID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
	}, tokenPair, nil
}

// RefreshTokens uses a refresh token to generate new access and permission tokens
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*RefreshTokenResponse, *api_models.TokenPair, error) {
	// Validate refresh token and generate new tokens
	tokenPair, err := s.jwtService.RefreshTokens(refreshToken, s.userRepo)
	if err != nil {
		return nil, nil, err
	}

	return &RefreshTokenResponse{
		AccessToken: tokenPair.AccessToken,
		TokenID:     tokenPair.TokenID,
		ExpiresAt:   tokenPair.ExpiresAt,
	}, tokenPair, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userId string) (*auth_models.User, error) {
	return s.userRepo.GetByID(ctx, userId)
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// UpdateUser updates a user in the database
func (s *AuthService) UpdateUser(ctx context.Context, user *auth_models.User) (*auth_models.User, error) {
	err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// VerifyEmailRequest represents the request to verify email with OTP
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}

// VerifyEmail verifies the OTP, creates the user in DB, and returns auth tokens
func (s *AuthService) VerifyEmail(ctx context.Context, req VerifyEmailRequest) (*AuthResponse, *api_models.TokenPair, error) {
	meta, err := s.verificationService.ValidateOTPAndGetSignupMetadata(ctx, req.Email, req.OTP)
	if err != nil || meta == nil {
		return nil, nil, errors.New("invalid or expired OTP")
	}

	// Create user only now, after successful OTP verification
	now := time.Now()
	user := auth_models.NewUser(meta.Username, req.Email, meta.PasswordHash, meta.Role, meta.Latitude, meta.Longitude)
	user.Active = true
	user.EmailVerifiedAt = &now
	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create account: %w", err)
	}

	tokenPair, err := s.jwtService.GenerateTokens(user.UserID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AuthResponse{
		AccessToken: tokenPair.AccessToken,
		TokenID:     tokenPair.TokenID,
		ExpiresAt:   tokenPair.ExpiresAt,
		UserID:      user.UserID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
	}, tokenPair, nil
}

// ResendOTP sends a new OTP for pending signup
func (s *AuthService) ResendOTP(ctx context.Context, email string) error {
	return s.verificationService.ResendSignupOTP(ctx, email)
}

// RequestPasswordReset sends a password reset email (idempotent - always returns success to prevent enumeration)
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Idempotent: don't reveal if user exists
		return nil
	}
	return s.verificationService.GenerateAndStorePasswordResetToken(ctx, email, user.Username)
}

// ResetPasswordRequest represents the request to reset password with token
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword validates the token and updates the user's password
func (s *AuthService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	email, err := s.verificationService.ValidateAndConsumePasswordResetToken(ctx, req.Token)
	if err != nil || email == "" {
		return errors.New("invalid or expired reset link")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}
