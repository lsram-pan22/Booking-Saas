package repositories

import (
	"booking-saas/config"
	accountModels "booking-saas/controllers/tenant/account/models"
	accountRepositories "booking-saas/controllers/tenant/account/repositories"
	"booking-saas/controllers/tenant/shops/models"
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddShopMemberIntegration(t *testing.T) {
	if os.Getenv("TENANT_REGISTRATION_DB_TEST") != "1" {
		t.Skip("set TENANT_REGISTRATION_DB_TEST=1 to test PostgreSQL temporary tables")
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
	poolCfg := utils.DB.Config()
	poolCfg.MaxConns = 1
	poolCfg.MinConns = 0
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, query := range []string{
		`CREATE TEMP TABLE users (LIKE public.users INCLUDING DEFAULTS INCLUDING CONSTRAINTS)`,
		`CREATE UNIQUE INDEX test_email_unique ON users(LOWER(email))`,
		`CREATE UNIQUE INDEX users_username_lower_unique ON users(LOWER(username))`,
		`CREATE TEMP TABLE packages (LIKE public.packages INCLUDING ALL)`,
		`CREATE TEMP TABLE subscriptions (LIKE public.subscriptions INCLUDING ALL)`,
		`CREATE TEMP TABLE tenants (LIKE public.tenants INCLUDING ALL)`,
		`CREATE TEMP TABLE tenant_members (LIKE public.tenant_members INCLUDING DEFAULTS INCLUDING CONSTRAINTS)`,
		`CREATE UNIQUE INDEX test_member_tenant_user ON tenant_members(tenant_id,user_id)`,
		`CREATE UNIQUE INDEX tenant_members_one_staff_shop ON tenant_members(user_id) WHERE role='staff'`,
		`INSERT INTO packages SELECT * FROM public.packages WHERE code='free'`,
		`UPDATE packages SET user_limit=2,shop_limit=1,status='active',monthly_price_satang=0`,
	} {
		if _, err = pool.Exec(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	owner, err := accountRepositories.NewRegisterRepository(pool).Register(ctx, accountModels.Registration{FullName: "Owner", Email: "owner@example.test", Username: "owner@example.test", Phone: "0812345678", PasswordHash: "unused-owner-hash"})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewShopRepository(pool)
	shop, err := repo.Create(ctx, owner.UserID, models.NewShop{Name: "Member Test Shop", Slug: "member-test-shop", BusinessType: "salon"})
	if err != nil {
		t.Fatal(err)
	}
	hash, err := utils.HashPassword("staff-password-123")
	if err != nil {
		t.Fatal(err)
	}
	input := models.NewMember{Username: "staff01", Mode: "new", Email: "staff@example.test", FullName: "Staff", Phone: "0891234567", PasswordHash: hash}
	staff, err := repo.AddMember(ctx, owner.UserID, shop.ID, input)
	if err != nil || staff.Role != "staff" || staff.UserID == owner.UserID {
		t.Fatalf("create staff failed: %v", err)
	}
	if staff.Username != "staff01" {
		t.Fatal("staff username was not saved")
	}
	staffAccount, err := accountRepositories.NewLoginRepository(pool).FindByIdentity(ctx, "staff01")
	if err != nil || staffAccount.User.ID != staff.UserID {
		t.Fatalf("staff username lookup failed: %v", err)
	}
	var storedHash string
	if err = pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1`, staff.UserID).Scan(&storedHash); err != nil || !utils.CheckPassword("staff-password-123", storedHash) {
		t.Fatal("staff credentials not saved")
	}
	profile, err := accountRepositories.NewLoginRepository(pool).Profile(ctx, staff.UserID)
	if err != nil || len(profile.Tenants) != 1 || profile.Tenants[0].Role != "staff" || profile.Subscription != nil {
		t.Fatalf("staff cannot access its shop: %v", err)
	}
	if profile.WorkspaceRole != "staff" {
		t.Fatal("staff workspace role is incorrect")
	}
	if profile.User.Role != "staff" || profile.User.SystemRole != "user" {
		t.Fatal("staff session must carry tenant role separately from system role")
	}
	if _, err = repo.Create(ctx, staff.UserID, models.NewShop{Name: "Unauthorized shop", Slug: "staff-owned-shop", BusinessType: "salon"}); !errors.Is(err, ErrOwnerAccountRequired) {
		t.Fatal("staff could create its own shop")
	}
	input.Email = "over-quota@example.test"
	if _, err = repo.AddMember(ctx, owner.UserID, shop.ID, input); !errors.Is(err, ErrMemberQuota) {
		t.Fatal("quota was exceeded")
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil || count != 2 {
		t.Fatal("quota failure left an orphan account")
	}
	if _, err = repo.AddMember(ctx, staff.UserID, shop.ID, input); !errors.Is(err, ErrShopNotFound) {
		t.Fatal("staff could add members")
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET user_limit=5`); err != nil {
		t.Fatal(err)
	}
	var existingID string
	if err = pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash,full_name,phone,system_role,status) VALUES('existing@example.test','existing@example.test','original-hash','Existing','0811111111','user','active') RETURNING id::text`).Scan(&existingID); err != nil {
		t.Fatal(err)
	}
	existing := models.NewMember{Mode: "existing", Email: "existing@example.test", PasswordHash: "must-not-replace", FullName: "must-not-replace"}
	linked, err := repo.AddMember(ctx, owner.UserID, shop.ID, existing)
	if err != nil || linked.UserID != existingID || linked.FullName != "Existing" {
		t.Fatalf("link existing account failed: %v", err)
	}
	// A subscribed account that is only a staff member must not expose packages.
	if _, err = pool.Exec(ctx, `INSERT INTO subscriptions(user_id,package_id,package_code,package_name,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,shop_limit) SELECT $1,package_id,package_code,package_name,monthly_price_satang,user_limit,staff_management_level,report_level,export_excel,api_webhook,shop_limit FROM subscriptions WHERE user_id=$2`, existingID, owner.UserID); err != nil {
		t.Fatal(err)
	}
	existingProfile, err := accountRepositories.NewLoginRepository(pool).Profile(ctx, existingID)
	if err != nil || existingProfile.WorkspaceRole != "staff" || existingProfile.Subscription != nil {
		t.Fatalf("staff subscription leaked: %v", err)
	}
	if err = pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1`, existingID).Scan(&storedHash); err != nil || storedHash != "original-hash" {
		t.Fatal("existing password was changed")
	}
	if _, err = repo.AddMember(ctx, owner.UserID, shop.ID, existing); !errors.Is(err, ErrMemberExists) {
		t.Fatal("duplicate membership allowed")
	}
	input.Email = "another-staff@example.test"
	input.Username = "STAFF01"
	if _, err = repo.AddMember(ctx, owner.UserID, shop.ID, input); !errors.Is(err, ErrMemberUsernameExists) {
		t.Fatalf("duplicate username was not rejected: %v", err)
	}
	input.Username = "duplicate_email_staff"
	input.Email = existing.Email
	if _, err = repo.AddMember(ctx, owner.UserID, shop.ID, input); !errors.Is(err, ErrMemberEmailExists) {
		t.Fatal("duplicate email allowed")
	}
	// Force the membership insert to fail after creating an account.
	var ownerMemberID string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM tenant_members WHERE tenant_id=$1 AND user_id=$2`, shop.ID, owner.UserID).Scan(&ownerMemberID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, ownerMemberID, "inactive"); !errors.Is(err, ErrMemberProtected) {
		t.Fatal("owner could be disabled")
	}
	if _, err = repo.SetMemberStatus(ctx, staff.UserID, shop.ID, staff.ID, "inactive"); !errors.Is(err, ErrShopNotFound) {
		t.Fatal("staff could change membership status")
	}
	var otherShopID, otherMembershipID string
	var outsiderID string
	if err = pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash,full_name,system_role,status) VALUES('outsider','outsider@example.test','unused','Outsider','user','active') RETURNING id::text`).Scan(&outsiderID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO tenants(name,slug,business_type,status) VALUES('Other shop','other-member-shop','salon','active') RETURNING id::text`).Scan(&otherShopID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'staff','active') RETURNING id::text`, otherShopID, outsiderID).Scan(&otherMembershipID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'owner','active')`, otherShopID, owner.UserID); err != nil {
		t.Fatal(err)
	}
	mixedProfile, err := accountRepositories.NewLoginRepository(pool).Profile(ctx, owner.UserID)
	if err != nil || mixedProfile.WorkspaceRole != "owner" || mixedProfile.Subscription == nil {
		t.Fatal("owner with multiple shops lost owner access")
	}
	if _, err = repo.AddMember(ctx, owner.UserID, otherShopID, models.NewMember{Mode: "existing", Email: staff.Email}); !errors.Is(err, ErrStaffShopAssigned) {
		t.Fatalf("staff was allowed to join a second shop: %v", err)
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, otherMembershipID, "inactive"); !errors.Is(err, ErrShopNotFound) {
		t.Fatal("cross-shop membership could be changed")
	}
	disabled, err := repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "inactive")
	if err != nil || disabled.Member.Status != "inactive" || disabled.Member.UserStatus != "active" || disabled.ActiveCount != 2 {
		t.Fatalf("disable staff failed: %v", err)
	}
	if _, err = repo.Get(ctx, staff.UserID, shop.ID); !errors.Is(err, ErrShopNotFound) {
		t.Fatal("disabled member retained access")
	}
	if _, err = repo.AddMember(ctx, owner.UserID, otherShopID, models.NewMember{Mode: "existing", Email: staff.Email}); !errors.Is(err, ErrStaffShopAssigned) {
		t.Fatalf("inactive staff was allowed to join a second shop: %v", err)
	}
	noShopProfile, err := accountRepositories.NewLoginRepository(pool).Profile(ctx, staff.UserID)
	if err != nil || noShopProfile.WorkspaceRole != "staff" || len(noShopProfile.Tenants) != 0 || noShopProfile.Subscription != nil {
		t.Fatal("disabled staff gained owner access")
	}
	if repeated, err := repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "inactive"); err != nil || repeated.ActiveCount != 2 {
		t.Fatal("repeated disable changed seat count")
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET user_limit=2`); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "active"); !errors.Is(err, ErrMemberQuota) {
		t.Fatal("reactivation bypassed quota")
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET user_limit=3`); err != nil {
		t.Fatal(err)
	}
	if enabled, err := repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "active"); err != nil || enabled.ActiveCount != 3 || enabled.Member.Status != "active" {
		t.Fatalf("reactivation failed: %v", err)
	}
	if repeated, err := repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "active"); err != nil || repeated.ActiveCount != 3 {
		t.Fatal("repeated activation changed seat count")
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET status='expired'`); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "inactive"); err != nil {
		t.Fatal("expired plan prevented disabling a member")
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "active"); !errors.Is(err, ErrSubscriptionRequired) {
		t.Fatal("expired subscription allowed reactivation")
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET status='active',user_limit=5`); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE users SET status='suspended' WHERE id=$1`, staff.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "active"); !errors.Is(err, ErrMemberAccountUnavailable) {
		t.Fatal("suspended account could be reactivated")
	}
	if _, err = pool.Exec(ctx, `UPDATE users SET status='active' WHERE id=$1`, staff.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetMemberStatus(ctx, owner.UserID, shop.ID, staff.ID, "active"); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `ALTER TABLE tenant_members ADD CONSTRAINT test_reject_staff CHECK(role<>'staff') NOT VALID`); err != nil {
		t.Fatal(err)
	}
	input.Username = "rollback_staff"
	input.Email = "rollback@example.test"
	if _, err = repo.AddMember(ctx, owner.UserID, shop.ID, input); err == nil {
		t.Fatal("membership insert should fail")
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email=$1`, input.Email).Scan(&count); err != nil || count != 0 {
		t.Fatal("account was not rolled back")
	}
	if _, err = pool.Exec(ctx, `ALTER TABLE tenant_members DROP CONSTRAINT test_reject_staff; UPDATE tenants SET status='inactive'`); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.AddMember(ctx, owner.UserID, shop.ID, input); !errors.Is(err, ErrShopNotFound) {
		t.Fatal("disabled shop accepted a new member")
	}
}
