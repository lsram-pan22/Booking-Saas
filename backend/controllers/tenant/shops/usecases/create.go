package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidShop = errors.New("invalid shop data")

type ShopRepository interface {
	Create(context.Context, string, models.NewShop) (models.Shop, error)
}
type ShopUsecase struct{ Repo ShopRepository }

func NewShopUsecase(repo ShopRepository) *ShopUsecase { return &ShopUsecase{Repo: repo} }
func (u *ShopUsecase) Create(ctx context.Context, userID string, req models.CreateShopRequest) (models.Shop, error) {
	req.Name = strings.TrimSpace(req.Name)
	validType := req.BusinessType == "salon" || req.BusinessType == "spa" || req.BusinessType == "clinic" || req.BusinessType == "fitness" || req.BusinessType == "other"
	if !utils.ValidUserID(userID) || req.Name == "" || utf8.RuneCountInString(req.Name) > 150 || !validType {
		return models.Shop{}, ErrInvalidShop
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return models.Shop{}, err
	}
	input := models.NewShop{Name: req.Name, BusinessType: req.BusinessType, Slug: "shop-" + hex.EncodeToString(random[:])}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return u.Repo.Create(ctx, userID, input)
}
