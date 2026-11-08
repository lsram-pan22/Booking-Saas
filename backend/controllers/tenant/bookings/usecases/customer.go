package usecases

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/utils"
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidCustomer = errors.New("invalid booking customer details")

func (u *BookingUsecase) UpdateCustomer(ctx context.Context, userID, shopID, bookingID string, req models.CustomerRequest) (models.Booking, error) {
	var valid bool
	req, valid = normalizeCustomer(req)
	if !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) || !utils.ValidUserID(bookingID) || !valid {
		return models.Booking{}, ErrInvalidCustomer
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.UpdateCustomer(ctx, userID, shopID, bookingID, req)
}

func normalizeCustomer(req models.CustomerRequest) (models.CustomerRequest, bool) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.CustomerPhone = strings.TrimSpace(req.CustomerPhone)
	req.Note = strings.TrimSpace(req.Note)
	digits := 0
	for _, c := range req.CustomerPhone {
		if c >= '0' && c <= '9' {
			digits++
		}
	}
	valid := req.CustomerName != "" && utf8.RuneCountInString(req.CustomerName) <= 150 && phonePattern.MatchString(req.CustomerPhone) && digits >= 7 && digits <= 15 && utf8.RuneCountInString(req.Note) <= 2000
	return req, valid
}
