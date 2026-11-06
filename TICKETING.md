# TICKETING.md: Go ticketing service + booking UI

## 1. Scope (now)
In: generated seat plans with tiers (built from a Backstage form), per-show seat
inventory, availability API, seat holds, orders, manual (WhatsApp/bank-transfer)
orders and ticket issue, signed QR tickets, ticket page, public booking UI, wiring
EVERY existing Reserve/Book button to the booking flow (per show, behind a flag),
Backstage ticketing pages (seat plan builder, manual orders), PayHere SANDBOX (last).
Out: scanner UI (a separate mobile app comes later), PayHere live mode, refund
automation, discount codes, public accounts, multi-currency, drawn or free-form seat
maps, deployment automation.
The backend exposes an "admit" endpoint (step 7, optional) for the future scanner app.

## 2. Architecture
- Repository: the existing Next.js + Payload app stays at the repo root; `ticketing/`
  (Go) and `deploy/` (later) sit beside it. Wherever this document says `site/`, read
  "the existing Next.js app at the repo root". Root `tsconfig.json` and
  `eslint.config.mjs` exclude `ticketing/`.
- Go 1.26+ (goose v3.29 requires it; pgx v5.12 requires 1.25), stdlib `net/http`
  (method+path routing), pgx v5, plain SQL migrations
  (golang-migrate or goose), `log/slog`, config from env, graceful shutdown.
- Free tooling only. Database: PostgreSQL 16 or newer, installed natively (Docker is
  NOT required; `docker-compose` may be added later). Use only standard Postgres
  features, nothing version- or provider-specific.
  Local: `DATABASE_URL=postgres:///ticketing_dev?host=/var/run/postgresql`.
  `make db-create` runs `createdb ticketing_dev`. `sslmode` must work via DATABASE_URL
  for remote hosts later.
- The site's Payload also moves to Postgres (database `payload_dev`); SQLite is removed
  from the site. The Go service uses schema `ticketing` with its own migrations.
  Never join across Payload's tables.
- Money: integer minor units (`*_cents`, currency LKR). Format "2500.00" only at the
  PayHere boundary. Times: timestamptz UTC; display in Asia/Colombo.
- Layout:
```
ticketing/
  cmd/ticketing/main.go       # serve | migrate
  cmd/seed/main.go            # dev seed
  internal/config  internal/db  internal/domain   # state machines, types
  internal/seatplan  internal/store  internal/service
  internal/httpapi            # handlers, auth, rate limit, request ids, CORS
  internal/payhere  internal/qr
  migrations/  Makefile       # make dev | test | migrate | db-create
```
- Env: DATABASE_URL, INTERNAL_TOKEN, QR_SECRET, PUBLIC_BASE_URL, HOLD_MINUTES=10,
  CHECKOUT_HOLD_MINUTES=15, MANUAL_HOLD_HOURS=24, MAX_SEATS_PER_ORDER=6,
  MAX_PLAN_SEATS=500, CORS_ALLOWED_ORIGINS (default empty), PAYHERE_MERCHANT_ID,
  PAYHERE_MERCHANT_SECRET, PAYHERE_SANDBOX=true, PAYHERE_ALLOW_LIVE=false.
  Fail fast if required vars are missing.

## 3. Data model (schema `ticketing`)
- seat_plans(id, key unique, name, layout jsonb, created_at)
- shows(id text pk = Payload show row id, production_slug, title, starts_at, venue,
  seat_plan_id, status in [draft,on_sale,closed,cancelled], currency, updated_at,
  scanner_token_hash null, scanner_token_set_at null)
- show_seats(show_id, seat_code, tier_key, tier_name, label null, price_cents,
  sort_index, status in [available,held,sold,blocked], order_id null,
  held_until null, pk(show_id, seat_code)). CHECKs: held => order_id and held_until
  set; sold => order_id set; available => both null. Tier and price are snapshots
  taken when the show's seat rows are created.
