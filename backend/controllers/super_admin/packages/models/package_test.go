package models

import "testing"

func TestPackageValidation(t *testing.T) {
	base := Package{ShopLimit: 1, Code: "growth", Name: "Growth", MonthlyPriceSatang: 69900, UserLimit: 15, StaffManagement: "full", Reports: "advanced"}
	tests := []struct {
		name   string
		mutate func(*Package)
		valid  bool
	}{
		{"valid", func(p *Package) {}, true},
		{"free", func(p *Package) { p.MonthlyPriceSatang = 0 }, true},
		{"negative price", func(p *Package) { p.MonthlyPriceSatang = -1 }, false},
		{"missing owner seat", func(p *Package) { p.UserLimit = 0 }, false},
		{"missing shop quota", func(p *Package) { p.ShopLimit = 0 }, false},
		{"invalid report", func(p *Package) { p.Reports = "unknown" }, false},
		{"invalid staff", func(p *Package) { p.StaffManagement = "unknown" }, false},
		{"blank name", func(p *Package) { p.Name = "  " }, false},
		{"invalid code", func(p *Package) { p.Code = "Growth plan" }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := base
			tt.mutate(&p)
			if p.Valid() != tt.valid {
				t.Fatalf("valid = %v, want %v", p.Valid(), tt.valid)
			}
		})
	}
}
