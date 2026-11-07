package repositories

import (
	"booking-saas/controllers/tenant/account/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountNotFound = errors.New("account not found")

type LoginRepository struct{ DB *pgxpool.Pool }

func NewLoginRepository(db *pgxpool.Pool) *LoginRepository { return &LoginRepository{DB: db} }
func (r *LoginRepository) FindByIdentity(ctx context.Context, email string) (models.Account, error) {
	var a models.Account
	err := r.DB.QueryRow(ctx, `SELECT id,email,full_name,password_hash,system_role,status FROM users WHERE system_role='user' AND (LOWER(email)=$1 OR LOWER(username)=$1)`, email).Scan(&a.User.ID, &a.User.Email, &a.User.FullName, &a.PasswordHash, &a.SystemRole, &a.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, ErrAccountNotFound
	}
	return a, err
}
func (r *LoginRepository) Profile(ctx context.Context, id string) (models.Profile, error) {
	p := models.Profile{Tenants: make([]models.Tenant, 0)}
	err := r.DB.QueryRow(ctx, `SELECT id,email,full_name,COALESCE(phone,'') FROM users WHERE id=$1 AND status='active' AND system_role='user'`, id).Scan(&p.User.ID, &p.User.Email, &p.User.FullName, &p.User.Phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrAccountNotFound
	}
	if err != nil {
		return p, err
	}
	var subscription models.Subscription
	var staffMemberships int
	err = r.DB.QueryRow(ctx, `SELECT count(*) FILTER (WHERE role='owner'),count(*) FILTER (WHERE role='staff') FROM tenant_members WHERE user_id=$1`, id).Scan(&p.OwnedShopCount, &staffMemberships)
	if err != nil {
		return p, err
	}
	err = r.DB.QueryRow(ctx, `SELECT id::text,package_id::text,package_code,package_name,status,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,starts_at,ends_at,shop_limit
        FROM subscriptions WHERE user_id=$1 AND status='active' AND starts_at<=NOW() AND (ends_at IS NULL OR ends_at>NOW())`, id).Scan(&subscription.ID, &subscription.PackageID, &subscription.PackageCode, &subscription.PackageName, &subscription.Status, &subscription.MonthlyPriceSatang, &subscription.UserLimit, &subscription.StaffManagementLevel, &subscription.ReportLevel, &subscription.ExportExcel, &subscription.APIWebhook, &subscription.StartsAt, &subscription.EndsAt, &subscription.ShopLimit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return p, err
	}
	if err == nil {
		p.Subscription = &subscription
	}
	rows, err := r.DB.Query(ctx, `SELECT t.id::text,t.name,t.status,m.role FROM tenant_members m JOIN tenants t ON t.id=m.tenant_id WHERE m.user_id=$1 AND m.status='active' AND m.role IN ('owner','staff') AND (t.status='active' OR (m.role='owner' AND t.status IN ('pending','inactive'))) ORDER BY t.created_at,t.id`, id)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var tenant models.Tenant
		if err = rows.Scan(&tenant.ID, &tenant.Name, &tenant.Status, &tenant.Role); err != nil {
			return p, err
		}
		p.Tenants = append(p.Tenants, tenant)
	}
	if err = rows.Err(); err != nil {
		return p, err
	}
	p.WorkspaceRole = "staff"
	if p.OwnedShopCount > 0 || (staffMemberships == 0 && p.Subscription != nil) {
		p.WorkspaceRole = "owner"
	} else {
		// Staff must not receive account subscription details in /login or /me.
		p.Subscription = nil
	}
	p.User.Role = p.WorkspaceRole
	p.User.SystemRole = "user"
	return p, nil
}
