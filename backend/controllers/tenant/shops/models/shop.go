package models

type CreateShopRequest struct {
	Name         string `json:"name" binding:"required"`
	BusinessType string `json:"business_type" binding:"required"`
}

type NewShop struct {
	Name         string
	BusinessType string
	Slug         string
}

type UpdateShopRequest struct {
	Name         string  `json:"name" binding:"required"`
	BusinessType string  `json:"business_type" binding:"required"`
	Status       *string `json:"status"`
}

type Shop struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BusinessType string `json:"business_type"`
	Slug         string `json:"slug"`
	Status       string `json:"status"`
	Role         string `json:"role"`
}
