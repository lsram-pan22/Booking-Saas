package utils

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrFreePlanUnavailable = errors.New("active free package unavailable")

// Run before opening a feature transaction. Requests reconcile the requester
// and the owners of their shops, so staff cannot bypass an expired owner plan.
func EnsureUserSubscriptions(ctx context.Context, db *pgxpool.Pool, userID string) error {
	rows, err := db.Query(ctx, `SELECT $1::uuid UNION SELECT owner.user_id FROM tenant_members actor JOIN tenant_members owner ON owner.tenant_id=actor.tenant_id WHERE actor.user_id=$1 AND owner.role='owner'`, userID)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err = EnsureFreeFallback(ctx, db, id); err != nil && !errors.Is(err, ErrFreePlanUnavailable) {
			return err
		}
	}
	return nil
}

func EnsureFreeFallback(ctx context.Context, db *pgxpool.Pool, ownerID string) error {
	var due bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM subscriptions s WHERE s.user_id=$1 AND s.package_code<>'free' AND s.status IN ('active','expired') AND s.ends_at<=statement_timestamp()) AND NOT EXISTS(SELECT 1 FROM subscriptions s WHERE s.user_id=$1 AND s.status='active' AND (s.ends_at IS NULL OR s.ends_at>statement_timestamp()))`, ownerID).Scan(&due)
	if err != nil || !due {
		return err
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, ownerID); err != nil {
		return err
	}
	// Match the tenant-first lock order used by booking/member writes.
	rows, err := tx.Query(ctx, `SELECT t.id FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id WHERE m.user_id=$1 AND m.role='owner' ORDER BY t.id FOR UPDATE OF t`, ownerID)
	if err != nil {
		return err
	}
	for rows.Next() {
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var latestID, code, status string
	var expired bool
	err = tx.QueryRow(ctx, `SELECT id::text,package_code,status,ends_at IS NOT NULL AND ends_at<=statement_timestamp() FROM subscriptions WHERE user_id=$1 ORDER BY starts_at DESC,id DESC LIMIT 1 FOR UPDATE`, ownerID).Scan(&latestID, &code, &status, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if code == "free" || !expired || (status != "active" && status != "expired") {
		return nil
	}
	// Another request may have already installed the Free subscription.
	var current bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE user_id=$1 AND status='active' AND (ends_at IS NULL OR ends_at>statement_timestamp()))`, ownerID).Scan(&current); err != nil {
		return err
	}
	if current {
		return nil
	}
	var packageID string
	var shops, users int
	err = tx.QueryRow(ctx, `SELECT id::text,shop_limit,user_limit FROM packages WHERE code='free' AND status='active' AND monthly_price_satang=0 AND shop_limit>=1 AND user_limit>=1 FOR SHARE`).Scan(&packageID, &shops, &users)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrFreePlanUnavailable
	}
	if err != nil {
		return err
	}
	var selection bool
	err = tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM tenant_members WHERE user_id=$1 AND role='owner')>$2 OR EXISTS(SELECT 1 FROM tenant_members owner JOIN tenant_members member ON member.tenant_id=owner.tenant_id WHERE owner.user_id=$1 AND owner.role='owner' AND member.status='active' GROUP BY owner.tenant_id HAVING count(*)>$3)`, ownerID, shops, users).Scan(&selection)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE subscriptions SET status='expired' WHERE user_id=$1 AND status='active' AND ends_at<=statement_timestamp()`, ownerID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO subscriptions(user_id,package_id,package_code,package_name,status,monthly_price_satang,shop_limit,user_limit,staff_management_level,report_level,export_excel,api_webhook,starts_at,ends_at,selection_required)
 SELECT $1,id,code,name,'active',monthly_price_satang,shop_limit,user_limit,staff_management_level,report_level,export_excel,api_webhook,statement_timestamp(),NULL,$3 FROM packages WHERE id=$2`, ownerID, packageID, selection)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE tenants SET plan_enabled=NOT $2 WHERE id IN(SELECT tenant_id FROM tenant_members WHERE user_id=$1 AND role='owner')`, ownerID, selection); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE tenant_members SET plan_enabled=(role='owner' OR NOT $2) WHERE tenant_id IN(SELECT tenant_id FROM tenant_members WHERE user_id=$1 AND role='owner')`, ownerID, selection); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