- orders(id uuid, show_id, status in [pending,paid,expired,cancelled,needs_review],
  source in [online,manual], payment_method in [payhere,bank_transfer,cash,whatsapp],
  buyer_name null, buyer_phone null, buyer_email null, total_cents,
  access_token_hash, idempotency_key null, expires_at, paid_at, notes, created_at;
  unique(show_id, idempotency_key))
- order_items(order_id, show_id, seat_code, price_cents, pk(order_id, seat_code),
  FK (show_id, seat_code) -> show_seats)
- tickets(id uuid, order_id, show_id, seat_code, status in [issued,used,void],
  issued_at, used_at)
- payments(id, order_id, provider, provider_payment_id, status_code, amount_cents,
  currency, raw jsonb, received_at, unique(provider, provider_payment_id))
- audit_log(id, at, actor, action, entity, entity_id, data jsonb)
Key constraint, one live ticket per seat:
```sql
CREATE UNIQUE INDEX one_live_ticket_per_seat
  ON ticketing.tickets (show_id, seat_code) WHERE status IN ('issued','used');
```
Ticket wording: label present => "Seat {label} ({tier_name})"; otherwise
"{tier_name} admission".

## 4. Generated seat plans
The plan is generated from parameters, NOT drawn. It does not mimic the real room.
Layout jsonb:
```json
{ "stage_label": "Stage", "columns": 10, "numbering": "sequential",
  "tiers": [{"key":"front","name":"Front","price_cents":300000,"seat_count":20},
            {"key":"standard","name":"Standard","price_cents":250000,"seat_count":30}],
  "seats": [{"code":"S001","tier":"front","label":"1"},
            {"code":"S021","tier":"standard","label":"21"}] }
```
- `code` is a stable internal id (S001, S002, ...). `label` is an optional display
  number (may be null). Seats are listed in display order; tiers display in the listed
  order (closest to the stage first), each as its own section.
