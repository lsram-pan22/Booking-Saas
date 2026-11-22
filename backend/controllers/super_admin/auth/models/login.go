package models

type SuperAdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SuperAdminLoginResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    SuperAdminLoginData `json:"data,omitempty"`
}

type SuperAdminLoginData struct {
	ExpiresIn   int64          `json:"expires_in"`
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	User        SuperAdminUser `json:"user"`
}

type SuperAdminUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type SuperAdminAccount struct {
	SessionVersion int64
	ID             string
	Username       string
	Email          string
	PasswordHash   string
	FullName       string
	SystemRole     string
	Status         string
}
