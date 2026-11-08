package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/controllers/tenant/shops/repositories"
	"context"
	"errors"
	"testing"
)

type manageStub struct {
	shop    models.Shop
	err     error
	updated bool
	input   models.UpdateShopRequest
}

func (s *manageStub) Get(context.Context, string, string) (models.Shop, error) { return s.shop, s.err }
func (s *manageStub) Update(_ context.Context, _, _ string, req models.UpdateShopRequest) (models.Shop, error) {
	s.updated = true
	s.input = req
	return s.shop, nil
}
func TestManageShopPermissions(t *testing.T) {
	for _, tt := range []struct {
		name, role, status string
		repoErr, wantErr   error
	}{
		{"owner", "owner", "active", nil, nil},
		{"inactive owner can reopen", "owner", "inactive", nil, nil},
		{"suspended shop cannot reopen", "owner", "suspended", nil, ErrShopReadOnly},
		{"staff", "staff", "active", nil, ErrShopReadOnly},
		{"pending", "owner", "pending", nil, ErrShopReadOnly},
		{"stranger", "", "", repositories.ErrShopNotFound, repositories.ErrShopNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &manageStub{shop: models.Shop{Role: tt.role, Status: tt.status}, err: tt.repoErr}
			_, err := NewManageUsecase(repo).Update(context.Background(), "11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222", models.UpdateShopRequest{Name: " New Name ", BusinessType: "salon"})
			if !errors.Is(err, tt.wantErr) || repo.updated != (tt.wantErr == nil) {
				t.Fatal("incorrect shop update authorization")
			}
			if repo.updated && repo.input.Name != "New Name" {
				t.Fatal("name not normalized")
			}
		})
	}
}

func TestManageShopHours(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	for _, tc := range []struct {
		open, close string
		valid       bool
	}{{"09:00", "18:00", true}, {"00:00", "24:00", true}, {"18:00", "09:00", false}, {"9:00", "18:00", false}} {
		repo := &manageStub{shop: models.Shop{Role: "owner", Status: "active"}}
		req := models.UpdateShopRequest{Name: "Shop", BusinessType: "salon", OpeningTime: &tc.open, ClosingTime: &tc.close}
		_, err := NewManageUsecase(repo).Update(context.Background(), id, id, req)
		if (err == nil) != tc.valid || repo.updated != tc.valid {
			t.Errorf("hours %s–%s: %v", tc.open, tc.close, err)
		}
	}
	open := "09:00"
	repo := &manageStub{}
	_, err := NewManageUsecase(repo).Update(context.Background(), id, id, models.UpdateShopRequest{Name: "Shop", BusinessType: "salon", OpeningTime: &open})
	if !errors.Is(err, ErrInvalidShop) || repo.updated {
		t.Fatal("partial hours update accepted")
	}
}
