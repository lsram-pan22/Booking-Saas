package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"booking-saas/config"
	"booking-saas/controllers/super_admin/auth/models"
	"booking-saas/controllers/super_admin/auth/repositories"
	"booking-saas/utils"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type UserRepository interface {
	FindByUsername(
		ctx context.Context,
		email string,
	) (*models.SuperAdminAccount, error)
}

type LoginUsecase struct {
	UserRepo UserRepository
	Config   *config.Config
}

func NewLoginUsecase(
	userRepo UserRepository,
	cfg *config.Config,
) *LoginUsecase {
	return &LoginUsecase{
		UserRepo: userRepo,
		Config:   cfg,
	}
}

func (u *LoginUsecase) Login(
	ctx context.Context,
	req models.SuperAdminLoginRequest,
) (*models.SuperAdminLoginData, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	if req.Username == "" || len(req.Username) > 254 || len(req.Password) == 0 || len(req.Password) > 72 {
		return nil, ErrInvalidCredentials
	}

	user, err := u.UserRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			utils.CheckUnknownAccountPassword(req.Password)
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	passwordMatches := utils.CheckPassword(req.Password, user.PasswordHash)
	// ตรวจสอบสิทธิ์ Super Admin
	if user.SystemRole != utils.RoleSuperAdmin {
		return nil, ErrInvalidCredentials
	}

	// ตรวจสอบสถานะบัญชี
	if user.Status != "active" {
		return nil, ErrInvalidCredentials
	}

	// ตรวจสอบ Password
	if !passwordMatches {
		return nil, ErrInvalidCredentials
	}

	// สร้าง JWT Token
	token, err := utils.GenerateJWT(
		u.Config.JWT.Secret,
		u.Config.JWT.ExpiresIn,
		user.ID,
		user.Email,
		user.SystemRole,
		user.SessionVersion,
	)

	if err != nil {
		return nil, err
	}

	return &models.SuperAdminLoginData{
		ExpiresIn:   int64(u.Config.JWT.ExpiresIn / time.Second),
		AccessToken: token,
		TokenType:   "Bearer",
		User: models.SuperAdminUser{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.SystemRole,
		},
	}, nil
}
