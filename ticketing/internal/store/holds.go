package store

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const sweepLockKey int64 = 0x5449434b45543032 // "TICKET02"

var (
	// ErrShowNotFound is returned for an unknown show id.
	ErrShowNotFound = errors.New("store: show not found")
	// ErrSalesClosed is returned when a show is not on sale.
	ErrSalesClosed = errors.New("store: show is not on sale")
	// ErrSeatTaken is returned when any requested seat is no longer claimable.
	ErrSeatTaken = errors.New("store: one or more seats are taken")
	// ErrUnknownSeat is returned when a requested seat does not exist.
	ErrUnknownSeat = errors.New("store: unknown seat")
	// ErrIdempotencyMismatch is returned when a key is reused for different seats.
	ErrIdempotencyMismatch = errors.New("store: idempotency key used with different seats")
	// ErrOrderNotFound is returned for an unknown order id.
	ErrOrderNotFound = errors.New("store: order not found")
	// ErrOrderTokenMismatch is returned when X-Order-Token is wrong.
	ErrOrderTokenMismatch = errors.New("store: order token mismatch")
	// ErrOrderNotPending is returned when a paid order is asked to release.
	ErrOrderNotPending = errors.New("store: order is not pending")
)

// ReleaseHold cancels a pending hold after verifying its access token. Releasing
// an already expired or cancelled order is a no-op.
func (s *Store) ReleaseHold(ctx context.Context, orderID, tokenHash string) error {
	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var (
			status   string
			stored   string
			parsedID string
		)
		if err := tx.QueryRow(ctx, `
			SELECT id::text, status, access_token_hash
			FROM ticketing.orders
			WHERE id = $1::uuid
			FOR UPDATE
		`, orderID).Scan(&parsedID, &status, &stored); err != nil {
			if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
				return ErrOrderNotFound
			}
			return fmt.Errorf("lock order: %w", err)
		}

		if subtle.ConstantTimeCompare([]byte(stored), []byte(tokenHash)) != 1 {
			return ErrOrderTokenMismatch
		}

		switch status {
		case "pending":
			if _, err := tx.Exec(ctx, `
				UPDATE ticketing.show_seats
				SET status = 'available', order_id = NULL, held_until = NULL
				WHERE order_id = $1::uuid AND status = 'held'
			`, parsedID); err != nil {
				return fmt.Errorf("release seats: %w", err)
			}

			if _, err := tx.Exec(ctx, `
				UPDATE ticketing.orders SET status = 'cancelled' WHERE id = $1::uuid
			`, parsedID); err != nil {
				return fmt.Errorf("cancel order: %w", err)
			}
			return nil

		case "expired", "cancelled":
			return nil

		default:
			return ErrOrderNotPending
		}
	})

	return err
}

// ExpireHolds marks overdue pending orders expired and frees their held seats.
// It is safe to run from several instances; pg_try_advisory_xact_lock elects one.
// Returns the number of orders transitioned to expired.
func (s *Store) ExpireHolds(ctx context.Context) (int, error) {
	expired := 0

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var locked bool
		if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, sweepLockKey).Scan(&locked); err != nil {
			return fmt.Errorf("acquire sweep lock: %w", err)
		}
		if !locked {
			return nil
		}

		tag, err := tx.Exec(ctx, `
			UPDATE ticketing.orders
			SET status = 'expired'
			WHERE status = 'pending' AND expires_at <= now()
		`)
		if err != nil {
			return fmt.Errorf("expire orders: %w", err)
		}
		expired = int(tag.RowsAffected())

		// Only seats still pointing at an expired order are freed. If another
		// order already re-claimed them, order_id points elsewhere and they are
		// untouched.
		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.show_seats ss
			SET status = 'available', order_id = NULL, held_until = NULL
			FROM ticketing.orders o
			WHERE ss.order_id = o.id
			  AND o.status = 'expired'
			  AND ss.status = 'held'
		`); err != nil {
			return fmt.Errorf("free expired seats: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return expired, nil
}
