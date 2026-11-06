# 0024 — Event page booking through SeasonBooking

- **Status:** accepted
- **Covers:** `src/components/ticketing/SeasonBooking.tsx`, `src/designs/`, `src/lib/ticketing/booking.ts`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0002](0002-content-only-cms.md), [0003](0003-design-contract.md), `src/components/ticketing/SeasonBooking.tsx`

## Context

Each design had its own show list and reserve buttons. Multiple reserve actions made it unclear which show a buyer would book and broke URL sharing.

## Decision

Use one `SeasonBooking` client component for event show state. Default to the next show, use the most recent show with a season-ended state when none are upcoming, keep `?show=` in the URL, and make the show list a selector. The booking card is the only reserve action. Hero and mobile-bar links scroll to the card.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

One state model and one booking action across designs. The card markup stays design-owned through a render prop.

## Revisit when

Revisit if an event needs multiple simultaneous booking cards or a non-show-based ticket product.

## Related

[0002](0002-content-only-cms.md), [0003](0003-design-contract.md), `src/components/ticketing/SeasonBooking.tsx`
