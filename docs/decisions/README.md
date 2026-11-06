# Decision records

- **Status:** implemented
- **Covers:** `docs/decisions/`
- **Last checked:** 8de3320

Decision records explain why the project is shaped the way it is. They are historical:
when a decision changes, add a new record that supersedes the old one and update this
index. Do not rewrite an accepted record to match later code.

Use [0000-template.md](0000-template.md) as the shape for a new record. Records are
numbered in order; the status column says whether a record still governs the code.

| Number | Title | Status | Implementation |
| --- | --- | --- | --- |
| [0001](0001-payload-inside-nextjs.md) | Payload CMS inside Next.js (one codebase) | accepted | done |
| [0002](0002-content-only-cms.md) | Content-only CMS; custom design component per event | accepted | done |
| [0003](0003-design-contract.md) | Design contract { Page, Shell, Booking? } and CSS-variable theming | accepted | done |
| [0004](0004-admin-at-backstage.md) | Admin hidden at /backstage | accepted | done |
| [0005](0005-backstage-name.md) | 'Backstage' names the admin panel only | accepted | done |
| [0006](0006-testimonials-deferred.md) | Testimonials deferred | deferred | planned |
| [0007](0007-site-settings-and-fallbacks.md) | Site Settings feeds public chrome and fallbacks | accepted | partial |
| [0008](0008-postgres-everywhere.md) | Postgres everywhere; SQLite dropped from the site | accepted | partial |
| [0009](0009-separate-go-service.md) | Separate Go ticketing service | accepted | done |
| [0010](0010-build-ticketing-ourselves.md) | Build the ticketing system ourselves | accepted | done |
| [0011](0011-hosting-portability.md) | Hosting kept portable | accepted | partial |
| [0012](0012-repo-layout.md) | Repo layout: Next.js at the root, ticketing beside it | accepted | done |
| [0013](0013-api-routing.md) | API routing through same-origin /ticket-api | accepted | partial |
| [0014](0014-money-and-time.md) | Money and time representation | accepted | done |
| [0015](0015-atomic-holds.md) | Atomic seat holds | accepted | done |
| [0016](0016-hold-policy.md) | Hold policy: 10 minutes, checkout extension, 24 hours for bank transfer | accepted | done |
| [0017](0017-order-access-token.md) | Order access via X-Order-Token | accepted | done |
| [0018](0018-state-machines.md) | State machines and database constraints | accepted | done |
| [0019](0019-qr-tickets.md) | HMAC-signed QR tickets | accepted | done |
| [0020](0020-payhere.md) | PayHere gateway, sandbox only for now | accepted | planned |
| [0021](0021-generated-seating.md) | Seating generated from event tiers | accepted | partial |
| [0022](0022-shows-in-event.md) | Runs merged into the event; Props removed | accepted | done |
| [0023](0023-payment-modes.md) | Payment modes per event | accepted | partial |
| [0024](0024-event-page-booking.md) | Event page booking through SeasonBooking | accepted | done |
| [0025](0025-frontend-first-mock.md) | Frontend-first development with a mock ticket API | accepted | done |
| [0026](0026-staff-tooling.md) | Staff tooling under /backstage-tools | accepted | planned |
| [0027](0027-lean-principles.md) | Lean principles and the parity guardrail | accepted | done |
| [0028](0028-development-workflow.md) | Development workflow: specs, plan mode, step gates | accepted | done |
| [0029](0029-deferred-scope.md) | Deferred or rejected scope | deferred | planned |
| [0030](0030-access-control.md) | Access control | accepted | partial |

The template itself is kept separately: [0000-template.md](0000-template.md).
