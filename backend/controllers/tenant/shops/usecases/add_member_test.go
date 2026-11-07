package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"errors"
	"testing"
)

type addMemberStub struct {
	called bool
	input  models.NewMember
}

func (s *addMemberStub) AddMember(_ context.Context, _, _ string, input models.NewMember) (models.ShopMember, error) {
	s.called = true
	s.input = input
	return models.ShopMember{Role: "staff"}, nil
}
func TestAddMemberValidation(t *testing.T) {
	for _, tt := range []struct {
		name, mode, email, password string
		valid                       bool
	}{
		{"new", "new", " Staff@Example.com ", "staff-password-123", true},
		{"existing without password", "existing", "existing@example.com", "", true},
		{"invalid email", "existing", "invalid", "", false},
		{"short password", "new", "staff@example.com", "short", false},
		{"invalid mode", "owner", "staff@example.com", "staff-password-123", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &addMemberStub{}
			_, err := NewAddMemberUsecase(repo).Add(context.Background(), "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222", models.AddMemberRequest{Username: " Staff01 ", Mode: tt.mode, Email: tt.email, Password: tt.password, FullName: " Staff ", Phone: "081-234-5678"})
			if !tt.valid {
				if !errors.Is(err, ErrInvalidMember) || repo.called {
					t.Fatal("invalid member reached repository")
				}
				return
			}
			if err != nil || !repo.called {
				t.Fatal(err)
			}
			if tt.mode == "new" && (!utils.CheckPassword(tt.password, repo.input.PasswordHash) || repo.input.Email != "staff@example.com" || repo.input.Phone != "0812345678") {
				t.Fatal("new account was not normalized and hashed")
			}
			if tt.mode == "existing" && (repo.input.PasswordHash != "" || repo.input.FullName != "") {
				t.Fatal("existing account update data must be discarded")
			}
		})
	}
}
