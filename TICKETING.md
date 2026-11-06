# TICKETING.md: Go ticketing service + booking UI

## 1. Scope (now)
In: seats generated from each event's tier list when a show is upserted, per-show
seat inventory, availability API, tier and seat holds, orders, manual
(WhatsApp/bank-transfer) orders and ticket issue, signed QR tickets, ticket page,
public booking UI, ONE booking card per event wired to the booking flow (behind a
derived active check), Backstage ticketing actions (orders, mark paid, cancel,
re-check payment, block seats), PayHere SANDBOX (last).
Out: scanner UI (a separate mobile app comes later), PayHere live mode, refund
automation, discount codes, public accounts, multi-currency, drawn or free-form seat
maps, seat-plan templates/builder/plan keys, deployment automation.
The backend exposes an "admit" endpoint (step 7, optional) for the future scanner app.
The revamp decisions that supersede older parts of this document are recorded in §14.

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
- The site's Payload keeps its current database adapter for the revamp (SQLite for
  local dev; Postgres remains future work). The Go service uses schema `ticketing`
  with its own migrations. Never join across Payload's tables, and the Go service
  never reads Payload tables directly.
- Money: integer minor units (`*_cents`, currency LKR). Format "2500.00" only at the
  PayHere boundary. Times: timestamptz UTC; display in Asia/Colombo.
- Layout:
```
ticketing/
  cmd/ticketing/main.go       # serve | migrate
  cmd/seed/main.go            # dev seed
  internal/config  internal/db  internal/domain   # state machines, types
  internal/seatgen  internal/store  internal/service
  internal/httpapi            # handlers, auth, rate limit, request ids, CORS
  internal/payhere  internal/qr
  migrations/  Makefile       # make dev | test | migrate | db-create
```
- Env: DATABASE_URL, INTERNAL_TOKEN, QR_SECRET, PUBLIC_BASE_URL, HOLD_MINUTES=10,
  CHECKOUT_HOLD_MINUTES=15, MANUAL_HOLD_HOURS=24, MAX_SEATS_PER_ORDER=6,
  MAX_SHOW_SEATS=500, CORS_ALLOWED_ORIGINS (default empty), PAYHERE_MERCHANT_ID,
  PAYHERE_MERCHANT_SECRET, PAYHERE_SANDBOX=true, PAYHERE_ALLOW_LIVE=false.
  Fail fast if required vars are missing.

## 3. Data model (schema `ticketing`)
- shows(id text pk = Payload show row id, production_slug, title, starts_at, venue,
  payment_mode in [bank_transfer,online,off], selection_mode in [quantity,seats],
  stage_position in [front,back,none], columns int default 10, numbering boolean,
  currency, status in [draft,on_sale,closed,cancelled], updated_at,
  scanner_token_hash null, scanner_token_set_at null). There is no stored
  ticketing-enabled flag: the site only syncs active shows (§6, §9.3).
- show_tiers(show_id, tier_id text, name, price_cents, seat_count, sort_index,
  pk(show_id, tier_id), FK show_id -> shows). `tier_id` is the Payload tier array
  row id; `sort_index` follows the tier array order (first = nearest the stage).
  Names and prices are snapshots kept in sync as described in §6.
- show_seats(show_id, seat_code, tier_id, tier_name, label null, price_cents,
  sort_index, status in [available,held,sold,blocked], order_id null,
  held_until null, pk(show_id, seat_code)), FK (show_id, tier_id) -> show_tiers.
  CHECKs: held => order_id and held_until set; sold => order_id set; available =>
  both null. `sort_index` is the global display order (tier order first, then seat
  order). Tier name and price are snapshots taken when the show's seat rows are
  created and updated per §6.
- There is NO `seat_plans` table, plan key, template, builder page, or plan endpoint.
  Seats are generated from the event's tiers when a show is upserted (§4).
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
Ticket wording: quantity mode => "{tier_name} admission"; seats mode with a label =>
"Seat {label} ({tier_name})"; seats mode without a label => "{tier_name} admission".

## 4. Seats generated from tiers
There is no plan template, builder, duplicate action, label editor, or plan key.
Seats are generated from the event's ticketing group every time a show is upserted:
- `tiers`: `[{tier_id, name, price_cents, count}]` in display order (first = nearest
  the stage). `tier_id` is the Payload tier array row id.
