# 0005 — 'Backstage' names the admin panel only

- **Status:** accepted
- **Covers:** `src/collections/Productions.ts`, `src/designs/`, `docs/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0002](0002-content-only-cms.md), `src/collections/Productions.ts`

## Context

Theatrical language is useful in the product, but using 'Backstage' for both the admin and public gallery sections would confuse editors and developers.

## Decision

Use 'Backstage' only for the admin panel. Gallery categories are 'On stage' and 'Behind the scenes'. Do not use 'Backstage' for public photos or content types.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

The word has one meaning in UI copy, code, and documentation. Editors see 'On stage' and 'Behind the scenes' in the gallery.

## Revisit when

Revisit only if the admin panel is renamed for another reason.

## Related

[0002](0002-content-only-cms.md), `src/collections/Productions.ts`
