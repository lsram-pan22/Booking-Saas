package models

type ServiceRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	PriceSatang     int64  `json:"price_satang"`
	Status          string `json:"status" binding:"required"`
}

type Service struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	PriceSatang     int64  `json:"price_satang"`
	Status          string `json:"status"`
}

type ServiceList struct {
	ShopName   string    `json:"shop_name"`
	ShopStatus string    `json:"shop_status"`
	Role       string    `json:"role"`
	Items      []Service `json:"items"`
}
