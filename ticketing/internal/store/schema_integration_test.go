//go:build integration

package store_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhanujanarada/basementplays/ticketing/internal/db"
)

// schemaTestPool migrates a disposable database and returns a pool. Every test
// runs inside a transaction that is rolled back, so the database stays clean.
func schemaTestPool(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping Postgres integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := db.Migrate(ctx, dsn, logger); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse database URL: %v", err)
	}
	// The hold concurrency tests need real parallelism, not a 4-connection queue.
	poolConfig.MaxConns = 50

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool, ctx
}

func begin(t *testing.T, ctx context.Context, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return tx
}

// fixture creates one show and one pending order for that show.
func fixture(t *testing.T, ctx context.Context, tx pgx.Tx) (showID, orderID string) {
	t.Helper()

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	showID = "show-" + suffix

	if _, err := tx.Exec(ctx, `
		INSERT INTO ticketing.shows (id, production_slug, title, starts_at, venue)
		VALUES ($1, 'schema-test', 'Schema test', now() + interval '7 days', 'Test venue')
	`, showID); err != nil {
		t.Fatalf("insert show: %v", err)
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO ticketing.orders (show_id, source, access_token_hash, expires_at)
		VALUES ($1, 'online', 'test-token-hash', now() + interval '10 minutes')
		RETURNING id::text
	`, showID).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}

	return showID, orderID
}

// insertSeat inserts a seat in an arbitrary (possibly invalid) shape so the
// database CHECK constraints can be exercised.
func insertSeat(
	ctx context.Context,
	tx pgx.Tx,
	showID, seatCode, status string,
	orderID any,
	heldUntil any,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO ticketing.show_seats (
			show_id, seat_code, tier_key, tier_name, label,
			price_cents, sort_index, status, order_id, held_until
		)
		VALUES ($1, $2, 'standard', 'Standard', NULL, 250000, 0, $3, $4::uuid, $5)
	`, showID, seatCode, status, orderID, heldUntil)

	return err
}

func expectSQLState(t *testing.T, err error, code, constraint string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected SQLSTATE %s, got nil", code)
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected *pgconn.PgError, got %T: %v", err, err)
	}
	if pgErr.Code != code {
		t.Fatalf("SQLSTATE = %s (%s), want %s", pgErr.Code, pgErr.Message, code)
	}
	if constraint != "" && pgErr.ConstraintName != constraint {
		t.Fatalf("constraint = %q, want %q", pgErr.ConstraintName, constraint)
	}
}

func TestShowSeatStatusShape(t *testing.T) {
	pool, ctx := schemaTestPool(t)

	heldUntil := time.Now().Add(10 * time.Minute)

	tests := []struct {
		name      string
		status    string
		withOrder bool
		withHeld  bool
		wantErr   bool
	}{
		{name: "held with order and expiry", status: "held", withOrder: true, withHeld: true},
		{name: "held without order", status: "held", withHeld: true, wantErr: true},
		{name: "held without expiry", status: "held", withOrder: true, wantErr: true},
		{name: "sold with order", status: "sold", withOrder: true},
		{name: "sold without order", status: "sold", wantErr: true},
		{name: "sold with expiry", status: "sold", withOrder: true, withHeld: true, wantErr: true},
		{name: "available bare", status: "available"},
		{name: "available with order", status: "available", withOrder: true, wantErr: true},
		{name: "available with expiry", status: "available", withHeld: true, wantErr: true},
		{name: "blocked bare", status: "blocked"},
		{name: "blocked with order", status: "blocked", withOrder: true, withHeld: true, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tx := begin(t, ctx, pool)
			showID, orderID := fixture(t, ctx, tx)

			var orderArg any
			if tt.withOrder {
				orderArg = orderID
			}
			var heldArg any
			if tt.withHeld {
				heldArg = heldUntil
			}

			err := insertSeat(ctx, tx, showID, "A1", tt.status, orderArg, heldArg)
			if tt.wantErr {
				expectSQLState(t, err, "23514", "show_seats_status_shape")
				return
			}
			if err != nil {
				t.Fatalf("insert valid %s seat: %v", tt.status, err)
			}
		})
	}
}

func TestShowSeatOrderMustBelongToSameShow(t *testing.T) {
	pool, ctx := schemaTestPool(t)

	tx := begin(t, ctx, pool)
	_, orderID := fixture(t, ctx, tx)

	var otherShowID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO ticketing.shows (id, production_slug, title, starts_at, venue)
		VALUES ('other-show-' || $1, 'schema-test', 'Other show', now() + interval '8 days', 'Other venue')
		RETURNING id
	`, strconv.FormatInt(time.Now().UnixNano(), 36)).Scan(&otherShowID); err != nil {
		t.Fatalf("insert other show: %v", err)
	}

	err := insertSeat(ctx, tx, otherShowID, "A1", "held", orderID, time.Now().Add(10*time.Minute))
	expectSQLState(t, err, "23503", "show_seats_order_fkey")
}

func TestOrderItemsMustMatchOrderAndSeatShow(t *testing.T) {
	pool, ctx := schemaTestPool(t)

	tx := begin(t, ctx, pool)
	_, orderID := fixture(t, ctx, tx)

	var otherShowID, otherOrderID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO ticketing.shows (id, production_slug, title, starts_at, venue)
		VALUES ('other-show-' || $1, 'schema-test', 'Other show', now() + interval '8 days', 'Other venue')
		RETURNING id
	`, strconv.FormatInt(time.Now().UnixNano(), 36)).Scan(&otherShowID); err != nil {
		t.Fatalf("insert other show: %v", err)
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO ticketing.orders (show_id, source, access_token_hash, expires_at)
		VALUES ($1, 'online', 'other-token-hash', now() + interval '10 minutes')
		RETURNING id::text
	`, otherShowID).Scan(&otherOrderID); err != nil {
		t.Fatalf("insert other order: %v", err)
	}
	if err := insertSeat(ctx, tx, otherShowID, "A1", "available", nil, nil); err != nil {
		t.Fatalf("insert other seat: %v", err)
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO ticketing.order_items (order_id, show_id, seat_code, price_cents)
		VALUES ($1::uuid, $2, 'A1', 250000)
	`, orderID, otherShowID)
	expectSQLState(t, err, "23503", "order_items_order_id_show_id_fkey")
}

