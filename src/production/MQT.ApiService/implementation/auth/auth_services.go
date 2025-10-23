package auth

import (
	"context"
	"errors"

	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	jwt "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/jwt"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"

	"golang.org/x/crypto/bcrypt"
)

// AuthService aggregates auth operations
type AuthService struct {
	userRepo    interfaces.UserRepository
	roleRepo    interfaces.RoleRepository
	jwtService  *jwt.Service
	rbacService *rbac.Service
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
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
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		jwtService:  jwtService,
		rbacService: rbacService,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*auth_models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
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

	// Create user
	user := auth_models.NewUser(req.Username, req.Email, string(hashedPassword), req.Role)
	return s.userRepo.Create(ctx, user)
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, *api_models.TokenPair, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
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
