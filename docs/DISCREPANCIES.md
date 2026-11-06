# Code / spec discrepancies

- **Status:** implemented
- **Covers:** `TICKETING.md`, `DESIGN.md`, `src/`, `ticketing/`
- **Last checked:** 8de3320

This is a temporary list of places where the written specs and the code disagree.
The documentation in `docs/` describes the **code as it is today**. Where the code
is behind the revamp target, that is recorded here instead of being presented as
implemented.

Delete this file when every item is either implemented or the spec is corrected.

## D-01 Payload still uses SQLite

- **Code:** `src/payload.config.ts` uses `sqliteAdapter`; `.env.example` defaults
  to `file:./payload.db`.
- **Spec:** Decision [0008](decisions/0008-postgres-everywhere.md) says Postgres
  everywhere and SQLite dropped from the site.
- **State:** open. The Go service uses Postgres; the site switch has not happened.

## D-02 Go still uses seat plans and `tier_key`

- **Code:** `ticketing/migrations/00002_schema.sql` has `seat_plans`,
  `shows.seat_plan_id`, and `show_seats.tier_key`; `internal/seatplan` generates
  layouts.
- **Spec:** `TICKETING.md` §3, §4, and §14 remove seat plans, add `show_tiers`,
  and use `tier_id` from event tiers.
- **State:** open. The site and mock use tiers; Go still uses plans.

## D-03 Show upsert takes a plan key, not tiers

- **Code:** `internal/service/shows.go` `UpsertShowInput` has `seat_plan_key`,
  `status`, and currency. It has no tiers, payment mode, selection mode, stage
  position, numbering, or columns.
- **Spec:** `TICKETING.md` §6 and §7 say the upsert takes the ticketing
  configuration and tiers and generates seats from them.
- **State:** open.

## D-04 Seat-plan internal endpoints still exist

- **Code:** `internal/httpapi/internal.go` mounts `POST/GET /seat-plans`,
  `GET /seat-plans/{key}`, `PATCH /seat-plans/{key}/labels`, and
  `POST /seat-plans/{key}/duplicate`.
- **Spec:** `TICKETING.md` §7 and §14 say there are no seat-plan endpoints.
- **State:** open.

## D-05 Holds accept seats only

- **Code:** `POST /v1/public/shows/{id}/holds` decodes `{"seats":[...]}` only.
  `Holds.Create` takes a seat list.
- **Spec:** `TICKETING.md` §6 and §14 accept `{"tiers":[{"tier_id","quantity"}]}`
  in quantity mode and `{"seats":[...]}` in seats mode.
- **State:** open.

## D-06 Quantity-mode allocation and SKIP LOCKED are missing

- **Code:** the hold query locks requested seat codes with `FOR UPDATE`, sorted by
  code. It does not select the lowest `sort_index` available seats per tier, and
  it does not use `SKIP LOCKED`.
- **Spec:** `TICKETING.md` §6 says quantity mode picks the lowest available seats
  per tier using `FOR UPDATE SKIP LOCKED` and returns `insufficient_inventory`
  when it cannot satisfy the request.
- **State:** open.

## D-07 Public show response does not match the type contract

- **Code:** Go `ShowSummary` and `SeatMap` use `key`/`tier_key`. They omit
  `payment_mode`, `selection_mode`, `stage_position`, and `numbering`.
- **Spec:** `src/lib/ticketing/types.ts` requires `tier_id` and those four
  fields.
- **State:** open.

## D-08 Public order and ticket shapes lack contract fields

- **Code:** Go `PublicOrder` has no `payment_method` or `bank_transfer`. Go
  `Ticket` uses `tier_key`.
- **Spec:** `src/lib/ticketing/types.ts` requires `payment_method`,
  `bank_transfer`, and `tier_id`.
- **State:** open. The mock returns bank details; the real API does not yet.

## D-09 PayHere is configuration only

- **Code:** `internal/config` validates `PAYHERE_*` and the live-mode guard. There
  is no `internal/payhere` package, checkout hash, notify route, or re-check
  endpoint.
- **Spec:** `TICKETING.md` §8 and §14 specify sandbox checkout and notify.
- **State:** open.

## D-10 No `/ticket-api` proxy route

