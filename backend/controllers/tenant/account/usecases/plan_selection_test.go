package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"testing"
)

func TestPlanSelectionIDs(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	for _, tc := range []struct {
		name  string
		input models.PlanSelection
		valid bool
	}{
		{"empty", models.PlanSelection{}, true},
		{"valid", models.PlanSelection{ShopIDs: []string{id}}, true},
		{"duplicate", models.PlanSelection{ShopIDs: []string{id, id}}, false},
		{"invalid member", models.PlanSelection{MemberIDs: []string{"invalid"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if ValidPlanSelection(tc.input) != tc.valid {
				t.Fatal("unexpected validation result")
			}
		})
	}
}
