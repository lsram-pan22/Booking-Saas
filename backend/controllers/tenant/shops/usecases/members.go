package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"time"
)

type MembersRepository interface {
	Members(context.Context, string, string) (models.ShopMembers, error)
}
type MembersUsecase struct{ Repo MembersRepository }

func NewMembersUsecase(repo MembersRepository) *MembersUsecase { return &MembersUsecase{Repo: repo} }
func (u *MembersUsecase) List(ctx context.Context, userID, shopID string) (models.ShopMembers, error) {
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) {
		return models.ShopMembers{}, ErrInvalidShop
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Members(ctx, userID, shopID)
}