- `selection_mode`: `quantity` or `seats`.
- `stage_position`: `front`, `back`, or `none`.
- `numbering`: seats mode only; true = labels "1".."N", false = null labels.
- `columns`: grid width for seats mode (default 10).
- `generateSeats(params)` is a pure function in `internal/seatgen` with tests; the
  show upsert calls it. Input validation returns typed errors for an empty tier list,
  an empty tier name, duplicate tier ids, a count below 1, a negative price, or more
  than `MAX_SHOW_SEATS` (500) seats.
- One `show_seats` row is created per ticket. Codes are `S001`, `S002`, ... across all
  tiers in tier order. `sort_index` follows the same global order: tier order first,
  then seat order.
- `tier_id` and `tier_name` are copied from the tier; `price_cents` comes from the
  tier; labels follow the numbering rule. Quantity-mode seats always have null labels.
- Display settings (`columns`, `stage_position`, `numbering`) and tier names always
  apply. Price changes apply only to seats whose status is `available`; held, sold,
  and blocked seats keep their price snapshot.
- The stage bar is a UI concern: `front` = above the tiers; `back` = below, with the
  tiers rendered reversed so the first tier is still nearest the stage; `none` = no
  bar. The label is "Stage".
- There is no reusable sample plan. The dev seed may create a placeholder event, but
  it is not a seat-plan template.

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
- Hold request shapes: quantity mode accepts
  `{"tiers":[{"tier_id":"...","quantity":2}, ...]}`; seats mode accepts
  `{"seats":["S001","S002"]}`. Reject the wrong shape for the show with 400.
- Hold (atomic, no double booking): in one transaction, claim the requested inventory;
  retry up to 3 times on deadlock (40P01) or serialization failure (40001).
  - Seats mode: lock the requested seats in SORTED order, then claim them.
    ```sql
    SELECT seat_code FROM ticketing.show_seats
     WHERE show_id=$1 AND seat_code = ANY($2) ORDER BY seat_code FOR UPDATE;
    UPDATE ticketing.show_seats SET status='held', order_id=$3, held_until=$4
     WHERE show_id=$1 AND seat_code = ANY($2)
       AND (status='available' OR (status='held' AND held_until < now()));
    -- rows affected must equal len(seats), otherwise roll back and return 409 seat_taken
    ```
  - Quantity mode: for each requested tier, pick the lowest `sort_index` available
    seats (expired holds count as available) in the same transaction.
    ```sql
    SELECT seat_code FROM ticketing.show_seats
     WHERE show_id=$1 AND tier_id=$2
       AND (status='available' OR (status='held' AND held_until < now()))
     ORDER BY sort_index
     FOR UPDATE SKIP LOCKED
     LIMIT $3;
    -- if fewer rows than requested, roll back and return 409 insufficient_inventory
    -- otherwise UPDATE the returned seat codes to held, then continue with the next tier
    ```
    The total requested quantity must stay within MAX_SEATS_PER_ORDER.
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
- Show upsert: idempotent PUT by show id. Send `payment_mode`, `selection_mode`,
  `stage_position`, `columns`, `numbering`, and `tiers` (see §7). Generate seat rows
  on the first upsert (§4). On later upserts:
  - tier names and display settings always apply;
  - price changes apply only to seats whose status is `available`;
  - a structural change (tier add/remove/reorder, or a count change) is rejected with
    409 `seating_locked` once the show has any order or any seat that is not
    `available`. The site surfaces that warning to the editor and does not silently
    pretend the sync succeeded;
  - an event whose derived ticketing state becomes inactive is set to `closed`, not
    deleted, so existing orders and tickets stay valid. `closed` stops new holds.
- QR token: `v1.<ticket_uuid>.<base64url(HMAC-SHA256(QR_SECRET, "v1|"+ticket_uuid))>`,
  verified with constant-time compare. The QR encodes `{PUBLIC_BASE_URL}/t/<token>`.
- Admit (for the later app): atomic, single statement.
```sql
UPDATE ticketing.tickets SET status='used', used_at=now()
 WHERE id=$1 AND status='issued';  -- 1 row = admit; 0 rows = already used or void
```

