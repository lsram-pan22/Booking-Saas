package repositories

import (
	"booking-saas/controllers/tenant/account/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountExists = errors.New("account already exists")
var ErrFreePackageUnavailable = errors.New("free package is unavailable")

type RegisterRepository struct{ DB *pgxpool.Pool }

func NewRegisterRepository(db *pgxpool.Pool) *RegisterRepository { return &RegisterRepository{DB: db} }

func (r *RegisterRepository) Register(ctx context.Context, input models.Registration) (models.RegisterData, error) {
	var result models.RegisterData
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.Background())
	// Tenant and owner membership are created later, when adding a shop.
	err = tx.QueryRow(ctx, `INSERT INTO users (username,email,password_hash,full_name,phone,system_role,status) VALUES ($1,$2,$3,$4,$5,'user','active') RETURNING id`, input.Username, input.Email, input.PasswordHash, input.FullName, input.Phone).Scan(&result.UserID)
	if err != nil {
		return result, accountError(err)
	}
	s := &result.Subscription
	err = tx.QueryRow(ctx, `INSERT INTO subscriptions
        (user_id,package_id,package_code,package_name,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,shop_limit)
        SELECT $1,id,code,name,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,shop_limit
        FROM packages WHERE code='free' AND status='active' AND monthly_price_satang=0
        RETURNING id::text,package_id::text,package_code,package_name,status,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,starts_at,ends_at,shop_limit`, result.UserID).Scan(&s.ID, &s.PackageID, &s.PackageCode, &s.PackageName, &s.Status, &s.MonthlyPriceSatang, &s.UserLimit, &s.StaffManagementLevel, &s.ReportLevel, &s.ExportExcel, &s.APIWebhook, &s.StartsAt, &s.EndsAt, &s.ShopLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.RegisterData{}, ErrFreePackageUnavailable
	}
	if err != nil {
		return models.RegisterData{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return models.RegisterData{}, err
	}
	return result, nil
}

func accountError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrAccountExists
	}
	return err
}
