# SQL and Postgres in this repo

- **Status:** partial
- **Covers:** `ticketing/internal/store/`, `ticketing/migrations/`, `ticketing/internal/db/`
- **Last checked:** 8de3320

The Go service talks to Postgres through `pgx`. SQL lives in
`ticketing/internal/store/`; handlers never write SQL.

## Transactions

`Store.InTx` begins a transaction, runs the callback, rolls back on error or
panic, and commits only on success. This is the unit of correctness for holds,
checkout, payment, and cancellation.

```go
// ticketing/internal/store/store.go — Store.InTx
tx, err := s.pool.Begin(ctx)
if err != nil {
    return fmt.Errorf("begin transaction: %w", err)
}
defer func() { _ = tx.Rollback(ctx) }()

if err := fn(tx); err != nil {
    return err
}
return tx.Commit(ctx)
```

## FOR UPDATE

The hold path locks the show seats in sorted order before checking claimability.
Locking in a stable order avoids deadlocks between overlapping requests.

```sql
-- ticketing/internal/store/orders.go — Store.CreateOrder
SELECT seat_code, price_cents,
       (status = 'available' OR (status = 'held' AND held_until < now())) AS claimable
FROM ticketing.show_seats
WHERE show_id = $1 AND seat_code = ANY($2)
ORDER BY seat_code
FOR UPDATE
```

The order token check also uses `FOR UPDATE` on `orders`, so two staff actions
cannot race.

## SKIP LOCKED

`SKIP LOCKED` is not in the current Go queries. The revamp target uses it for
quantity-mode holds: pick the lowest available seats per tier without waiting on
rows another transaction is already holding. Until that code exists, this page
does not claim it is implemented.

## Partial unique indexes

One live ticket per seat is a partial unique index, not a nullable column hack:

```sql
CREATE UNIQUE INDEX one_live_ticket_per_seat
  ON ticketing.tickets (show_id, seat_code)
  WHERE status IN ('issued', 'used');
```

Voided tickets do not block a replacement ticket. A second issued ticket for the
same seat fails at the database, even if application code is wrong.

## CHECK constraints

The schema repeats critical rules as CHECKs: valid order/seat/ticket statuses,
non-negative money, a paid order must have `paid_at`, and a held seat must have
an order and expiry. These are the second line of defence after the domain
state machines.

## Advisory locks

The sweeper uses a transaction-scoped advisory lock so several API instances can
run it without duplicating work:

```sql
SELECT pg_try_advisory_xact_lock(0x5449434b45543032);
```

The lock is released automatically at commit or rollback. Correctness does not
depend on the sweeper: the claim query also treats expired holds as available.

## jsonb

`seat_plans.layout`, `payments.raw`, and `audit_log.data` are `jsonb`. The layout
is the current plan representation; the target tier model would remove
`seat_plans`. JSONB is used for flexible provider payloads and audit data, not
for money or state fields.

## pgx patterns

- `QueryRow(...).Scan(...)` for one row.
- `Query(...)` plus `rows.Next()` for seat sets.
- `pgx.CopyFrom` to bulk-insert generated seats.
- `pgconn.PgError` codes `40P01` and `40001` trigger retries.
- `pgx.ErrNoRows` maps to not-found errors.
- `tx.Rollback` is deferred even after a successful callback.

## Related

- [Data model](../architecture/data-model.md)
- [State machines](../architecture/state-machines.md)
- [Go primer](go-primer.md)