## 7. API surfaces
- Public `/v1/public/...` (no auth, rate-limited per IP, body size limits, strict JSON):
  GET shows/{id} (summary: status, payment_mode, selection_mode, stage_position,
  columns, currency, tiers with tier_id/name/price_cents/seats_total/seats_remaining,
  total_seats, seats_remaining, price_from),
  GET shows/{id}/seats (tiers + seats with status, label, tier_id, tier_name,
  sort_index; ETag),
  POST shows/{id}/holds (body per §6), DELETE orders/{id} (release),
  POST orders/{id}/checkout, GET orders/{id}, GET tickets/{token} (includes
  production_slug so the site can theme the ticket page). `DELETE orders/{id}`,
  `POST orders/{id}/checkout` and `GET orders/{id}` all require the `X-Order-Token`
  header; store only its hash and compare in constant time. The order response for a
  bank-transfer checkout includes the bank details, order reference, and payment
  deadline the buyer must see.
- Internal `/v1/internal/...` (Bearer INTERNAL_TOKEN, Next.js server only; never
  exposed publicly):
  shows: PUT shows/{id} (idempotent upsert taking payment_mode, selection_mode,
  stage_position, columns, numbering, tiers; generates/updates show_tiers and
  show_seats per §4 and §6);
  orders: POST orders (manual), POST orders/{id}/mark-paid, POST orders/{id}/cancel,
  POST orders/{id}/recheck-payment, GET shows/{id}/orders; PUT seats/{code}/block.
  There are no seat-plan endpoints.
- Callbacks: POST /payhere/notify (public, signature verified).
- Scanner (optional step 7): POST /v1/scan/admit with a per-show scanner token in
  `X-Scanner-Token`. Internal `POST/DELETE /v1/internal/shows/{id}/scanner-token`
  issues (plaintext returned once) or revokes it; only the hash is stored.
- Standard errors: 400 validation or wrong hold shape, 401/403 auth, 404, 409
  `seat_taken`, `insufficient_inventory`, `seating_locked`, or bad state, 429 rate
  limit. Request IDs on every response.

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
Do not rebuild the UI. The revamp replaces Runs with shows inside the Events collection
and adds one shared event-page booking component. Every existing Reserve/Book button is
either removed or becomes an anchor to that single booking card.

9.1 Design contract and `EventView`. Each entry in `designs` stays
`{ Page, Shell, Booking? }`:
- `src/lib/payload.ts` returns a normalized `EventView`: the production fields plus
  `shows` flattened from the event's shows array with derived status, the ticketing
  settings/tiers, and the people, partners, and gallery content. No design imports a
  Run or reads raw Payload rows.
- `Page` receives `{ event }`; `Shell` receives `{ event, children }`; `Booking`
  receives `{ event, show }`.
- An `EventShow` carries `id`, `startsAt`, `venue`, `cancelled`, derived status
  (`open`, `few-left`, `sold-out`, `cancelled`, `past`), price/tier display data, and
  availability when ticketing is active.
- `Shell` provides the design's CSS variables, fonts, textures, header and themed
  footer. It is used by Page AND by the booking and ticket routes.
- `Booking` (optional): a fully custom booking layout for that event. If missing, the
  shared `DefaultBooking` renders inside the design's `Shell`.

9.2 One booking card and one Reserve action.
- The shared client component `SeasonBooking` owns the event's shows, the selected
  show (default = next upcoming; otherwise the most recent with a "season ended"
  state), and `?show=<id>` synced with `history.replaceState` and honored on load.
- Each design passes its existing card markup as a render prop, so the Dear Daughter
  card keeps its current look. The show-dates list becomes the selector: selecting a
  show updates the card and `?show=`. Past shows are dimmed and not selectable;
  cancelled shows are marked.
- `bookingHref(event, show)` is the only source of the reserve link:
  - ticketing active => `/events/<slug>/book?show=<show.id>`, label "Reserve seats";
  - ticketing inactive, or the show is not bookable => the Site Settings WhatsApp
    fallback with a prefilled message containing the title and date.
- The booking card is the ONLY place with a reserve action. The hero CTA and the
  mobile sticky bar become "Get tickets" anchor links that scroll to the card. Past,
  cancelled, and sold-out shows render a disabled state with their status text.

