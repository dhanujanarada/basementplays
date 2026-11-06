# Go primer for this repo

- **Status:** implemented
- **Covers:** `ticketing/cmd/`, `ticketing/internal/`, `ticketing/Makefile`
- **Last checked:** 8de3320

This is not a general Go tutorial. Each section shows a feature the ticketing
service actually uses and why it matters here.

## Packages and imports

Go code is grouped by package. `cmd/ticketing` imports the internal packages and
wires them together; internal packages cannot be imported from outside the
module.

```go
// ticketing/cmd/ticketing/main.go — runServer
handler := httpapi.New(httpapi.Options{
    Logger:        logger,
    Pinger:        pool,
    InternalToken: cfg.InternalToken,
    // ...
})
```

## Structs and methods

`Store` is a small struct holding the database pool. Methods such as `InTx` keep
the SQL layer in one package.

```go
// ticketing/internal/store/store.go — Store.InTx
func (s *Store) InTx(ctx context.Context, fn func(pgx.Tx) error) error {
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

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}
```

## Interfaces

Small interfaces let services be tested without Postgres. `TicketLookup` is all
the ticket service needs from storage.

```go
// ticketing/internal/service/tickets.go — TicketLookup
type TicketLookup interface {
    TicketByID(ctx context.Context, ticketID string) (store.TicketInfo, error)
}
```

## Error wrapping with %w

`%w` keeps the original error in the chain, so callers can use `errors.Is` and
`errors.As` instead of matching strings.

```go
// ticketing/internal/db/db.go — Open
return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
```

## defer

`defer` runs on every return path. It is used for rollback, closing rows and
pools, stopping tickers, and canceling contexts.

```go
// ticketing/internal/service/sweeper.go — Sweeper.sweep
ctx, cancel := context.WithTimeout(parent, sweepTimeout)
defer cancel()
```

## context.Context

Database calls, HTTP handlers, and background work take a context so deadlines
and cancellation propagate. A shutdown cancels the root context; the server and
sweeper stop.

## Goroutines, channels, and select

The HTTP server runs in a goroutine and reports its error on a channel. `select`
waits for either a server error or shutdown.

```go
// ticketing/cmd/ticketing/main.go — runServer
serverErr := make(chan error, 1)
go func() { serverErr <- server.ListenAndServe() }()

select {
case err := <-serverErr:
    // handle server error
case <-ctx.Done():
    // graceful shutdown
}
```

## sync and atomics

The rate limiter uses a `sync.Mutex` because it is shared by concurrent
requests. Tests use `sync/atomic` for counters and channels to wait for a
background loop.

```go
// ticketing/internal/httpapi/guards.go — rateLimiter.allow
l.mu.Lock()
defer l.mu.Unlock()
```

## Table-driven tests

State-machine tests enumerate every from/to pair and assert both the boolean
and the error path.

```go
// ticketing/internal/domain/status_test.go — TestSeatTransitions
t.Run(name, func(t *testing.T) {
    want := allowed[[2]SeatStatus{from, to}]
    if got := from.CanTransitionTo(to); got != want {
        t.Fatalf("%s -> %s = %v, want %v", from, to, got, want)
    }
})
```

## Race detector

`make test-race` and `make test-integration-race` run `go test -race`. They are
used for the concurrency-sensitive hold and sweeper code.

## pgx

The service uses `pgx` directly, not an ORM. `InTx` begins/commits, `QueryRow`
scans one row, `Query` iterates seats, and `CopyFrom` bulk-inserts generated
seats. SQL lives in `internal/store/`.

## Related

- [Ticketing walkthrough](ticketing-walkthrough.md)
- [SQL and Postgres](sql-and-postgres.md)
- [Commands](../reference/commands.md)
