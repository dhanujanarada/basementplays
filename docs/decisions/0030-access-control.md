# 0030 — Access control

- **Status:** accepted
- **Covers:** `src/collections/access.ts`, `src/collections/Users.ts`, `ticketing/internal/httpapi/guards.go`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0004](0004-admin-at-backstage.md), [0017](0017-order-access-token.md), `src/collections/access.ts`

## Context

Content editors need access without user-management power. The public needs read-only content, and ticketing data must not be exposed through the CMS.

## Decision

Use admin and member roles. Only admins manage users. Content collections are publicly readable. Ticketing data is reachable only through the public API and authenticated staff pages. No public sign-up.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

A small permission model that covers the current team. Staff pages are still planned, so the ticketing side is not fully wired.

## Revisit when

Revisit when more staff roles or a public account system are introduced.

## Related

[0004](0004-admin-at-backstage.md), [0017](0017-order-access-token.md), `src/collections/access.ts`
