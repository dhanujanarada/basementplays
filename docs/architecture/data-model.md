# Data model

- **Status:** partial
- **Covers:** `src/collections/`, `src/globals/SiteSettings.ts`, `src/lib/payload.ts`, `src/lib/event-view.ts`, `ticketing/migrations/`
- **Last checked:** 8de3320

There are two deliberately separate data stores: Payload owns content, and the Go
service owns ticketing. They do not join across databases.

## Payload: Events (`productions`)

The collection slug stays `productions`, but the admin label is **Events**. One
editor holds everything about an event in four tabs.

| Tab | Fields |
| --- | --- |
| Details | `title`, `slug`, `tagline`, `presentedBy`, `writtenBy`, `directedBy`, `synopsis`, `poster`, `heroImage` |
| Shows & tickets | `venue`, `shows[]`, `ticketing` group |
| People | `cast[]`, `crew[]`, `partners[]` |
| Gallery | `gallery[]` |

`shows[]` rows contain `startsAt`, an optional `venue` override, and
`cancelled`. The `ticketing` group contains `paymentMode`, `selectionMode`,
`stagePosition`, `numbering`, `columns`, and `tiers[]`. Tiers contain `name`,
whole-rupee `price`, and `count`.

Show row ids and tier row ids are Payload-generated stable strings. They are the
show ids used in `?show=` and the ids the ticketing service is meant to receive.
The current Go service uses a different show/seat model; see the gap section.

## Payload: other collections and global

- **Media** — uploads with required `alt` text; `thumbnail` (400 px), `card`
  (800 px), and `full` (1920 px) variants.
- **Users** — auth collection with `admin` and `member` roles, five login
  attempts, and a ten-minute lock.
- **Site Settings** — public singleton with group name, tagline, about text and
  image, email, phone, WhatsApp number, location, maps URL, social links, and
  the `bankTransfer` group.

The `EventView` in `src/lib/event-view.ts` is the normalized shape designs
consume. It flattens shows, attaches the default or overridden venue, derives
`open`/`few-left`/`sold-out`/`cancelled`/`past`, and derives whether ticketing is
active. It does not read Payload directly.

## Ticketing: current Postgres schema

Migrations live in `ticketing/migrations/`. The current schema is the earlier
seat-plan design, not the tier-generated target.

| Table | Purpose |
| --- | --- |
| `seat_plans` | `key`, `name`, JSONB `layout`; one plan can be reused by shows |
| `shows` | show id, production slug, title, time, venue, `seat_plan_id`, status, currency, scanner token fields |
| `show_seats` | one row per seat: `tier_key`, `tier_name`, label, price, sort order, status, order link, hold expiry |
| `orders` | pending/paid/expired/cancelled/needs_review order, buyer fields, total, access-token hash, expiry |
| `order_items` | one row per ordered seat, linked to the order and show seat |
| `tickets` | issued/used/void tickets, one live ticket per seat |
| `payments` | provider notifications and amounts |
| `audit_log` | append-only staff/system actions |

Important constraints:

- A `show_seats` row is available, held, sold, or blocked. Held requires an
  order and expiry; sold requires an order; available and blocked have neither.
- An order and its seats must belong to the same show.
- One live ticket per seat is enforced by a partial unique index on
  `tickets (show_id, seat_code)` where status is `issued` or `used`.
- Money is integer cents; times are `timestamptz`.

## Target ticketing model

The revamp target is tier-generated seating: `show_tiers`, `show_seats` with
`tier_id`, and show payment/selection/stage settings. That model is described in
[0021](../decisions/0021-generated-seating.md) and
[0023](../decisions/0023-payment-modes.md). It is not in the migrations yet; the
current Go code still contains `seat_plans` and `tier_key` fields. The
temporary discrepancy list will be created in D5.
