# 0016 — Hold policy: 10 minutes, checkout extension, 24 hours for bank transfer

- **Status:** accepted
- **Covers:** `ticketing/internal/config/config.go`, `ticketing/internal/service/holds.go`, `src/lib/ticketing/types.ts`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0015](0015-atomic-holds.md), [0017](0017-order-access-token.md), `ticketing/internal/config/config.go`

## Context

Holds need to be long enough to enter buyer details but short enough not to strand inventory. Manual bank transfers need much longer.

## Decision

Hold for 10 minutes with no personal data. At checkout, capture buyer details and extend the hold to 15 minutes from that moment. Bank transfer holds for 24 hours. Allow at most 6 tickets per order, rate-limit per IP, and accept `Idempotency-Key` on holds and checkout.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Inventory is released predictably, and no open-ended holds exist. Buyers can still lose a hold if they take too long, which is surfaced in the UI.

## Revisit when

Revisit if real checkout times show 15 minutes is too short or if manual confirmation becomes much slower.

## Related

[0015](0015-atomic-holds.md), [0017](0017-order-access-token.md), `ticketing/internal/config/config.go`
