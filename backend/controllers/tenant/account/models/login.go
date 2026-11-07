package models

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}
type Tenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Role   string `json:"role"`
}
type TenantUser struct {
	Role       string `json:"role,omitempty"`
	SystemRole string `json:"system_role,omitempty"`
	ID         string `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
}
type Account struct {
	User         TenantUser
	PasswordHash string
	SystemRole   string
	Status       string
}
type Profile struct {
	WorkspaceRole  string        `json:"workspace_role"`
	OwnedShopCount int           `json:"owned_shop_count"`
	Subscription   *Subscription `json:"subscription"`
	User           TenantUser    `json:"user"`
	Tenants        []Tenant      `json:"tenants"`
}
type LoginData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Profile
}
