# 0026 — Staff tooling under /backstage-tools

- **Status:** accepted
- **Covers:** `src/app/(frontend)/`, `ticketing/internal/httpapi/internal.go`
- **Last checked:** 8de3320
- **Implementation:** planned
- **Date:** 2026-11
- **Related:** [0004](0004-admin-at-backstage.md), [0023](0023-payment-modes.md), `ticketing/internal/httpapi/internal.go`

## Context

Staff need order lists, manual orders, mark-paid, cancel, re-check payment, and seat blocking. A separate Go admin UI would duplicate authentication and layout.

## Decision

Build standalone Next.js pages under `/backstage-tools/...`, protected by the Payload session, calling Go server-side with the internal token. There is no separate Go admin. Deleting an event or show with orders is blocked and fails closed if Go is unreachable. Unreachable Go on save shows a warning; seating lock shows a warning.

## Alternatives considered

- **A Payload custom admin view** — more import-map and component-path setup for the same result.

## Consequences and trade-offs accepted

One login and one UI stack. The staff pages and event-sync hook are not implemented yet.

## Revisit when

Revisit when the staff pages are built, or if the internal API needs a richer permission model.

## Related

[0004](0004-admin-at-backstage.md), [0023](0023-payment-modes.md), `ticketing/internal/httpapi/internal.go`
