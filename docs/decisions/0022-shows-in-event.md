# 0022 — Runs merged into the event; Props removed

- **Status:** accepted
- **Covers:** `src/collections/Productions.ts`, `src/lib/event-view.ts`, `src/designs/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0002](0002-content-only-cms.md), [0021](0021-generated-seating.md), `src/lib/event-view.ts`

## Context

The separate Runs collection and the ticketing-enabled flag duplicated data already implied by the event, its show dates, and its tiers. Props were not part of the current product.

## Decision

Put a shows array inside the event. Remove Runs, Props, `ticketingEnabled`, `seatPlanKey`, `ticketUrl`, price text, seats notes, and manual status. Derive ticketing active, badges, and the WhatsApp fallback.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Fewer collections and fewer flags to keep consistent. Event data is edited in one place. Show row ids remain the ids used in `?show=` and the ticketing service.

## Revisit when

Revisit only if shows need to be reused across events, which the current model does not support.

## Related

[0002](0002-content-only-cms.md), [0021](0021-generated-seating.md), `src/lib/event-view.ts`