- `GeneratePlan(params)` is a pure function in `internal/seatplan` with tests. Input:
  name, key (slug), tiers [{name, price_cents, seat_count}] (at least one; "everything
  one common price" is a single tier), columns (display grid width, default 10),
  numbering in [none, sequential], optional stage_label. Total seats <= MAX_PLAN_SEATS.
  Sequential numbering continues across tiers ("1".."N"); "none" leaves labels null.
  Return typed validation errors (empty tier name, zero seats, duplicate tier names,
  negative price, too many seats).
- Individual labels can be edited or cleared while no show references the plan.
- A plan is a TEMPLATE: editable only while unused. To change tier prices for another
  show, duplicate the plan. Tier prices live in the plan (no per-show price override).
- When a show is linked to a plan, `show_seats` rows are created from it (snapshot).
- Seed a PLACEHOLDER plan `sample-50`: one tier "Standard", 50 seats, Rs. 2500
  (250000 cents), sequential numbering. The seed output must say it is a placeholder.

## 5. State machines (implement in `internal/domain`, table-driven tests)
- Seat: available -> held -> sold; held -> available (release/expiry/cancel);
  available <-> blocked (admin). sold -> available only via admin cancel of a paid order.
- Order: pending -> paid | expired | cancelled | needs_review; expired -> paid |
  needs_review (a late payment may be re-claimed); paid -> cancelled (admin; voids
  tickets); needs_review -> paid | cancelled (admin).
- Ticket: issued -> used | void. `used` and `void` are terminal.
Illegal transitions return typed errors. DB CHECKs are the second line of defense.

## 6. Core behaviour
- Hold policy: a hold stores NO personal details. TTL is HOLD_MINUTES (10). Max
  MAX_SEATS_PER_ORDER (6) seats per order. Rate-limit hold creation per IP (generous,
  since a venue may share one network). Buyer details are captured at checkout, which
  extends the hold to CHECKOUT_HOLD_MINUTES (15) from that moment. There are no
  open-ended holds: release early on PayHere `cancelled`/`failed` notifications and on
  the cancel URL. A successful payment arriving after expiry follows the re-claim /
  `needs_review` logic below. Manual orders hold for MANUAL_HOLD_HOURS (24).
- Hold (atomic, no double booking): in one transaction, lock the requested seats in
  SORTED order, then claim them; retry up to 3 times on deadlock (40P01) or
  serialization failure (40001).
```sql
SELECT seat_code FROM ticketing.show_seats
 WHERE show_id=$1 AND seat_code = ANY($2) ORDER BY seat_code FOR UPDATE;
UPDATE ticketing.show_seats SET status='held', order_id=$3, held_until=$4
 WHERE show_id=$1 AND seat_code = ANY($2)
   AND (status='available' OR (status='held' AND held_until < now()));
-- rows affected must equal len(seats), otherwise roll back and return 409 seat_taken
```
- Expiry: expired holds count as available (checked in the claim query), plus a sweeper
  every 30s that sets expired pending orders to `expired` and frees seats. Correctness
  must never depend on the sweeper.
- Idempotency: POST holds and checkout accept an `Idempotency-Key` header, stored on
  the order (unique per show). The same key returns the same order.
- MarkPaid(orderID): one transaction; `SELECT ... FOR UPDATE` on the order; idempotent
  if already paid; seats -> sold; one ticket per seat. If the hold expired but the
  seats are still free, re-claim and complete. If the seats were taken, set the order
  to `needs_review` (payment received, seats gone) for a manual refund decision.
- Never trust client amounts or prices. Totals come from `show_seats.price_cents`.
- Linking a plan to a show or updating a show: idempotent upsert. Setting a show to
  `closed` stops new holds; existing paid orders stay valid.
- QR token: `v1.<ticket_uuid>.<base64url(HMAC-SHA256(QR_SECRET, "v1|"+ticket_uuid))>`,
  verified with constant-time compare. The QR encodes `{PUBLIC_BASE_URL}/t/<token>`.
- Admit (for the later app): atomic, single statement.
```sql
UPDATE ticketing.tickets SET status='used', used_at=now()
 WHERE id=$1 AND status='issued';  -- 1 row = admit; 0 rows = already used or void
```

## 7. API surfaces
- Public `/v1/public/...` (no auth, rate-limited per IP, body size limits, strict JSON):
  GET shows/{id} (summary: status, tiers with price and seats remaining, price_from),
  GET shows/{id}/seats (tiers + seats with status, label, tier; ETag),
  POST shows/{id}/holds, DELETE orders/{id} (release), POST orders/{id}/checkout,
  GET orders/{id}, GET tickets/{token} (includes production_slug so the site can theme
  the ticket page). `DELETE orders/{id}`, `POST orders/{id}/checkout` and
  `GET orders/{id}` all require the `X-Order-Token` header; store only its hash and
  compare in constant time.
- Internal `/v1/internal/...` (Bearer INTERNAL_TOKEN, Next.js server only; never
  exposed publicly):
  seat plans: POST seat-plans (generator params), GET seat-plans, GET seat-plans/{key},
  PATCH seat-plans/{key}/labels (only while unused), POST seat-plans/{key}/duplicate;
  shows: PUT shows/{id} (idempotent upsert, takes seat_plan_key);
  orders: POST orders (manual), POST orders/{id}/mark-paid, POST orders/{id}/cancel,
  POST orders/{id}/recheck-payment, GET shows/{id}/orders; PUT seats/{code}/block.
- Callbacks: POST /payhere/notify (public, signature verified).
- Scanner (optional step 7): POST /v1/scan/admit with a per-show scanner token in
  `X-Scanner-Token`. Internal `POST/DELETE /v1/internal/shows/{id}/scanner-token`
  issues (plaintext returned once) or revokes it; only the hash is stored.
- Standard errors: 400 validation, 401/403 auth, 404, 409 seat_taken or bad state,
  429 rate limit. Request IDs on every response.

## 8. PayHere (SANDBOX only for now)
- The service refuses to start with PAYHERE_SANDBOX=false unless
  PAYHERE_ALLOW_LIVE=true is also set.
- Check the current PayHere docs for field names, limits, and URLs. Sandbox and live
  are different checkout URLs.
- return_url => `/events/<slug>/book/confirmation?order=<id>` on the site;
  cancel_url => back to the booking page; notify_url => the Go service's public
  /payhere/notify (use a tunnel in development; notifications cannot reach localhost).
- Checkout hash computed ONLY on the server:
  upper(md5(merchant_id + order_id + amount("2500.00") + currency + upper(md5(secret)))).
- Notify handler: verify md5sig = upper(md5(merchant_id + order_id + payhere_amount +
  payhere_currency + status_code + upper(md5(secret)))) with constant-time compare;
  check merchant_id; check amount and currency against the STORED order; status_code 2
  => MarkPaid. Handle the other codes per the docs (pending, cancelled, failed,
  chargeback => needs_review). Record every notification in `payments`, idempotent on
  provider_payment_id, respond 200 fast, never log secrets.
- Never mark an order paid from return_url alone. Add a Backstage "Re-check payment"
  action. Only automate reconciliation if the current PayHere docs confirm a retrieval
  API and how it authenticates.
- Tests: forged signature, wrong amount, duplicate, late (after expiry), out-of-order.

## 9. Integration with the EXISTING site (site/)
Do not rebuild the UI. First inspect the repo and confirm or correct these assumptions:
designs live in `src/designs/<slug>/`; `src/designs/index.ts` maps slug -> design;
`PlainDesign` is the fallback; Reserve/Book buttons currently link to `show.ticketUrl`
(a wa.me link). List EVERY place a Reserve/Book button or ticketUrl is used (hero,
show-dates rows, booking block, mobile sticky bar, events listing, home page, etc.).

9.1 Design contract. Each entry in `designs` becomes `{ Page, Shell, Booking? }`:
- `Page`: today's component, unchanged.
- `Shell`: wrapper providing the design's CSS variables, fonts, textures, header and
  themed footer. Used by Page AND by the booking and ticket routes.
- `Booking` (optional): a fully custom booking layout for that event. If missing, the
  shared `DefaultBooking` renders inside the design's `Shell`.

9.2 One Reserve entry point. Create `bookingHref(production, show?)` and a
`<ReserveButton>` used by EVERY Reserve/Book button in every design.
- If `show.ticketingEnabled`: link to `/events/<slug>/book?show=<show.id>`, label
  "Reserve seats".
- Otherwise: link to `show.ticketUrl` (WhatsApp) with the existing label.
- Past, cancelled, or sold-out shows render a disabled state with their status text.

9.3 Payload changes.
- Site database moves to Postgres; remove SQLite; delete the dev DB file and re-seed.
- Show: add `ticketingEnabled` (checkbox, default false). Run: add `seatPlanKey`
  (text), validated in the admin by calling Go's GET seat-plans/{key}; an unknown key
  blocks the save. There is NO price field: price comes from the plan's tiers.
- When ticketing is enabled, the displayed price ("From Rs. X" or the tier list) and the
  availability badge and "seats left" note come from Go's show summary. The old text
  `price`, `status`, and `seatsNote` fields apply only while ticketing is disabled
  (status stays as a manual override for cancelled or sold-out).
- Sync: an `afterChange` hook on Runs calls the internal PUT shows/{id} for each show
  with ticketingEnabled (id = the show's stable row id; send production_slug, title,
  starts_at, venue, seat_plan_key, status). Server-side only. If Go is unreachable,
  do not lose the save: log it, show an admin warning, and provide a "Re-sync" action.
  Unchecking ticketingEnabled sets the Go show to `closed`.
- Deleting a Run: block the delete when Go reports any orders (paid, pending, or held
  seats) for its shows. With none, allow the delete and set each Go show to `closed`.
  If Go is unreachable, block the delete (fail closed).
- `seatPlanKey` validation: block the save only when Go is reachable and reports the
  key as unknown. If Go is unreachable, allow the save, show an admin warning, and mark
  the show as not synced (the Re-sync action fixes it). Exception: turning
  `ticketingEnabled` on requires a successful validation (fail closed).

9.4 Routes, all rendered inside the production's Shell:
- `/events/[slug]/book?show=<id>`: if there is no `show` param and exactly one show is
  on sale, use it; if several, show a date picker step first.
- `/events/[slug]/book/confirmation?order=<id>`: polls order status. The order access
  token lives in sessionStorage and is sent as a header.
- `/t/[token]`: ticket page. Look up the ticket via Go, then render inside that
  production's Shell, with seat or tier text, show, a short order reference (last six
  characters of the order UUID, uppercased), and a QR image. No buyer name.
