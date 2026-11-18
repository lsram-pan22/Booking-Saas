package repositories

import (
	"booking-saas/controllers/tenant/account/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanRepository struct{ DB *pgxpool.Pool }

func NewPlanRepository(db *pgxpool.Pool) *PlanRepository { return &PlanRepository{DB: db} }
func (r *PlanRepository) Overview(ctx context.Context, userID string) (models.PlanOverview, error) {
	result := models.PlanOverview{History: []models.Subscription{}, Shops: []models.PlanShop{}}
	var allowed bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members WHERE user_id=$1 AND role='owner') OR (NOT EXISTS(SELECT 1 FROM tenant_members WHERE user_id=$1 AND role='staff') AND EXISTS(SELECT 1 FROM subscriptions WHERE user_id=$1))`, userID).Scan(&allowed)
	if err != nil {
		return result, err
	}
	if !allowed {
		return result, models.ErrPlanSelection
	}
	rows, err := r.DB.Query(ctx, `SELECT id::text,package_id::text,package_code,package_name,status,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,starts_at,ends_at,shop_limit,selection_required FROM subscriptions WHERE user_id=$1 ORDER BY starts_at DESC,id DESC LIMIT 100`, userID)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var s models.Subscription
		err = rows.Scan(&s.ID, &s.PackageID, &s.PackageCode, &s.PackageName, &s.Status, &s.MonthlyPriceSatang, &s.UserLimit, &s.StaffManagementLevel, &s.ReportLevel, &s.ExportExcel, &s.APIWebhook, &s.StartsAt, &s.EndsAt, &s.ShopLimit, &s.SelectionRequired)
		if err != nil {
			rows.Close()
			return result, err
		}
		result.History = append(result.History, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	rows, err = r.DB.Query(ctx, `SELECT t.id::text,t.name,t.status,t.plan_enabled FROM tenants t JOIN tenant_members owner ON owner.tenant_id=t.id WHERE owner.user_id=$1 AND owner.role='owner' AND owner.status='active' AND t.status IN ('active','inactive','pending') ORDER BY t.created_at,t.id`, userID)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		s := models.PlanShop{Members: []models.PlanMember{}}
		if err = rows.Scan(&s.ID, &s.Name, &s.Status, &s.Enabled); err != nil {
			rows.Close()
			return result, err
		}
		result.Shops = append(result.Shops, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	for i := range result.Shops {
		rows, err = r.DB.Query(ctx, `SELECT m.id::text,u.full_name,m.role,m.plan_enabled FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.tenant_id=$1 AND m.status='active' AND u.status='active' AND m.role IN ('owner','staff') ORDER BY m.role,u.full_name,m.id`, result.Shops[i].ID)
		if err != nil {
			return result, err
		}
		for rows.Next() {
			var m models.PlanMember
			if err = rows.Scan(&m.ID, &m.Name, &m.Role, &m.Enabled); err != nil {
				rows.Close()
				return result, err
			}
			result.Shops[i].Members = append(result.Shops[i].Members, m)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (r *PlanRepository) Select(ctx context.Context, userID string, input models.PlanSelection) error {
	if input.ShopIDs == nil {
		input.ShopIDs = []string{}
	}
	if input.MemberIDs == nil {
		input.MemberIDs = []string{}
	}
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, userID); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT t.id FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id WHERE m.user_id=$1 AND m.role='owner' ORDER BY t.id FOR UPDATE OF t`, userID)
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
	var subID string
	var shopLimit, userLimit int
	err = tx.QueryRow(ctx, `SELECT id::text,shop_limit,user_limit FROM subscriptions WHERE user_id=$1 AND package_code='free' AND status='active' AND starts_at<=statement_timestamp() AND (ends_at IS NULL OR ends_at>statement_timestamp()) FOR UPDATE`, userID).Scan(&subID, &shopLimit, &userLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ErrPlanSelection
	}
	if err != nil {
		return err
	}
	if len(input.ShopIDs) > shopLimit {
		return models.ErrPlanSelection
	}
	var valid int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM tenants t JOIN tenant_members m ON m.tenant_id=t.id JOIN users u ON u.id=m.user_id WHERE t.id=ANY($1::uuid[]) AND m.user_id=$2 AND m.role='owner' AND m.status='active' AND u.status='active' AND t.status IN ('active','inactive','pending')`, input.ShopIDs, userID).Scan(&valid)
	if err != nil {
		return err
	}
	if valid != len(input.ShopIDs) {
		return models.ErrPlanSelection
	}
	err = tx.QueryRow(ctx, `SELECT count(*) FROM tenant_members m JOIN users u ON u.id=m.user_id WHERE m.id=ANY($1::uuid[]) AND m.tenant_id=ANY($2::uuid[]) AND m.role='staff' AND m.status='active' AND u.status='active'`, input.MemberIDs, input.ShopIDs).Scan(&valid)
	if err != nil {
		return err
	}
	if valid != len(input.MemberIDs) {
		return models.ErrPlanSelection
	}
	var exceeded bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_members WHERE tenant_id=ANY($1::uuid[]) AND id=ANY($2::uuid[]) GROUP BY tenant_id HAVING count(*)+1>$3)`, input.ShopIDs, input.MemberIDs, userLimit).Scan(&exceeded)
	if err != nil {
		return err
	}
	if exceeded || userLimit < 1 {
		return models.ErrPlanSelection
	}
	if _, err = tx.Exec(ctx, `UPDATE tenants SET plan_enabled=(id=ANY($2::uuid[])) WHERE id IN(SELECT tenant_id FROM tenant_members WHERE user_id=$1 AND role='owner')`, userID, input.ShopIDs); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE tenant_members SET plan_enabled=(role='owner' OR id=ANY($2::uuid[])) WHERE tenant_id IN(SELECT tenant_id FROM tenant_members WHERE user_id=$1 AND role='owner')`, userID, input.MemberIDs); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE subscriptions SET selection_required=FALSE WHERE id=$1`, subID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
