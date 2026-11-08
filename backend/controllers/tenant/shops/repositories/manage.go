package repositories

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

var ErrShopNotFound = errors.New("shop not found or inaccessible")

func scanShop(row pgx.Row) (models.Shop, error) {
	var shop models.Shop
	err := row.Scan(&shop.ID, &shop.Name, &shop.BusinessType, &shop.Slug, &shop.Status, &shop.Role, &shop.OpeningTime, &shop.ClosingTime)
	if errors.Is(err, pgx.ErrNoRows) {
		return shop, ErrShopNotFound
	}
	return shop, err
}

func (r *ShopRepository) Get(ctx context.Context, userID, shopID string) (models.Shop, error) {
	return scanShop(r.DB.QueryRow(ctx, `SELECT t.id::text,t.name,t.business_type,t.slug,t.status,m.role,t.opening_time,t.closing_time
 FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id
 WHERE t.id=$1 AND u.id=$2 AND u.system_role='user' AND u.status='active'
 AND m.status='active' AND m.role='owner' AND t.status IN ('pending','active','inactive')`, shopID, userID))
}

func (r *ShopRepository) Update(ctx context.Context, userID, shopID string, req models.UpdateShopRequest) (models.Shop, error) {
	// Scope the write itself to current account, active owner membership and shop.
	return scanShop(r.DB.QueryRow(ctx, `UPDATE tenants t SET name=$1,business_type=$2,status=COALESCE($5::varchar,t.status),opening_time=COALESCE($6::varchar,t.opening_time),closing_time=COALESCE($7::varchar,t.closing_time),updated_at=NOW()
 FROM tenant_members m, users u
 WHERE t.id=$3 AND m.tenant_id=t.id AND m.user_id=$4 AND u.id=m.user_id
 AND u.status='active' AND u.system_role='user' AND m.status='active' AND m.role='owner' AND t.status IN ('active','inactive')
 RETURNING t.id::text,t.name,t.business_type,t.slug,t.status,m.role,t.opening_time,t.closing_time`, req.Name, req.BusinessType, shopID, userID, req.Status, req.OpeningTime, req.ClosingTime))
}
