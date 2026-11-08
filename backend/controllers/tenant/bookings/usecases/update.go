package usecases

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/utils"
	"context"
	"time"
)

func (u *BookingUsecase) Update(ctx context.Context, userID, shopID, bookingID string, req models.UpdateBookingRequest) (models.Booking, error) {
	var valid bool
	req.CustomerRequest, valid = normalizeCustomer(req.CustomerRequest)
	if !valid || !utils.ValidUserID(userID) || !utils.ValidUserID(shopID) || !utils.ValidUserID(bookingID) {
		return models.Booking{}, ErrInvalidCustomer
	}
	if req.ServiceID != nil && !utils.ValidUserID(*req.ServiceID) {
		return models.Booking{}, ErrInvalid
	}
	if req.StartsAt != nil && (req.StartsAt.IsZero() || req.StartsAt.Second() != 0 || req.StartsAt.Nanosecond() != 0) {
		return models.Booking{}, ErrInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return u.Repo.Update(ctx, userID, shopID, bookingID, req)
}
