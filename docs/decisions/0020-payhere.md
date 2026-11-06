# 0020 — PayHere gateway, sandbox only for now

- **Status:** accepted
- **Covers:** `ticketing/internal/config/config.go`, `ticketing/internal/httpapi/`
- **Last checked:** 8de3320
- **Implementation:** planned
- **Date:** 2026-11
- **Related:** [0014](0014-money-and-time.md), [0018](0018-state-machines.md), `ticketing/internal/config/config.go`

## Context

The group is a registered entity and PayHere eligibility was confirmed. Online card payment is required later, but live money is out of scope for the current demo.

## Decision

Use PayHere in sandbox mode. Compute the checkout hash only on the server. Verify the notify signature, merchant, amount, and currency against the stored order. Record notifications idempotently. Never mark paid from the return URL. Refuse live mode unless explicitly allowed.

## Alternatives considered

- **Stripe** — a secondary source says Sri Lanka is not supported as a payout country (unverified).

## Consequences and trade-offs accepted

A local payment method that fits the group. PayHere code and notify handling are not implemented yet; only configuration guards exist.

## Revisit when

Revisit before live payments, after confirming current PayHere terms, URLs, and merchant configuration.

## Related

[0014](0014-money-and-time.md), [0018](0018-state-machines.md), `ticketing/internal/config/config.go`
