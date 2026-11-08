package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrShopReadOnly = errors.New("shop is read only")

type ManageRepository interface {
	Get(context.Context, string, string) (models.Shop, error)
	Update(context.Context, string, string, models.UpdateShopRequest) (models.Shop, error)
}
type ManageUsecase struct{ Repo ManageRepository }

func NewManageUsecase(repo ManageRepository) *ManageUsecase { return &ManageUsecase{Repo: repo} }
func (u *ManageUsecase) Get(ctx context.Context, userID, shopID string) (models.Shop, error) {
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) {
		return models.Shop{}, ErrInvalidShop
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Get(ctx, userID, shopID)
}
func (u *ManageUsecase) Update(ctx context.Context, userID, shopID string, req models.UpdateShopRequest) (models.Shop, error) {
	req.Name = strings.TrimSpace(req.Name)
	if (req.OpeningTime == nil) != (req.ClosingTime == nil) || (req.OpeningTime != nil && !utils.ValidShopHours(*req.OpeningTime, *req.ClosingTime)) {
		return models.Shop{}, ErrInvalidShop
	}
	validType := req.BusinessType == "salon" || req.BusinessType == "spa" || req.BusinessType == "clinic" || req.BusinessType == "fitness" || req.BusinessType == "other"
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) || req.Name == "" || utf8.RuneCountInString(req.Name) > 150 || !validType || (req.Status != nil && *req.Status != "active" && *req.Status != "inactive") {
		return models.Shop{}, ErrInvalidShop
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	shop, err := u.Repo.Get(ctx, userID, shopID)
	if err != nil {
		return models.Shop{}, err
	}
	if shop.Role != "owner" || (shop.Status != "active" && shop.Status != "inactive") {
		return models.Shop{}, ErrShopReadOnly
	}
	return u.Repo.Update(ctx, userID, shopID, req)
}
