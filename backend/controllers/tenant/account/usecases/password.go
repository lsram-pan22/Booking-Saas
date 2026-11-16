package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidPassword = errors.New("invalid password")
var ErrPasswordDenied = models.ErrPasswordDenied
var ErrCurrentPassword = models.ErrCurrentPassword

type PasswordRepository interface {
	Reset(context.Context, models.PasswordActor, string, string, string) error
	Change(context.Context, models.PasswordActor, string, string, string) error
	Owners(context.Context, string) ([]models.ResettableOwner, error)
}
type PasswordUsecase struct{ Repo PasswordRepository }

func NewPasswordUsecase(repo PasswordRepository) *PasswordUsecase {
	return &PasswordUsecase{Repo: repo}
}
func ValidPassword(value string) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= 12 && len(value) <= 72 && strings.TrimSpace(value) != ""
}
func (u *PasswordUsecase) Reset(ctx context.Context, actor models.PasswordActor, shopID, target string, req models.ResetPasswordRequest) error {
	if !utils.ValidUserID(actor.ID) || !utils.ValidUserID(target) || actor.ID == target || !ValidPassword(req.Password) {
		return ErrInvalidPassword
	}
	if actor.Role != "super_admin" && (actor.Role != "user" || !utils.ValidUserID(shopID)) {
		return ErrPasswordDenied
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return u.Repo.Reset(ctx, actor, shopID, target, hash)
}
func (u *PasswordUsecase) Change(ctx context.Context, actor models.PasswordActor, req models.ChangePasswordRequest) error {
	if !utils.ValidUserID(actor.ID) || !ValidPassword(req.NewPassword) || len(req.CurrentPassword) == 0 || len(req.CurrentPassword) > 72 {
		return ErrInvalidPassword
	}
	if req.CurrentPassword == req.NewPassword {
		return ErrCurrentPassword
	}
	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return u.Repo.Change(ctx, actor, req.CurrentPassword, req.NewPassword, hash)
}