- Production content comes from Payload on the server; availability and holds come
  from the browser via `NEXT_PUBLIC_TICKET_API_BASE` (see section 12).

9.5 Seat map theming. One shared client `SeatMap` component rendering each tier as a
section (tier name, price, seats remaining) containing a CSS grid of seat tiles
(`columns` per row, responsive), with the stage label at the top. Styled ONLY through
CSS variables that each design's Shell defines, with neutral fallbacks: --seat-available,
--seat-selected, --seat-held, --seat-sold, --seat-blocked, --seat-stroke, --seat-label,
--stage-fill. State must not rely on color alone (add a pattern or shape plus a legend).
Tap targets at least 32px, keyboard accessible, poll availability every ~10s, selection
list with tier subtotals and a total. Blank tiles (no label) are still selectable.
- dear-daughter mapping: available = --paper with a --wood outline; selected =
  --accent; held = hatched --accent at reduced opacity; sold = muted --bg-deep with
  hatch; stage = --wood. Booking header: compact damask strip with the title, date, and
  venue; torn-paper top edge; panel on --paper; terracotta primary buttons; display font
  headings; themed footer.
- PlainDesign provides neutral variables.

9.6 Booking flow: select seats -> Hold (countdown timer) -> buyer form (name, phone,
email) -> payment method (PayHere sandbox, or reserve and pay by bank transfer) ->
confirmation page. Handle 409 seat_taken gracefully (refresh, keep other selections).

