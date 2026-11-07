package repositories

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

// Members exposes account/contact details only to an active owner of this shop.
func (r *ShopRepository) Members(ctx context.Context, userID, shopID string) (models.ShopMembers, error) {
	result := models.ShopMembers{Members: make([]models.ShopMember, 0)}
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.Background())
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members m JOIN users u ON u.id=m.user_id JOIN tenants t ON t.id=m.tenant_id WHERE m.tenant_id=$1 AND m.user_id=$2 AND m.role='owner' AND m.status='active' AND u.status='active' AND u.system_role='user' AND (t.status IN ('pending','active') OR (t.status='inactive' AND m.role='owner')))`, shopID, userID).Scan(&allowed)
	if err != nil {
		return result, err
	}
	if !allowed {
		return result, ErrShopNotFound
	}
	var userLimit int
	err = tx.QueryRow(ctx, `SELECT user_limit FROM subscriptions WHERE user_id=$1 AND status='active' AND starts_at<=NOW() AND (ends_at IS NULL OR ends_at>NOW())`, userID).Scan(&userLimit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	if err == nil {
		result.UserLimit = &userLimit
	}
	rows, err := tx.Query(ctx, `SELECT m.id::text,u.id::text,u.full_name,u.email,COALESCE(u.phone,''),m.role,m.status,u.status,u.username FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=$1 ORDER BY CASE WHEN m.role='owner' THEN 0 ELSE 1 END,m.created_at,m.id`, shopID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var member models.ShopMember
		if err = rows.Scan(&member.ID, &member.UserID, &member.FullName, &member.Email, &member.Phone, &member.Role, &member.Status, &member.UserStatus, &member.Username); err != nil {
			return result, err
		}
		if member.Status == "active" {
			result.ActiveCount++
		}
		result.Members = append(result.Members, member)
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return result, err
	}
	return result, nil
}
