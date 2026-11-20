# Security and permission boundaries

This is an application hardening baseline, not a claim of a completed penetration
test or production security certification. Use `go run ./cmd/migrate` from
`backend/` to create a new public schema and apply pending migrations before
starting the API. Database integration checks remain explicitly opt-in.

On 2026-11-19, all six database integration scenarios passed against the configured
Supabase PostgreSQL instance: registration/Free subscription, shop members,
services, bookings (including concurrent creation and rescheduling), password
reset/session revocation, and expired-plan Free fallback/scope selection. Fixtures
used temporary tables or isolated schemas. Hosted-database scenarios have longer
overall deadlines to accommodate network round trips; assertions are unchanged.
Testing also found and fixed a PostgreSQL parameter-type ambiguity in staff
activation/deactivation. Browser end-to-end testing and a production security
review remain separate work.

Migration 022 enables RLS and removes direct table grants for `anon`,
`authenticated` and `PUBLIC` on application tables, including the migration
ledger. No browser policies are installed: use the Go API for application data.
The backend database role must own the tables or have explicit privileges and
appropriate RLS access. Supabase's `auth` schema is not modified.

## Access matrix

| Operation | Owner | Staff | Super Admin |
| --- | --- | --- | --- |
| Read shop booking calendar | Own shop | Assigned shop | No implicit tenant access |
| Create booking | Eligible provider in own shop | Self only | No implicit tenant access |
| Edit booking | Own shop, pending/confirmed | Own assigned pending/confirmed | No implicit tenant access |
| Confirm/complete/cancel | Own shop | Own assigned booking | No implicit tenant access |
| Manage shop/services/members | Own shop only | Denied | No implicit tenant access |
| Reset staff password | Staff in own shop, with actor password | Denied | Owner recovery endpoint only |
| Package catalog/platform recovery | Denied | Denied | Verified active Super Admin |

All private endpoints require a signed HS256 JWT with the expected issuer,
expiry, issued-at time, a valid user UUID and a nonnegative session version.
Tokens over 4 KB, duplicate Authorization headers and lifetimes over 12 hours are
rejected. Normal tokens expire after one hour. Every protected request checks the
current account status, system role and version in PostgreSQL. Business roles
are loaded from shop membership, never accepted from client JSON or local storage.

Shop endpoints have a common membership guard in addition to repository SQL
scoping. Unrelated/nonexistent shop identifiers receive the same denial. Inactive
members lose access even while their token is otherwise valid. Staff can read the
shop calendar as required by the booking workflow, but cannot manage coworkers'
appointments. Write transactions recheck scope and lock authorization rows.
Subscription entitlements are an additional requirement for new work; they never
grant role or membership permissions.

## Passwords and sessions

Passwords are hashed with bcrypt. New/reset passwords need 12 characters and at
most 72 UTF-8 bytes. Login rejects oversized passwords before database work and
performs a dummy bcrypt comparison for unknown accounts. Authentication failures
do not disclose which credential was wrong.

Resetting another account requires the authenticated actor's current password,
rechecks account/version/ownership inside the transaction, hashes the temporary
password and records the existing password audit event atomically. Temporary
passwords expire after 24 hours and cannot access normal application operations
until changed. Successful password changes clear the reset deadline and revoke
prior sessions. A fresh reset is needed after the temporary password expires.

Both tenant and Super Admin logout buttons revoke **all devices** by incrementing
the account session version on the server. The frontend clears its session only
after confirmation (or an already-invalid token). Network/database failures are
reported instead of claiming server-side logout succeeded. Super Admin login
issues the current database version and returns the expiry to the browser.
Revocation takes effect on subsequent requests; requests already executing are
not forcibly interrupted.

## Abuse and request controls

Counters are bounded, concurrent-safe, in-memory and shared across the login
surfaces of one API process:

| Action | Limit |
| --- | --- |
| Tenant and Super Admin login, by IP | 30 attempts / 5 minutes |
| Login, by normalized identity across IPs | 10 attempts / 5 minutes |
| Registration, by IP | 5 attempts / hour |
| Reset/change password, by verified actor | 10 attempts / 15 minutes |

Limits include successful and failed attempts and return 429 with Retry-After.
Identity keys are hashed in memory, not written to logs. Authentication bodies
are capped at 16 KB; the global body cap is 1 MiB with smaller endpoint limits
where appropriate. JSON binding rejects unknown fields, arrays/null in place of
objects and trailing documents, retaining required-field validation.

HTTP header/read/write/idle timeouts and a 32 KB header cap bound slow/oversized
requests. CORS uses exact configured origins without credentials or wildcards.
No forwarding headers are trusted by default; configure TRUSTED_PROXIES only for
actual proxy peers, and configure the proxy to sanitize client-supplied forwarding
headers. Trusting all addresses is rejected at startup. CORS is not authorization.

API responses disable caching and include nosniff, frame denial, referrer and
feature-policy headers. HSTS is set for direct TLS; a TLS-terminating proxy must
set its own HTTPS security headers. JWT_SECRET must be a non-placeholder secret
of at least 32 bytes. It is never generated or printed automatically.

Security logs contain generated request IDs, route patterns, method, status,
client IP and verified actor IDs. Bodies, credentials, tokens and query strings
are omitted. Default access/request dump logging is replaced; panic responses
and panic logs do not print attacker-controlled panic values or headers.

## Verification

Run `go test ./...` in backend, then build and lint in frontend/web. Tests cover
JWT tampering/role rejection, expiry, reset confirmation, rate-limit concurrency
and identity/IP boundaries, payload parsing, headers, secret-free audit/recovery
and the common tenant permission guard.

`PASSWORD_RESET_DB_TEST=1` enables isolated-schema PostgreSQL checks for reset
ownership, actor confirmation, expiry, session revocation and live membership
denial. `TENANT_REGISTRATION_DB_TEST=1` covers shop/member/service/booking access
and concurrency. `SUBSCRIPTION_FALLBACK_DB_TEST=1` covers plan scope changes.
Set these environment variables only when deliberately running the database
tests. Fixtures create and remove unique schemas without modifying real accounts.

## Work remaining before production

- Repeat the database integration suite for the target deployment and complete
  a browser end-to-end role/tenant security review.
- Configure HTTPS, frontend CSP, private database networking, a least-privilege
  database user, managed secrets, encrypted off-machine backups and restore tests.
- Add MFA, especially for Super Admin, and decide verification/invitation policies
  for accounts added by email. Existing-account staff addition remains the current
  owner-driven business workflow; no invitation consent flow was introduced here.
- Replace process-local throttles with a shared store or edge enforcement before
  running multiple API replicas; restarts currently reset counters.
- Set log retention/access rules and monitoring for repeated denials/reset attempts.
- Browser tokens remain in sessionStorage. Frontend XSS protection and deployment
  CSP are still required; this change does not migrate authentication to cookies.
