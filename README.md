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
and toggle active/inactive; staff can read active services in their active shop.
Prices use integer satang. Apply `backend/migrations/013_shop_services.sql` once
to create the services table in a new environment. Service endpoints are
`GET/POST /api/v1/tenant/shops/:shopID/services` and
`PATCH /api/v1/tenant/shops/:shopID/services/:serviceID`.

## Checks

Run `go test ./...` from `backend`, and `npm run build`, `npm run lint` from
`frontend/web`.
