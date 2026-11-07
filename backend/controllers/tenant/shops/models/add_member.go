package models

type AddMemberRequest struct {
	Username string `json:"username"`
	Mode     string `json:"mode" binding:"required"`
	Email    string `json:"email" binding:"required"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type NewMember struct {
	Username     string
	Mode         string
	Email        string
	FullName     string
	Phone        string
	PasswordHash string
}
