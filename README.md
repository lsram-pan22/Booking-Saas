# Booking-Saas

## Project layout

```text
backend/
  controllers/
    super_admin/
      auth/
      packages/
    tenant/
      account/       # Registration, login, profile and subscription
      shops/         # Create, read and update shops
      services/      # Shop services, prices, duration and availability
      bookings/      # Daily bookings, scheduling conflicts and status history
  routes/            # Route registration and dependency wiring
  middleware/        # JWT and role checks
  config/
  utils/

frontend/web/src/
  api/
    super-admin/
    tenant/
  pages/
    super-admin/
    tenant/
  components/
    layouts/
      super-admin/
      tenant/
    shops/tenant/
    services/tenant/
    bookings/tenant/
  types/
    shared/
    super-admin/
    tenant/
  routes/
  utils/
```

Each backend feature has `ctrls`, `usecases`, `repositories` and `models` folders.
HTTP handlers belong in `ctrls`, business rules in `usecases`, SQL in
`repositories`, and data structures in `models`.

Frontend interfaces and types belong in `types`, separate from pages, components
and API functions. Keep account-level pages separate from shop-level pages.
Use Ant Design `Row` and `Col` for responsive layouts; editors open in a right
Drawer. API URLs use `/api/v1/`.

Each tenant represents one shop/branch. There is no separate branch layer.
Subscriptions use `shop_limit` for the number of shops an owner can create;
`user_limit` applies to each shop and includes its owner. Staff belong to one shop.

Shop services belong directly to `tenant_id`. Owners can create and edit services
and toggle active/inactive. The services page and service management APIs are
owner-only. Staff obtain active service choices through the booking API.
Prices use integer satang. Apply `backend/migrations/013_shop_services.sql` once
to create the services table in a new environment. Service endpoints are
`GET/POST /api/v1/tenant/shops/:shopID/services` and
`PATCH /api/v1/tenant/shops/:shopID/services/:serviceID`.

Apply `backend/migrations/014_shop_bookings.sql` once after the services migration.
Bookings currently contain one service and one provider (an active owner/staff
membership). Customer contact details, service name, duration, price and provider
name are saved on the booking. Later service edits do not change existing bookings.
Owners can choose any active provider in their shop; staff can create bookings
only for themselves. Provider choices and booking writes enforce this on the server.
New bookings must start in the future, at minute precision, within two years.
The daily list uses Asia/Bangkok dates and includes bookings that cross midnight.
The booking page defaults to a seven-day calendar with days as rows and time as
the horizontal axis. Its time range follows the shop's opening/closing times;
without configured hours it displays 00:00–24:00. Simultaneous appointments use
separate lanes in the same day. Providers are selected in the booking form, not
used as calendar columns. Click an appointment for details and actions in a right
Drawer. A list view includes cancelled bookings and remarks.
Active shop members can list bookings. Owners can change statuses for every
booking in their shop; staff can change statuses only for bookings assigned to
their own membership. Each booking response includes `can_manage` for the UI;
the status API independently checks the current membership and assignment inside
its transaction. An inactive shop is read-only for its owner. Staff cannot access
another shop.

Booking routes:
- `GET /api/v1/tenant/shops/:shopID/bookings?date=YYYY-MM-DD`
- `GET /api/v1/tenant/shops/:shopID/bookings?date=YYYY-MM-DD&days=7`
- `POST /api/v1/tenant/shops/:shopID/bookings`
- `PATCH /api/v1/tenant/shops/:shopID/bookings/:bookingID`
- `PATCH /api/v1/tenant/shops/:shopID/bookings/:bookingID/status`
- `PATCH /api/v1/tenant/shops/:shopID/bookings/:bookingID/customer`

Status transitions are pending → confirmed → completed, or pending/confirmed →
cancelled. Every transition is recorded in `booking_status_history`. All booking
writes lock the tenant row before checking availability. Overlapping appointments
for the same provider are rejected, including concurrent requests. Adjacent
appointments are allowed; cancelled appointments release their slot.

Apply `backend/migrations/015_booking_cancellation_remark.sql` after the booking
migration. Cancelling requires a nonblank `remark` (up to 1,000 characters) in
the status PATCH body. It is saved atomically in `bookings.remark`
and the status history's `remark`, and shown in the expanded booking row.
Historical cancellations remain intact without inventing a reason.
Apply `backend/migrations/016_booking_remark.sql` after migration 015 to rename
the booking remark column while preserving existing values and constraints.

Apply `backend/migrations/017_shop_opening_hours.sql` to add nullable opening and
closing times to tenants. Owners configure them in shop details using HH:mm,
with the same hours every day. Closing must follow opening on the same day;
24:00 is supported as closing time. Existing bookings remain unchanged. New
bookings must start and end within configured hours. Legacy appointments outside
current hours are flagged on the calendar and remain accessible in list view.

Apply `backend/migrations/018_booking_edit_history.sql` before editing booking
customer details. The customer PATCH takes `customer_name`, `customer_phone`
and `note`; an empty note clears it. Owners can edit any pending/confirmed booking
in their shop and staff can edit their own assigned bookings. Completed/cancelled
bookings are read-only. Each actual change saves the editor and before/after
customer values in `booking_edit_history` in the same transaction. Saving unchanged
values creates no history entry. Customer edits preserve service snapshots,
scheduling, assignment, status and cancellation remarks.

The main booking PATCH also accepts `service_id` and `starts_at` with customer
details. Pending and confirmed bookings remain in their current status after an
edit. A different service refreshes the stored name, price and duration; choosing
the original service retains its snapshot, including when its catalog price has
changed. Moving the start time retains the snapshot duration. Schedule changes
recalculate the end, check shop hours and provider availability, and reject overlaps
while excluding the booking being edited. A moved start must be in the future
within two years. Scheduling and service changes are stored with before/after
values in the existing edit history transaction. The provider remains assigned.

## Checks

Run `go test ./...` from `backend`, and `npm run build`, `npm run lint` from
`frontend/web`.

Set `TENANT_REGISTRATION_DB_TEST=1` when running the booking repository integration
test to verify PostgreSQL access checks, snapshot values, time boundaries and
concurrent requests. It creates and removes a uniquely named test schema without
using shop records from the public schema.
