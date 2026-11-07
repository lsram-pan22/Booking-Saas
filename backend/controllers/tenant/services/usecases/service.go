package usecases

import (
	"booking-saas/controllers/tenant/services/models"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidService = errors.New("invalid service data")

type ServiceRepository interface {
	List(context.Context, string, string) (models.ServiceList, error)
	Save(context.Context, string, string, string, models.ServiceRequest) (models.Service, error)
}
type ServiceUsecase struct{ Repo ServiceRepository }

func NewServiceUsecase(repo ServiceRepository) *ServiceUsecase { return &ServiceUsecase{Repo: repo} }
func (u *ServiceUsecase) List(ctx context.Context, userID, shopID string) (models.ServiceList, error) {
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) {
		return models.ServiceList{}, ErrInvalidService
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.List(ctx, userID, shopID)
}
func (u *ServiceUsecase) Save(ctx context.Context, userID, shopID, serviceID string, req models.ServiceRequest) (models.Service, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) || (serviceID != "" && !utils.ValidUserID(serviceID)) || req.Name == "" || utf8.RuneCountInString(req.Name) > 150 || utf8.RuneCountInString(req.Description) > 2000 || req.DurationMinutes < 1 || req.DurationMinutes > 1440 || req.PriceSatang < 0 || req.PriceSatang > 100000000 || (req.Status != "active" && req.Status != "inactive") {
		return models.Service{}, ErrInvalidService
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Save(ctx, userID, shopID, serviceID, req)
}
