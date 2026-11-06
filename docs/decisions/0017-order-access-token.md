# 0017 — Order access via X-Order-Token

- **Status:** accepted
- **Covers:** `ticketing/internal/httpapi/guards.go`, `src/lib/ticketing/order-session.ts`, `src/components/ticketing/OrderConfirmation.tsx`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0004](0004-admin-at-backstage.md), [0016](0016-hold-policy.md), `ticketing/internal/httpapi/guards.go`

## Context

Buyers need to retrieve, release, and check out an order without creating public accounts. The order identifier alone must not grant access.

## Decision

Generate an order access token, store only its hash, and require `X-Order-Token` for get, delete, and checkout. Compare in constant time. Keep the token in the browser session. On the public ticket page, show no buyer name and only the short order reference.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

No public account system and no cookie session. A buyer who loses the browser session cannot recover the order without staff help.

## Revisit when

Revisit only if buyer accounts or email magic links are introduced.

## Related

[0004](0004-admin-at-backstage.md), [0016](0016-hold-policy.md), `ticketing/internal/httpapi/guards.go`