9.3 Payload changes.
- The collection slug stays `productions`; the admin label becomes "Events". Tabs:
  - **Details**: title, slug, tagline, presentedBy (default "A Basement production"),
    writtenBy, directedBy, synopsis, poster.
  - **Shows & tickets**: venue (default for all shows); shows array (startsAt
    required, optional venue override, cancelled checkbox); ticketing group:
    paymentMode (`bank_transfer` default | `online` | `off`), selectionMode
    (`quantity` default | `seats`), stagePosition (`front` default | `back` | `none`),
    numbering (checkbox, seats mode only), columns (number, default 10, seats mode
    only), tiers array (name, price in rupees, count; min 1 each).
  - **People**: cast, crew, partners.
  - **Gallery**: gallery (image, caption, category On stage / Behind the scenes).
- Remove the Runs and Props collections and everything attached to them (hooks, seeds,
  docs). Remove `ticketingEnabled`, `seatPlanKey`, `ticketUrl`, `price`, `seatsNote`,
  and the manual status text.
- Show array row ids are the show ids used in `?show=` and in Go. Tier array row ids
  are the tier ids and stay stable across rename and reorder. This was confirmed
  empirically against the current SQLite adapter (cast and show rows preserved their
  ids through a reverse and a rename, then were restored).
- Site Settings gains a `bankTransfer` group: bankName, accountName, accountNumber,
  branch, instructions. Keep `whatsappNumber` for the fallback.
- Backstage navigation: Events, Media, Users, and the Site Settings global. Nothing
  else.
- Derived ticketing state (no stored flag): active iff paymentMode != `off` AND at
  least one tier has price > 0 and count > 0.
- Sync: an `afterChange` hook on Events computes the derived state and calls the
  internal PUT shows/{id} for each active show (id = the show's stable array row id;
  send production_slug, title, starts_at, venue, payment_mode, selection_mode,
  stage_position, columns, numbering, tiers). Server-side only. If Go is unreachable,
  do not lose the save: log it, show an admin warning, and provide a "Re-sync" action.
  An event that becomes inactive sets its Go shows to `closed`.
- Deleting an event: block the delete when Go reports any orders (paid, pending, or
  held seats) for its shows. With none, allow the delete and set each Go show to
  `closed`. If Go is unreachable, block the delete (fail closed).
- If Go rejects a structural tier change with 409 `seating_locked`, surface the
  warning to the editor. There is no plan-key validation and no plan builder.

9.4 Routes, all rendered inside the production's Shell:
- `/events/[slug]`: the event page (SeasonBooking). The selected show comes from
  `?show=`; without it, use the next upcoming show, or the most recent show with a
  "season ended" state.
- `/events/[slug]/book?show=<id>`: if there is no `show` param and exactly one show is
  on sale, use it; if several, show a date picker step first.
- `/events/[slug]/book/confirmation?order=<id>`: polls order status. The order access
  token lives in sessionStorage and is sent as a header. For bank transfer it shows the
  bank details, order reference, and deadline; once paid it reveals the ticket link.
- `/t/[token]`: ticket page. Look up the ticket via Go, then render inside that
  production's Shell, with seat or tier text, show, a short order reference (last six
  characters of the order UUID, uppercased), and a QR image. No buyer name.
- Production content comes from Payload on the server; availability and holds come
  from the browser via `NEXT_PUBLIC_TICKET_API_BASE` (see section 12).

9.5 Selection UI and stage bar.
- Quantity mode: one shared tier-box UI showing each tier's name, price, remaining
  count, a quantity stepper, and a sold-out state, plus the stage bar. It is themed
  through the existing CSS variables.
- Seats mode: keep the existing shared `SeatMap` (tier sections, seat tiles, legend,
  keyboard support, ~10s polling) and add the stage bar.
- Stage bar: `front` = above the tiers; `back` = below with the tiers rendered in
  reverse so the first tier is still nearest the stage; `none` = no bar. Label is
  "Stage".
- Both modes allow at most `MAX_SEATS_PER_ORDER` (6) tickets per order, show a
  selection list with subtotals and a total, do not rely on colour alone, and have
  tap targets of at least 32px.
- dear-daughter mapping: available = --paper with a --wood outline; selected =
  --accent; held = hatched --accent at reduced opacity; sold = muted --bg-deep with
  hatch; stage = --wood. Booking header: compact damask strip with the title, date, and
  venue; torn-paper top edge; panel on --paper; terracotta primary buttons; display font
  headings; themed footer.
