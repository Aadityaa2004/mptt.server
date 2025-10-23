package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/google/uuid"
	api_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/api"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

// Service provides JWT operations
type Service struct {
	config api_models.Config
}

// NewService creates a new JWT service
func NewService(config api_models.Config) *Service {
	return &Service{
		config: config,
	}
}

// GenerateTokens creates a new set of tokens: access and refresh
func (s *Service) GenerateTokens(userID, role string) (*api_models.TokenPair, error) {
	tokenID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(s.config.AccessTokenDuration)

	// Generate access token
	accessClaims := api_models.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
		},
		UserID:  userID,
		Role:    role,
		TokenID: tokenID,
	}

	// Generate refresh token
	refreshClaims := api_models.RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
		},
		UserID:  userID,
		TokenID: tokenID,
	}

	// Sign the tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessTokenString, err := accessToken.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return nil, err
	}

	return &api_models.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenID:      tokenID,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *Service) ValidateAccessToken(tokenString string) (*api_models.AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &api_models.AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*api_models.AccessClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *Service) ValidateRefreshToken(tokenString string) (*api_models.RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &api_models.RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*api_models.RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// RefreshTokens generates new access token using a refresh token
func (s *Service) RefreshTokens(refreshTokenString string, userRepo interfaces.UserRepository) (*api_models.TokenPair, error) {
	// Validate the refresh token
	refreshClaims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Extract information from the refresh token
	userID := refreshClaims.UserID

	// Get user data from the database
	user, err := userRepo.FindByID(context.Background(), userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new tokens with the user information
	newTokens, err := s.GenerateTokens(userID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate new tokens: " + err.Error())
	}

	return newTokens, nil
}
