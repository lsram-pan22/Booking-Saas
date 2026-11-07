package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"testing"
)

type memberStatusStub struct {
	called bool
	status string
}

func (s *memberStatusStub) SetMemberStatus(_ context.Context, _, _, _, status string) (models.MemberStatusResult, error) {
	s.called = true
	s.status = status
	return models.MemberStatusResult{}, nil
}
func TestMemberStatusValidation(t *testing.T) {
	for _, status := range []string{"active", "inactive", "suspended", "owner", ""} {
		t.Run(status, func(t *testing.T) {
			repo := &memberStatusStub{}
			_, err := NewMemberStatusUsecase(repo).Set(context.Background(), "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222", "33333333-3333-4333-8333-333333333333", models.MemberStatusRequest{Status: status})
			valid := status == "active" || status == "inactive"
			if repo.called != valid || (!valid && !errors.Is(err, ErrInvalidMember)) || (valid && err != nil) {
				t.Fatal("unexpected membership status validation")
			}
		})
	}
}
