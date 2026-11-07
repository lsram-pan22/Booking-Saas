package models

import (
	"regexp"
	"strings"
)

type Package struct {
	ShopLimit          int    `json:"shop_limit"`
	ID                 string `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	MonthlyPriceSatang int64  `json:"monthly_price_satang"`
	UserLimit          int    `json:"user_limit"`
	StaffManagement    string `json:"staff_management"`
	Reports            string `json:"reports"`
	ExportExcel        bool   `json:"export_excel"`
	APIWebhook         bool   `json:"api_webhook"`
	Active             bool   `json:"active"`
	SortOrder          int    `json:"sort_order"`
}

var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ValidID(id string) bool { return uuidPattern.MatchString(id) }

func (p Package) Valid() bool {
	return codePattern.MatchString(p.Code) && len(strings.TrimSpace(p.Name)) > 0 && len(p.Name) <= 100 && len(p.Description) <= 2000 &&
		p.MonthlyPriceSatang >= 0 && p.MonthlyPriceSatang <= 100000000 &&
		p.ShopLimit >= 1 && p.ShopLimit <= 100000 && p.UserLimit >= 1 && p.UserLimit <= 100000 && p.SortOrder >= 0 && p.SortOrder <= 100000 &&
		(p.StaffManagement == "basic" || p.StaffManagement == "full") && (p.Reports == "basic" || p.Reports == "advanced")
}
