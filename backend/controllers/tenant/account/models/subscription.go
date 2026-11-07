package models

import "time"

type Subscription struct {
	ShopLimit            int        `json:"shop_limit"`
	ID                   string     `json:"id"`
	PackageID            string     `json:"package_id"`
	PackageCode          string     `json:"package_code"`
	PackageName          string     `json:"package_name"`
	Status               string     `json:"status"`
	MonthlyPriceSatang   int64      `json:"monthly_price_satang"`
	UserLimit            int        `json:"user_limit"`
	StaffManagementLevel string     `json:"staff_management_level"`
	ReportLevel          string     `json:"report_level"`
	ExportExcel          bool       `json:"export_excel"`
	APIWebhook           bool       `json:"api_webhook"`
	StartsAt             time.Time  `json:"starts_at"`
	EndsAt               *time.Time `json:"ends_at"`
}
