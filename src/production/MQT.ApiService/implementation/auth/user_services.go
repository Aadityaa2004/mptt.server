package auth

import (
	"context"

	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides user management operations
type UserService struct {
	userRepo interfaces.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*auth_models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*auth_models.User, error) {
	return s.userRepo.GetAll(ctx)
}

// UpdateUserRole updates a user's role
func (s *UserService) UpdateUserRole(ctx context.Context, userID string, newRole string) (*auth_models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Role = newRole

	// Update user
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user in the database
func (s *UserService) UpdateUser(ctx context.Context, user *auth_models.User) (*auth_models.User, error) {
	err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser deletes a user from the database
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID, true) // hard delete
}

// HashPassword hashes a password using bcrypt
func (s *UserService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
