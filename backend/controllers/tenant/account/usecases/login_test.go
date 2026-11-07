package usecases

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/repositories"
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

type loginStub struct {
	account models.Account
	profile models.Profile
	err     error
	email   string
}

func (s *loginStub) FindByIdentity(_ context.Context, email string) (models.Account, error) {
	s.email = email
	return s.account, s.err
}
func (s *loginStub) Profile(context.Context, string) (models.Profile, error) { return s.profile, s.err }
func TestTenantLogin(t *testing.T) {
	hash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, status, role, password string
		tenants                      bool
		repoErr                      error
		want                         bool
	}{
		{"owner", "active", "user", "correct-password", true, nil, true},
		{"wrong password", "active", "user", "incorrect", true, nil, false},
		{"suspended", "suspended", "user", "correct-password", true, nil, false},
		{"super admin", "active", "super_admin", "correct-password", true, nil, false},
		{"new account without shop", "active", "user", "correct-password", false, nil, true},
		{"not found", "active", "user", "correct-password", true, repositories.ErrAccountNotFound, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &loginStub{account: models.Account{User: models.TenantUser{ID: "11111111-1111-4111-8111-111111111111", Email: "owner@example.com"}, PasswordHash: hash, Status: tt.status, SystemRole: tt.role}, err: tt.repoErr}
			if tt.tenants {
				repo.profile = models.Profile{User: repo.account.User, Tenants: []models.Tenant{{ID: "tenant-id", Role: "owner", Status: "pending"}}}
			}
			cfg := &config.Config{JWT: config.JWTConfig{Secret: "tenant-test-secret", ExpiresIn: time.Hour}}
			data, err := NewLoginUsecase(repo, cfg).Login(context.Background(), models.LoginRequest{Email: " Owner@Example.com ", Password: tt.password})
			if !tt.want {
				if !errors.Is(err, ErrInvalidCredentials) || data.AccessToken != "" {
					t.Fatal("unauthorized login accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			claims := &utils.JWTClaims{}
			_, err = jwt.ParseWithClaims(data.AccessToken, claims, func(_ *jwt.Token) (interface{}, error) { return []byte(cfg.JWT.Secret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("booking-saas"))
			if err != nil || claims.Role != "user" || claims.UserID != "11111111-1111-4111-8111-111111111111" || data.ExpiresIn != 3600 || repo.email != "owner@example.com" {
				t.Fatal("incorrect tenant session")
			}
		})
	}
}
