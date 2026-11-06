# 0018 — State machines and database constraints

- **Status:** accepted
- **Covers:** `ticketing/internal/domain/status.go`, `ticketing/migrations/`, `ticketing/internal/store/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0015](0015-atomic-holds.md), [0017](0017-order-access-token.md), `ticketing/internal/domain/status.go`

## Context

Seats, orders, and tickets move through states that must not be corrupted by a crash, a late payment, or two staff actions at once.

## Decision

Implement explicit seat, order, and ticket state machines with table-driven tests. Add database CHECK constraints as the second line of defense. Use `needs_review` when payment arrives after the seats are gone. Enforce one live ticket per seat with a partial unique index.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Illegal transitions fail early and are hard to introduce. The schema and domain code must stay in sync.

## Revisit when

Revisit when a new state is added; update both the machine and the database constraints together.

## Related

[0015](0015-atomic-holds.md), [0017](0017-order-access-token.md), `ticketing/internal/domain/status.go`