9.7 Backstage ticketing pages. Standalone Next.js pages under `/backstage-tools/...`
that require a logged-in Payload user (redirect unauthenticated visitors to the Payload
login) and call the Go internal API server-side with INTERNAL_TOKEN (no separate Go
admin UI). This is the chosen option over a Payload custom admin view: no import map or
component-path setup, and the existing `Disallow: /backstage` robots rule covers the
path. No public links to these pages. Pages:
- Seat plans: list (with usage), create via the BUILDER FORM, duplicate, edit labels.
  Builder form: plan name, key (auto from name, editable), toggle "common price for all
  seats" (collapses to one tier), tier rows (name, price in LKR, seat count), seats per
  row, numbering (none or sequential), stage label, and a LIVE PREVIEW of the generated
  grid. Saving calls the generator.
- Show ticketing: orders list, create a manual order for a WhatsApp reservation by
  picking seats on the same seat map, mark paid, cancel, re-check payment, block seats.

9.8 Add a Privacy Policy page before public launch (we collect names, phones, emails).

## 10. Build steps (stop after each, say how to verify, wait for "next")
1. Scaffold: Go module, config, native-Postgres setup docs and `make db-create`,
   migrations runner (`ticketing migrate`, advisory-lock guarded at startup), /healthz,
   /readyz, slog, Dockerfile, `.env.example`; root `tsconfig.json` and ESLint exclude
   `ticketing/`. Verify: `make dev`, probes OK, migrations apply.
2. Schema + domain state machines + tests (no HTTP).
3. `GeneratePlan` + tests; plan CRUD; show upsert creating `show_seats`; `sample-50` seed.
4. Availability + holds + expiry + release. REQUIRED tests under `go test -race`:
   200 goroutines racing for one seat give exactly one winner; overlapping multi-seat
   holds do not deadlock or oversell; expired holds are reclaimable; correctness holds
   with the sweeper disabled. The sweeper uses `pg_try_advisory_lock`.
5. Orders, checkout, manual orders, MarkPaid (idempotent), cancel, needs_review path,
   Idempotency-Key handling.
