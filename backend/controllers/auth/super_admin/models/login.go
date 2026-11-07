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
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	User        SuperAdminUser `json:"user"`
}

type SuperAdminUser struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type SuperAdminAccount struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	SystemRole   string
	Status       string
}