- PlainDesign provides neutral variables.

9.6 Booking flow: choose tiers (quantity) or seats -> hold (countdown timer) -> buyer
form (name, phone, email) -> pay using the event's `paymentMode` -> confirmation page.
The buyer does NOT choose the payment method:
- `bank_transfer`: the confirmation shows the Site Settings bank details, the short
  order reference, and the deadline; staff mark the order paid; the page then shows the
  ticket link.
- `online`: PayHere sandbox per §8.
- `off`: the card uses the WhatsApp fallback and no booking page is offered.
Handle 409 `seat_taken` (seats mode) and `insufficient_inventory` (quantity mode)
gracefully by refreshing and keeping the rest of the selection.

9.7 Backstage ticketing pages. Standalone Next.js pages under `/backstage-tools/...`
that require a logged-in Payload user (redirect unauthenticated visitors to the Payload
login) and call the Go internal API server-side with INTERNAL_TOKEN (no separate Go
admin UI). This is the chosen option over a Payload custom admin view: no import map or
component-path setup, and the existing `Disallow: /backstage` robots rule covers the
path. No public links to these pages. Pages:
- Show ticketing: orders list, create a manual order for a WhatsApp reservation by
  picking tiers or seats, mark paid, cancel, re-check payment, and block seats. Show the
  bank transfer reference and deadline for manual orders, and let staff copy the ticket
  link after payment.
- There are no seat-plan pages. Seats come from the event's tiers (§4).

9.8 Add a Privacy Policy page before public launch (we collect names, phones, emails).

## 10. Build steps (stop after each, say how to verify, wait for "next")
1. Scaffold: Go module, config, native-Postgres setup docs and `make db-create`,
   migrations runner (`ticketing migrate`, advisory-lock guarded at startup), /healthz,
   /readyz, slog, Dockerfile, `.env.example`; root `tsconfig.json` and ESLint exclude
   `ticketing/`. Verify: `make dev`, probes OK, migrations apply.
2. Schema + domain state machines + tests (no HTTP).
3. Seat generation inside the show upsert + tests; persist tiers; no plan CRUD and no
   sample plan. Verify: upserting a show creates/updates `show_tiers` and `show_seats`
   in tier order, and structural changes are locked once inventory is used.
4. Availability + holds + expiry + release. REQUIRED tests under `go test -race`:
   200 goroutines racing for one seat give exactly one winner; overlapping multi-seat
   holds do not deadlock or oversell; quantity-mode picks the lowest available seats per
   tier with no double booking; expired holds are reclaimable; correctness holds with the
   sweeper disabled. The sweeper uses `pg_try_advisory_lock`.
5. Orders, checkout, manual orders, MarkPaid (idempotent), cancel, needs_review path,
   Idempotency-Key handling.
6. QR sign/verify, ticket lookup endpoint, internal auth, rate limiting, CORS allowlist.
7. (Optional) admit endpoint + tests.
8. Site integration, in this order:
   8a. Audit the existing UI, introduce `EventView` and the `{Page, Shell, Booking?}`
       contract, add `SeasonBooking`, `bookingHref`, and the one-card rule. With
       ticketing inactive, behaviour must be identical to today. (Any time after step 1.)
   8b. Payload Events collection fields, derived ticketing state, sync hook, and the
       Site Settings bank-transfer group. Keep the current dev database adapter.
   8d. Booking UI on `/events/<slug>/book` for both modes (tier boxes for quantity,
       seat map for seats, stage bar), with `NEXT_PUBLIC_TICKET_API_BASE` and the runtime
       proxy route.
   8e. Holds, countdown, buyer form, confirmation page (including bank transfer details,
       reference, deadline, and ticket link when paid).
   8f. Ticket page `/t/[token]` with QR.
   8g. Backstage show-ticketing page (manual orders, mark paid, cancel, re-check
       payment, block seats, copy ticket link).
   8h. MEDIA_STORAGE switch for Payload (any time after step 1).
9. PayHere sandbox integration + "Re-check payment" + tests.
10. `deploy/README.md` documenting setups A/B/C, plus `deploy/backup.sh`. Stop.

