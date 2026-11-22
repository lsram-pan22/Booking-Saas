package usecases

import (
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"testing"
)

type passwordStub struct {
	called bool
	hash   string
}

func (s *passwordStub) Reset(_ context.Context, _ models.PasswordActor, _, _, hash, _ string) error {
	s.called = true
	s.hash = hash
	return nil
}
func (s *passwordStub) Change(_ context.Context, _ models.PasswordActor, _, _, hash string) error {
	s.called = true
	s.hash = hash
	return nil
}
func (s *passwordStub) Owners(context.Context, string) ([]models.ResettableOwner, error) {
	return nil, nil
}
func TestResetPasswordValidationAndHash(t *testing.T) {
	actor := models.PasswordActor{ID: "11111111-1111-4111-8111-111111111111", Role: "user"}
	target := "22222222-2222-4222-8222-222222222222"
	for _, password := range []string{"short", strings.Repeat("a", 73), strings.Repeat("ก", 25), strings.Repeat(" ", 12)} {
		repo := &passwordStub{}
		err := NewPasswordUsecase(repo).Reset(context.Background(), actor, target, target, models.ResetPasswordRequest{CurrentPassword: "owner-password", Password: password})
		if !errors.Is(err, ErrInvalidPassword) || repo.called {
			t.Fatal("invalid password reached storage")
		}
	}
	repo := &passwordStub{}
	if err := NewPasswordUsecase(repo).Reset(context.Background(), actor, target, actor.ID, models.ResetPasswordRequest{CurrentPassword: "owner-password", Password: "temporary-password"}); !errors.Is(err, ErrInvalidPassword) || repo.called {
		t.Fatal("self reset accepted")
	}
	if err := NewPasswordUsecase(repo).Reset(context.Background(), actor, target, target, models.ResetPasswordRequest{CurrentPassword: "owner-password", Password: "temporary-password"}); err != nil || !utils.CheckPassword("temporary-password", repo.hash) {
		t.Fatalf("hash not stored: %v", err)
	}
}
func TestChangeCannotReuseTemporaryPassword(t *testing.T) {
	repo := &passwordStub{}
	err := NewPasswordUsecase(repo).Change(context.Background(), models.PasswordActor{ID: "11111111-1111-4111-8111-111111111111", Role: "user"}, models.ChangePasswordRequest{CurrentPassword: "temporary-password", NewPassword: "temporary-password"})
	if !errors.Is(err, ErrCurrentPassword) || repo.called {
		t.Fatal("temporary password reused")
	}
}

func TestResetRequiresActorPassword(t *testing.T) {
	repo := &passwordStub{}
	err := NewPasswordUsecase(repo).Reset(context.Background(), models.PasswordActor{ID: "11111111-1111-4111-8111-111111111111", Role: "user"}, "33333333-3333-4333-8333-333333333333", "22222222-2222-4222-8222-222222222222", models.ResetPasswordRequest{Password: "temporary-password"})
	if !errors.Is(err, ErrInvalidPassword) || repo.called {
		t.Fatal("reset accepted without actor confirmation")
	}
}
