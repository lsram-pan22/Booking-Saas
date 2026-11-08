package usecases

import (
	"booking-saas/controllers/tenant/bookings/models"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestBookingCustomerValidation(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	valid := models.CustomerRequest{CustomerName: " ลูกค้า ", CustomerPhone: " 0812345678 ", Note: " หมายเหตุ "}
	for _, tc := range []struct {
		name   string
		change func(*models.CustomerRequest)
	}{
		{"blank name", func(r *models.CustomerRequest) { r.CustomerName = "  " }},
		{"long name", func(r *models.CustomerRequest) { r.CustomerName = strings.Repeat("ก", 151) }},
		{"invalid phone", func(r *models.CustomerRequest) { r.CustomerPhone = "abcdefghi" }},
		{"phone without digits", func(r *models.CustomerRequest) { r.CustomerPhone = "-------" }},
		{"short phone", func(r *models.CustomerRequest) { r.CustomerPhone = "123456" }},
		{"long phone", func(r *models.CustomerRequest) { r.CustomerPhone = "1234567890123456" }},
		{"long note", func(r *models.CustomerRequest) { r.Note = strings.Repeat("ก", 2001) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			req := valid
			tc.change(&req)
			_, err := NewBookingUsecase(repo).UpdateCustomer(context.Background(), id, id, id, req)
			if !errors.Is(err, ErrInvalidCustomer) || repo.calls != 0 {
				t.Fatal("invalid edit reached repository", err)
			}
		})
	}
	repo := &fakeRepo{}
	u := NewBookingUsecase(repo)
	if _, err := u.UpdateCustomer(context.Background(), id, id, id, valid); err != nil || repo.customer.CustomerName != "ลูกค้า" || repo.customer.CustomerPhone != "0812345678" || repo.customer.Note != "หมายเหตุ" {
		t.Fatal("customer edit not normalized", err)
	}
	valid.Note = "   "
	if _, err := u.UpdateCustomer(context.Background(), id, id, id, valid); err != nil || repo.customer.Note != "" {
		t.Fatal("cannot clear note", err)
	}
	if _, err := u.UpdateCustomer(context.Background(), id, id, "bad-id", valid); !errors.Is(err, ErrInvalidCustomer) {
		t.Fatal("invalid booking ID accepted")
	}
}
