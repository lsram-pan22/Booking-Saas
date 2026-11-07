package models

type ShopMember struct {
	Username   string `json:"username"`
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	UserStatus string `json:"user_status"`
}

type ShopMembers struct {
	ActiveCount int          `json:"active_count"`
	UserLimit   *int         `json:"user_limit"`
	Members     []ShopMember `json:"members"`
}
