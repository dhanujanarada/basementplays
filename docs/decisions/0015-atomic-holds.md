# 0015 — Atomic seat holds

- **Status:** accepted
- **Covers:** `ticketing/internal/store/holds.go`, `ticketing/internal/service/holds.go`, `ticketing/internal/store/holds_integration_test.go`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0016](0016-hold-policy.md), [0018](0018-state-machines.md), `ticketing/internal/store/holds.go`

## Context

Two buyers must never hold the same seat. Expiry and the sweeper must not be required for correctness.

## Decision

Lock rows in sorted order and claim them in one transaction. Retry on deadlock or serialization failure. Treat expired holds as available in the claim query. Use a Postgres advisory lock for the sweeper. Require race tests, including many goroutines competing for one seat.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Stronger correctness under concurrency and a clear recovery path. The implementation is more complex than a single update and needs integration tests.

## Revisit when

Revisit if the seat model changes enough that row-level locking no longer maps to the inventory unit.

## Related

[0016](0016-hold-policy.md), [0018](0018-state-machines.md), `ticketing/internal/store/holds.go`
