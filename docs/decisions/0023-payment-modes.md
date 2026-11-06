# 0023 — Payment modes per event

- **Status:** accepted
- **Covers:** `src/collections/Productions.ts`, `src/globals/SiteSettings.ts`, `src/components/ticketing/BookingFlow.tsx`, `src/components/ticketing/OrderConfirmation.tsx`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0020](0020-payhere.md), [0016](0016-hold-policy.md), `src/components/ticketing/BookingFlow.tsx`

## Context

Different events may take bank transfers, online payments, or no online booking at all. The buyer should not choose the method; the event controls it.

## Decision

Add per-event `paymentMode`: `bank_transfer` (default), `online` (PayHere), or `off` (WhatsApp fallback). Put bank details in Site Settings. Bank-transfer orders hold for 24 hours; staff mark them paid; the ticket link appears after payment. No payment selector in the buyer form.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

One booking flow with event-controlled payment. The frontend and mock implement all three modes; the Go API and staff mark-paid pages still need the matching sync and UI.

## Revisit when

Revisit if a single event needs mixed payment methods or deposits.

## Related

[0020](0020-payhere.md), [0016](0016-hold-policy.md), `src/components/ticketing/BookingFlow.tsx`
