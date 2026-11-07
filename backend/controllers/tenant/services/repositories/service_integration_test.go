package repositories

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/services/models"
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceAccessIntegration(t *testing.T) {
	if os.Getenv("TENANT_REGISTRATION_DB_TEST") != "1" {
		t.Skip("set TENANT_REGISTRATION_DB_TEST=1 for PostgreSQL temporary-table tests")
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
	cfg := utils.DB.Config()
	cfg.MaxConns = 1
	cfg.MinConns = 0
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, table := range []string{"users", "tenants", "tenant_members", "services"} {
		exec("CREATE TEMP TABLE " + table + " (LIKE public." + table + " INCLUDING ALL)")
	}
	user := func(name string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash,full_name,phone,system_role,status) VALUES($1,$1,'unused','Test','0812345678','user','active') RETURNING id::text`, name+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	owner, staff, other := user("owner"), user("staff"), user("other")
	shop := func(slug string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO tenants(name,slug,business_type,status) VALUES('Test',$1,'salon','active') RETURNING id::text`, slug).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first, second := shop("first"), shop("second")
	exec(`INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'owner','active'),($1,$3,'staff','active'),($4,$5,'owner','active')`, first, owner, staff, second, other)
	repo := NewServiceRepository(pool)
	input := models.ServiceRequest{Name: "Haircut", Description: "Test", DurationMinutes: 30, PriceSatang: 29950, Status: "active"}
	active, err := repo.Save(ctx, owner, first, "", input)
	if err != nil || active.PriceSatang != 29950 {
		t.Fatalf("create: %+v %v", active, err)
	}
	input.Status = "inactive"
	input.PriceSatang = 0
	inactive, err := repo.Save(ctx, owner, first, "", input)
	if err != nil {
		t.Fatal(err)
	}
	ownerList, err := repo.List(ctx, owner, first)
	if err != nil || len(ownerList.Items) != 2 || ownerList.Role != "owner" {
		t.Fatalf("owner list: %+v %v", ownerList, err)
	}
	staffList, err := repo.List(ctx, staff, first)
	if err != nil || len(staffList.Items) != 1 || staffList.Items[0].ID != active.ID {
		t.Fatalf("staff list: %+v %v", staffList, err)
	}
	denied := func(err error) {
		t.Helper()
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected access denied, got %v", err)
		}
	}
	_, err = repo.Save(ctx, staff, first, "", input)
	denied(err)
	_, err = repo.Save(ctx, staff, first, active.ID, input)
	denied(err)
	_, err = repo.List(ctx, other, first)
	denied(err)
	_, err = repo.Save(ctx, other, first, "", input)
	denied(err)
	_, err = repo.Save(ctx, other, second, active.ID, input)
	denied(err)
	input.Status = "active"
	input.Name = "Updated"
	updated, err := repo.Save(ctx, owner, first, inactive.ID, input)
	if err != nil || updated.Name != "Updated" || updated.PriceSatang != 0 {
		t.Fatalf("update: %+v %v", updated, err)
	}
	staffList, err = repo.List(ctx, staff, first)
	if err != nil || len(staffList.Items) != 2 {
		t.Fatal("reactivated service not visible", err)
	}
	exec(`UPDATE tenants SET status='inactive' WHERE id=$1`, first)
	_, err = repo.List(ctx, staff, first)
	denied(err)
	if _, err = repo.Save(ctx, owner, first, active.ID, input); err != nil {
		t.Fatal("inactive shop owner cannot edit", err)
	}
	exec(`UPDATE tenants SET status='suspended' WHERE id=$1`, first)
	_, err = repo.List(ctx, owner, first)
	denied(err)
	_, err = repo.Save(ctx, owner, first, "", input)
	denied(err)
	exec(`UPDATE tenants SET status='active' WHERE id=$1`, first)
	exec(`UPDATE tenant_members SET status='inactive' WHERE user_id=$1`, staff)
	_, err = repo.List(ctx, staff, first)
	denied(err)
	exec(`UPDATE users SET status='suspended' WHERE id=$1`, owner)
	_, err = repo.Save(ctx, owner, first, active.ID, input)
	denied(err)
}
