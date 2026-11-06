# 0003 — Design contract { Page, Shell, Booking? } and CSS-variable theming

- **Status:** accepted
- **Covers:** `src/lib/types.ts`, `src/designs/`, `src/components/ticketing/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0002](0002-content-only-cms.md), [0024](0024-event-page-booking.md), `src/designs/index.ts`

## Context

The event page, booking page, confirmation page, and ticket page must all wear the same event-specific look without duplicating layout code.

## Decision

Each design exports `Page`, `Shell`, and optionally `Booking`. `Shell` owns the palette, fonts, textures, header, footer, and CSS variables. Booking and ticket routes reuse that shell.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

A new design is a self-contained folder. Shared ticketing UI must be themed only through variables, which limits design-specific hacks but keeps the routes consistent.

## Revisit when

Revisit if the shell contract cannot express a design without duplicating the booking routes.

## Related

[0002](0002-content-only-cms.md), [0024](0024-event-page-booking.md), `src/designs/index.ts`
