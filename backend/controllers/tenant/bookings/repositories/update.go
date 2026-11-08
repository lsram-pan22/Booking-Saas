package repositories

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

var ErrInvalidSchedule = errors.New("new booking time must be in the future within two years")

func editSnapshot(b models.Booking) models.BookingEditSnapshot {
	return models.BookingEditSnapshot{CustomerRequest: models.CustomerRequest{CustomerName: b.CustomerName, CustomerPhone: b.CustomerPhone, Note: b.Note}, ServiceID: b.ServiceID, ServiceName: b.ServiceName, DurationMinutes: b.DurationMinutes, PriceSatang: b.PriceSatang, StartsAt: b.StartsAt, EndsAt: b.EndsAt, Status: b.Status}
}

func (r *BookingRepository) Update(ctx context.Context, userID, shopID, bookingID string, req models.UpdateBookingRequest) (models.Booking, error) {
	var empty models.Booking
	tx, err := r.beginWrite(ctx, userID, shopID)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(context.Background())
	before, err := lockManageableBooking(ctx, tx, userID, shopID, bookingID)
	if err != nil {
		return empty, err
	}
	if before.Status != "pending" && before.Status != "confirmed" {
		return empty, ErrReadOnly
	}
	after := before
	after.CustomerName = req.CustomerName
	after.CustomerPhone = req.CustomerPhone
	after.Note = req.Note
	serviceChanged := req.ServiceID != nil && *req.ServiceID != before.ServiceID
	timeChanged := req.StartsAt != nil && !req.StartsAt.Equal(before.StartsAt)
	if serviceChanged {
		after.ServiceID = *req.ServiceID
		err = tx.QueryRow(ctx, `SELECT name,duration_minutes,price_satang FROM services WHERE id=$1 AND tenant_id=$2 AND status='active' FOR SHARE`, after.ServiceID, shopID).Scan(&after.ServiceName, &after.DurationMinutes, &after.PriceSatang)
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrUnavailable
		}
		if err != nil {
			return empty, err
		}
	}
	if timeChanged {
		if req.StartsAt.Before(time.Now()) || req.StartsAt.After(time.Now().AddDate(2, 0, 0)) {
			return empty, ErrInvalidSchedule
		}
		after.StartsAt = *req.StartsAt
	}
	if serviceChanged || timeChanged {
		after.EndsAt = after.StartsAt.Add(time.Duration(after.DurationMinutes) * time.Minute)
		var available bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.id=$1 AND m.tenant_id=$2 AND m.status='active' AND m.role IN ('owner','staff') AND u.status='active' AND u.system_role='user')`, after.StaffMemberID, shopID).Scan(&available); err != nil {
			return empty, err
		}
		if !available {
			return empty, ErrUnavailable
		}
		var opening, closing *string
		if err = tx.QueryRow(ctx, `SELECT opening_time,closing_time FROM tenants WHERE id=$1`, shopID).Scan(&opening, &closing); err != nil {
			return empty, err
		}
		if opening != nil && closing != nil && !utils.WithinShopHours(after.StartsAt, after.EndsAt, *opening, *closing) {
			return empty, ErrOutsideHours
		}
		var overlap bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings WHERE tenant_id=$1 AND staff_member_id=$2 AND id<>$5 AND status<>'cancelled' AND starts_at<$4 AND ends_at>$3)`, shopID, after.StaffMemberID, after.StartsAt, after.EndsAt, bookingID).Scan(&overlap); err != nil {
			return empty, err
		}
		if overlap {
			return empty, ErrConflict
		}
	}
	if !serviceChanged && !timeChanged && before.CustomerName == after.CustomerName && before.CustomerPhone == after.CustomerPhone && before.Note == after.Note {
		before.CanManage = true
		return before, tx.Commit(ctx)
	}
	updated, err := scan(tx.QueryRow(ctx, `UPDATE bookings SET customer_name=$3,customer_phone=$4,note=$5,service_id=$6,service_name=$7,duration_minutes=$8,price_satang=$9,starts_at=$10,ends_at=$11,updated_at=NOW() WHERE id=$1 AND tenant_id=$2 RETURNING `+columns, bookingID, shopID, after.CustomerName, after.CustomerPhone, after.Note, after.ServiceID, after.ServiceName, after.DurationMinutes, after.PriceSatang, after.StartsAt, after.EndsAt))
	if err != nil {
		return empty, err
	}
	beforeJSON, err := json.Marshal(editSnapshot(before))
	if err != nil {
		return empty, err
	}
	afterJSON, err := json.Marshal(editSnapshot(updated))
	if err != nil {
		return empty, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO booking_edit_history(booking_id,changed_by,before_data,after_data) VALUES($1,$2,$3::jsonb,$4::jsonb)`, bookingID, userID, string(beforeJSON), string(afterJSON)); err != nil {
		return empty, err
	}
	if err = tx.Commit(ctx); err != nil {
		return empty, err
	}
	updated.CanManage = true
	return updated, nil
}
