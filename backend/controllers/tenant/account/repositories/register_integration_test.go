package repositories

import (
	"booking-saas/config"
	packageRepositories "booking-saas/controllers/super_admin/packages/repositories"
	"booking-saas/controllers/tenant/account/models"
	shopModels "booking-saas/controllers/tenant/shops/models"
	shopRepositories "booking-saas/controllers/tenant/shops/repositories"
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Uses session-local temporary tables; real accounts and subscriptions are untouched.
func TestRegistrationFreeSubscriptionIntegration(t *testing.T) {
	if os.Getenv("TENANT_REGISTRATION_DB_TEST") != "1" {
		t.Skip("set TENANT_REGISTRATION_DB_TEST=1 to test against PostgreSQL temporary tables")
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
		`CREATE TEMP TABLE users (LIKE public.users INCLUDING ALL)`,
		`CREATE UNIQUE INDEX test_email_lower ON users (LOWER(email))`,
		`CREATE TEMP TABLE packages (LIKE public.packages INCLUDING ALL)`,
		`CREATE TEMP TABLE subscriptions (LIKE public.subscriptions INCLUDING ALL)`,
		`CREATE TEMP TABLE tenants (LIKE public.tenants INCLUDING ALL)`,
		`CREATE TEMP TABLE tenant_members (LIKE public.tenant_members INCLUDING ALL)`,
		`INSERT INTO packages SELECT * FROM public.packages WHERE code='free'`,
	} {
		if _, err = pool.Exec(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRegisterRepository(pool)
	input := models.Registration{FullName: "Test Account", Email: "subscription-test@example.com", Username: "subscription-test@example.com", Phone: "0812345678", PasswordHash: "test-only-hash"}
	result, err := repo.Register(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Subscription.PackageCode != "free" || result.Subscription.Status != "active" {
		t.Fatal("missing free subscription")
	}
	if !utils.ValidUserID(result.UserID) {
		t.Fatal("registration did not return a user UUID")
	}
	loginRepo := NewLoginRepository(pool)
	account, err := loginRepo.FindByIdentity(ctx, input.Email)
	if err != nil || account.User.ID != result.UserID {
		t.Fatalf("UUID account lookup failed: %v", err)
	}
	profile, err := loginRepo.Profile(ctx, result.UserID)
	if err != nil || profile.Subscription == nil || profile.Subscription.ID != result.Subscription.ID {
		t.Fatalf("UUID profile/subscription lookup failed: %v", err)
	}
	if profile.User.Phone != input.Phone {
		t.Fatalf("profile phone = %q, want %q", profile.User.Phone, input.Phone)
	}
	if profile.WorkspaceRole != "owner" {
		t.Fatal("new owner account must retain onboarding access before creating a shop")
	}
	if profile.User.Role != "owner" || profile.User.SystemRole != "user" {
		t.Fatal("owner session roles are incorrect")
	}
	if profile.Subscription.ShopLimit < 1 || profile.Subscription.UserLimit < 1 || profile.Subscription.StartsAt.IsZero() {
		t.Fatal("subscription entitlements were not loaded")
	}
	updated, err := NewSettingsRepository(pool).Update(ctx, result.UserID, models.UpdateProfileRequest{FullName: "Updated Owner", Phone: "0898765432"})
	if err != nil || updated.FullName != "Updated Owner" || updated.Phone != "0898765432" || updated.Email != input.Email {
		t.Fatalf("profile settings update failed: %v", err)
	}
	for _, query := range []string{
		`ALTER TABLE tenant_members ADD CONSTRAINT test_reject_owner CHECK (role <> 'owner')`,
	} {
		if _, err = pool.Exec(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	shopRepo := shopRepositories.NewShopRepository(pool)
	shopInput := shopModels.NewShop{Name: "Test Shop", BusinessType: "salon", Slug: "test-shop"}
	if _, err = shopRepo.Create(ctx, result.UserID, shopInput); err == nil {
		t.Fatal("membership failure was not propagated")
	}
	var shopCount int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM tenants`).Scan(&shopCount); err != nil || shopCount != 0 {
		t.Fatal("shop creation was not rolled back")
	}
	if _, err = pool.Exec(ctx, `ALTER TABLE tenant_members DROP CONSTRAINT test_reject_owner`); err != nil {
		t.Fatal(err)
	}
	shop, err := shopRepo.Create(ctx, result.UserID, shopInput)
	if err != nil || shop.Role != "owner" || shop.Status != "active" {
		t.Fatalf("shop creation failed: %v", err)
	}
	var ownerID string
	if err = pool.QueryRow(ctx, `SELECT user_id::text FROM tenant_members WHERE tenant_id=$1 AND role='owner'`, shop.ID).Scan(&ownerID); err != nil || ownerID != result.UserID {
		t.Fatal("shop was not linked to its authenticated owner")
	}
	if _, err = shopRepo.Get(ctx, result.UserID, shop.ID); err != nil {
		t.Fatal(err)
	}
	memberList, err := shopRepo.Members(ctx, result.UserID, shop.ID)
	if err != nil || memberList.ActiveCount != 1 || len(memberList.Members) != 1 || memberList.UserLimit == nil || *memberList.UserLimit != result.Subscription.UserLimit {
		t.Fatalf("owner member quota is incorrect: %v", err)
	}
	strangerID := "99999999-9999-4999-8999-999999999999"
	if _, err = shopRepo.Members(ctx, strangerID, shop.ID); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("stranger could read member contact details")
	}
	if _, err = shopRepo.Get(ctx, strangerID, shop.ID); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("stranger could read shop")
	}
	edit := shopModels.UpdateShopRequest{Name: "Edited Shop", BusinessType: "spa"}
	if _, err = shopRepo.Update(ctx, strangerID, shop.ID, edit); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("stranger could update shop")
	}
	var staffID string
	if err = pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash,full_name,system_role,status) VALUES('staff-test','staff@example.test','unused','Staff','user','active') RETURNING id::text`).Scan(&staffID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'staff','active')`, shop.ID, staffID); err != nil {
		t.Fatal(err)
	}
	if _, err = shopRepo.Get(ctx, staffID, shop.ID); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("staff could access owner shop details")
	}
	if _, err = shopRepo.Members(ctx, staffID, shop.ID); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("staff could read owner-only member list")
	}
	memberList, err = shopRepo.Members(ctx, result.UserID, shop.ID)
	if err != nil || memberList.ActiveCount != 2 || len(memberList.Members) != 2 {
		t.Fatalf("active member count is incorrect: %v", err)
	}
	if _, err = shopRepo.Update(ctx, staffID, shop.ID, edit); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("staff could update shop")
	}
	inactiveStatus, activeStatus := "inactive", "active"
	toggle := shopModels.UpdateShopRequest{Name: shop.Name, BusinessType: shop.BusinessType, Status: &inactiveStatus}
	if closed, err := shopRepo.Update(ctx, result.UserID, shop.ID, toggle); err != nil || closed.Status != "inactive" {
		t.Fatalf("owner could not disable shop: %v", err)
	}
	if _, err = shopRepo.Get(ctx, result.UserID, shop.ID); err != nil {
		t.Fatal("owner cannot access disabled shop to reopen it")
	}
	if _, err = shopRepo.Get(ctx, staffID, shop.ID); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("staff can access disabled shop")
	}
	if preserved, err := shopRepo.Update(ctx, result.UserID, shop.ID, edit); err != nil || preserved.Status != "inactive" {
		t.Fatal("omitted status did not preserve inactive shop")
	}
	inactiveProfile, err := loginRepo.Profile(ctx, result.UserID)
	if err != nil || len(inactiveProfile.Tenants) != 1 || inactiveProfile.Tenants[0].Status != "inactive" {
		t.Fatal("owner profile lost disabled shop")
	}
	toggle.Status = &activeStatus
	if reopened, err := shopRepo.Update(ctx, result.UserID, shop.ID, toggle); err != nil || reopened.Status != "active" {
		t.Fatalf("owner could not reopen shop: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE tenant_members SET status='inactive' WHERE user_id=$1`, staffID); err != nil {
		t.Fatal(err)
	}
	if _, err = shopRepo.Get(ctx, staffID, shop.ID); !errors.Is(err, shopRepositories.ErrShopNotFound) {
		t.Fatal("inactive member could read shop")
	}
	memberList, err = shopRepo.Members(ctx, result.UserID, shop.ID)
	if err != nil || memberList.ActiveCount != 1 || len(memberList.Members) != 2 {
		t.Fatalf("inactive memberships should remain listed without counting toward active seats: %v", err)
	}
	if _, err = pool.Exec(ctx, `DELETE FROM tenant_members WHERE user_id=$1`, staffID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, staffID); err != nil {
		t.Fatal(err)
	}
	changed, err := shopRepo.Update(ctx, result.UserID, shop.ID, edit)
	if err != nil || changed.Name != edit.Name || changed.BusinessType != "spa" || changed.Slug != shop.Slug {
		t.Fatalf("owner update failed: %v", err)
	}
	if _, err = shopRepo.Create(ctx, result.UserID, shopInput); !errors.Is(err, shopRepositories.ErrShopLimit) {
		t.Fatal("second shop was not rejected")
	}
	packageRepo := packageRepositories.NewPackageRepository(pool)
	plans, err := packageRepo.List(ctx, true)
	if err != nil || len(plans) != 1 || plans[0].ShopLimit != 1 {
		t.Fatalf("package shop limit lookup failed: %v", err)
	}
	plan := plans[0]
	plan.ShopLimit = 2
	if _, err = packageRepo.Save(ctx, plan, plan.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET shop_limit=2`); err != nil {
		t.Fatal(err)
	}
	shopInput.Slug = "second-test-shop"
	if _, err = shopRepo.Create(ctx, result.UserID, shopInput); err != nil {
		t.Fatalf("higher shop quota was not honored: %v", err)
	}
	if _, err = shopRepo.Create(ctx, result.UserID, shopInput); !errors.Is(err, shopRepositories.ErrShopLimit) {
		t.Fatal("shop quota was exceeded")
	}
	profile, err = loginRepo.Profile(ctx, result.UserID)
	if err != nil || profile.OwnedShopCount != 2 || profile.Subscription.ShopLimit != 2 {
		t.Fatalf("shop counter is incorrect: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET status='expired'`); err != nil {
		t.Fatal(err)
	}
	if _, err = shopRepo.Create(ctx, result.UserID, shopInput); !errors.Is(err, shopRepositories.ErrSubscriptionRequired) {
		t.Fatal("expired subscription was accepted")
	}
	if _, err = pool.Exec(ctx, `UPDATE subscriptions SET status='active'`); err != nil {
		t.Fatal(err)
	}
	var users, subscriptions int
	if err = pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM subscriptions WHERE monthly_price_satang=0 AND ends_at IS NULL)`).Scan(&users, &subscriptions); err != nil {
		t.Fatal(err)
	}
	if users != 1 || subscriptions != 1 {
		t.Fatal("registration did not create exactly one account and free subscription")
	}
	if _, err = repo.Register(ctx, input); !errors.Is(err, ErrAccountExists) {
		t.Fatal("duplicate account was not rejected")
	}
	if _, err = pool.Exec(ctx, `UPDATE packages SET status='inactive' WHERE code='free'`); err != nil {
		t.Fatal(err)
	}
	input.Email = "another@example.com"
	input.Username = input.Email
	if _, err = repo.Register(ctx, input); !errors.Is(err, ErrFreePackageUnavailable) {
		t.Fatal("unavailable free plan was not rejected")
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&users); err != nil {
		t.Fatal(err)
	}
	if users != 1 {
		t.Fatal("account was not rolled back when free subscription failed")
	}
}
