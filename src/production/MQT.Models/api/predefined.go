package api_models

// PredefinedRole represents a predefined role
type PredefinedRole struct {
	Name        string
	Description string
}

// GetPredefinedRoles returns a list of predefined roles
func GetPredefinedRoles() []PredefinedRole {
	return []PredefinedRole{
		{
			Name:        "admin",
			Description: "Administrator with full access to all resources",
		},
		{
			Name:        "user",
			Description: "Regular user with read-only access to assigned resources",
		},
	}
}
