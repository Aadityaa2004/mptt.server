package auth

import (
	"context"
	"fmt"

	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	auth_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/auth"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
	rbac "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.ApiService/implementation/rbac"

	"golang.org/x/crypto/bcrypt"
)

// RoleInitializerService handles initializing roles
type RoleInitializerService struct {
	roleRepo    interfaces.RoleRepository
	userRepo    interfaces.UserRepository
	rbacService *rbac.Service
	logger      *logger.Logger
	adminConfig AdminConfig
}

// AdminConfig holds admin user configuration
type AdminConfig struct {
	Username string
	Email    string
	Password string
}

// NewRoleInitializerService creates a new role initializer service
func NewRoleInitializerService(
	roleRepo interfaces.RoleRepository,
	userRepo interfaces.UserRepository,
	rbacService *rbac.Service,
	logger *logger.Logger,
	adminConfig AdminConfig,
) *RoleInitializerService {
	return &RoleInitializerService{
		roleRepo:    roleRepo,
		userRepo:    userRepo,
		rbacService: rbacService,
		logger:      logger,
		adminConfig: adminConfig,
	}
}

// InitializeRoles initializes roles from the database or creates default admin role if none exist
func (s *RoleInitializerService) InitializeRoles(ctx context.Context) error {
	// First check if there are any roles in the database
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	// If no roles exist, create the admin and user roles
	if len(roles) == 0 {
		s.logger.Logger.Info().Msg("No roles found in database. Creating admin and user roles...")

		adminRole := auth_models.NewRole(
			"admin",
			"Administrator with full access to all resources",
		)

		_, err := s.roleRepo.Create(ctx, adminRole)
		if err != nil {
			return err
		}
		s.logger.Logger.Info().Msg("Admin role created successfully")

		// Add admin role to the RBAC service
		s.rbacService.AddRole("admin")

		// Also create a default user role
		userRole := auth_models.NewRole(
			"user",
			"Regular user with read-only access to assigned resources",
		)

		_, err = s.roleRepo.Create(ctx, userRole)
		if err != nil {
			return err
		}
		s.logger.Logger.Info().Msg("User role created successfully")

		// Add user role to the RBAC service
		s.rbacService.AddRole("user")
	} else {
		// Load all roles from the database into the RBAC service
		s.logger.Logger.Info().Int("count", len(roles)).Msg("Loading roles from database")
		for _, role := range roles {
			s.rbacService.AddRole(role.Name)
		}
		s.logger.Logger.Info().Msg("Roles loaded successfully")
	}

	return nil
}

// InitializeAdminUser creates the first admin user if no admin users exist
func (s *RoleInitializerService) InitializeAdminUser(ctx context.Context) error {
	// Check if any admin users exist
	adminUsers, err := s.userRepo.GetByRole(ctx, "admin")
	if err != nil {
		return err
	}

	// If admin users exist, no need to create one
	if len(adminUsers) > 0 {
		s.logger.Logger.Info().Int("count", len(adminUsers)).Msg("Admin users already exist, skipping admin user creation")
		return nil
	}

	// Create the first admin user
	s.logger.Logger.Info().Msg("No admin users found. Creating first admin user...")

	// Hash the admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.adminConfig.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin user with configured credentials
	adminUser := auth_models.NewUser(
		s.adminConfig.Username,
		s.adminConfig.Email,
		string(hashedPassword),
		"admin",
	)

	_, err = s.userRepo.Create(ctx, adminUser)
	if err != nil {
		return err
	}

	s.logger.Logger.Info().Msg("First admin user created successfully")
	s.logger.Logger.Info().Str("username", s.adminConfig.Username).Str("email", s.adminConfig.Email).Msg("Admin user created with configured credentials")
	s.logger.Logger.Warn().Msg("IMPORTANT: Change the admin password after first login for security!")

	return nil
}
