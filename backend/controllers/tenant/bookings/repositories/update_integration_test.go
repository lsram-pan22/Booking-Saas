package repositories

import (
	"booking-saas/controllers/tenant/bookings/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"
)

// Runs in the isolated schema created by TestBookingIntegration.
func testBookingReschedule(t *testing.T, ctx context.Context, pool *pgxpool.Pool, repo *BookingRepository, owner, staff, other, shopID, serviceID, staffID, otherServiceID string) {
	t.Helper()
	expect := func(err, want error) {
		t.Helper()
		if !errors.Is(err, want) {
			t.Fatalf("reschedule error %v, want %v", err, want)
		}
	}
	zone := time.FixedZone("Bangkok", 7*3600)
	future := time.Now().In(zone).AddDate(0, 0, 10)
	day := time.Date(future.Year(), future.Month(), future.Day(), 0, 0, 0, 0, zone)
	req := models.CreateBookingRequest{ServiceID: serviceID, StaffMemberID: staffID, CustomerName: "Reschedule", CustomerPhone: "0812345678", StartsAt: day.Add(10 * time.Hour)}
	booking, err := repo.Create(ctx, owner, shopID, req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetStatus(ctx, owner, shopID, booking.ID, "confirmed", ""); err != nil {
		t.Fatal(err)
	}
	input := models.UpdateBookingRequest{CustomerRequest: models.CustomerRequest{CustomerName: booking.CustomerName, CustomerPhone: booking.CustomerPhone, Note: booking.Note}, ServiceID: &serviceID, StartsAt: &booking.StartsAt}
	assertNoChange := func() {
		t.Helper()
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM booking_edit_history WHERE booking_id=$1`, booking.ID).Scan(&count); err != nil || count != 0 {
			t.Fatal("rejected or unchanged edit wrote history", err, count)
		}
	}
	if _, err = repo.Update(ctx, staff, shopID, booking.ID, input); err != nil {
		t.Fatal("same booking conflicts with itself", err)
	}
	assertNoChange()
	blocker := req
	blocker.StartsAt = day.Add(11*time.Hour + 30*time.Minute)
	if _, err = repo.Create(ctx, owner, shopID, blocker); err != nil {
		t.Fatal(err)
	}
	conflicting := day.Add(11 * time.Hour)
	input.StartsAt = &conflicting
	_, err = repo.Update(ctx, staff, shopID, booking.ID, input)
	expect(err, ErrConflict)
	assertNoChange()
	outside := day.Add(17*time.Hour + 30*time.Minute)
	input.StartsAt = &outside
	_, err = repo.Update(ctx, owner, shopID, booking.ID, input)
	expect(err, ErrOutsideHours)
	assertNoChange()
	past := time.Now().Add(-time.Hour).Truncate(time.Minute)
	input.StartsAt = &past
	_, err = repo.Update(ctx, owner, shopID, booking.ID, input)
	expect(err, ErrInvalidSchedule)
	tooFar := time.Now().AddDate(3, 0, 0).Truncate(time.Minute)
	input.StartsAt = &tooFar
	_, err = repo.Update(ctx, owner, shopID, booking.ID, input)
	expect(err, ErrInvalidSchedule)
	moved := day.Add(9 * time.Hour)
	input.StartsAt = &moved
	_, err = repo.Update(ctx, other, shopID, booking.ID, input)
	expect(err, ErrNotFound)
	rescheduled, err := repo.Update(ctx, staff, shopID, booking.ID, input)
	if err != nil || rescheduled.Status != "confirmed" || !rescheduled.StartsAt.Equal(moved) || rescheduled.PriceSatang != booking.PriceSatang || rescheduled.DurationMinutes != booking.DurationMinutes || rescheduled.StaffMemberID != booking.StaffMemberID {
		t.Fatal("time edit changed snapshot/status/provider", err)
	}
	input.StartsAt = &moved
	input.ServiceID = &otherServiceID
	_, err = repo.Update(ctx, owner, shopID, booking.ID, input)
	expect(err, ErrUnavailable)
	var newService string
	if err = pool.QueryRow(ctx, `INSERT INTO services(tenant_id,name,duration_minutes,price_satang) VALUES($1,'New service',45,45000) RETURNING id::text`, shopID).Scan(&newService); err != nil {
		t.Fatal(err)
	}
	input.ServiceID = &newService
	changed, err := repo.Update(ctx, owner, shopID, booking.ID, input)
	if err != nil || changed.Status != "confirmed" || changed.ServiceID != newService || changed.ServiceName != "New service" || changed.PriceSatang != 45000 || changed.DurationMinutes != 45 || !changed.EndsAt.Equal(moved.Add(45*time.Minute)) {
		t.Fatal("new service did not refresh snapshot/end or changed status", err)
	}
	var oldPrice, newPrice int64
	var oldStatus, newStatus string
	if err = pool.QueryRow(ctx, `SELECT (before_data->>'price_satang')::bigint,(after_data->>'price_satang')::bigint,before_data->>'status',after_data->>'status' FROM booking_edit_history WHERE booking_id=$1 AND after_data->>'service_id'=$2`, booking.ID, newService).Scan(&oldPrice, &newPrice, &oldStatus, &newStatus); err != nil || oldPrice != booking.PriceSatang || newPrice != 45000 || oldStatus != "confirmed" || newStatus != "confirmed" {
		t.Fatal("edit audit lost price/status", err)
	}
	// A later price change must not alter this booking on a time-only edit.
	if _, err = pool.Exec(ctx, `UPDATE services SET duration_minutes=120,price_satang=99999,status='inactive' WHERE id=$1`, newService); err != nil {
		t.Fatal(err)
	}
	next := day.Add(10 * time.Hour)
	input.StartsAt = &next
	unchanged, err := repo.Update(ctx, staff, shopID, booking.ID, input)
	if err != nil || unchanged.PriceSatang != 45000 || unchanged.DurationMinutes != 45 || !unchanged.EndsAt.Equal(next.Add(45*time.Minute)) {
		t.Fatal("time-only edit repriced original service", err)
	}
	// Replacing with a different inactive service is rejected.
	input.ServiceID = &serviceID
	if _, err = pool.Exec(ctx, `UPDATE services SET status='inactive' WHERE id=$1`, serviceID); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Update(ctx, owner, shopID, booking.ID, input)
	expect(err, ErrUnavailable)
	if _, err = pool.Exec(ctx, `UPDATE services SET status='active' WHERE id=$1`, serviceID); err != nil {
		t.Fatal(err)
	}
	// Simultaneous moves into one provider's slot must have exactly one winner.
	var candidates []models.Booking
	for _, hour := range []int{13, 14} {
		candidate := req
		candidate.StartsAt = day.Add(time.Duration(hour) * time.Hour)
		b, err := repo.Create(ctx, owner, shopID, candidate)
		if err != nil {
			t.Fatal(err)
		}
		candidates = append(candidates, b)
	}
	target := day.Add(16 * time.Hour)
	gate := make(chan struct{})
	results := make(chan error, 2)
	for _, candidate := range candidates {
		go func(b models.Booking) {
			<-gate
			_, err := repo.Update(ctx, staff, shopID, b.ID, models.UpdateBookingRequest{CustomerRequest: models.CustomerRequest{CustomerName: b.CustomerName, CustomerPhone: b.CustomerPhone, Note: b.Note}, StartsAt: &target})
			results <- err
		}(candidate)
	}
	close(gate)
	wins, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			wins++
		} else if errors.Is(err, ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("concurrent moves: %d wins, %d conflicts", wins, conflicts)
	}
	input.ServiceID = &newService
	if _, err = repo.SetStatus(ctx, owner, shopID, booking.ID, "completed", ""); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Update(ctx, owner, shopID, booking.ID, input)
	expect(err, ErrReadOnly)
}
