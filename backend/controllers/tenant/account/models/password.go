package models

import "errors"

var ErrPasswordDenied = errors.New("password operation denied")
var ErrCurrentPassword = errors.New("incorrect current password or reused password")

type ResetPasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	Password        string `json:"password"`
}
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type PasswordActor struct {
	ID      string
	Role    string
	Version int64
}
type ResettableOwner struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
