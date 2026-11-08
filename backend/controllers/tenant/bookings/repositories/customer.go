package repositories

import (
	"booking-saas/controllers/tenant/bookings/models"
	"context"

	"errors"
	"github.com/jackc/pgx/v5"
)

var ErrReadOnly = errors.New("only pending or confirmed bookings can be edited")

func lockManageableBooking(ctx context.Context, tx pgx.Tx, userID, shopID, bookingID string) (models.Booking, error) {
	return scan(tx.QueryRow(ctx, `SELECT `+columns+` FROM bookings WHERE id=$1 AND tenant_id=$2
 AND EXISTS(SELECT 1 FROM tenant_members actor WHERE actor.tenant_id=bookings.tenant_id AND actor.user_id=$3 AND actor.status='active'
 AND (actor.role='owner' OR (actor.role='staff' AND actor.id=bookings.staff_member_id))) FOR UPDATE`, bookingID, shopID, userID))
}

func (r *BookingRepository) UpdateCustomer(ctx context.Context, userID, shopID, bookingID string, req models.CustomerRequest) (models.Booking, error) {
	return r.Update(ctx, userID, shopID, bookingID, models.UpdateBookingRequest{CustomerRequest: req})
}
