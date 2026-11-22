package repositories

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/utils"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Explicit opt-in. All fixture tables and the migration live in a disposable
// schema; application accounts and subscriptions are never changed.
func TestFreeFallbackAndSelectionIntegration(t *testing.T) {
	if os.Getenv("SUBSCRIPTION_FALLBACK_DB_TEST") != "1" {
		t.Skip("set SUBSCRIPTION_FALLBACK_DB_TEST=1 for isolated PostgreSQL tests")
	}
	envFile, err := filepath.Abs("../../../../.env")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV_FILE", envFile)
	if err = utils.ConnectDatabase(config.LoadConfig()); err != nil {
		t.Fatal(err)
	}
	defer utils.CloseDatabase()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	schema := fmt.Sprintf("plan_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = utils.DB.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := utils.DB.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	cfg := utils.DB.Config()
	cfg.MinConns = 0
	cfg.MaxConns = 4
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",pg_catalog"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE TABLE users(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),system_role TEXT DEFAULT 'user',status TEXT DEFAULT 'active',full_name TEXT DEFAULT 'Test');
 CREATE TABLE tenants(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),name TEXT DEFAULT 'Shop',status TEXT DEFAULT 'active',created_at TIMESTAMPTZ DEFAULT NOW());
 CREATE TABLE tenant_members(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),tenant_id UUID,user_id UUID,role TEXT,status TEXT DEFAULT 'active');
 CREATE TABLE packages(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),code TEXT,name TEXT,status TEXT,monthly_price_satang BIGINT,shop_limit INT,user_limit INT,staff_management_level TEXT DEFAULT 'basic',report_level TEXT DEFAULT 'basic',export_excel BOOLEAN DEFAULT FALSE,api_webhook BOOLEAN DEFAULT FALSE);
 CREATE TABLE subscriptions(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),user_id UUID UNIQUE,package_id UUID,package_code TEXT,package_name TEXT,status TEXT DEFAULT 'active',monthly_price_satang BIGINT,shop_limit INT,user_limit INT,staff_management_level TEXT,report_level TEXT,export_excel BOOLEAN,api_webhook BOOLEAN,starts_at TIMESTAMPTZ DEFAULT NOW(),ends_at TIMESTAMPTZ);`)
	migration, err := os.ReadFile("../../../../migrations/020_subscription_free_fallback.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(string(migration))
	var owner, staff, stranger, first, second, member string
	for _, dest := range []*string{&owner, &staff, &stranger} {
		if err = pool.QueryRow(ctx, `INSERT INTO users DEFAULT VALUES RETURNING id::text`).Scan(dest); err != nil {
			t.Fatal(err)
		}
	}
	for _, dest := range []*string{&first, &second} {
		if err = pool.QueryRow(ctx, `INSERT INTO tenants DEFAULT VALUES RETURNING id::text`).Scan(dest); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO tenant_members(tenant_id,user_id,role) VALUES($1,$3,'owner'),($2,$3,'owner')`, first, second, owner)
	if err = pool.QueryRow(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role) VALUES($1,$2,'staff') RETURNING id::text`, first, staff).Scan(&member); err != nil {
		t.Fatal(err)
	}
	exec(`INSERT INTO packages(code,name,status,monthly_price_satang,shop_limit,user_limit) VALUES('free','Free','active',0,1,2),('growth','Growth','active',99000,3,10)`)
	exec(`INSERT INTO subscriptions(user_id,package_id,package_code,package_name,monthly_price_satang,shop_limit,user_limit,staff_management_level,report_level,export_excel,api_webhook,starts_at,ends_at)
 SELECT $1,id,code,name,monthly_price_satang,shop_limit,user_limit,staff_management_level,report_level,export_excel,api_webhook,NOW()-INTERVAL '31 days',NOW()-INTERVAL '1 second' FROM packages WHERE code='growth'`, owner)
	var wg sync.WaitGroup
	errs := make(chan error, 3)
	for range 3 {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- utils.EnsureFreeFallback(ctx, pool, owner) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var active, expired int
	var selection bool
	if err = pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE status='active' AND package_code='free' AND ends_at IS NULL),count(*) FILTER(WHERE status='expired' AND package_code='growth' AND monthly_price_satang=99000) FROM subscriptions WHERE user_id=$1`, owner).Scan(&active, &expired); err != nil || active != 1 || expired != 1 {
		t.Fatalf("history/fallback: %d %d %v", active, expired, err)
	}
	if err = pool.QueryRow(ctx, `SELECT selection_required FROM subscriptions WHERE user_id=$1 AND status='active'`, owner).Scan(&selection); err != nil || !selection {
		t.Fatal("over quota was not restricted", err)
	}
	enabled, err := utils.ShopSubscriptionActive(ctx, pool, first, staff)
	if err != nil || enabled {
		t.Fatal("staff bypassed required scope selection", err)
	}
	repo := NewPlanRepository(pool)
	if err = repo.Select(ctx, stranger, models.PlanSelection{ShopIDs: []string{first}}); !errors.Is(err, models.ErrPlanSelection) {
		t.Fatal("stranger changed shop", err)
	}
	if err = repo.Select(ctx, owner, models.PlanSelection{ShopIDs: []string{first, second}}); !errors.Is(err, models.ErrPlanSelection) {
		t.Fatal("shop quota bypass", err)
	}
	if err = repo.Select(ctx, owner, models.PlanSelection{ShopIDs: []string{first}, MemberIDs: []string{member}}); err != nil {
		t.Fatal(err)
	}
	enabled, err = utils.ShopSubscriptionActive(ctx, pool, first, staff)
	if err != nil || !enabled {
		t.Fatal("selected staff denied", err)
	}
	enabled, err = utils.ShopSubscriptionActive(ctx, pool, second, owner)
	if err != nil || enabled {
		t.Fatal("unselected shop allowed", err)
	}
	// Re-selection can exclude a staff member without deactivating the account.
	if err = repo.Select(ctx, owner, models.PlanSelection{ShopIDs: []string{first}}); err != nil {
		t.Fatal(err)
	}
	enabled, err = utils.ShopSubscriptionActive(ctx, pool, first, staff)
	if err != nil || enabled {
		t.Fatal("excluded staff allowed", err)
	}
	var staffStatus string
	if err = pool.QueryRow(ctx, `SELECT status FROM tenant_members WHERE id=$1`, member).Scan(&staffStatus); err != nil || staffStatus != "active" {
		t.Fatal("membership status destroyed", err)
	}
	if err = utils.EnsureFreeFallback(ctx, pool, owner); err != nil {
		t.Fatal(err)
	}
	var records int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM subscriptions WHERE user_id=$1`, owner).Scan(&records); err != nil || records != 2 {
		t.Fatal("duplicate Free subscription", err)
	}
	// A paid account with no excess resources can use Free immediately.
	exec(`INSERT INTO subscriptions(user_id,package_id,package_code,package_name,monthly_price_satang,shop_limit,user_limit,starts_at,ends_at)
 SELECT $1,id,code,name,monthly_price_satang,shop_limit,user_limit,NOW()-INTERVAL '31 days',NOW()-INTERVAL '1 second' FROM packages WHERE code='growth'`, stranger)
	if err = utils.EnsureFreeFallback(ctx, pool, stranger); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT selection_required FROM subscriptions WHERE user_id=$1 AND status='active' AND package_code='free' AND ends_at IS NULL`, stranger).Scan(&selection); err != nil || selection {
		t.Fatal("within-quota Free account incorrectly requires selection", err)
	}
}