> **Frontend-first track.** The current frontend work covers step 8a, the frontend part
> of 8b, and mock-backed versions of 8d–8f. It runs against a mock client so the
> booking UI can be demoed before the Go service is deployed. Real wiring is a separate
> later step: set the API mode to `http`, add the runtime proxy route, and reconcile the
> TypeScript contract with the Go API.
>
> **Demoing the mock locally.** The seeded Dear Daughter event is active (quantity
> mode, bank transfer), so the event page's booking card opens the booking flow. No Go
> service is required. The mock keeps holds, orders, and buyer data in `sessionStorage`,
> so book and confirm in the same tab (a new tab cannot recover the order token). It
> supports both selection modes, the existing `?scenario=` values, and a dev-only
> "mark paid" button on bank-transfer confirmations. Seat-plan templates are neither
> needed nor supported. `NEXT_PUBLIC_TICKET_API_MODE=http` intentionally throws until
> the HTTP client and `/ticket-api` proxy are implemented.


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
2. **Deleting an event.** Block the delete when Go reports any orders (paid, pending,
   or held seats) for the event's shows. With no such orders, allow the delete and set
   each Go show to `closed`. If Go is unreachable, block the delete (fail closed).
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
7. **Show upsert and structural locking.** The event save sends the full ticketing
   configuration and tiers to Go. Names and display settings always apply; prices apply
   only to available seats. A structural tier change is rejected with 409
   `seating_locked` once the show has any order or any non-available seat. The site
   surfaces the warning to the editor. There is no plan key and no plan validation.
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

## 14. Revamp decisions

This section records the ticketing/UI revamp. Where it conflicts with an older
paragraph above, this section wins. The revamp changes ONLY what is listed here;
everything else must look and behave as it did at tag `pre-revamp`. The parity
baseline is `docs/ui-parity.md`.

### 14.1 Events data model

- The Payload collection slug stays `productions`; the admin label is **Events**. One
  editor, four tabs:
  - **Details:** title, slug, tagline, presentedBy (default "A Basement production"),
    writtenBy, directedBy, synopsis, poster.
  - **Shows & tickets:** venue (default for all shows); shows array (startsAt
    required, optional venue override, cancelled checkbox); ticketing group:
    paymentMode (`bank_transfer` default | `online` | `off`), selectionMode
    (`quantity` default | `seats`), stagePosition (`front` default | `back` | `none`),
    numbering (checkbox, seats mode only), columns (number, default 10, seats mode
    only), tiers array (name, price in rupees, count; min 1 each).
  - **People:** cast, crew, partners.
  - **Gallery:** gallery (image, caption, category On stage / Behind the scenes).
- Remove the Runs and Props collections and everything attached to them (hooks,
  seeds, docs). Remove `ticketingEnabled`, `seatPlanKey`, `ticketUrl`, `price`,
  `seatsNote`, and the manual status text.
- The show array row id is the show id used in `?show=` and in Go. The tier array row
  id is the tier id. Both are Payload-generated 24-hex text and were confirmed to stay
  stable across rename and reorder against the current SQLite adapter.
- Backstage navigation is Events, Media, Users, and the Site Settings global. Nothing
  else.
- Site Settings gains a `bankTransfer` group: bankName, accountName, accountNumber,
  branch, instructions. `whatsappNumber` stays for the fallback.

### 14.2 Derived rules (no stored flags)

- Ticketing is ACTIVE for an event iff `paymentMode != off` AND at least one tier has
  price > 0 and count > 0.
- Inactive events show the WhatsApp fallback: a `wa.me` link built from the Site
  Settings `whatsappNumber` with a prefilled message that includes the title and date.
- Badges:
  - **cancelled** from the show's checkbox;
  - **past** from the show's date;
  - **sold out / few left** from Go availability when ticketing is active, and hidden
    when it is not.

### 14.3 Seating, inventory, and holds

- Go generates seats from the tiers when a show is upserted: codes `S001`, `S002`,
  ... and `sort_index` in tier order (§4). Quantity-mode seats have null labels.
- There are NO seat-plan templates, builder page, plan key, duplicate/label actions,
  or plan endpoints. `seat_plans` is removed from the spec.
- Structural changes (tier add/remove/reorder/count) are rejected with 409
  `seating_locked` once the show has any order or any non-available seat. Prices apply
  to available seats only; names and display settings always apply.
