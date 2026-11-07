package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"context"
	"errors"
	"testing"
)

type settingsStub struct {
	id     string
	input  models.UpdateProfileRequest
	called bool
}

func (s *settingsStub) Update(_ context.Context, id string, req models.UpdateProfileRequest) (models.TenantUser, error) {
	s.id = id
	s.input = req
	s.called = true
	return models.TenantUser{ID: id, FullName: req.FullName, Phone: req.Phone}, nil
}
func TestUpdateProfileValidation(t *testing.T) {
	for _, tt := range []struct {
		name, fullName, phone string
		valid                 bool
	}{
		{"valid", " Owner ", "081-234-5678", true},
		{"blank name", " ", "0812345678", false},
		{"invalid phone", "Owner", "123", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &settingsStub{}
			_, err := NewSettingsUsecase(repo).Update(context.Background(), "11111111-1111-4111-8111-111111111111", models.UpdateProfileRequest{FullName: tt.fullName, Phone: tt.phone})
			if !tt.valid {
				if !errors.Is(err, ErrInvalidRegistration) || repo.called {
					t.Fatal("invalid update reached repository")
				}
				return
			}
			if err != nil || repo.input.FullName != "Owner" || repo.input.Phone != "0812345678" || repo.id != "11111111-1111-4111-8111-111111111111" {
				t.Fatal("incorrect profile update")
			}
		})
	}
}
