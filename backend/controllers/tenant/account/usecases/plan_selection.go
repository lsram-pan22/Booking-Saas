package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/utils"
	"context"
)

type PlanRepository interface {
	Overview(context.Context, string) (models.PlanOverview, error)
	Select(context.Context, string, models.PlanSelection) error
}
type PlanUsecase struct{ Repo PlanRepository }

func NewPlanUsecase(repo PlanRepository) *PlanUsecase { return &PlanUsecase{Repo: repo} }
func ValidPlanSelection(input models.PlanSelection) bool {
	if len(input.ShopIDs) > 100000 || len(input.MemberIDs) > 100000 {
		return false
	}
	seen := map[string]bool{}
	for _, id := range append(append([]string{}, input.ShopIDs...), input.MemberIDs...) {
		if !utils.ValidUserID(id) || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
func (u *PlanUsecase) Select(ctx context.Context, id string, input models.PlanSelection) error {
	if !ValidPlanSelection(input) {
		return models.ErrPlanSelection
	}
	if input.ShopIDs == nil {
		input.ShopIDs = []string{}
	}
	if input.MemberIDs == nil {
		input.MemberIDs = []string{}
	}
	return u.Repo.Select(ctx, id, input)
}
