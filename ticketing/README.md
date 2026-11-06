# Ticketing service

Go service for seat plans, inventory, holds, orders, and tickets. It owns the
`ticketing` schema in PostgreSQL and talks to the Next.js site only over HTTP. See
`../TICKETING.md` for the full design.

Implemented so far: step 1 (scaffold, configuration, migrations, health probes,
logging, container image), step 2 (the full schema plus the seat/order/ticket state
machines), step 3 (seat-plan generator and CRUD, show upsert with seat snapshots,
placeholder seed), step 4 (public availability, atomic seat holds, expiry and
release), step 5 (checkout, manual orders, idempotent MarkPaid, cancellation, and
ticket issuance), step 6 (signed QR tokens, public ticket lookup, per-IP rate
limits, and the CORS allowlist), and step 7 (per-show scanner credentials and the
atomic admit endpoint). PayHere callbacks and the booking UI are still to come.

## Requirements

- Go 1.26 or newer (goose v3.29 requires 1.26; pgx v5.12 requires 1.25).
- PostgreSQL 16 or newer, installed natively.
- `make`, `psql`, `createdb`, `dropdb`.

## Local PostgreSQL

### Fedora / Nobara / RHEL

```sh
sudo dnf install -y postgresql-server postgresql
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql
# Give your OS user a role so `createdb` works without sudo.
sudo -u postgres createuser -s "$USER"
```

### Debian / Ubuntu

```sh
sudo apt install -y postgresql postgresql-client
sudo systemctl enable --now postgresql
sudo -u postgres createuser -s "$USER"
```

Create the database (idempotent):

```sh
make db-create
```

The default DSN uses the local Unix socket:

```
postgres:///ticketing_dev?host=/var/run/postgresql
```

For a remote host use a TCP URL and an explicit TLS mode:

```
postgres://user:password@db.example.com:5432/ticketing?sslmode=verify-full
```

## Configure and run

```sh
cp .env.example .env
# edit .env: set INTERNAL_TOKEN, QR_SECRET, PUBLIC_BASE_URL
make dev
```

`make dev` loads `.env` and runs `go run ./cmd/ticketing serve`. On startup the
service applies pending migrations (disable with `RUN_MIGRATIONS_ON_START=false`),
opens the pool, starts the expired-hold sweeper, and listens on `PORT`.

Probes:

```sh
curl -i http://localhost:8080/healthz   # 200 while the process is alive
curl -i http://localhost:8080/readyz    # 200 when Postgres is reachable, else 503
```

Migrations can also be run explicitly:

```sh
make migrate
```

They are guarded by a PostgreSQL advisory lock, so concurrent `serve` instances and
`make migrate` calls serialise instead of racing. The runner creates the `ticketing`
schema; goose keeps its version table at `ticketing.goose_db_version`.

## Seat plans and shows

Seed the placeholder 50-seat plan (idempotent):

```sh
make seed
```

Internal endpoints require `Authorization: Bearer $INTERNAL_TOKEN`:

```sh
TOKEN=$(grep '^INTERNAL_TOKEN=' .env | cut -d= -f2-)

curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/v1/internal/seat-plans
curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/v1/internal/seat-plans/sample-50

# Link a show to a plan; repeat this call to confirm it is idempotent.
curl -s -X PUT -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  localhost:8080/v1/internal/shows/<payload-show-id> \
  -d '{"production_slug":"dear-daughter","title":"Dear Daughter",
       "starts_at":"2026-12-01T19:30:00Z","venue":"Basement",
       "seat_plan_key":"sample-50","status":"on_sale","currency":"LKR"}'
```

`PUT` snapshots the plan into `show_seats`. A plan cannot be edited once any show uses
it; duplicate it instead. A show can only switch plans while it has no order history
and every seat is still available.

## Public availability and holds

These endpoints need no authentication.

```sh
SHOW=show-demo-1

# Badge: status, tiers, seats remaining, price_from.
curl -s localhost:8080/v1/public/shows/$SHOW

# Full seat map with an ETag. Repeating the request with the ETag returns 304.
curl -si localhost:8080/v1/public/shows/$SHOW/seats
curl -si -H 'If-None-Match: "<etag>"' localhost:8080/v1/public/shows/$SHOW/seats

# Hold seats. The response contains an order_token exactly once; keep it.
curl -s -X POST -H 'Content-Type: application/json' -H 'Idempotency-Key: demo-1' \
  localhost:8080/v1/public/shows/$SHOW/holds \
  -d '{"seats":["S001","S002"]}'

# Repeating the same Idempotency-Key and seats returns the same order (200) and
# sets Idempotent-Replayed: true. Reusing it with different seats returns 409.

# Release a pending hold. A wrong or missing X-Order-Token is rejected.
curl -s -o /dev/null -w '%{http_code}\n' -X DELETE \
  -H "X-Order-Token: $ORDER_TOKEN" \
  localhost:8080/v1/public/orders/$ORDER_ID
```

