# 0019 — HMAC-signed QR tickets

- **Status:** accepted
- **Covers:** `ticketing/internal/qr/qr.go`, `ticketing/internal/service/tickets.go`, `src/app/(frontend)/t/[token]/`, `src/components/ticketing/TicketView.tsx`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0018](0018-state-machines.md), [0030](0030-access-control.md), `ticketing/internal/qr/qr.go`

## Context

Door staff need a fast, offline-tolerant check without exposing buyer accounts. The website is not the scanner.

## Decision

Issue a v1 HMAC-signed token in `/t/<token>`. Verify with a constant-time compare. Admit with one atomic update that only accepts an issued ticket for that show. Build the scanner as a separate mobile app later, not inside the website.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

The ticket URL is verifiable and does not expose a buyer name. Revocation and scanning behavior depend on the Go service and the future scanner app.

## Revisit when

Revisit if the scanner app needs a different token format or offline verification.

## Related

[0018](0018-state-machines.md), [0030](0030-access-control.md), `ticketing/internal/qr/qr.go`