- `POST holds` accepts `{"tiers":[{"tier_id","quantity"}]}` for quantity mode or
  `{"seats":[codes]}` for seats mode. Quantity mode picks the lowest `sort_index`
  available seats per tier using `FOR UPDATE SKIP LOCKED` in the same transaction. If
  fewer than requested are obtainable, return 409.
- Ticket wording: quantity mode `"{tier} admission"`; seats mode
  `"Seat {label} ({tier})"`, or `"{tier} admission"` when unlabeled.
- Stage bar: `front` above the tiers; `back` below with the tiers reversed so the first
  tier stays nearest the stage; `none` no bar.

### 14.4 Payment modes

- `bank_transfer`: the hold lasts `MANUAL_HOLD_HOURS`. The buyer sees the Site Settings
  bank details, a short order reference, and the deadline. Staff mark the order paid in
  Backstage; the confirmation page then shows the ticket link, which staff can copy.
- `online`: PayHere sandbox per §8.
- `off`: the WhatsApp fallback; no booking page is offered.

### 14.5 Event page

- One shared client component `SeasonBooking` owns the event's shows, the selected
  show (default = next upcoming; if none, the most recent with a "season ended" state),
  and `?show=<id>` kept in sync with `history.replaceState` and honored on load.
- Each design passes its existing card markup as a render prop. The Dear Daughter card
  must look exactly as it did at `pre-revamp`.
- The show-dates list is the selector. Selecting a show updates the card and `?show=`;
  past shows are dimmed and not selectable; cancelled shows are marked.
- The booking card is the ONLY reserve action ("Reserve seats" ->
  `/events/<slug>/book?show=<id>`). Every other Reserve button is removed. The hero
  CTA and mobile sticky bar become "Get tickets" anchors that scroll to the card.
- Events listing and home cards link to the event page.

### 14.6 Staged delivery

Work in stages and stop after each one with exact verification URLs and clicks:

- **R0 (docs only):** commit the pre-revamp state, tag it `pre-revamp`, write
  `docs/ui-parity.md`, and update this document and `DESIGN.md`.
- **R1 (Payload + data layer):** implement §14.1 and §14.2. `src/lib/payload.ts`
  returns a normalized `EventView` (flattened shows with derived status); designs
  consume only `EventView`, with zero visual change. Re-seed Dear Daughter with three
  clearly-marked placeholder future shows, venue "The Black Bean Cafe, Kandy", one
  "General" tier at Rs. 2500 × 50, quantity mode, stage front, bank transfer; add a
  fictional dev-only event on PlainDesign if none exists. Resetting the dev database
  is fine. Verify one Events editor with four tabs and the parity checklist.
- **R2 (contract, mock, booking page):** update the types, mock client, and HTTP
  client for §14.3 and §14.4; the mock supports both selection modes and the existing
  `?scenario=` values. Default booking UI is tier boxes (name, price, remaining,
  quantity stepper, sold-out state, stage bar) themed through the existing CSS
  variables. Keep `SeatMap` for seats mode and add the stage bar. Bank-transfer
  confirmation shows details/reference/deadline, polls status, and reveals the ticket
  link when paid (the mock gets a dev-only "mark paid" button). Keep the PayHere mock
  for online mode and the WhatsApp fallback.
- **R3 (event page):** implement §14.5 for every design. Verify on
  `/events/dear-daughter`: default selection is the next show, selecting another show
  updates the card and `?show=`, past shows are dimmed, the card is the only reserve
  action, and the anchors scroll to it.
- **R4 (lean cleanup):** knip and depcheck; remove unused files, exports, and
  dependencies; enable TypeScript unused-locals checks; run a one-time bundle analysis
  and remove the tooling; lazy-load the lightbox and QR libraries; keep the mock client
  out of the production bundle; remove provably duplicate/unused CSS. Go changes are
  docs, `go mod tidy`, and `go vet` only. Report before/after lines of code,
  dependency count, and `next build` route sizes.

### 14.7 Non-negotiables

- Keep the WhatsApp fallback; no new paid services.
- Money is whole rupees in Payload and integer cents in Go and the API.
- Typecheck, lint, and build pass at the end of every stage.
- Do not touch Go code beyond docs, `go mod tidy`, and `go vet` during this revamp.
- Everything not listed in this section keeps the pre-revamp look and behaviour.