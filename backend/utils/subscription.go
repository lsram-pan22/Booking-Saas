package utils

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

var ErrSubscriptionRequired = errors.New("active shop owner subscription required")

type SubscriptionQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// A staff account's personal subscription never grants rights to its shop.
const shopSubscriptionQuery = `SELECT s.id::text FROM subscriptions s
 JOIN tenant_members owner ON owner.user_id=s.user_id
 JOIN users u ON u.id=owner.user_id
 WHERE owner.tenant_id=$1 AND owner.role='owner' AND owner.status='active'
 AND u.system_role='user' AND u.status='active' AND s.status='active'
 AND s.starts_at<=statement_timestamp()
 AND (s.ends_at IS NULL OR s.ends_at>statement_timestamp())`

func ShopSubscriptionActive(ctx context.Context, db SubscriptionQuerier, shopID string) (bool, error) {
	var id string
	err := db.QueryRow(ctx, shopSubscriptionQuery, shopID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Call after shop authorization, inside the write transaction. Locks prevent
// a concurrent subscription change from racing the protected write.
func RequireShopSubscription(ctx context.Context, tx pgx.Tx, shopID string) error {
	var id string
	err := tx.QueryRow(ctx, shopSubscriptionQuery+` FOR SHARE OF s,owner,u`, shopID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSubscriptionRequired
	}
	return err
}
