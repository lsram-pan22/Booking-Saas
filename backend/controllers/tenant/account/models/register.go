package models

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Registration contains server-generated values only after input validation.
type Registration struct {
	FullName     string
	Email        string
	Phone        string
	PasswordHash string
	Username     string
}

type RegisterData struct {
	UserID       string       `json:"user_id"`
	Subscription Subscription `json:"subscription"`
}