Holds last `HOLD_MINUTES` (10 by default). A held seat whose `held_until` has passed
is treated as available by the hold transaction even if the sweeper has not run yet.
The sweeper runs every 30 seconds, marks overdue pending orders `expired`, and frees
only seats still pointing at those orders. It uses a PostgreSQL advisory lock so
several API instances can run safely. Requests are capped at `MAX_SEATS_PER_ORDER`
(6) and 1 MiB of JSON.

## Tests

```sh
make test                 # unit tests (no database)
make test-race            # unit tests under the race detector
make test-integration     # Postgres integration tests
make test-integration-race  # Postgres integration tests under -race
```

## Orders and checkout

Checkout captures buyer details, extends the hold, and returns the order. The
`X-Order-Token` from the hold is required. `payhere` keeps the order `online` and uses
`CHECKOUT_HOLD_MINUTES`; `bank_transfer` becomes a `manual` order and uses
`MANUAL_HOLD_HOURS`.

```sh
# Capture buyer details and extend the hold.
curl -s -X POST -H 'Content-Type: application/json' -H "X-Order-Token: $ORDER_TOKEN" \
  -H 'Idempotency-Key: checkout-1' \
  localhost:8080/v1/public/orders/$ORDER_ID/checkout \
  -d '{"buyer_name":"Nimal","buyer_phone":"0771234567",
       "buyer_email":"nimal@example.com","payment_method":"payhere"}'

# Read the order back; buyer phone and email are not exposed on this route.
curl -s -H "X-Order-Token: $ORDER_TOKEN" localhost:8080/v1/public/orders/$ORDER_ID
```

Internal order endpoints (Bearer token) let staff create manual orders, mark them paid,
cancel them, and list a show's orders:

```sh
# Manual order: buyer name and phone are required, email is optional.
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  localhost:8080/v1/internal/orders \
  -d '{"show_id":"'$SHOW'","seats":["S003"],"buyer_name":"Kasun",
       "buyer_phone":"0771234567","payment_method":"cash","notes":"door sale"}'

# Idempotent: marking an already-paid order paid returns the same tickets.
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  localhost:8080/v1/internal/orders/$ORDER_ID/mark-paid \
  -d '{"payment_method":"cash"}'

curl -s -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  localhost:8080/v1/internal/orders/$ORDER_ID/cancel -d '{"reason":"refunded"}'

curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/v1/internal/shows/$SHOW/orders
```

`MarkPaid` sells the seats and issues one ticket per seat. If a late payment arrives
after the hold expired, it re-claims the seats when they are still free; if any seat
was taken, the order becomes `needs_review` and is not sold. Cancelling a paid order
voids its `issued` tickets and frees the seats; it is rejected with `tickets_in_use`
once any ticket is `used`.

## Tickets and QR tokens

A ticket QR token is `v1.<ticket_uuid>.<signature>`, where the signature is
`base64url(HMAC-SHA256(QR_SECRET, "v1|" + ticket_uuid))`. The QR image should encode
the `qr_url` returned by the lookup endpoint (`PUBLIC_BASE_URL/t/<token>`). Tokens are
verified with a constant-time comparison before the ticket is loaded.

```sh
# The token is the path segment from the QR URL. No auth is required, but a
# malformed or tampered token returns 404 ticket_not_found.
curl -s localhost:8080/v1/public/tickets/$QR_TOKEN
```

The response includes `production_slug` (for the site theme), show title/date/venue,
seat or tier wording, `price_cents`, `status` (`issued`, `used`, or `void`), the short
`order_reference` (last six order-UUID characters, uppercased), and `qr_token` /
`qr_url`. It never includes buyer details or the full order id.

## Scanner and admission

Scanner credentials are per show. Create or rotate one through the internal API; the
plaintext is returned exactly once and only its SHA-256 hash is stored.

```sh
# Issue (or rotate) a show's scanner token.
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  localhost:8080/v1/internal/shows/$SHOW/scanner-token

# Revoke it.
curl -s -o /dev/null -w '%{http_code}\n' -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  localhost:8080/v1/internal/shows/$SHOW/scanner-token
```

The future scanner app calls the admit endpoint with the show token in
`X-Scanner-Token` and the QR token it scanned:

