package repositories

import (
	"booking-saas/controllers/tenant/shops/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

var ErrMemberProtected = errors.New("owner or suspended membership is protected")

func (r *ShopRepository) SetMemberStatus(ctx context.Context, ownerID, shopID, memberID, status string) (models.MemberStatusResult, error) {
	var result models.MemberStatusResult
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.Background())
	var shop string
	// Same shop lock as AddMember: reactivation and creation share the quota lock.
	err = tx.QueryRow(ctx, `SELECT t.id::text FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id WHERE t.id=$1 AND m.user_id=$2 AND m.role='owner' AND m.status='active' AND u.status='active' AND u.system_role='user' AND t.status IN ('active','inactive') FOR UPDATE OF t FOR SHARE OF m,u`, shopID, ownerID).Scan(&shop)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrShopNotFound
	}
	if err != nil {
		return result, err
	}
	member := &result.Member
	err = tx.QueryRow(ctx, `SELECT m.id::text,u.id::text,u.username,u.full_name,u.email,COALESCE(u.phone,''),m.role,m.status,u.status FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.id=$1 AND m.tenant_id=$2 FOR UPDATE OF m FOR SHARE OF u`, memberID, shopID).Scan(&member.ID, &member.UserID, &member.Username, &member.FullName, &member.Email, &member.Phone, &member.Role, &member.Status, &member.UserStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrShopNotFound
	}
	if err != nil {
		return result, err
	}
	if member.Role != "staff" || member.Status == "suspended" {
		return result, ErrMemberProtected
	}
	var limit int
	err = tx.QueryRow(ctx, `SELECT user_limit FROM subscriptions WHERE user_id=$1 AND status='active' AND starts_at<=NOW() AND (ends_at IS NULL OR ends_at>NOW()) FOR SHARE`, ownerID).Scan(&limit)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	if err == nil {
		result.UserLimit = &limit
	}
	err = tx.QueryRow(ctx, `SELECT count(*) FROM tenant_members WHERE tenant_id=$1 AND status='active'`, shopID).Scan(&result.ActiveCount)
	if err != nil {
		return result, err
	}
	if status == "active" {
		if member.UserStatus != "active" {
			return result, ErrMemberAccountUnavailable
		}
		if result.UserLimit == nil {
			return result, ErrSubscriptionRequired
		}
		if member.Status != "active" && result.ActiveCount >= limit {
			return result, ErrMemberQuota
		}
	}
	if member.Status != status {
		if status == "active" {
			result.ActiveCount++
		} else if member.Status == "active" {
			result.ActiveCount--
		}
		_, err = tx.Exec(ctx, `UPDATE tenant_members SET status=$1,updated_at=NOW() WHERE id=$2 AND tenant_id=$3 AND role='staff'`, status, memberID, shopID)
		if err != nil {
			return result, err
		}
		member.Status = status
	}
	if err = tx.Commit(ctx); err != nil {
		return models.MemberStatusResult{}, err
	}
	return result, nil
}
