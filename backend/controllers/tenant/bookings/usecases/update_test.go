package usecases

import (
	"booking-saas/controllers/tenant/bookings/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestBookingUpdateValidation(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	serviceID := id
	start := time.Now().Add(time.Hour).Truncate(time.Minute)
	valid := models.UpdateBookingRequest{CustomerRequest: models.CustomerRequest{CustomerName: " ลูกค้า ", CustomerPhone: " 0812345678 ", Note: " note "}, ServiceID: &serviceID, StartsAt: &start}
	repo := &fakeRepo{}
	if _, err := NewBookingUsecase(repo).Update(context.Background(), id, id, id, valid); err != nil || repo.update.CustomerName != "ลูกค้า" || repo.update.Note != "note" {
		t.Fatal("valid edit not normalized", err)
	}
	badID := "not-a-uuid"
	invalidTime := start.Add(time.Second)
	for _, req := range []models.UpdateBookingRequest{
		{CustomerRequest: valid.CustomerRequest, ServiceID: &badID},
		{CustomerRequest: valid.CustomerRequest, StartsAt: &invalidTime},
		{CustomerRequest: models.CustomerRequest{CustomerName: " ", CustomerPhone: "0812345678"}},
	} {
		repo := &fakeRepo{}
		_, err := NewBookingUsecase(repo).Update(context.Background(), id, id, id, req)
		if err == nil || repo.calls != 0 {
			t.Fatal("invalid edit reached SQL", err)
		}
	}
	if _, err := NewBookingUsecase(&fakeRepo{}).Update(context.Background(), id, id, badID, valid); !errors.Is(err, ErrInvalidCustomer) {
		t.Fatal("bad booking ID accepted")
	}
}
