# 0021 — Seating generated from event tiers

- **Status:** accepted
- **Covers:** `src/collections/Productions.ts`, `src/lib/event-view.ts`, `src/components/ticketing/BookingFlow.tsx`, `ticketing/internal/seatplan/`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0022](0022-shows-in-event.md), [0023](0023-payment-modes.md), `src/collections/Productions.ts`

## Context

A separate seat-plan builder added a second content model and a plan key to keep in sync. The event already knows its tiers, prices, and counts.

## Decision

Generate seating from event tiers: name, price, and count. Remove plan templates, the builder, and plan keys. Default to quantity mode (server assigns the lowest available seats); allow seats mode with a numbered grid. Support stage positions front, back, and none. Lock structural changes once sales exist.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

One source of truth for prices and inventory. The site and frontend mock implement tiers, but the Go service still contains the earlier seatplan design, so this decision is only partly implemented.

## Revisit when

Revisit when the Go show upsert and public API are moved to tiers, or if quantity mode proves insufficient for a venue.

## Related

[0022](0022-shows-in-event.md), [0023](0023-payment-modes.md), `src/collections/Productions.ts`
