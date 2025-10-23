package api_models


type RegisterRequest struct {
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}