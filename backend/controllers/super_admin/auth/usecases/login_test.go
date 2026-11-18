package usecases

import (
	"booking-saas/config"
	"booking-saas/controllers/super_admin/auth/models"
	"booking-saas/utils"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

type adminLoginStub struct {
	user     *models.SuperAdminAccount
	identity string
	called   bool
}

func (s *adminLoginStub) FindByUsername(_ context.Context, value string) (*models.SuperAdminAccount, error) {
	s.called = true
	s.identity = value
	return s.user, nil
}
func TestAdminLoginUsesCurrentSessionVersion(t *testing.T) {
	hash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	repo := &adminLoginStub{user: &models.SuperAdminAccount{ID: "11111111-1111-4111-8111-111111111111", Email: "admin@example.test", PasswordHash: hash, Status: "active", SystemRole: "super_admin", SessionVersion: 7}}
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "admin-login-test-secret", ExpiresIn: time.Hour}}
	data, err := NewLoginUsecase(repo, cfg).Login(context.Background(), models.SuperAdminLoginRequest{Username: " Admin ", Password: "correct-password"})
	if err != nil {
		t.Fatal(err)
	}
	var claims utils.JWTClaims
	if _, err = jwt.ParseWithClaims(data.AccessToken, &claims, func(*jwt.Token) (any, error) { return []byte(cfg.JWT.Secret), nil }); err != nil {
		t.Fatal(err)
	}
	if claims.SessionVersion != 7 || data.ExpiresIn != 3600 || repo.identity != "admin" {
		t.Fatal("admin session did not refresh version/expiry")
	}
	repo.user.SystemRole = "user"
	if _, err = NewLoginUsecase(repo, cfg).Login(context.Background(), models.SuperAdminLoginRequest{Username: "admin", Password: "correct-password"}); err == nil {
		t.Fatal("tenant account became super admin")
	}
}
