package usecases

import (
	"context"
	"errors"

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

	user, err := u.UserRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// ตรวจสอบสิทธิ์ Super Admin
	if user.SystemRole != utils.RoleSuperAdmin {
		return nil, ErrInvalidCredentials
	}

	// ตรวจสอบสถานะบัญชี
	if user.Status != "active" {
		return nil, ErrInvalidCredentials
	}

	// ตรวจสอบ Password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// สร้าง JWT Token
	token, err := utils.GenerateJWT(
		u.Config.JWT.Secret,
		u.Config.JWT.ExpiresIn,
		user.ID,
		user.Email,
		user.SystemRole,
	)

	if err != nil {
		return nil, err
	}

	return &models.SuperAdminLoginData{
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
