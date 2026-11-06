# Ticketing service walkthrough

- **Status:** partial
- **Covers:** `ticketing/cmd/`, `ticketing/internal/config/`, `ticketing/internal/db/`, `ticketing/internal/domain/`, `ticketing/internal/httpapi/`, `ticketing/internal/service/`, `ticketing/internal/store/`, `ticketing/internal/qr/`, `ticketing/internal/seatplan/`
- **Last checked:** 8de3320

This is a tour of the Go service as it exists now. The revamp target removes
seat plans and generates seats from event tiers; the current code still contains
the earlier `seatplan` model.

## Package map

| Package | Role |
| --- | --- |
| `cmd/ticketing` | process entrypoint: `serve` or `migrate`, HTTP server, shutdown |
| `cmd/seed` | dev seed for the sample seat plan |
| `internal/config` | environment parsing and fail-fast validation |
| `internal/db` | pgx pool, migrations, advisory-lock migration runner |
| `internal/domain` | pure state machines, validation errors, ticket wording |
| `internal/httpapi` | public, internal, scanner routes, guards, request ids, CORS |
| `internal/service` | use cases: shows, availability, holds, orders, tickets, scanner, sweeper |
| `internal/store` | every SQL statement and transaction |
| `internal/qr` | HMAC sign/verify for ticket tokens |
| `internal/seatplan` | current plan generation/validation (target: remove) |

## Request path

A public request follows a consistent path:

1. `httpapi` middleware adds a request id, CORS, rate limiting, logging, and
   panic recovery.
2. A handler decodes a bounded JSON body and calls a service method.
3. The service validates input and calls the store.
4. The store runs SQL in a transaction through `Store.InTx`.
5. The handler maps typed errors to a JSON error envelope.

```go
// ticketing/internal/store/store.go — Store.InTx
tx, err := s.pool.Begin(ctx)
if err != nil {
    return fmt.Errorf("begin transaction: %w", err)
}
defer func() {
    _ = tx.Rollback(ctx)
}()

if err := fn(tx); err != nil {
    return err
}
```

## Show upsert today

`PUT /v1/internal/shows/{id}` currently accepts `seat_plan_key`, not tiers. The
store locks the plan, validates its JSON layout, and copies seats into
`show_seats`. Relinking is blocked once orders or non-available seats exist.
This is the `show_in_use` conflict in the API.

The target is different: the event's tiers become `show_tiers`, and seats are
generated from those tiers. That change has not reached the Go service.

## Holds and orders today

Holds accept a list of seat codes. The store locks the requested seats in sorted
order, checks that each is available or has an expired hold, creates the order
and order items, then claims the seats with one update. Deadlocks and
serialization failures retry up to three times.

`MarkPaid` sells seats and issues tickets. `Cancel` releases or voids according
to state. `needs_review` protects the case where payment arrives after the seats
are gone.

## QR and ticket lookup

`internal/qr` signs `v1.<ticket-id>.<hmac>` and verifies it with a constant-time
compare. `Tickets.Lookup` verifies the token, loads the ticket, and returns a
public payload with a short order reference and no buyer details.

## Planned pieces

- Tier-generated seating and quantity holds.
- Payment mode and stage position on shows.
- PayHere checkout and notify.
- Staff pages under `/backstage-tools`.
- Event save sync and delete guards in Payload.

## Related

- [Go primer](go-primer.md)
- [SQL and Postgres](sql-and-postgres.md)
- [API contract](../architecture/api-contract.md)
