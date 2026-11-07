package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"time"
)

type MemberStatusRepository interface {
	SetMemberStatus(context.Context, string, string, string, string) (models.MemberStatusResult, error)
}
type MemberStatusUsecase struct{ Repo MemberStatusRepository }

func NewMemberStatusUsecase(repo MemberStatusRepository) *MemberStatusUsecase {
	return &MemberStatusUsecase{Repo: repo}
}
func (u *MemberStatusUsecase) Set(ctx context.Context, ownerID, shopID, memberID string, req models.MemberStatusRequest) (models.MemberStatusResult, error) {
	if !utils.ValidUserID(ownerID) || !utils.ValidUserID(shopID) || !utils.ValidUserID(memberID) || (req.Status != "active" && req.Status != "inactive") {
		return models.MemberStatusResult{}, ErrInvalidMember
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return u.Repo.SetMemberStatus(ctx, ownerID, shopID, memberID, req.Status)
}