```sh
curl -s -X POST -H "X-Scanner-Token: $SCANNER_TOKEN" \
  -H 'Content-Type: application/json' \
  localhost:8080/v1/scan/admit \
  -d '{"qr_token":"'"$QR_TOKEN"'"}'
```

A successful admit atomically changes the ticket from `issued` to `used` and returns
`{"result":"admitted","ticket":{...}}`. Errors:

- `401 invalid_scanner_token` — missing or unknown show token.
- `404 ticket_not_found` — malformed, tampered, or unknown QR token.
- `403 wrong_show` — the ticket belongs to another show.
- `409 ticket_already_used` / `ticket_void` — the ticket cannot be admitted again.

Two scanners racing on the same ticket produce one `200` and one `409`, because the
state change is a single `UPDATE ... WHERE status = 'issued'`.

## CORS and rate limits

`CORS_ALLOWED_ORIGINS` is an explicit comma-separated allowlist applied only to
`/v1/public/*`. The service never allows credentials or `*`, returns
`Access-Control-Allow-Origin` only for allowlisted origins, and answers preflight
requests with the public methods (`GET`, `POST`, `DELETE`, `OPTIONS`) and headers
(`Content-Type`, `Idempotency-Key`, `If-None-Match`, `X-Order-Token`,
`X-Request-Id`).

Rate limiting is in-memory and per IP, which is sufficient for the documented
single-instance deployment. Public routes allow 1200 requests per minute per IP; hold,
checkout, and release writes allow 300 per minute per IP. A denied request returns
`429 rate_limited` with `Retry-After`. The limiter prefers `X-Forwarded-For` and
`X-Real-IP`, so the Go service must stay behind the site's reverse proxy rather than
being exposed directly.

`make test-integration*` need `TEST_DATABASE_URL` (defaults to `DATABASE_URL`).
Integration tests are build-tagged `integration` and skip when `TEST_DATABASE_URL`
is unset. Point them at a disposable database, never at `ticketing_dev` or production;
the tests create rows and do not clean them up.

The required concurrency tests are in
`internal/store/holds_integration_test.go`: 200 goroutines racing for one seat must
produce exactly one winner, and overlapping multi-seat holds must neither deadlock
nor oversell. Order lifecycle coverage, including late re-claims, `needs_review`
recovery, and cancellation, is in `internal/store/orders_integration_test.go`; ticket
lookup is covered by `internal/store/tickets_integration_test.go`.

## Container

Docker is not required. Podman works:

```sh
make docker-build
# The default Unix-socket DSN does not reach the host from inside a container.
# Use a TCP DSN, e.g. postgres://user:pass@host:5432/ticketing?sslmode=disable
podman run --rm -p 8080:8080 \
  -e DATABASE_URL='postgres://user:pass@host:5432/ticketing?sslmode=disable' \
  -e INTERNAL_TOKEN='...' -e QR_SECRET='...' -e PUBLIC_BASE_URL='http://localhost:3000' \
  basementplays/ticketing:dev
```

The image is a static binary on `scratch`, runs as UID 65532, and embeds time zone
data.

## Configuration

Required: `DATABASE_URL`, `INTERNAL_TOKEN` (>= 16 chars), `QR_SECRET` (>= 16 chars),
`PUBLIC_BASE_URL` (absolute http/https).

Defaults: `PORT=8080`, `LOG_LEVEL=info`, `HOLD_MINUTES=10`,
`CHECKOUT_HOLD_MINUTES=15`, `MANUAL_HOLD_HOURS=24`, `MAX_SEATS_PER_ORDER=6`,
`MAX_PLAN_SEATS=500`, `CORS_ALLOWED_ORIGINS=` (empty), `PAYHERE_SANDBOX=true`,
`PAYHERE_ALLOW_LIVE=false`, `RUN_MIGRATIONS_ON_START=true`, `SHUTDOWN_TIMEOUT=10s`.

`PAYHERE_SANDBOX=false` is refused unless `PAYHERE_ALLOW_LIVE=true` is also set.

## Troubleshooting

- `psql: error: connection to server on socket "/var/run/postgresql/.s.PGSQL.5432" failed`:
  the cluster is not running (`sudo systemctl start postgresql`) or the socket path
  differs.
- `createdb: error: role "you" does not exist`: run
  `sudo -u postgres createuser -s "$USER"`.
- `FATAL: password authentication failed`: for local socket access, `pg_hba.conf`
  should use `peer` for your user; for TCP, set a password or use `sslmode`.
- Migrations wait for the advisory lock: another instance is migrating. The wait is
  capped at 30 seconds.
