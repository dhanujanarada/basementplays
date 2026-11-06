# Architecture overview

- **Status:** partial
- **Covers:** `src/app/`, `src/lib/`, `src/collections/`, `src/globals/`, `ticketing/`, `package.json`
- **Last checked:** 8de3320

Basement Plays is one repository with two runnable services:

- the **site**, a Next.js App Router application that also runs Payload CMS;
- the **ticketing service**, a Go API with its own Postgres schema.

The split is deliberate: content and presentation change often and are safe to
deploy together, while money and inventory need transactions and a small,
stable surface. See [0009](../decisions/0009-separate-go-service.md).

## Services

```
Browser
  |
  |-- public pages ----------> Next.js site (server components)
  |                              |-- Payload CMS ----------> Payload database
  |                              |                           (SQLite today)
  |                              `-- /t/<token> -----------> Go public API
  |
  |-- booking UI (client) ----> ticketing client
  |                              |-- mock mode (default)
  |                              `-- /ticket-api (planned) --> Go public API --> Postgres
  |
  `-- /backstage -------------> Payload admin

Staff
  `-- /backstage-tools (planned) --> Next.js server --> Go internal API
```

The current demo uses mock mode, so the browser keeps holds, orders, and buyer
details in `sessionStorage` and no Go service is required. The HTTP client and
the Go service exist, but the browser-facing `/ticket-api` proxy is not built
yet. The account for that gap is in `docs/DISCREPANCIES.md` (D5).

## Request flow: public content

1. A request reaches a Next.js server component in `src/app/(frontend)/`.
2. The component calls `src/lib/payload.ts`, the only module that talks to
   Payload.
3. Payload reads its local database and returns content.
4. `src/lib/payload.ts` flattens the raw event into an `EventView`.
5. A design in `src/designs/` renders the event. Booking and ticket routes reuse
   that design's `Shell` so they inherit the same palette and footer.

The current Payload adapter is SQLite. Decision
[0008](../decisions/0008-postgres-everywhere.md) says Postgres everywhere, but
the site switch has not happened yet.

## Request flow: booking

1. The event page's `SeasonBooking` component selects a show and links to
   `/events/<slug>/book?show=<id>`.
2. `BookingFlow` loads a show summary and, in seats mode, a seat map through
   `src/lib/ticketing/client.ts`.
3. The client is a lazy proxy: it loads the mock module or the HTTP module only
   when a method is called.
4. A hold creates a pending order. Checkout captures buyer details.
5. The confirmation page reads the order token from `sessionStorage`, polls the
   order, and reveals ticket links when the order is paid.

The mock is intentionally excluded from first-load route bundles. An HTTP-mode
production build contains no mock chunks.

## Request flow: ticket page

`/t/<token>` is a server route. It asks the ticketing client for the ticket,
verifies the signed token in Go (or the mock), loads the production content from
Payload, and renders `TicketView` inside that production's design shell. The
public ticket never exposes a buyer name.

## Boundaries and failures

- Payload and ticketing never share tables. The Go service reads its own
  `ticketing` schema only.
- The browser does not call the Go internal API. The Next.js server does, with a
  service token.
- If the ticketing service is down, content pages still render. Booking shows an
  error and the event page keeps the WhatsApp fallback when ticketing is off.
- Staff pages are planned. Until they exist, manual bank-transfer orders are
  marked paid through the Go internal API, not through the public site.

## Current gaps

The frontend revamp is ahead of the Go revamp. The site and mock understand
quantity tiers, payment modes, and stage positions; the current Go schema and
handlers still use seat plans and seat-only holds. The code is documented as it
is today, and the target contract is recorded separately in
[api-contract.md](api-contract.md). The temporary discrepancy list will be
created in D5.
