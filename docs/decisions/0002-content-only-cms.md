# 0002 — Content-only CMS; custom design component per event

- **Status:** accepted
- **Covers:** `src/collections/Productions.ts`, `src/designs/`, `src/lib/types.ts`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0003](0003-design-contract.md), [0024](0024-event-page-booking.md), `src/designs/index.ts`

## Context

Each production has a bespoke visual identity. The CMS should hold content, not page design, because an admin cannot design a page from a few theme fields.

## Decision

Keep content in Payload. Keep presentation in a code component under `src/designs/<slug>/`, registered by slug. Events without a custom design use `PlainDesign`.

## Alternatives considered

- **Theme fields in the CMS** — every show has a bespoke look; exposing colors, fonts, and layout controls would produce weak pages and a large admin surface.

## Consequences and trade-offs accepted

Designers and developers still edit code for visual work. Content editors can change text, people, shows, tiers, and gallery items without a deploy.

## Revisit when

Revisit only if events become visually repetitive enough that a constrained theme editor would produce equally good pages.

## Related

[0003](0003-design-contract.md), [0024](0024-event-page-booking.md), `src/designs/index.ts`
