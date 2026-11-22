package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/utils"
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidRegistration = errors.New("invalid registration data")
var phonePattern = regexp.MustCompile(`^(0[0-9]{8,9}|\+66[0-9]{8,9})$`)
var registrationPasswordPattern = regexp.MustCompile(`^[0-9]*[A-Za-z][0-9]*[A-Za-z][0-9]*[A-Za-z][0-9]*$`)

type RegisterRepository interface {
	Register(context.Context, models.Registration) (models.RegisterData, error)
}
type RegisterUsecase struct{ Repo RegisterRepository }

func NewRegisterUsecase(repo RegisterRepository) *RegisterUsecase {
	return &RegisterUsecase{Repo: repo}
}

func (u *RegisterUsecase) Register(ctx context.Context, req models.RegisterRequest) (models.RegisterData, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.NewReplacer(" ", "", "-", "", "\t", "").Replace(strings.TrimSpace(req.Phone))
	address, err := mail.ParseAddress(req.Email)
	validPassword := len(req.Password) == 8 && registrationPasswordPattern.MatchString(req.Password)
	if err != nil || address.Address != req.Email || len(req.Email) > 254 || req.FullName == "" || utf8.RuneCountInString(req.FullName) > 100 || !phonePattern.MatchString(req.Phone) || !validPassword {
		return models.RegisterData{}, ErrInvalidRegistration
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return models.RegisterData{}, err
	}
	input := models.Registration{FullName: req.FullName, Email: req.Email, Phone: req.Phone, PasswordHash: hash, Username: req.Email}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return u.Repo.Register(ctx, input)
}