func TestOneLiveTicketPerSeat(t *testing.T) {
	pool, ctx := schemaTestPool(t)

	insertSoldSeat := func(t *testing.T, ctx context.Context, tx pgx.Tx, showID, orderID string) {
		t.Helper()
		if err := insertSeat(ctx, tx, showID, "A1", "sold", orderID, nil); err != nil {
			t.Fatalf("insert sold seat: %v", err)
		}
	}

	t.Run("second issued ticket is rejected", func(t *testing.T) {
		tx := begin(t, ctx, pool)
		showID, orderID := fixture(t, ctx, tx)
		insertSoldSeat(t, ctx, tx, showID, orderID)

		if _, err := tx.Exec(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code)
			VALUES ($1::uuid, $2, 'A1')
		`, orderID, showID); err != nil {
			t.Fatalf("insert first ticket: %v", err)
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code)
			VALUES ($1::uuid, $2, 'A1')
		`, orderID, showID)
		expectSQLState(t, err, "23505", "one_live_ticket_per_seat")
	})

	t.Run("used ticket blocks a new issue", func(t *testing.T) {
		tx := begin(t, ctx, pool)
		showID, orderID := fixture(t, ctx, tx)
		insertSoldSeat(t, ctx, tx, showID, orderID)

		var ticketID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code)
			VALUES ($1::uuid, $2, 'A1')
			RETURNING id::text
		`, orderID, showID).Scan(&ticketID); err != nil {
			t.Fatalf("insert ticket: %v", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.tickets SET status = 'used', used_at = now() WHERE id = $1::uuid
		`, ticketID); err != nil {
			t.Fatalf("admit ticket: %v", err)
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code)
			VALUES ($1::uuid, $2, 'A1')
		`, orderID, showID)
		expectSQLState(t, err, "23505", "one_live_ticket_per_seat")
	})

	t.Run("a void ticket does not block a reissue", func(t *testing.T) {
		tx := begin(t, ctx, pool)
		showID, orderID := fixture(t, ctx, tx)
		insertSoldSeat(t, ctx, tx, showID, orderID)

		if _, err := tx.Exec(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code, status)
			VALUES ($1::uuid, $2, 'A1', 'void')
		`, orderID, showID); err != nil {
			t.Fatalf("insert void ticket: %v", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code)
			VALUES ($1::uuid, $2, 'A1')
		`, orderID, showID); err != nil {
			t.Fatalf("insert issued ticket after void: %v", err)
		}
	})
}

func TestOrderIdempotencyKeyIsUniquePerShow(t *testing.T) {
	pool, ctx := schemaTestPool(t)

	t.Run("duplicate key on one show is rejected", func(t *testing.T) {
		tx := begin(t, ctx, pool)
		showID, _ := fixture(t, ctx, tx)

		for i := 0; i < 2; i++ {
			_, err := tx.Exec(ctx, `
				INSERT INTO ticketing.orders (
					show_id, source, access_token_hash, expires_at, idempotency_key
				)
				VALUES ($1, 'online', 'test-token-hash', now() + interval '10 minutes', 'same-key')
			`, showID)
			if i == 0 && err != nil {
				t.Fatalf("insert first order: %v", err)
			}
			if i == 1 {
				expectSQLState(t, err, "23505", "orders_show_id_idempotency_key_key")
			}
		}
	})

	t.Run("duplicate NULL keys are allowed", func(t *testing.T) {
		tx := begin(t, ctx, pool)
		showID, _ := fixture(t, ctx, tx)

		for i := 0; i < 2; i++ {
			if _, err := tx.Exec(ctx, `
				INSERT INTO ticketing.orders (show_id, source, access_token_hash, expires_at)
				VALUES ($1, 'manual', 'test-token-hash', now() + interval '24 hours')
			`, showID); err != nil {
				t.Fatalf("insert manual order %d: %v", i, err)
			}
		}
	})

	t.Run("the same key on another show is allowed", func(t *testing.T) {
		tx := begin(t, ctx, pool)
		showID, _ := fixture(t, ctx, tx)

		var otherShowID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO ticketing.shows (id, production_slug, title, starts_at, venue)
			VALUES ('other-show-' || $1, 'schema-test', 'Other show', now() + interval '8 days', 'Other venue')
			RETURNING id
		`, strconv.FormatInt(time.Now().UnixNano(), 36)).Scan(&otherShowID); err != nil {
			t.Fatalf("insert other show: %v", err)
		}

		for _, id := range []string{showID, otherShowID} {
			if _, err := tx.Exec(ctx, `
				INSERT INTO ticketing.orders (
					show_id, source, access_token_hash, expires_at, idempotency_key
				)
				VALUES ($1, 'online', 'test-token-hash', now() + interval '10 minutes', 'shared-key')
			`, id); err != nil {
				t.Fatalf("insert order for %s: %v", id, err)
			}
		}
	})
}
