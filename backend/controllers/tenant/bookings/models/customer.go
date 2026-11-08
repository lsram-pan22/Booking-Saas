package models

type CustomerRequest struct {
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	Note          string `json:"note"`
}
