package models

import "time"

type CreateBookingRequest struct {
	ServiceID     string    `json:"service_id"`
	StaffMemberID string    `json:"staff_member_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	Note          string    `json:"note"`
	StartsAt      time.Time `json:"starts_at"`
}
type StatusRequest struct {
	Status string `json:"status"`
	Remark string `json:"remark"`
}
type Booking struct {
	CanManage       bool      `json:"can_manage"`
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	ServiceID       string    `json:"service_id"`
	StaffMemberID   string    `json:"staff_member_id"`
	CustomerName    string    `json:"customer_name"`
	CustomerPhone   string    `json:"customer_phone"`
	Note            string    `json:"note"`
	ServiceName     string    `json:"service_name"`
	StaffName       string    `json:"staff_name"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceSatang     int64     `json:"price_satang"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Status          string    `json:"status"`
	Remark          string    `json:"remark"`
}
type ServiceOption struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"duration_minutes"`
	PriceSatang     int64  `json:"price_satang"`
}
type StaffOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}
type BookingList struct {
	OpeningTime *string         `json:"opening_time"`
	ClosingTime *string         `json:"closing_time"`
	Role        string          `json:"role"`
	ShopName    string          `json:"shop_name"`
	ShopStatus  string          `json:"shop_status"`
	Items       []Booking       `json:"items"`
	Services    []ServiceOption `json:"services"`
	Staff       []StaffOption   `json:"staff"`
}