- **Code:** `src/lib/ticketing/http.ts` defaults its base to `/ticket-api`, but
  `src/app` has no catch-all route that forwards to Go.
- **Spec:** `TICKETING.md` §12 requires a runtime catch-all proxy or an equivalent
  reverse-proxy rule.
- **State:** open.

## D-11 No `/backstage-tools` staff pages

- **Code:** there are no staff pages under `src/app`. Internal Go endpoints exist,
  but no logged-in Next.js UI uses them.
- **Spec:** `TICKETING.md` §9.7 and decision 0026 require pages for manual orders,
  mark paid, cancel, re-check payment, block seats, and copy ticket links.
- **State:** open.

## D-12 Event sync, delete guard, and admin warnings are missing

- **Code:** the Events collection has a slug hook only. There is no `afterChange`
  sync to Go, no delete guard, no re-sync action, and no `seating_locked` warning.
- **Spec:** `TICKETING.md` §9.3 and decision 0026 require them.
- **State:** open.

## D-13 Privacy Policy page is missing

- **Code:** no `/privacy` route exists.
- **Spec:** `TICKETING.md` §9.8 requires a Privacy Policy page before public
  booking because names, phones, and emails are collected.
- **State:** open; launch blocker.

## D-14 `MAX_PLAN_SEATS` vs `MAX_SHOW_SEATS`

- **Code:** `internal/config` and `ticketing/.env.example` use `MAX_PLAN_SEATS`.
- **Spec:** `TICKETING.md` §2 and §4 use `MAX_SHOW_SEATS`.
- **State:** open; part of the seat-plan removal.

## D-15 Framer Motion is still listed but unused

- **Code:** `framer-motion` was removed from `package.json`; no source imports it.
- **Spec:** `DESIGN.md` §2 lists Framer Motion in the tech stack.
- **State:** open documentation mismatch. The docs/code pages describe the
  current no-Framer-Motion code.

## D-16 HTTP mode is documented as failing fast

- **Code:** `src/lib/ticketing/client.ts` now supports `http` and dynamically loads
  `src/lib/ticketing/http.ts`.
- **Spec/comment:** `.env.example` says HTTP is not wired and fails fast;
  `TICKETING.md`'s demo note says `http` intentionally throws until the client and
  proxy are implemented.
- **State:** open. The missing part is the proxy route (D-10), not the client.

## D-17 Torn edge uses background layers, not mask-image

- **Code:** `dear-daughter.css` `.dd-torn` uses a texture plus radial-gradient
  background. `mask-image` is used only on `.site-grid`.
- **Spec:** the D3 documentation brief described the torn edge as `mask-image`.
- **State:** resolved in `docs/code/css-patterns.md`; no code change.

## D-18 AGENTS.md current-state block is stale

- **Code/repo:** the repository has a full Next.js/Payload/Go codebase.
- **Doc:** `AGENTS.md` still says the repo is a placeholder containing only
  `README.md`.
- **State:** open. D5 only adds the documentation rule to `AGENTS.md`; the
  current-state block was not rewritten.

## D-19 Ticket scan endpoint is implemented but not part of the site

- **Code:** `POST /v1/scan/admit` exists and is tested.
- **Spec:** `TICKETING.md` calls it optional step 7 and says the scanner UI is a
  separate mobile app later.
- **State:** consistent; noted to avoid treating the absence of site UI as a bug.

## D-20 The site and Go are not wired together yet

- **Code:** mock is the default. The HTTP client exists, the Go service exists,
  but there is no proxy and no event sync.
- **Spec:** the target is a browser booking flow backed by Go.
- **State:** open; this is the main integration gap behind D-02–D-12.

## D-21 Ticketing README says the booking UI is still to come

- **Code:** the public booking UI, confirmation page, and ticket page exist under
  `src/app/(frontend)`.
- **Doc:** `ticketing/README.md` says “PayHere callbacks and the booking UI are
  still to come.”
- **State:** open documentation mismatch. The README predates the frontend-first
  revamp.

## Related

- [Documentation index](README.md)
- [Decision records](decisions/README.md)
- [API contract](architecture/api-contract.md)
- [Environment variables](reference/env-vars.md)
