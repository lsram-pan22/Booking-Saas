package usecases

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/utils"
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid booking request")
var ErrInvalidRemark = errors.New("cancellation remark required, maximum 1000 characters")

type BookingRepository interface {
	Update(context.Context, string, string, string, models.UpdateBookingRequest) (models.Booking, error)
	UpdateCustomer(context.Context, string, string, string, models.CustomerRequest) (models.Booking, error)
	List(context.Context, string, string, time.Time) (models.BookingList, error)
	ListWeek(context.Context, string, string, time.Time) (models.BookingList, error)
	Create(context.Context, string, string, models.CreateBookingRequest) (models.Booking, error)
	SetStatus(context.Context, string, string, string, string, string) (models.Booking, error)
}
type BookingUsecase struct{ Repo BookingRepository }

func NewBookingUsecase(repo BookingRepository) *BookingUsecase { return &BookingUsecase{Repo: repo} }

var bangkok = time.FixedZone("Asia/Bangkok", 7*60*60)
var phonePattern = regexp.MustCompile(`^[+0-9() -]{7,30}$`)

func (u *BookingUsecase) List(ctx context.Context, userID, shopID, date string) (models.BookingList, error) {
	return u.list(ctx, userID, shopID, date, false)
}
func (u *BookingUsecase) ListWeek(ctx context.Context, userID, shopID, date string) (models.BookingList, error) {
	return u.list(ctx, userID, shopID, date, true)
}
func (u *BookingUsecase) list(ctx context.Context, userID, shopID, date string, week bool) (models.BookingList, error) {
	day, err := time.ParseInLocation("2006-01-02", date, bangkok)
	if err != nil || day.Year() < 1 || day.Year() > 9998 || !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) {
		return models.BookingList{}, ErrInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if week {
		return u.Repo.ListWeek(ctx, userID, shopID, day)
	}
	return u.Repo.List(ctx, userID, shopID, day)
}
func (u *BookingUsecase) Create(ctx context.Context, userID, shopID string, req models.CreateBookingRequest) (models.Booking, error) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.CustomerPhone = strings.TrimSpace(req.CustomerPhone)
	req.Note = strings.TrimSpace(req.Note)
	digits := 0
	for _, c := range req.CustomerPhone {
		if c >= '0' && c <= '9' {
			digits++
		}
	}
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) || !utils.ValidUserID(req.ServiceID) || !utils.ValidUserID(req.StaffMemberID) || req.CustomerName == "" || utf8.RuneCountInString(req.CustomerName) > 150 || utf8.RuneCountInString(req.Note) > 2000 || !phonePattern.MatchString(req.CustomerPhone) || digits < 7 || digits > 15 || req.StartsAt.IsZero() || req.StartsAt.Before(time.Now()) || req.StartsAt.After(time.Now().AddDate(2, 0, 0)) || req.StartsAt.Second() != 0 || req.StartsAt.Nanosecond() != 0 {
		return models.Booking{}, ErrInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Create(ctx, userID, shopID, req)
}
func (u *BookingUsecase) SetStatus(ctx context.Context, userID, shopID, bookingID, status, remark string) (models.Booking, error) {
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) || !utils.ValidUserID(bookingID) || (status != "confirmed" && status != "completed" && status != "cancelled") {
		return models.Booking{}, ErrInvalid
	}
	remark = strings.TrimSpace(remark)
	if status == "cancelled" && (remark == "" || utf8.RuneCountInString(remark) > 1000) {
		return models.Booking{}, ErrInvalidRemark
	}
	if status != "cancelled" {
		remark = ""
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.SetStatus(ctx, userID, shopID, bookingID, status, remark)
}