6. QR sign/verify, ticket lookup endpoint, internal auth, rate limiting, CORS allowlist.
7. (Optional) admit endpoint + tests.
8. Site integration, in this order:
   8a. Audit the existing UI, refactor to the `{Page, Shell, Booking?}` contract, add
       `bookingHref` and `ReserveButton`, replace every Reserve/Book button. With the
       flag off, behaviour must be IDENTICAL to today. (Any time after step 1.)
   8b. Switch Payload to Postgres and remove SQLite; Payload fields, sync hook, plan key
       validation.
   8c. Backstage seat plan builder + list pages.
   8d. Read-only seat map on `/events/<slug>/book` (dear-daughter theme, then
       PlainDesign), with `NEXT_PUBLIC_TICKET_API_BASE` and the runtime proxy route.
   8e. Holds, countdown, buyer form, confirmation page.
   8f. Ticket page `/t/[token]` with QR.
   8g. Backstage show-ticketing page (manual orders, mark paid, cancel).
   8h. MEDIA_STORAGE switch for Payload (any time after step 1).
9. PayHere sandbox integration + "Re-check payment" + tests.
10. `deploy/README.md` documenting setups A/B/C, plus `deploy/backup.sh`. Stop.

> **Frontend-first track.** The current frontend work covers step 8a (F1), the
> `ticketingEnabled` part of 8b (F2), and mock-backed versions of 8d–8f (F3–F4).
> It runs against a mock client so the booking UI can be demoed before the Go
> service is deployed. Real wiring is a separate later step: set the API mode to
> `http`, add the runtime proxy route, and reconcile the TypeScript contract with
> the Go API.
>
> **Demoing the mock locally.** Enable **Use ticketing service** on a show in
> Backstage → Runs, then use the event page's Reserve seats button. No Go service
> is required. The mock keeps holds, orders, and buyer data in `sessionStorage`,
> so book and confirm in the same tab (a new tab cannot recover the order token).
> Its seat plan is fixed at 50 seats (Front, then Standard) with a few seats
> pre-sold. `NEXT_PUBLIC_TICKET_API_MODE=http` intentionally throws until the HTTP
> client and `/ticket-api` proxy are implemented.


## 11. Conventions
Small packages, no globals, context on every DB call, errors wrapped with %w, SQL in
`store/` only, no business logic in handlers, no secrets in logs, tests alongside code.

## 12. Portability (keep hosting options open)
Goal: moving between (A) one VPS with Docker Compose + Caddy, (B) separate hosts for
the site and the Go service, and (C) a hybrid must be a configuration change, not a
code change.
1. All config via environment variables. Ship `.env.example` for each service. No
   host-specific code or SDKs.
2. A Dockerfile per service (Go: multi-stage static binary, CGO_ENABLED=0, non-root;
   Next.js: standalone output). Compose for the single-VPS case comes later.
3. URLs: the site reaches Go server-to-server via TICKETING_INTERNAL_URL. The browser
   reaches Go via ONE setting, NEXT_PUBLIC_TICKET_API_BASE, defaulting to `/ticket-api`
   (same origin). The site includes a runtime catch-all route handler
   `/ticket-api/[...path]` that proxies to TICKETING_INTERNAL_URL (reads env at runtime)
   for setups where no reverse proxy does the routing. In production on one VPS, the
   reverse proxy routes `/ticket-api/*` to Go and only public paths are exposed.
