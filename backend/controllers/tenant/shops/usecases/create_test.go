package usecases

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"strings"
	"testing"
)

type shopStub struct {
	called bool
	input  models.NewShop
	userID string
}

func (s *shopStub) Create(_ context.Context, id string, input models.NewShop) (models.Shop, error) {
	s.called = true
	s.input = input
	s.userID = id
	return models.Shop{ID: "shop-id", Name: input.Name, Role: "owner"}, nil
}
func TestCreateShopValidation(t *testing.T) {
	for _, tt := range []struct {
		name, shopName, businessType string
		valid                        bool
	}{
		{"valid", " Test Shop ", "salon", true},
		{"blank", " ", "salon", false},
		{"long name", strings.Repeat("ก", 151), "salon", false},
		{"unknown type", "Test", "invalid", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &shopStub{}
			id := "11111111-1111-4111-8111-111111111111"
			_, err := NewShopUsecase(repo).Create(context.Background(), id, models.CreateShopRequest{Name: tt.shopName, BusinessType: tt.businessType})
			if !tt.valid {
				if !errors.Is(err, ErrInvalidShop) || repo.called {
					t.Fatal("invalid shop reached database")
				}
				return
			}
			if err != nil || repo.userID != id || repo.input.Name != "Test Shop" || !strings.HasPrefix(repo.input.Slug, "shop-") {
				t.Fatal("invalid normalized shop data")
			}
		})
	}
}

func TestCreateShopHours(t *testing.T) {
	opening, closing, earlyClose := "09:00", "24:00", "08:00"
	for _, test := range []struct {
		name    string
		closing *string
		valid   bool
	}{
		{"valid hours", &closing, true},
		{"missing closing time", nil, false},
		{"closing before opening", &earlyClose, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &shopStub{}
			_, err := NewShopUsecase(repo).Create(context.Background(), "11111111-1111-4111-8111-111111111111", models.CreateShopRequest{
				Name: "Test Shop", BusinessType: "salon", OpeningTime: &opening, ClosingTime: test.closing,
			})
			if !test.valid {
				if !errors.Is(err, ErrInvalidShop) || repo.called {
					t.Fatal("invalid hours reached the repository")
				}
				return
			}
			if err != nil || !repo.called || repo.input.OpeningTime == nil || repo.input.ClosingTime == nil || *repo.input.OpeningTime != opening || *repo.input.ClosingTime != closing {
				t.Fatalf("repository did not receive the opening and closing times: %v", err)
			}
		})
	}
}
