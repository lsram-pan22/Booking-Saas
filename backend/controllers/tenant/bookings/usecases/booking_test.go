package usecases

import (
	"booking-saas/controllers/tenant/bookings/models"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeRepo struct {
	update   models.UpdateBookingRequest
	calls    int
	req      models.CreateBookingRequest
	day      time.Time
	remark   string
	customer models.CustomerRequest
}

func (r *fakeRepo) Update(_ context.Context, _, _, _ string, req models.UpdateBookingRequest) (models.Booking, error) {
	r.calls++
	r.update = req
	return models.Booking{}, nil
}

func (r *fakeRepo) UpdateCustomer(_ context.Context, _, _, _ string, req models.CustomerRequest) (models.Booking, error) {
	r.calls++
	r.customer = req
	return models.Booking{}, nil
}

func (r *fakeRepo) List(_ context.Context, _, _ string, day time.Time) (models.BookingList, error) {
	r.calls++
	r.day = day
	return models.BookingList{}, nil
}
func (r *fakeRepo) ListWeek(ctx context.Context, userID, shopID string, day time.Time) (models.BookingList, error) {
	return r.List(ctx, userID, shopID, day)
}
func (r *fakeRepo) Create(_ context.Context, _, _ string, req models.CreateBookingRequest) (models.Booking, error) {
	r.calls++
	r.req = req
	return models.Booking{}, nil
}
func (r *fakeRepo) SetStatus(_ context.Context, _, _, _, _, remark string) (models.Booking, error) {
	r.calls++
	r.remark = remark
	return models.Booking{}, nil
}
func TestBookingValidation(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	valid := models.CreateBookingRequest{ServiceID: id, StaffMemberID: id, CustomerName: " ลูกค้า ", CustomerPhone: " 0812345678 ", Note: " รายละเอียด ", StartsAt: time.Now().Add(time.Hour).Truncate(time.Minute)}
	for _, tc := range []struct {
		name   string
		change func(*models.CreateBookingRequest)
	}{
		{"service ID", func(r *models.CreateBookingRequest) { r.ServiceID = "bad" }},
		{"staff ID", func(r *models.CreateBookingRequest) { r.StaffMemberID = "bad" }},
		{"blank name", func(r *models.CreateBookingRequest) { r.CustomerName = " " }},
		{"invalid phone", func(r *models.CreateBookingRequest) { r.CustomerPhone = "abcdefghi" }},
		{"phone without digits", func(r *models.CreateBookingRequest) { r.CustomerPhone = "-------" }},
		{"past date", func(r *models.CreateBookingRequest) { r.StartsAt = time.Now().Add(-time.Hour) }},
		{"too far ahead", func(r *models.CreateBookingRequest) { r.StartsAt = time.Now().AddDate(3, 0, 0) }},
		{"seconds", func(r *models.CreateBookingRequest) { r.StartsAt = r.StartsAt.Add(time.Second) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			req := valid
			tc.change(&req)
			_, err := NewBookingUsecase(repo).Create(context.Background(), id, id, req)
			if !errors.Is(err, ErrInvalid) || repo.calls != 0 {
				t.Fatal("invalid request reached repository", err)
			}
		})
	}
	repo := &fakeRepo{}
	u := NewBookingUsecase(repo)
	if _, err := u.Create(context.Background(), id, id, valid); err != nil || repo.req.CustomerName != "ลูกค้า" || repo.req.CustomerPhone != "0812345678" || repo.req.Note != "รายละเอียด" {
		t.Fatal("valid request not normalized", err)
	}
	if _, err := u.List(context.Background(), id, id, "2026-11-07"); err != nil || repo.day.UTC().Format(time.RFC3339) != "2026-11-06T17:00:00Z" {
		t.Fatal("Thai day boundary incorrect", err, repo.day)
	}
	if _, err := u.List(context.Background(), id, id, "2026-02-30"); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid date accepted")
	}
	if _, err := u.SetStatus(context.Background(), id, id, id, "pending", ""); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid transition target accepted")
	}
}

func TestCancellationRemark(t *testing.T) {
	const id = "11111111-1111-4111-8111-111111111111"
	for _, remark := range []string{"", " \n\t ", strings.Repeat("ก", 1001)} {
		repo := &fakeRepo{}
		_, err := NewBookingUsecase(repo).SetStatus(context.Background(), id, id, id, "cancelled", remark)
		if !errors.Is(err, ErrInvalidRemark) || repo.calls != 0 {
			t.Fatal("invalid reason reached repository", err)
		}
	}
	repo := &fakeRepo{}
	u := NewBookingUsecase(repo)
	if _, err := u.SetStatus(context.Background(), id, id, id, "cancelled", "  ลูกค้าขอยกเลิก  "); err != nil || repo.remark != "ลูกค้าขอยกเลิก" {
		t.Fatal("remark not normalized", err)
	}
	if _, err := u.SetStatus(context.Background(), id, id, id, "confirmed", "ignored"); err != nil || repo.remark != "" {
		t.Fatal("confirmation saved cancellation remark", err)
	}
}
