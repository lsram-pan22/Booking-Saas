package models

type MemberStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
type MemberStatusResult struct {
	Member      ShopMember `json:"member"`
	ActiveCount int        `json:"active_count"`
	UserLimit   *int       `json:"user_limit"`
}
