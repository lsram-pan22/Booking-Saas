package repositories

import (
	"booking-saas/config"
	"booking-saas/controllers/tenant/bookings/models"
	shopModels "booking-saas/controllers/tenant/shops/models"
	shopRepositories "booking-saas/controllers/tenant/shops/repositories"
	"booking-saas/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBookingIntegration(t *testing.T) {
	if os.Getenv("TENANT_REGISTRATION_DB_TEST") != "1" {
		t.Skip("set TENANT_REGISTRATION_DB_TEST=1 for isolated PostgreSQL tests")
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// A generated, isolated schema allows multiple connections to exercise races.
	var entropy [12]byte
	if _, err = rand.Read(entropy[:]); err != nil {
		t.Fatal(err)
	}
	schema := "booking_test_" + hex.EncodeToString(entropy[:])
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err = utils.DB.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := utils.DB.Exec(context.Background(), "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
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
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, table := range []string{"users", "tenants", "tenant_members", "services", "bookings", "booking_status_history", "booking_edit_history"} {
		exec("CREATE TABLE " + table + " (LIKE public." + table + " INCLUDING ALL)")
	}
	newUser := func(name string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO users(username,email,password_hash,full_name,phone,system_role,status) VALUES($1,$1,'unused',$1,'0812345678','user','active') RETURNING id::text`, name+"@example.test").Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	owner, staff, other := newUser("owner"), newUser("staff"), newUser("other")
	newShop := func(slug string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO tenants(name,slug,business_type,status) VALUES('Test',$1,'salon','active') RETURNING id::text`, slug).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	first, second := newShop("first"), newShop("second")
	exec(`INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'owner','active'),($1,$3,'staff','active'),($4,$5,'owner','active')`, first, owner, staff, second, other)
	var staffID, ownerMemberID, otherStaffID, serviceID, otherServiceID string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM tenant_members WHERE user_id=$1`, staff).Scan(&staffID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT id::text FROM tenant_members WHERE user_id=$1`, owner).Scan(&ownerMemberID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT id::text FROM tenant_members WHERE user_id=$1`, other).Scan(&otherStaffID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO services(tenant_id,name,duration_minutes,price_satang) VALUES($1,'Haircut',30,29950) RETURNING id::text`, first).Scan(&serviceID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO services(tenant_id,name,duration_minutes,price_satang) VALUES($1,'Other',30,0) RETURNING id::text`, second).Scan(&otherServiceID); err != nil {
		t.Fatal(err)
	}
	zone := time.FixedZone("Bangkok", 7*3600)
	day := time.Date(2030, 1, 2, 0, 0, 0, 0, zone)
	req := models.CreateBookingRequest{ServiceID: serviceID, StaffMemberID: staffID, CustomerName: "Customer", CustomerPhone: "0812345678", StartsAt: day.Add(10 * time.Hour)}
	repo := NewBookingRepository(pool)
	firstBooking, err := repo.Create(ctx, owner, first, req)
	if err != nil || firstBooking.Status != "pending" || firstBooking.PriceSatang != 29950 || !firstBooking.EndsAt.Equal(req.StartsAt.Add(30*time.Minute)) {
		t.Fatalf("create: %+v %v", firstBooking, err)
	}
	expect := func(err, want error) {
		t.Helper()
		if !errors.Is(err, want) {
			t.Fatalf("error %v, want %v", err, want)
		}
	}
	_, err = repo.Create(ctx, staff, first, req)
	expect(err, ErrConflict)
	overlapping := req
	overlapping.StartsAt = req.StartsAt.Add(-15 * time.Minute)
	_, err = repo.Create(ctx, staff, first, overlapping)
	expect(err, ErrConflict)
	adjacent := req
	adjacent.StartsAt = firstBooking.EndsAt
	if _, err = repo.Create(ctx, staff, first, adjacent); err != nil {
		t.Fatal("adjacent booking blocked", err)
	}
	another := req
	another.StaffMemberID = ownerMemberID
	_, err = repo.Create(ctx, staff, first, another)
	expect(err, ErrUnavailable)
	if _, err = repo.Create(ctx, owner, first, another); err != nil {
		t.Fatal("another provider blocked", err)
	}
	_, err = repo.Create(ctx, other, first, req)
	expect(err, ErrNotFound)
	_, err = repo.List(ctx, other, first, day)
	expect(err, ErrNotFound)
	_, err = repo.SetStatus(ctx, other, second, firstBooking.ID, "confirmed", "")
	expect(err, ErrNotFound)
	invalid := req
	invalid.ServiceID = otherServiceID
	_, err = repo.Create(ctx, owner, first, invalid)
	expect(err, ErrUnavailable)
	invalid = req
	invalid.StaffMemberID = otherStaffID
	_, err = repo.Create(ctx, owner, first, invalid)
	expect(err, ErrUnavailable)
	exec(`UPDATE services SET name='Changed',duration_minutes=60,price_satang=99900 WHERE id=$1`, serviceID)
	list, err := repo.List(ctx, staff, first, day)
	if err != nil || len(list.Items) != 3 || len(list.Services) != 1 || len(list.Staff) != 1 || list.Staff[0].ID != staffID || list.Role != "staff" {
		t.Fatalf("list: %+v %v", list, err)
	}
	ownerList, err := repo.List(ctx, owner, first, day)
	if err != nil || ownerList.Role != "owner" || len(ownerList.Staff) != 2 {
		t.Fatalf("owner provider choices: %+v %v", ownerList, err)
	}
	for _, b := range list.Items {
		if b.CanManage != (b.StaffMemberID == staffID) {
			t.Fatal("staff management capability does not match assignment")
		}
		if b.ID == firstBooking.ID && (b.ServiceName != "Haircut" || b.PriceSatang != 29950 || b.DurationMinutes != 30) {
			t.Fatal("snapshot changed")
		}
	}
	for _, b := range ownerList.Items {
		if !b.CanManage {
			t.Fatal("owner cannot manage all shop bookings")
		}
	}
	_, err = repo.SetStatus(ctx, owner, first, firstBooking.ID, "completed", "")
	expect(err, ErrTransition)
	customerInput := models.CustomerRequest{CustomerName: "Updated Customer", CustomerPhone: "0891234567", Note: "Customer preference"}
	_, err = repo.UpdateCustomer(ctx, other, second, firstBooking.ID, customerInput)
	expect(err, ErrNotFound)
	_, err = repo.UpdateCustomer(ctx, other, first, firstBooking.ID, customerInput)
	expect(err, ErrNotFound)
	edited, err := repo.UpdateCustomer(ctx, staff, first, firstBooking.ID, customerInput)
	if err != nil || edited.CustomerName != customerInput.CustomerName || edited.CustomerPhone != customerInput.CustomerPhone || edited.Note != customerInput.Note || !edited.CanManage {
		t.Fatal("assigned staff customer edit", err)
	}
	if edited.ServiceID != firstBooking.ServiceID || edited.StaffMemberID != firstBooking.StaffMemberID || edited.ServiceName != firstBooking.ServiceName || edited.DurationMinutes != firstBooking.DurationMinutes || edited.PriceSatang != firstBooking.PriceSatang || !edited.StartsAt.Equal(firstBooking.StartsAt) || !edited.EndsAt.Equal(firstBooking.EndsAt) || edited.Status != "pending" {
		t.Fatal("customer edit changed scheduling, pricing or status")
	}
	if _, err = repo.UpdateCustomer(ctx, staff, first, firstBooking.ID, customerInput); err != nil {
		t.Fatal("no-op customer edit", err)
	}
	var editCount int
	var editor, beforeName, afterName string
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM booking_edit_history WHERE booking_id=$1`, firstBooking.ID).Scan(&editCount); err != nil || editCount != 1 {
		t.Fatal("no-op created duplicate history", err)
	}
	if err = pool.QueryRow(ctx, `SELECT changed_by::text,before_data->>'customer_name',after_data->>'customer_name' FROM booking_edit_history WHERE booking_id=$1`, firstBooking.ID).Scan(&editor, &beforeName, &afterName); err != nil || editor != staff || beforeName != firstBooking.CustomerName || afterName != customerInput.CustomerName {
		t.Fatal("customer audit data incorrect", err)
	}
	if _, err = repo.SetStatus(ctx, staff, first, firstBooking.ID, "confirmed", ""); err != nil {
		t.Fatal(err)
	}
	customerInput.Note = ""
	if edited, err = repo.UpdateCustomer(ctx, owner, first, firstBooking.ID, customerInput); err != nil || edited.Note != "" || edited.Status != "confirmed" {
		t.Fatal("owner edit/clear note on confirmed booking", err)
	}
	if _, err = repo.SetStatus(ctx, owner, first, firstBooking.ID, "completed", ""); err != nil {
		t.Fatal(err)
	}
	_, err = repo.UpdateCustomer(ctx, owner, first, firstBooking.ID, customerInput)
	expect(err, ErrReadOnly)
	_, err = repo.SetStatus(ctx, owner, first, firstBooking.ID, "cancelled", "Customer cancelled")
	expect(err, ErrTransition)
	var historyCount int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM booking_status_history WHERE booking_id=$1`, firstBooking.ID).Scan(&historyCount); err != nil || historyCount != 3 {
		t.Fatal("missing status history", err, historyCount)
	}
	// Staff must never mutate another provider's bookings, even via direct API IDs.
	teammate := newUser("teammate")
	var teammateID string
	if err = pool.QueryRow(ctx, `INSERT INTO tenant_members(tenant_id,user_id,role,status) VALUES($1,$2,'staff','active') RETURNING id::text`, first, teammate).Scan(&teammateID); err != nil {
		t.Fatal(err)
	}
	assigned := req
	assigned.StaffMemberID = teammateID
	assigned.StartsAt = day.Add(15 * time.Hour)
	otherBooking, err := repo.Create(ctx, owner, first, assigned)
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.UpdateCustomer(ctx, staff, first, otherBooking.ID, customerInput)
	expect(err, ErrNotFound)
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM booking_edit_history WHERE booking_id=$1`, otherBooking.ID).Scan(&editCount); err != nil || editCount != 0 {
		t.Fatal("denied edit added customer history", err)
	}
	for _, status := range []string{"confirmed", "completed", "cancelled"} {
		_, err = repo.SetStatus(ctx, staff, first, otherBooking.ID, status, "unauthorized cancellation")
		expect(err, ErrNotFound)
	}
	var unchangedStatus, unchangedRemark string
	if err = pool.QueryRow(ctx, `SELECT status,remark FROM bookings WHERE id=$1`, otherBooking.ID).Scan(&unchangedStatus, &unchangedRemark); err != nil || unchangedStatus != "pending" || unchangedRemark != "" {
		t.Fatal("denied request mutated booking", err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM booking_status_history WHERE booking_id=$1`, otherBooking.ID).Scan(&historyCount); err != nil || historyCount != 1 {
		t.Fatal("denied request added status history", err)
	}
	if _, err = repo.SetStatus(ctx, owner, first, otherBooking.ID, "confirmed", ""); err != nil {
		t.Fatal("owner cannot confirm assigned staff work", err)
	}
	for _, status := range []string{"completed", "cancelled"} {
		_, err = repo.SetStatus(ctx, staff, first, otherBooking.ID, status, "not assigned")
		expect(err, ErrNotFound)
	}
	if _, err = repo.SetStatus(ctx, teammate, first, otherBooking.ID, "completed", ""); err != nil {
		t.Fatal("assigned staff cannot complete own work", err)
	}
	assigned.StartsAt = day.Add(16 * time.Hour)
	toCancel, err := repo.Create(ctx, owner, first, assigned)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetStatus(ctx, owner, first, toCancel.ID, "cancelled", "Owner cancelled assigned work"); err != nil {
		t.Fatal("owner cannot cancel staff work", err)
	}
	_, err = repo.UpdateCustomer(ctx, owner, first, toCancel.ID, customerInput)
	expect(err, ErrReadOnly)
	// Two independent connections attempt the same slot simultaneously.
	raceReq := req
	raceReq.StartsAt = day.Add(12 * time.Hour)
	gate := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { <-gate; _, err := repo.Create(ctx, staff, first, raceReq); results <- err }()
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
		t.Fatalf("race: %d created, %d conflicts", wins, conflicts)
	}
	var raceID string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM bookings WHERE starts_at=$1 AND staff_member_id=$2`, raceReq.StartsAt, staffID).Scan(&raceID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SetStatus(ctx, staff, first, raceID, "cancelled", "ลูกค้าขอยกเลิก"); err != nil {
		t.Fatal(err)
	}
	var savedReason, historyReason string
	if err = pool.QueryRow(ctx, `SELECT remark FROM bookings WHERE id=$1`, raceID).Scan(&savedReason); err != nil || savedReason != "ลูกค้าขอยกเลิก" {
		t.Fatal("missing booking cancellation reason", err)
	}
	if err = pool.QueryRow(ctx, `SELECT remark FROM booking_status_history WHERE booking_id=$1 AND to_status='cancelled'`, raceID).Scan(&historyReason); err != nil || historyReason != savedReason {
		t.Fatal("missing history cancellation reason", err)
	}
	afterCancel, err := repo.List(ctx, owner, first, day)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range afterCancel.Items {
		if b.ID == raceID && b.Remark != savedReason {
			t.Fatal("reason lost on list reload")
		}
	}
	if _, err = repo.Create(ctx, owner, first, raceReq); err != nil {
		t.Fatal("cancelled slot not released", err)
	}
	// A booking spanning midnight appears in both affected Thai dates.
	overnight := req
	overnight.StartsAt = day.Add(23*time.Hour + 30*time.Minute)
	if _, err = repo.Create(ctx, owner, first, overnight); err != nil {
		t.Fatal(err)
	}
	nextDay, err := repo.List(ctx, staff, first, day.AddDate(0, 0, 1))
	if err != nil || len(nextDay.Items) != 1 {
		t.Fatal("midnight/day boundary", err, len(nextDay.Items))
	}
	opening, closing := "09:00", "18:00"
	shopRepo := shopRepositories.NewShopRepository(pool)
	updatedShop, err := shopRepo.Update(ctx, owner, first, shopModels.UpdateShopRequest{Name: "Test", BusinessType: "salon", OpeningTime: &opening, ClosingTime: &closing})
	if err != nil || updatedShop.OpeningTime == nil || *updatedShop.OpeningTime != opening {
		t.Fatal("shop opening hours update", err)
	}
	loadedShop, err := shopRepo.Get(ctx, owner, first)
	if err != nil || loadedShop.ClosingTime == nil || *loadedShop.ClosingTime != closing {
		t.Fatal("shop opening hours reload", err)
	}
	_, err = shopRepo.Update(ctx, staff, first, shopModels.UpdateShopRequest{Name: "Test", BusinessType: "salon", OpeningTime: &opening, ClosingTime: &closing})
	expect(err, shopRepositories.ErrShopNotFound)
	early := req
	early.StartsAt = day.Add(8 * time.Hour)
	_, err = repo.Create(ctx, owner, first, early)
	expect(err, ErrOutsideHours)
	late := req
	late.StartsAt = day.Add(17*time.Hour + 30*time.Minute)
	_, err = repo.Create(ctx, owner, first, late)
	expect(err, ErrOutsideHours)
	lastSlot := req
	lastSlot.StartsAt = day.Add(17 * time.Hour)
	if _, err = repo.Create(ctx, owner, first, lastSlot); err != nil {
		t.Fatal("booking ending at closing rejected", err)
	}
	sixth := req
	sixth.StartsAt = day.AddDate(0, 0, 6).Add(10 * time.Hour)
	withinWeek, err := repo.Create(ctx, owner, first, sixth)
	if err != nil {
		t.Fatal(err)
	}
	seventh := req
	seventh.StartsAt = day.AddDate(0, 0, 7).Add(10 * time.Hour)
	afterWeek, err := repo.Create(ctx, owner, first, seventh)
	if err != nil {
		t.Fatal(err)
	}
	week, err := repo.ListWeek(ctx, staff, first, day)
	if err != nil || week.OpeningTime == nil || *week.OpeningTime != opening || week.ClosingTime == nil || *week.ClosingTime != closing {
		t.Fatal("weekly hours", err)
	}
	found := false
	for _, b := range week.Items {
		if b.ID == withinWeek.ID {
			found = true
		}
		if b.ID == afterWeek.ID {
			t.Fatal("eighth day leaked into week")
		}
	}
	if !found {
		t.Fatal("seventh calendar day missing")
	}
	testBookingReschedule(t, ctx, pool, repo, owner, staff, other, first, serviceID, staffID, otherServiceID)
	exec(`UPDATE services SET status='inactive' WHERE id=$1`, serviceID)
	_, err = repo.Create(ctx, owner, first, req)
	expect(err, ErrUnavailable)
	exec(`UPDATE services SET status='active' WHERE id=$1`, serviceID)
	exec(`UPDATE tenant_members SET status='inactive' WHERE id=$1`, staffID)
	_, err = repo.UpdateCustomer(ctx, staff, first, withinWeek.ID, customerInput)
	expect(err, ErrNotFound)
	_, err = repo.SetStatus(ctx, staff, first, withinWeek.ID, "confirmed", "")
	expect(err, ErrNotFound)
	_, err = repo.Create(ctx, owner, first, req)
	expect(err, ErrUnavailable)
	_, err = repo.List(ctx, staff, first, day)
	expect(err, ErrNotFound)
	exec(`UPDATE tenants SET status='inactive' WHERE id=$1`, first)
	_, err = repo.UpdateCustomer(ctx, owner, first, withinWeek.ID, customerInput)
	expect(err, ErrNotFound)
	inactiveList, err := repo.List(ctx, owner, first, day)
	if err != nil {
		t.Fatal("owner cannot read inactive shop", err)
	}
	for _, b := range inactiveList.Items {
		if b.CanManage {
			t.Fatal("inactive shop exposed write capability")
		}
	}
	_, err = repo.Create(ctx, owner, first, req)
	expect(err, ErrNotFound)
	_, err = repo.SetStatus(ctx, owner, first, raceID, "confirmed", "")
	expect(err, ErrNotFound)
	exec(`UPDATE tenants SET status='active' WHERE id=$1`, first)
	exec(`UPDATE users SET status='suspended' WHERE id=$1`, owner)
	_, err = repo.List(ctx, owner, first, day)
	expect(err, ErrNotFound)
}
