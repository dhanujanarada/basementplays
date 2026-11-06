# API contract

- **Status:** partial
- **Covers:** `src/lib/ticketing/types.ts`, `src/lib/ticketing/http.ts`, `src/lib/ticketing/client.ts`, `ticketing/internal/httpapi/`, `ticketing/internal/service/`
- **Last checked:** 8de3320

The canonical browser-facing contract is the TypeScript type file
`src/lib/ticketing/types.ts`. The Go service is intended to match it. Today the
frontend revamp is ahead of the Go service, so this page documents both the
typed contract and the currently implemented Go surface.

## Typed browser contract

The HTTP client calls `/v1/public/...` under `NEXT_PUBLIC_TICKET_API_BASE`
(default `/ticket-api`). All public fields are snake_case.

| Type | Shape |
| --- | --- |
| `ShowSummary` | show id, production slug, title, time, venue, status, `payment_mode`, `selection_mode`, `stage_position`, `numbering`, columns, totals, `price_from`, tiers |
| `TierSummary` | `tier_id`, name, `price_cents`, `seats_total`, `seats_remaining` |
| `SeatMap` | show summary fields plus every seat, tier totals, stage label |
| `Seat` | code, label, `tier_id`, tier name, price, status, sort index |
| `Hold` | order id, show id, status, source, total, currency, expiry, seats, optional `order_token` |
| `PublicOrder` | order id, show id, status, source, payment method, total, currency, expiry, paid time, created time, `bank_transfer`, seats, tickets |
| `Ticket` | ticket id, status, production slug, show, seat/tier, description, price, short reference, QR token/URL |

`HoldRequest` is a union:

- quantity mode: `{"tiers":[{"tier_id":"...","quantity":2}]}`;
- seats mode: `{"seats":["S001","S002"]}`.

`CheckoutInput` is `buyer_name`, `buyer_phone`, `buyer_email`, and
`payment_method` (`payhere` or `bank_transfer`). The browser does not let the
buyer choose the method; the event's `paymentMode` supplies it.

## Implemented Go public routes

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/v1/public/shows/{id}` | Availability summary; ETag is not set on this route |
| GET | `/v1/public/shows/{id}/seats` | Seat map with ETag and `If-None-Match` support |
| POST | `/v1/public/shows/{id}/holds` | Current body is `{"seats":["S001"]}` only |
| DELETE | `/v1/public/orders/{id}` | Requires `X-Order-Token`; returns 204 |
| POST | `/v1/public/orders/{id}/checkout` | Buyer details; extends the hold |
| GET | `/v1/public/orders/{id}` | Requires `X-Order-Token` |
| GET | `/v1/public/tickets/{token}` | Verifies the signed QR token |
| POST | `/v1/scan/admit` | Scanner endpoint; requires `X-Scanner-Token` |

The current Go summary and seat map use `tier_key` and do not include
`payment_mode`, `selection_mode`, `stage_position`, or `numbering`. The current
hold body is seats-only. These are implementation gaps, not the typed target.

## Implemented Go internal routes

Internal routes require `Authorization: Bearer <INTERNAL_TOKEN>` and are mounted
under `/v1/internal`.

| Method | Path |
| --- | --- |
| POST, GET | `/seat-plans`, `/seat-plans/{key}` |
| PATCH | `/seat-plans/{key}/labels` |
| POST | `/seat-plans/{key}/duplicate` |
| PUT | `/shows/{id}` |
| POST, DELETE | `/shows/{id}/scanner-token` |
| POST | `/orders` (manual order) |
| POST | `/orders/{id}/mark-paid` |
| POST | `/orders/{id}/cancel` |
| GET | `/shows/{id}/orders` |

The target revamp removes the seat-plan endpoints and changes show upsert to
tier configuration. Those routes are still present in the current code.

## Headers and idempotency

- `X-Order-Token` is required for release, checkout, and order lookup.
- `Idempotency-Key` is accepted on hold and checkout. It is stored per show.
- `X-Request-Id` is echoed, or generated when missing.
- `If-None-Match` is used on the seat map.
- `X-Scanner-Token` is used only by the scan API.

## Error envelope

Errors are JSON:

```json
{"error":"seat_taken","message":"one or more selected seats are no longer available"}
```

Current codes include:

`invalid_request`, `validation_error`, `not_found`, `unauthorized`,
`missing_order_token`, `invalid_order_token`, `order_not_pending`, `hold_expired`, `order_not_checkoutable`,
`order_needs_review`, `sales_closed`, `seat_taken`, `unknown_seat`,
`idempotency_mismatch`, `rate_limited`, `origin_not_allowed`, `internal_error`.

Internal and scanner routes add plan/order/staff codes such as
`plan_key_exists`, `plan_in_use`, `show_in_use`, `order_not_payable`,
`order_not_cancellable`, `tickets_in_use`, `invalid_scanner_token`,
`wrong_show`, `ticket_already_used`, and `ticket_void`.

The frontend maps HTTP errors into `TicketingError` with `code` and `status`.
The typed contract's quantity hold error is `insufficient_inventory`; the
current Go service does not implement quantity holds yet.

## Related

- [Data model](data-model.md)
- [State machines](state-machines.md)
- [Security](security.md)
- Temporary discrepancy list: `docs/DISCREPANCIES.md` (D5)
