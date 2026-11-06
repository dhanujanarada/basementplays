# 0012 — Repo layout: Next.js at the root, ticketing beside it

- **Status:** accepted
- **Covers:** `src/`, `ticketing/`, `package.json`, `tsconfig.json`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0009](0009-separate-go-service.md), [0011](0011-hosting-portability.md), `ticketing/`

## Context

The Next.js app is the main product and already at the repository root. Moving it would create churn while features are still in flight.

## Decision

Keep the Next.js app at the root and put the Go service in `ticketing/`. A `deploy/` directory can be added later.

## Alternatives considered

- **Move everything into `site/`** — churn while features are in flight.

## Consequences and trade-offs accepted

Short paths for the main app, a clear service boundary, and no big move. Root TypeScript and ESLint ignore `ticketing/`.

## Revisit when

Revisit only if more services make the root layout ambiguous.

## Related

[0009](0009-separate-go-service.md), [0011](0011-hosting-portability.md), `ticketing/`
