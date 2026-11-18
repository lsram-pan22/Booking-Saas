package repositories

import (
	"booking-saas/controllers/tenant/bookings/models"
	"booking-saas/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"
)

// Called from the existing isolated-schema integration test. The staff member
// has an active personal subscription throughout; only the owner's plan counts.
func testBookingSubscription(t *testing.T, ctx context.Context, pool *pgxpool.Pool, repo *BookingRepository, owner, staff, shop, service, member string, at time.Time) {
	t.Helper()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	request := models.CreateBookingRequest{ServiceID: service, StaffMemberID: member, CustomerName: "Subscription test", CustomerPhone: "0812345678", StartsAt: at}
	existing, err := repo.Create(ctx, owner, shop, request)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, sql string }{
		{"past end while still active", `UPDATE subscriptions SET status='active',starts_at=NOW()-INTERVAL '1 day',ends_at=NOW()-INTERVAL '1 second' WHERE user_id=$1`},
		{"expired status", `UPDATE subscriptions SET status='expired',ends_at=NULL WHERE user_id=$1`},
		{"future start", `UPDATE subscriptions SET status='active',starts_at=NOW()+INTERVAL '1 day',ends_at=NULL WHERE user_id=$1`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(tc.sql, owner)
			for _, actor := range []string{owner, staff} {
				request.StartsAt = at.Add(time.Hour)
				if _, err := repo.Create(ctx, actor, shop, request); !errors.Is(err, utils.ErrSubscriptionRequired) {
					t.Fatalf("create after expiry: %v", err)
				}
				change := models.UpdateBookingRequest{CustomerRequest: models.CustomerRequest{CustomerName: "Changed", CustomerPhone: "0812345678"}}
				if _, err := repo.Update(ctx, actor, shop, existing.ID, change); !errors.Is(err, utils.ErrSubscriptionRequired) {
					t.Fatalf("edit after expiry: %v", err)
				}
				if _, err := repo.UpdateCustomer(ctx, actor, shop, existing.ID, change.CustomerRequest); !errors.Is(err, utils.ErrSubscriptionRequired) {
					t.Fatalf("legacy customer endpoint bypass: %v", err)
				}
				list, err := repo.List(ctx, actor, shop, at)
				if err != nil || list.SubscriptionActive || len(list.Items) != 1 || !list.Items[0].CanManage {
					t.Fatalf("expired read/status permission: %+v %v", list, err)
				}
			}
		})
	}
	// Operational status transitions are intentionally still available.
	if _, err = repo.SetStatus(ctx, staff, shop, existing.ID, "confirmed", ""); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetStatus(ctx, staff, shop, existing.ID, "completed", ""); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE subscriptions SET status='active',starts_at=NOW()-INTERVAL '1 day',ends_at=NULL WHERE user_id=$1`, owner)
	request.StartsAt = at.Add(2 * time.Hour)
	toCancel, err := repo.Create(ctx, owner, shop, request)
	if err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE subscriptions SET ends_at=NOW()-INTERVAL '1 second' WHERE user_id=$1`, owner)
	if _, err = repo.SetStatus(ctx, staff, shop, toCancel.ID, "cancelled", "Customer cancelled"); err != nil {
		t.Fatal(err)
	}
	// Missing owner subscription also fails closed despite the staff's own plan.
	exec(`DELETE FROM subscriptions WHERE user_id=$1`, owner)
	if _, err = repo.Create(ctx, staff, shop, request); !errors.Is(err, utils.ErrSubscriptionRequired) {
		t.Fatalf("staff plan granted shop rights: %v", err)
	}
	exec(`INSERT INTO subscriptions(user_id) VALUES($1)`, owner)
	request.StartsAt = at.Add(3 * time.Hour)
	if _, err = repo.Create(ctx, staff, shop, request); err != nil {
		t.Fatalf("renewed subscription did not restore rights: %v", err)
	}
}
