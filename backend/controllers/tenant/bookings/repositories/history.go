package repositories

import (
	"context"
	"errors"

	"booking-saas/controllers/tenant/bookings/models"

	"github.com/jackc/pgx/v5"
)

func (r *BookingRepository) EditHistory(ctx context.Context, userID, shopID, bookingID string) ([]models.BookingEdit, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	_, _, role, err := authorize(ctx, tx, userID, shopID, false)
	if err != nil {
		return nil, err
	}
	var allowedBooking string
	err = tx.QueryRow(ctx, `SELECT b.id::text FROM bookings b
 JOIN tenant_members m ON m.tenant_id=b.tenant_id AND m.user_id=$3
 WHERE b.id=$1 AND b.tenant_id=$2 AND ($4='owner' OR b.staff_member_id=m.id)`, bookingID, shopID, userID, role).Scan(&allowedBooking)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT h.id::text,u.full_name,h.created_at,h.before_data,h.after_data
 FROM booking_edit_history h JOIN users u ON u.id=h.changed_by
 WHERE h.booking_id=$1 ORDER BY h.created_at DESC,h.id DESC LIMIT 50`, allowedBooking)
	if err != nil {
		return nil, err
	}
	history, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.BookingEdit, error) {
		var edit models.BookingEdit
		err := row.Scan(&edit.ID, &edit.EditorName, &edit.CreatedAt, &edit.Before, &edit.After)
		return edit, err
	})
	if err != nil {
		return nil, err
	}
	return history, tx.Commit(ctx)
}
