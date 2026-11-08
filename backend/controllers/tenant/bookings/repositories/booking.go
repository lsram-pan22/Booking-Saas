package repositories

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrNotFound     = errors.New("booking or shop not found or inaccessible")
	ErrUnavailable  = errors.New("service or staff unavailable")
	ErrConflict     = errors.New("staff booking overlaps")
	ErrTransition   = errors.New("invalid booking status transition")
	ErrOutsideHours = errors.New("booking outside shop opening hours")
)

type BookingRepository struct{ DB *pgxpool.Pool }

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository { return &BookingRepository{DB: db} }

const columns = `id::text,tenant_id::text,service_id::text,staff_member_id::text,customer_name,customer_phone,note,service_name,staff_name,duration_minutes,price_satang,starts_at,ends_at,status,remark`

func scan(row pgx.Row) (models.Booking, error) {
	var b models.Booking
	err := row.Scan(&b.ID, &b.TenantID, &b.ServiceID, &b.StaffMemberID, &b.CustomerName, &b.CustomerPhone, &b.Note, &b.ServiceName, &b.StaffName, &b.DurationMinutes, &b.PriceSatang, &b.StartsAt, &b.EndsAt, &b.Status, &b.Remark)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}
func authorize(ctx context.Context, tx pgx.Tx, userID, shopID string, write bool) (string, string, string, error) {
	var name, status, role string
	err := tx.QueryRow(ctx, `SELECT t.name,t.status,m.role FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id
 WHERE t.id=$1 AND m.user_id=$2 AND m.status='active' AND m.role IN ('owner','staff') AND u.status='active' AND u.system_role='user'
 AND (t.status='active' OR (NOT $3 AND m.role='owner' AND t.status='inactive'))`, shopID, userID, write).Scan(&name, &status, &role)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return name, status, role, err
}
func (r *BookingRepository) List(ctx context.Context, userID, shopID string, day time.Time) (models.BookingList, error) {
	return r.list(ctx, userID, shopID, day, 1)
}
func (r *BookingRepository) ListWeek(ctx context.Context, userID, shopID string, day time.Time) (models.BookingList, error) {
	return r.list(ctx, userID, shopID, day, 7)
}
func (r *BookingRepository) list(ctx context.Context, userID, shopID string, day time.Time, days int) (models.BookingList, error) {
	result := models.BookingList{Items: []models.Booking{}, Services: []models.ServiceOption{}, Staff: []models.StaffOption{}}
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.Background())
	result.ShopName, result.ShopStatus, result.Role, err = authorize(ctx, tx, userID, shopID, false)
	if err != nil {
		return result, err
	}
	var memberID string
	if err = tx.QueryRow(ctx, `SELECT t.opening_time,t.closing_time,m.id::text FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id WHERE t.id=$1 AND m.user_id=$2`, shopID, userID).Scan(&result.OpeningTime, &result.ClosingTime, &memberID); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT `+columns+` FROM bookings WHERE tenant_id=$1 AND starts_at<$3 AND ends_at>$2 ORDER BY starts_at,id`, shopID, day, day.AddDate(0, 0, days))
	if err != nil {
		return result, err
	}
	for rows.Next() {
		b, e := scan(rows)
		if e != nil {
			rows.Close()
			return result, e
		}
		b.CanManage = result.ShopStatus == "active" && (result.Role == "owner" || b.StaffMemberID == memberID)
		result.Items = append(result.Items, b)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	rows, err = tx.Query(ctx, `SELECT id::text,name,duration_minutes,price_satang FROM services WHERE tenant_id=$1 AND status='active' ORDER BY name,id`, shopID)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var s models.ServiceOption
		if err = rows.Scan(&s.ID, &s.Name, &s.DurationMinutes, &s.PriceSatang); err != nil {
			rows.Close()
			return result, err
		}
		result.Services = append(result.Services, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	rows, err = tx.Query(ctx, `SELECT m.id::text,u.full_name,m.role FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=$1 AND m.status='active' AND m.role IN ('owner','staff') AND u.status='active' AND u.system_role='user' AND ($2='owner' OR m.user_id=$3) ORDER BY u.full_name,m.id`, shopID, result.Role, userID)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var s models.StaffOption
		if err = rows.Scan(&s.ID, &s.Name, &s.Role); err != nil {
			rows.Close()
			return result, err
		}
		result.Staff = append(result.Staff, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	err = tx.Commit(ctx)
	return result, err
}

// Every booking write locks its tenant first. A waiting writer checks overlaps
// using a fresh READ COMMITTED snapshot after the previous writer commits.
func (r *BookingRepository) beginWrite(ctx context.Context, userID, shopID string) (pgx.Tx, error) {
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM tenants WHERE id=$1 FOR UPDATE`, shopID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	if err == nil {
		_, _, _, err = authorize(ctx, tx, userID, shopID, true)
	}
	if err != nil {
		tx.Rollback(context.Background())
		return nil, err
	}
	return tx, nil
}
func (r *BookingRepository) Create(ctx context.Context, userID, shopID string, req models.CreateBookingRequest) (models.Booking, error) {
	var empty models.Booking
	tx, err := r.beginWrite(ctx, userID, shopID)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(context.Background())
	var serviceName, staffName string
	var duration int
	var price int64
	err = tx.QueryRow(ctx, `SELECT name,duration_minutes,price_satang FROM services WHERE id=$1 AND tenant_id=$2 AND status='active' FOR SHARE`, req.ServiceID, shopID).Scan(&serviceName, &duration, &price)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrUnavailable
	}
	if err != nil {
		return empty, err
	}
	err = tx.QueryRow(ctx, `SELECT u.full_name FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.id=$1 AND m.tenant_id=$2 AND m.role IN ('owner','staff') AND m.status='active' AND u.status='active' AND u.system_role='user'
 AND (m.user_id=$3 OR EXISTS(SELECT 1 FROM tenant_members actor WHERE actor.tenant_id=m.tenant_id AND actor.user_id=$3 AND actor.role='owner' AND actor.status='active')) FOR SHARE OF m,u`, req.StaffMemberID, shopID, userID).Scan(&staffName)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrUnavailable
	}
	if err != nil {
		return empty, err
	}
	end := req.StartsAt.Add(time.Duration(duration) * time.Minute)
	var opening, closing *string
	if err = tx.QueryRow(ctx, `SELECT opening_time,closing_time FROM tenants WHERE id=$1`, shopID).Scan(&opening, &closing); err != nil {
		return empty, err
	}
	if opening != nil && closing != nil && !utils.WithinShopHours(req.StartsAt, end, *opening, *closing) {
		return empty, ErrOutsideHours
	}
	var overlap bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings WHERE tenant_id=$1 AND staff_member_id=$2 AND status<>'cancelled' AND starts_at<$4 AND ends_at>$3)`, shopID, req.StaffMemberID, req.StartsAt, end).Scan(&overlap)
	if err != nil {
		return empty, err
	}
	if overlap {
		return empty, ErrConflict
	}
	b, err := scan(tx.QueryRow(ctx, `INSERT INTO bookings(tenant_id,service_id,staff_member_id,customer_name,customer_phone,note,service_name,staff_name,duration_minutes,price_satang,starts_at,ends_at,created_by)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING `+columns, shopID, req.ServiceID, req.StaffMemberID, req.CustomerName, req.CustomerPhone, req.Note, serviceName, staffName, duration, price, req.StartsAt, end, userID))
	if err != nil {
		return empty, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO booking_status_history(booking_id,to_status,changed_by) VALUES($1,'pending',$2)`, b.ID, userID); err != nil {
		return empty, err
	}
	if err = tx.Commit(ctx); err != nil {
		return empty, err
	}
	b.CanManage = true
	return b, nil
}
func (r *BookingRepository) SetStatus(ctx context.Context, userID, shopID, bookingID, status, remark string) (models.Booking, error) {
	var empty models.Booking
	tx, err := r.beginWrite(ctx, userID, shopID)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(context.Background())
	b, err := lockManageableBooking(ctx, tx, userID, shopID, bookingID)
	if err != nil {
		return empty, err
	}
	allowed := (b.Status == "pending" && (status == "confirmed" || status == "cancelled")) || (b.Status == "confirmed" && (status == "completed" || status == "cancelled"))
	if !allowed {
		return empty, ErrTransition
	}
	updated, err := scan(tx.QueryRow(ctx, `UPDATE bookings SET status=$3::varchar,remark=CASE WHEN $3::varchar='cancelled' THEN $4::text ELSE '' END,updated_at=NOW() WHERE id=$1 AND tenant_id=$2 RETURNING `+columns, bookingID, shopID, status, remark))
	if err != nil {
		return empty, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO booking_status_history(booking_id,from_status,to_status,changed_by,remark) VALUES($1,$2,$3,$4,$5)`, b.ID, b.Status, status, userID, remark); err != nil {
		return empty, err
	}
	if err = tx.Commit(ctx); err != nil {
		return empty, err
	}
	updated.CanManage = true
	return updated, nil
}
