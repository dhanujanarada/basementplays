# State machines

- **Status:** implemented
- **Covers:** `ticketing/internal/domain/status.go`, `ticketing/migrations/00002_schema.sql`, `ticketing/internal/store/`
- **Last checked:** 8de3320

The Go service uses explicit, table-tested state machines. Database CHECK
constraints are the second line of defence, so an illegal state survives neither
the domain layer nor the schema.

## Seat

| From | Allowed next | Meaning |
| --- | --- | --- |
| `available` | `held`, `blocked` | buyer hold or staff block |
| `held` | `sold`, `available` | payment, release, expiry, or cancellation |
| `blocked` | `available` | staff unblock |
| `sold` | `available` | staff cancels a paid order |

A held seat must carry an order id and a hold expiry. A sold seat must carry an
order id and no hold expiry. Available and blocked seats carry neither.

## Order

| From | Allowed next | Meaning |
| --- | --- | --- |
| `pending` | `paid`, `expired`, `cancelled`, `needs_review` | normal checkout and expiry |
| `expired` | `paid`, `needs_review` | late payment may be re-claimed |
| `paid` | `cancelled` | staff cancel; tickets are voided |
| `needs_review` | `paid`, `cancelled` | manual decision |
| `cancelled` | — | terminal |

`needs_review` means payment arrived but the seats are no longer available. A
staff member decides whether to refund or find another seat. Upgrading to `paid`
is only possible when the seats can be claimed.

## Ticket

| From | Allowed next | Meaning |
| --- | --- | --- |
| `issued` | `used`, `void` | scan at the door, or staff void |
| `used` | — | terminal |
| `void` | — | terminal |

Admission is one atomic update from `issued` to `used`. The scan API returns a
conflict when a ticket is already used or void.

## Database constraints that matter

- `orders.status`, `orders.source`, and `orders.payment_method` have CHECK
  constraints matching the domain values.
- `show_seats` has a single shape check covering available, held, sold, and
  blocked.
- `tickets` has a single-live-ticket-per-seat partial unique index.
- `order_items` and `tickets` use composite foreign keys so the order, seat, and
  show must agree.

## Transition code

The transition functions are small and pure in
`ticketing/internal/domain/status.go`. Tests cover allowed and illegal moves.

```go
func TransitionOrder(from, to OrderStatus) error {
    if from.CanTransitionTo(to) {
        return nil
    }
    return &TransitionError{Entity: "order", From: string(from), To: string(to)}
}
```

## Related

- [Data model](data-model.md)
- [Payments](payments.md)
- [API contract](api-contract.md)
