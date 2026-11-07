package usecases

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/repositories"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type LoginRepository interface {
	FindByIdentity(context.Context, string) (models.Account, error)
	Profile(context.Context, string) (models.Profile, error)
}
type LoginUsecase struct {
	Repo   LoginRepository
	Config *config.Config
}

func NewLoginUsecase(repo LoginRepository, cfg *config.Config) *LoginUsecase {
	return &LoginUsecase{Repo: repo, Config: cfg}
}
func (u *LoginUsecase) Login(ctx context.Context, req models.LoginRequest) (models.LoginData, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	identity := strings.ToLower(strings.TrimSpace(req.Username))
	if identity == "" {
		identity = strings.ToLower(strings.TrimSpace(req.Email))
	}
	if identity == "" || len(identity) > 254 {
		return models.LoginData{}, ErrInvalidCredentials
	}
	a, err := u.Repo.FindByIdentity(ctx, identity)
	if errors.Is(err, repositories.ErrAccountNotFound) {
		return models.LoginData{}, ErrInvalidCredentials
	}
	if err != nil {
		return models.LoginData{}, err
	}
	if a.Status != "active" || a.SystemRole != "user" || !utils.CheckPassword(req.Password, a.PasswordHash) {
		return models.LoginData{}, ErrInvalidCredentials
	}
	p, err := u.Profile(ctx, a.User.ID)
	if err != nil {
		return models.LoginData{}, err
	}
	token, err := utils.GenerateJWT(u.Config.JWT.Secret, u.Config.JWT.ExpiresIn, a.User.ID, a.User.Email, "user")
	if err != nil {
		return models.LoginData{}, err
	}
	return models.LoginData{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(u.Config.JWT.ExpiresIn / time.Second), Profile: p}, nil
}
func (u *LoginUsecase) Profile(ctx context.Context, id string) (models.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	p, err := u.Repo.Profile(ctx, id)
	if errors.Is(err, repositories.ErrAccountNotFound) {
		return models.Profile{}, ErrInvalidCredentials
	}
	return p, err
}
