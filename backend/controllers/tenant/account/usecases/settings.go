package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"context"
	"strings"
	"time"
	"unicode/utf8"
)

type SettingsRepository interface {
	Update(context.Context, string, models.UpdateProfileRequest) (models.TenantUser, error)
}
type SettingsUsecase struct{ Repo SettingsRepository }

func NewSettingsUsecase(repo SettingsRepository) *SettingsUsecase {
	return &SettingsUsecase{Repo: repo}
}
func (u *SettingsUsecase) Update(ctx context.Context, id string, req models.UpdateProfileRequest) (models.TenantUser, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	req.Phone = strings.NewReplacer(" ", "", "-", "", "\t", "").Replace(strings.TrimSpace(req.Phone))
	if req.FullName == "" || utf8.RuneCountInString(req.FullName) > 100 || !phonePattern.MatchString(req.Phone) {
		return models.TenantUser{}, ErrInvalidRegistration
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Update(ctx, id, req)
}
