# 0006 — Testimonials deferred

- **Status:** deferred
- **Covers:** `src/collections/Productions.ts`, `src/designs/`
- **Last checked:** 8de3320
- **Implementation:** planned
- **Date:** 2026-11
- **Related:** [0029](0029-deferred-scope.md), `src/collections/Productions.ts`

## Context

Testimonials could add social proof, but no current page needs them and they would add another content shape to every event.

## Decision

Do not build testimonials now. If they return, model them as an event-level list rather than a separate collection.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

One less content type and one less public section. The design has no testimonial slot.

## Revisit when

Revisit when a production wants quotes on its event page and the content can be curated per event.

## Related

[0029](0029-deferred-scope.md), `src/collections/Productions.ts`
