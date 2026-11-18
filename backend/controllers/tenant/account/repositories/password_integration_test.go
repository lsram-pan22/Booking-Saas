package repositories

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/account/models"
	"booking-saas/middleware"
	"booking-saas/utils"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Explicit opt-in only; creates an isolated schema with synthetic accounts.
func TestPasswordResetIntegration(t *testing.T) {
	if os.Getenv("PASSWORD_RESET_DB_TEST") != "1" {
		t.Skip("set PASSWORD_RESET_DB_TEST=1 for isolated PostgreSQL password tests")
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
	schema := fmt.Sprintf("password_test_%d", time.Now().UnixNano())
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
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
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
	exec(`CREATE TABLE users(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),username TEXT,email TEXT,full_name TEXT,password_hash TEXT,system_role TEXT,status TEXT);
 CREATE TABLE tenants(id UUID PRIMARY KEY DEFAULT gen_random_uuid(),status TEXT);
 CREATE TABLE tenant_members(tenant_id UUID,user_id UUID,role TEXT,status TEXT);
 CREATE TABLE subscriptions(user_id UUID);`)
	migration, err := os.ReadFile("../../../../migrations/019_password_reset.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(string(migration))
	expiryMigration, err := os.ReadFile("../../../../migrations/021_temporary_password_expiry.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(string(expiryMigration))
	hash, err := utils.HashPassword("original-password")
	if err != nil {
		t.Fatal(err)
	}
	user := func(role string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO users(username,email,full_name,password_hash,system_role,status) VALUES(gen_random_uuid()::text,'test@example.test','Test',$1,$2,'active') RETURNING id::text`, hash, role).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	owner, staff, stranger, admin := user("user"), user("user"), user("user"), user("super_admin")
	var shop string
	if err = pool.QueryRow(ctx, `INSERT INTO tenants(status) VALUES('active') RETURNING id::text`).Scan(&shop); err != nil {
		t.Fatal(err)
	}
	exec(`INSERT INTO tenant_members VALUES($1,$2,'owner','active'),($1,$3,'staff','active')`, shop, owner, staff)
	repo := NewPasswordRepository(pool)
	temp, err := utils.HashPassword("temporary-password")
	if err != nil {
		t.Fatal(err)
	}
	actor := models.PasswordActor{ID: owner, Role: "user"}
	if err = repo.Reset(ctx, actor, shop, staff, temp, "incorrect-owner-password"); !errors.Is(err, models.ErrCurrentPassword) {
		t.Fatal("actor password confirmation bypassed", err)
	}
	denied := func(actor models.PasswordActor, shop, target string) {
		t.Helper()
		if err := repo.Reset(ctx, actor, shop, target, temp, "original-password"); !errors.Is(err, models.ErrPasswordDenied) {
			t.Fatalf("unauthorized reset: %v", err)
		}
	}
	denied(models.PasswordActor{ID: staff, Role: "user"}, shop, owner)
	denied(actor, shop, stranger)
	denied(models.PasswordActor{ID: stranger, Role: "user"}, shop, staff)
	denied(models.PasswordActor{ID: admin, Role: "super_admin"}, "", staff)
	// A staff member with another shop membership is not resettable by this owner.
	exec(`INSERT INTO tenant_members VALUES(gen_random_uuid(),$1,'staff','active')`, staff)
	denied(actor, shop, staff)
	exec(`DELETE FROM tenant_members WHERE user_id=$1 AND tenant_id<>$2::uuid`, staff, shop)
	if err = repo.Reset(ctx, actor, shop, staff, temp, "original-password"); err != nil {
		t.Fatal(err)
	}
	var version int64
	var must bool
	var saved string
	state := func() {
		t.Helper()
		if err := pool.QueryRow(ctx, `SELECT session_version,must_change_password,password_hash FROM users WHERE id=$1`, staff).Scan(&version, &must, &saved); err != nil {
			t.Fatal(err)
		}
	}
	state()
	if version != 1 || !must || !utils.CheckPassword("temporary-password", saved) {
		t.Fatal("reset did not atomically revoke sessions and store temporary password")
	}
	next, err := utils.HashPassword("new-personal-password")
	if err != nil {
		t.Fatal(err)
	}
	staffActor := models.PasswordActor{ID: staff, Role: "user", Version: 1}
	if err = repo.Change(ctx, staffActor, "wrong-password", "new-personal-password", next); !errors.Is(err, models.ErrCurrentPassword) {
		t.Fatal("wrong current password accepted")
	}
	if err = repo.Change(ctx, models.PasswordActor{ID: staff, Role: "user"}, "temporary-password", "new-personal-password", next); !errors.Is(err, models.ErrPasswordDenied) {
		t.Fatal("revoked token accepted")
	}
	if err = repo.Change(ctx, staffActor, "temporary-password", "new-personal-password", next); err != nil {
		t.Fatal(err)
	}
	state()
	if version != 2 || must || !utils.CheckPassword("new-personal-password", saved) {
		t.Fatal("password change state incorrect")
	}
	if err = repo.Change(ctx, staffActor, "new-personal-password", "temporary-password", temp); !errors.Is(err, models.ErrPasswordDenied) {
		t.Fatal("change token was reusable")
	}
	if err = repo.Reset(ctx, models.PasswordActor{ID: admin, Role: "super_admin"}, "", owner, temp, "original-password"); err != nil {
		t.Fatal(err)
	}
	// An owner who must change their own password cannot reset staff.
	denied(models.PasswordActor{ID: owner, Role: "user", Version: 1}, shop, staff)
	var events int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM password_events`).Scan(&events); err != nil || events != 3 {
		t.Fatalf("audit count %d: %v", events, err)
	}
	exec(`UPDATE users SET password_reset_expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, owner)
	if err = repo.Change(ctx, models.PasswordActor{ID: owner, Role: "user", Version: 1}, "temporary-password", "new-personal-password", next); !errors.Is(err, models.ErrPasswordDenied) {
		t.Fatal("expired temporary password accepted", err)
	}
	// Exercise actual session/membership queries against the isolated tables.
	const secret = "integration-session-revocation-secret"
	oldToken, err := utils.GenerateJWT(secret, time.Hour, staff, "staff@example.test", "user", 2)
	if err != nil {
		t.Fatal(err)
	}
	revoked, err := utils.RevokeAccountSessions(ctx, pool, &utils.JWTClaims{UserID: staff, Role: "user", SessionVersion: 2})
	if err != nil || !revoked {
		t.Fatal("logout failed", err)
	}
	newToken, err := utils.GenerateJWT(secret, time.Hour, staff, "staff@example.test", "user", 3)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(middleware.RequireJWT(secret), middleware.RequireSession(pool, false, true))
	router.GET("/shops/:shopID", middleware.RequireShopAccess(pool, false), func(c *gin.Context) { c.Status(200) })
	router.GET("/manage/:shopID", middleware.RequireShopAccess(pool, true), func(c *gin.Context) { c.Status(200) })
	request := func(token, path string, want int) {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s status %d want %d", path, rec.Code, want)
		}
	}
	request(oldToken, "/shops/"+shop, 401)
	request(newToken, "/shops/"+shop, 200)
	request(newToken, "/shops/99999999-9999-4999-8999-999999999999", 404)
	request(newToken, "/manage/"+shop, 404)
	exec(`UPDATE tenant_members SET status='inactive' WHERE user_id=$1`, staff)
	request(newToken, "/shops/"+shop, 404)
}
