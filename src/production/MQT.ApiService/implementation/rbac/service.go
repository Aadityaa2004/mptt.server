package rbac

// Service provides RBAC operations
type Service struct {
	roles map[string]bool
}

// NewService creates a new RBAC service with predefined roles
func NewService() *Service {
	return &Service{
		roles: map[string]bool{
			"admin": true,
			"user":  true,
		},
	}
}

// IsValidRole checks if a role is valid
func (s *Service) IsValidRole(roleName string) bool {
	return s.roles[roleName]
}

// IsAdmin checks if a role is admin
func (s *Service) IsAdmin(roleName string) bool {
	return roleName == "admin"
}

// IsUser checks if a role is user
func (s *Service) IsUser(roleName string) bool {
	return roleName == "user"
}

// AddRole adds a new role
func (s *Service) AddRole(roleName string) {
	s.roles[roleName] = true
}

// GetValidRoles returns all valid roles
func (s *Service) GetValidRoles() []string {
	var roles []string
	for role := range s.roles {
		roles = append(roles, role)
	}
	return roles
}
