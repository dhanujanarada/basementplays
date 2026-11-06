# 0014 — Money and time representation

- **Status:** accepted
- **Covers:** `src/lib/format.ts`, `src/lib/ticketing/format.ts`, `ticketing/internal/domain/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0016](0016-hold-policy.md), [0020](0020-payhere.md), `src/lib/ticketing/format.ts`

## Context

Floating-point money and local-time storage cause avoidable errors at payment boundaries and across time zones.

## Decision

Store money as integer cents in Go and the API, and whole rupees in Payload. Store times in UTC and display in Asia/Colombo. Use `2500.00` only at the PayHere boundary.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Prices and totals are exact. The Payload editor sees friendly rupees while the API and database remain integer cents.

## Revisit when

Revisit only if a second currency is added, which would require an explicit currency field in more places.

## Related

[0016](0016-hold-policy.md), [0020](0020-payhere.md), `src/lib/ticketing/format.ts`