4. CORS: Go supports CORS_ALLOWED_ORIGINS, only on /v1/public/*. No cookies anywhere on
   the ticketing API; order access tokens are sent in a header.
5. Database: standard PostgreSQL features only. Payload's adapter is selected by env.
6. Media storage: Payload uses local disk when MEDIA_STORAGE=local and its
   S3-compatible storage adapter when MEDIA_STORAGE=s3 (endpoint, bucket, keys, public
   URL from env). Check the current Payload docs for the right adapter.
7. Stateless, sleep-tolerant Go: no local files; correctness never depends on the
   sweeper; scheduled work uses pg_try_advisory_lock; in-memory rate limiting is
   acceptable for one instance (document it). /healthz and /readyz (checks the DB).
8. UI resilience: the client retries idempotent GETs with backoff and shows a "waking up
   the booking system" state when a request is slow. POSTs are never retried without an
   Idempotency-Key.
9. Domain: one custom domain with optional subdomains. Public URLs come only from env
   (PUBLIC_BASE_URL, TICKETING_PUBLIC_URL). PayHere ties the merchant secret to a
   domain, so choose the final domain before using live mode.
10. `deploy/backup.sh` using pg_dump, plus a documented restore.

## 13. Decisions

1. **Repository layout.** The existing Next.js + Payload app stays at the repo root.
   `ticketing/` (and later `deploy/`) sit beside it. Wherever this document says
   `site/`, read "the existing Next.js app at the repo root". Root `tsconfig.json` and
   `eslint.config.mjs` exclude `ticketing/`.
2. **Deleting a run.** Block the delete when Go reports any orders (paid, pending, or
   held seats) for the run's shows. With no such orders, allow the delete and set each
   Go show to `closed`. If Go is unreachable, block the delete (fail closed).
3. **Checkout expiry.** Checkout extends the hold to `CHECKOUT_HOLD_MINUTES` (15,
   configurable) from the moment checkout starts. There are no open-ended holds. Holds
   are released early on PayHere `cancelled`/`failed` notifications and on the cancel
   URL. A successful payment that arrives after expiry follows the re-claim /
   `needs_review` logic in §6.
4. **Order access token.** `X-Order-Token` is required on `DELETE orders/{id}`,
   `GET orders/{id}`, and `POST orders/{id}/checkout`. Store only the token hash and
   compare in constant time.
5. **Public ticket page.** Show no buyer name. Show a short order reference: the last
   six characters of the order UUID, uppercased. Staff verification happens through the
   authenticated scanner API later.
6. **Staff pages.** Standalone staff pages live under `/backstage-tools/...` (covered
   by the existing `Disallow: /backstage` robots rule). They require a Payload session
   and redirect unauthenticated visitors to the Payload login. No public links.
7. **`seatPlanKey` validation.** Block the save only when Go is reachable and reports
   the key as unknown. If Go is unreachable, allow the save, show an admin warning, and
   mark the show as not synced (the Re-sync action fixes it). Exception: turning
   `ticketingEnabled` on requires a successful validation (fail closed).
8. **Late payment.** `expired` orders may transition to `paid` or `needs_review` when a
   payment arrives. `MarkPaid` re-claims the seats if they are all still free; if any
   seat was taken, the order becomes `needs_review` for a manual refund decision. The
   `needs_review` order can later become `paid` once the seats are free again, or be
   cancelled.
9. **Bank-transfer checkout.** A public checkout with `payment_method=bank_transfer`
   changes the order source to `manual` and extends the hold to `MANUAL_HOLD_HOURS`
   (24). `payhere` stays `online` and uses `CHECKOUT_HOLD_MINUTES` (15). Manual orders
   created by staff require name and phone, accept optional email, and allow
   `bank_transfer`, `cash`, or `whatsapp`.
10. **Ticket lookup and public guards.** `GET /v1/public/tickets/{token}` verifies the
    signed QR token, then returns the production slug, show, seat/tier wording, price,
    ticket status, the short order reference (last six order-UUID characters,
    uppercased), and the QR token/URL. Invalid or unknown tokens return 404. Public
    routes use an in-memory per-IP fixed window: 1200 requests per minute per IP, with
    300 per minute for holds, checkout, and release. CORS is allowlisted only on
    `/v1/public/*`, never allows credentials, and has no wildcard origin.
11. **Scanner admission.** Each show gets an independent scanner token through
    `POST /v1/internal/shows/{id}/scanner-token`; the plaintext is returned once and
    only a SHA-256 hash is stored. `POST /v1/scan/admit` requires `X-Scanner-Token`
    plus `{"qr_token": ...}`, verifies the QR signature, and admits only tickets that
    belong to that show. The state change is one atomic
    `UPDATE tickets SET status='used', used_at=now() WHERE id=$1 AND show_id=$2 AND
    status='issued'`. A second scan returns 409 (`ticket_already_used` or
    `ticket_void`), a ticket from another show returns 403, and an unknown or
    tampered QR token returns 404.