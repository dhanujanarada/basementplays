# 0010 — Build the ticketing system ourselves

- **Status:** accepted
- **Covers:** `ticketing/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0009](0009-separate-go-service.md), [0029](0029-deferred-scope.md), `ticketing/README.md`

## Context

The group needs seat holds, PayHere, QR tickets, and a WhatsApp-first manual flow. The public alternatives evaluated did not fit the per-event design and payment needs.

## Decision

Build a focused ticketing service instead of self-hosting a general platform. Fall back to pretix only if scope grows into refunds, waitlists, and vouchers.

## Alternatives considered

- **pretix** — AGPL, seating plans, many payment plugins, but heavy to host with Postgres plus Redis plus workers, no PayHere plugin found, and seating-plan licensing reports conflicted (unverified).

## Consequences and trade-offs accepted

Full control over wording, manual orders, QR tickets, and integration. The team owns security, payments, and all future ticketing features.

## Revisit when

Revisit if refund automation, vouchers, waitlists, or multiple payment providers turn into a large product surface.

## Related

[0009](0009-separate-go-service.md), [0029](0029-deferred-scope.md), `ticketing/README.md`
