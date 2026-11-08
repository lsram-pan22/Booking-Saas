package models

import "time"

type UpdateBookingRequest struct {
	CustomerRequest
	ServiceID *string    `json:"service_id"`
	StartsAt  *time.Time `json:"starts_at"`
}

type BookingEditSnapshot struct {
	CustomerRequest
	ServiceID       string    `json:"service_id"`
	ServiceName     string    `json:"service_name"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceSatang     int64     `json:"price_satang"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Status          string    `json:"status"`
}
