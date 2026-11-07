package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/controllers/tenant/account/repositories"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"testing"
)

type registrationStub struct {
	input  models.Registration
	called bool
	err    error
}

func (r *registrationStub) Register(_ context.Context, input models.Registration) (models.RegisterData, error) {
	r.called = true
	r.input = input
	return models.RegisterData{UserID: "11111111-1111-4111-8111-111111111111"}, r.err
}
func validRequest() models.RegisterRequest {
	return models.RegisterRequest{FullName: " Owner ", Email: " Owner@Example.com ", Phone: "081-234-5678", Password: "a-long-password"}
}
func TestRegisterHashesAndNormalizes(t *testing.T) {
	repo := &registrationStub{}
	data, err := NewRegisterUsecase(repo).Register(context.Background(), validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if data.UserID != "11111111-1111-4111-8111-111111111111" || repo.input.Email != "owner@example.com" || repo.input.Phone != "0812345678" || repo.input.FullName != "Owner" {
		t.Fatal("unexpected registration data")
	}
	if !utils.CheckPassword(validRequest().Password, repo.input.PasswordHash) {
		t.Fatal("password was not hashed correctly")
	}
	if repo.input.Username != repo.input.Email {
		t.Fatal("username must equal normalized email")
	}
}

func TestRegisterSupportsEmailLongerThanOldUsernameLimit(t *testing.T) {
	repo := &registrationStub{}
	req := validRequest()
	req.Email = strings.Repeat("a", 60) + "@example.com"
	_, err := NewRegisterUsecase(repo).Register(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if repo.input.Username != req.Email {
		t.Fatal("email was truncated or changed")
	}
}
func TestInvalidRegistrationNeverWrites(t *testing.T) {
	tests := []struct {
		name   string
		change func(*models.RegisterRequest)
	}{
		{"email", func(r *models.RegisterRequest) { r.Email = "invalid" }},
		{"name", func(r *models.RegisterRequest) { r.FullName = " " }},
		{"phone", func(r *models.RegisterRequest) { r.Phone = "123" }},
		{"short password", func(r *models.RegisterRequest) { r.Password = "short" }},
		{"bcrypt limit", func(r *models.RegisterRequest) { r.Password = strings.Repeat("ก", 25) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &registrationStub{}
			req := validRequest()
			tt.change(&req)
			_, err := NewRegisterUsecase(repo).Register(context.Background(), req)
			if !errors.Is(err, ErrInvalidRegistration) || repo.called {
				t.Fatal("invalid request reached repository")
			}
		})
	}
}
func TestDuplicateAccountPropagates(t *testing.T) {
	repo := &registrationStub{err: repositories.ErrAccountExists}
	_, err := NewRegisterUsecase(repo).Register(context.Background(), validRequest())
	if !errors.Is(err, repositories.ErrAccountExists) {
		t.Fatal("duplicate account error lost")
	}
}
