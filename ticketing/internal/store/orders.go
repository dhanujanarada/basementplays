package store

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
)

var (
	// ErrHoldExpired is returned when checkout starts after the hold lapsed.
	ErrHoldExpired = errors.New("store: hold has expired")
	// ErrOrderNotCheckoutable is returned for cancelled or otherwise closed orders.
	ErrOrderNotCheckoutable = errors.New("store: order is not awaiting checkout")
	// ErrOrderNeedsReview is returned when checkout is attempted on a review order.
	ErrOrderNeedsReview = errors.New("store: order needs review")
	// ErrOrderNotPayable is returned when a cancelled order is marked paid.
	ErrOrderNotPayable = errors.New("store: order cannot be marked paid")
	// ErrOrderNotCancellable is returned when a terminal order is cancelled.
	ErrOrderNotCancellable = errors.New("store: order cannot be cancelled")
	// ErrTicketsInUse is returned when a paid order with used tickets is cancelled.
	ErrTicketsInUse = errors.New("store: order has used tickets")
)

// OrderSeat is one seat snapshot attached to an order.
type OrderSeat struct {
	Code       string  `json:"code"`
	Label      *string `json:"label"`
	TierKey    string  `json:"tier_key"`
	TierName   string  `json:"tier_name"`
	PriceCents int     `json:"price_cents"`
}

// Ticket is one issued ticket row.
type Ticket struct {
	ID       string     `json:"id"`
	SeatCode string     `json:"seat_code"`
	Status   string     `json:"status"`
	IssuedAt time.Time  `json:"issued_at"`
	UsedAt   *time.Time `json:"used_at"`
}

// Order is a full order row with seat and ticket snapshots.
type Order struct {
	ID            string      `json:"id"`
	ShowID        string      `json:"show_id"`
	Status        string      `json:"status"`
	Source        string      `json:"source"`
	PaymentMethod *string     `json:"payment_method"`
	BuyerName     *string     `json:"buyer_name"`
	BuyerPhone    *string     `json:"buyer_phone"`
	BuyerEmail    *string     `json:"buyer_email"`
	TotalCents    int         `json:"total_cents"`
	Currency      string      `json:"currency"`
	ExpiresAt     time.Time   `json:"expires_at"`
	PaidAt        *time.Time  `json:"paid_at"`
	Notes         string      `json:"notes"`
	CreatedAt     time.Time   `json:"created_at"`
	Seats         []OrderSeat `json:"seats"`
	Tickets       []Ticket    `json:"tickets"`

	accessTokenHash string
	idempotencyKey  *string
}

// CreateOrderInput creates a pending online or manual order that holds seats.
type CreateOrderInput struct {
	ShowID          string
	SeatCodes       []string
	Source          string
	PaymentMethod   *string
	BuyerName       *string
	BuyerPhone      *string
	BuyerEmail      *string
	Notes           string
	ExpiresAt       time.Time
	IdempotencyKey  string
	AccessTokenHash string
}

// HoldInput creates a pending online order that holds seats.
type HoldInput struct {
	ShowID          string
	SeatCodes       []string
	ExpiresAt       time.Time
	IdempotencyKey  string
	AccessTokenHash string
}

// CreateHold is the online-hold wrapper around CreateOrder.
func (s *Store) CreateHold(ctx context.Context, input HoldInput) (Order, bool, error) {
	return s.CreateOrder(ctx, CreateOrderInput{
		ShowID:          input.ShowID,
		SeatCodes:       input.SeatCodes,
		Source:          "online",
		ExpiresAt:       input.ExpiresAt,
		IdempotencyKey:  input.IdempotencyKey,
		AccessTokenHash: input.AccessTokenHash,
	})
}

// CreateOrder atomically claims seats and creates a pending order. The second
// return value reports an idempotent replay (the order already existed).
func (s *Store) CreateOrder(ctx context.Context, input CreateOrderInput) (Order, bool, error) {
	seats := uniqueSorted(input.SeatCodes)
	if len(seats) == 0 {
		return Order{}, false, ErrUnknownSeat
	}

	source := input.Source
	if source == "" {
		source = "online"
	}

	if input.IdempotencyKey != "" {
		existing, found, err := s.orderByIdempotencyKey(ctx, input.ShowID, input.IdempotencyKey)
		if err != nil {
			return Order{}, false, err
		}
		if found {
			if !sameSeatSet(existing.Seats, seats) {
				return Order{}, false, ErrIdempotencyMismatch
			}
			return existing, true, nil
		}
	}

	var (
		order Order
		last  error
	)

	for attempt := 0; attempt < 3; attempt++ {
		last = s.InTx(ctx, func(tx pgx.Tx) error {
			var (
				status   string
				currency string
			)
			err := tx.QueryRow(ctx, `
				SELECT status, currency
				FROM ticketing.shows
				WHERE id = $1
				FOR SHARE
			`, input.ShowID).Scan(&status, &currency)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrShowNotFound
			}
			if err != nil {
				return fmt.Errorf("lock show: %w", err)
			}
			if status != "on_sale" {
				return ErrSalesClosed
			}

			rows, err := tx.Query(ctx, `
				SELECT seat_code, price_cents,
				       (status = 'available' OR (status = 'held' AND held_until < now())) AS claimable
				FROM ticketing.show_seats
				WHERE show_id = $1 AND seat_code = ANY($2)
				ORDER BY seat_code
				FOR UPDATE
			`, input.ShowID, seats)
			if err != nil {
				return fmt.Errorf("lock show seats: %w", err)
			}
			defer rows.Close()

			prices := make(map[string]int, len(seats))
			for rows.Next() {
				var (
					code      string
					price     int
					claimable bool
				)
				if err := rows.Scan(&code, &price, &claimable); err != nil {
					return fmt.Errorf("scan show seat: %w", err)
				}
				if !claimable {
					return ErrSeatTaken
				}
				prices[code] = price
			}
			if err := rows.Err(); err != nil {
				return fmt.Errorf("iterate show seats: %w", err)
			}
			if len(prices) != len(seats) {
				return ErrUnknownSeat
			}

			total := 0
			for _, code := range seats {
				total += prices[code]
			}

			var idempotencyKey any
			if input.IdempotencyKey != "" {
				idempotencyKey = input.IdempotencyKey
			}

			var orderID string
			if err := tx.QueryRow(ctx, `
				INSERT INTO ticketing.orders (
					show_id, status, source, payment_method,
					buyer_name, buyer_phone, buyer_email,
					total_cents, access_token_hash, idempotency_key,
					expires_at, notes
				)
				VALUES ($1, 'pending', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				RETURNING id::text
			`, input.ShowID, source, nullable(input.PaymentMethod),
				nullable(input.BuyerName), nullable(input.BuyerPhone), nullable(input.BuyerEmail),
				total, input.AccessTokenHash, idempotencyKey, input.ExpiresAt, input.Notes,
			).Scan(&orderID); err != nil {
				return fmt.Errorf("insert order: %w", err)
			}

			for _, code := range seats {
				if _, err := tx.Exec(ctx, `
					INSERT INTO ticketing.order_items (order_id, show_id, seat_code, price_cents)
					VALUES ($1::uuid, $2, $3, $4)
				`, orderID, input.ShowID, code, prices[code]); err != nil {
					return fmt.Errorf("insert order item: %w", err)
				}
			}

			tag, err := tx.Exec(ctx, `
				UPDATE ticketing.show_seats
				SET status = 'held', order_id = $2::uuid, held_until = $3
				WHERE show_id = $1 AND seat_code = ANY($4)
				  AND (status = 'available' OR (status = 'held' AND held_until < now()))
			`, input.ShowID, orderID, input.ExpiresAt, seats)
			if err != nil {
				return fmt.Errorf("claim show seats: %w", err)
			}
			if tag.RowsAffected() != int64(len(seats)) {
				return ErrSeatTaken
			}

			order, err = loadOrder(ctx, tx, orderID)
			if err != nil {
				return err
			}
			order.Currency = currency

			return nil
		})

		if last == nil {
			return order, false, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(last, &pgErr) {
			switch pgErr.Code {
			case "40P01", "40001": // deadlock, serialization failure
				continue
			}

			if pgErr.Code == "23505" && pgErr.ConstraintName == "orders_show_id_idempotency_key_key" && input.IdempotencyKey != "" {
				return s.replayOrder(ctx, input.ShowID, input.IdempotencyKey, seats, last)
			}
		}

		// Two identical requests can race: the loser blocks on the seat lock,
		// then sees the winner's hold as "taken". Re-check the key so a
		// concurrent retry still returns the same order instead of a conflict.
		if errors.Is(last, ErrSeatTaken) && input.IdempotencyKey != "" {
			return s.replayOrder(ctx, input.ShowID, input.IdempotencyKey, seats, last)
		}

		return Order{}, false, last
	}

	return Order{}, false, fmt.Errorf("create order after retries: %w", last)
}

// replayOrder returns the order already stored for an idempotency key. fallback
// is returned unchanged when no such order exists.
func (s *Store) replayOrder(
	ctx context.Context,
	showID, key string,
	seats []string,
	fallback error,
) (Order, bool, error) {
	existing, found, err := s.orderByIdempotencyKey(ctx, showID, key)
	if err != nil {
		return Order{}, false, err
	}
	if !found {
		return Order{}, false, fallback
	}
	if !sameSeatSet(existing.Seats, seats) {
		return Order{}, false, ErrIdempotencyMismatch
	}

	return existing, true, nil
}

// CheckoutInput captures buyer details and extends a pending hold.
type CheckoutInput struct {
	OrderID        string
	TokenHash      string
	BuyerName      string
	BuyerPhone     string
	BuyerEmail     string
	PaymentMethod  string
	Source         string
	ExpiresAt      time.Time
	IdempotencyKey string
}

// Checkout verifies the order token, captures buyer details, and extends the
// hold. Repeating checkout on a paid order returns it unchanged.
func (s *Store) Checkout(ctx context.Context, input CheckoutInput) (Order, error) {
	var order Order

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		locked, err := lockOrderForToken(ctx, tx, input.OrderID, input.TokenHash)
		if err != nil {
			return err
		}

		switch locked.Status {
		case "paid":
			order, err = loadOrder(ctx, tx, locked.ID)
			return err
		case "pending":
		case "needs_review":
			return ErrOrderNeedsReview
		default:
			return ErrOrderNotCheckoutable
		}

		if !locked.ExpiresAt.After(time.Now()) {
			return ErrHoldExpired
		}

		var items, held int
		if err := tx.QueryRow(ctx, `
			SELECT
				(SELECT count(*)::int FROM ticketing.order_items WHERE order_id = $1::uuid),
				(SELECT count(*)::int FROM ticketing.show_seats
				  WHERE order_id = $1::uuid AND status = 'held' AND held_until > now())
		`, locked.ID).Scan(&items, &held); err != nil {
			return fmt.Errorf("check checkout seats: %w", err)
		}
		if items != held {
			return ErrHoldExpired
		}

		var idempotencyKey any
		if input.IdempotencyKey != "" {
			if locked.idempotencyKey == nil || *locked.idempotencyKey != input.IdempotencyKey {
				var used bool
				if err := tx.QueryRow(ctx, `
					SELECT EXISTS (
						SELECT 1 FROM ticketing.orders
						WHERE show_id = $1 AND idempotency_key = $2 AND id <> $3::uuid
					)
				`, locked.ShowID, input.IdempotencyKey, locked.ID).Scan(&used); err != nil {
					return fmt.Errorf("check checkout idempotency key: %w", err)
				}
				if used {
					return ErrIdempotencyMismatch
				}
			}
			idempotencyKey = input.IdempotencyKey
		}

		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.orders
			SET buyer_name = $2,
			    buyer_phone = $3,
			    buyer_email = $4,
			    payment_method = $5,
			    source = $6,
			    expires_at = $7,
			    idempotency_key = COALESCE($8, idempotency_key)
			WHERE id = $1::uuid
		`, locked.ID, input.BuyerName, input.BuyerPhone, input.BuyerEmail,
			input.PaymentMethod, input.Source, input.ExpiresAt, idempotencyKey,
		); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrIdempotencyMismatch
			}
			return fmt.Errorf("update checkout: %w", err)
		}

		order, err = loadOrder(ctx, tx, locked.ID)
		return err
	})
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// GetOrderForToken returns an order after a constant-time token check.
func (s *Store) GetOrderForToken(ctx context.Context, orderID, tokenHash string) (Order, error) {
	var stored string
	err := s.pool.QueryRow(ctx, `
		SELECT access_token_hash FROM ticketing.orders WHERE id = $1::uuid
	`, orderID).Scan(&stored)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("read order token: %w", err)
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(tokenHash)) != 1 {
		return Order{}, ErrOrderTokenMismatch
	}

	order, err := loadOrder(ctx, s.pool, orderID)
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// MarkPaidInput marks an order paid and optionally records the payment method.
type MarkPaidInput struct {
	OrderID       string
	PaymentMethod *string
}

// MarkPaidResult reports whether the payment could not be seated.
type MarkPaidResult struct {
	Order       Order
	NeedsReview bool
}

// MarkPaid converts held or expired seats to sold, issues one ticket per seat,
// and marks the order paid. If any seat was taken, the order becomes
// needs_review instead of being paid.
func (s *Store) MarkPaid(ctx context.Context, input MarkPaidInput) (MarkPaidResult, error) {
	var result MarkPaidResult

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var (
			orderID string
			showID  string
			status  string
		)
		if err := tx.QueryRow(ctx, `
			SELECT id::text, show_id, status
			FROM ticketing.orders
			WHERE id = $1::uuid
			FOR UPDATE
		`, input.OrderID).Scan(&orderID, &showID, &status); err != nil {
			if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
				return ErrOrderNotFound
			}
			return fmt.Errorf("lock order for payment: %w", err)
		}

		switch status {
		case "paid":
			order, err := loadOrder(ctx, tx, orderID)
			if err != nil {
				return err
			}
			result = MarkPaidResult{Order: order}
			return nil
		case "cancelled":
			return ErrOrderNotPayable
		}

		if err := domain.TransitionOrder(domain.OrderStatus(status), domain.OrderPaid); err != nil {
			return ErrOrderNotPayable
		}

		var seatCodes []string
		rows, err := tx.Query(ctx, `
			SELECT seat_code FROM ticketing.order_items
			WHERE order_id = $1::uuid
			ORDER BY seat_code
		`, orderID)
		if err != nil {
			return fmt.Errorf("read order seat codes: %w", err)
		}
		for rows.Next() {
			var code string
			if err := rows.Scan(&code); err != nil {
				rows.Close()
				return fmt.Errorf("scan order seat code: %w", err)
			}
			seatCodes = append(seatCodes, code)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("iterate order seat codes: %w", err)
		}
		rows.Close()

		if len(seatCodes) == 0 {
			return fmt.Errorf("order %s has no seats", orderID)
		}

		type seatState struct {
			code      string
			status    string
			orderID   *string
			claimable bool
		}

		states := make([]seatState, 0, len(seatCodes))
		rows, err = tx.Query(ctx, `
			SELECT seat_code, status, order_id::text,
			       (status = 'available'
			        OR (status = 'held' AND (order_id = $1::uuid OR held_until < now()))
			        OR (status = 'sold' AND order_id = $1::uuid)) AS claimable
			FROM ticketing.show_seats
			WHERE show_id = $2 AND seat_code = ANY($3)
			ORDER BY seat_code
			FOR UPDATE
		`, orderID, showID, seatCodes)
		if err != nil {
			return fmt.Errorf("lock payment seats: %w", err)
		}
		for rows.Next() {
			var state seatState
			if err := rows.Scan(&state.code, &state.status, &state.orderID, &state.claimable); err != nil {
				rows.Close()
				return fmt.Errorf("scan payment seat: %w", err)
			}
			states = append(states, state)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("iterate payment seats: %w", err)
		}
		rows.Close()

		taken := len(states) != len(seatCodes)
		for _, state := range states {
			if !state.claimable {
				taken = true
				break
			}
		}

		if taken {
			if status != "needs_review" {
				if _, err := tx.Exec(ctx, `
					UPDATE ticketing.orders SET status = 'needs_review' WHERE id = $1::uuid
				`, orderID); err != nil {
					return fmt.Errorf("mark order needs_review: %w", err)
				}
			}

			order, err := loadOrder(ctx, tx, orderID)
			if err != nil {
				return err
			}
			result = MarkPaidResult{Order: order, NeedsReview: true}
			return nil
		}

		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.show_seats
			SET status = 'sold', order_id = $1::uuid, held_until = NULL
			WHERE show_id = $2 AND seat_code = ANY($3)
		`, orderID, showID, seatCodes); err != nil {
			return fmt.Errorf("mark seats sold: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO ticketing.tickets (order_id, show_id, seat_code)
			SELECT $1::uuid, $2, unnest($3::text[])
			ON CONFLICT (show_id, seat_code) WHERE status IN ('issued', 'used') DO NOTHING
		`, orderID, showID, seatCodes); err != nil {
			return fmt.Errorf("issue tickets: %w", err)
		}

		var liveTickets int
		if err := tx.QueryRow(ctx, `
			SELECT count(*)::int FROM ticketing.tickets
			WHERE order_id = $1::uuid AND status IN ('issued', 'used')
		`, orderID).Scan(&liveTickets); err != nil {
			return fmt.Errorf("count issued tickets: %w", err)
		}
		if liveTickets != len(seatCodes) {
			if status != "needs_review" {
				if _, err := tx.Exec(ctx, `
					UPDATE ticketing.orders SET status = 'needs_review' WHERE id = $1::uuid
				`, orderID); err != nil {
					return fmt.Errorf("mark ticket conflict needs_review: %w", err)
				}
			}

			order, err := loadOrder(ctx, tx, orderID)
			if err != nil {
				return err
			}
			result = MarkPaidResult{Order: order, NeedsReview: true}
			return nil
		}

		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.orders
			SET status = 'paid',
			    paid_at = now(),
			    payment_method = COALESCE($2, payment_method)
			WHERE id = $1::uuid
		`, orderID, nullable(input.PaymentMethod)); err != nil {
			return fmt.Errorf("mark order paid: %w", err)
		}

		order, err := loadOrder(ctx, tx, orderID)
		if err != nil {
			return err
		}
		result = MarkPaidResult{Order: order}
		return nil
	})
	if err != nil {
		return MarkPaidResult{}, err
	}

	return result, nil
}

// CancelOrder cancels a pending, paid, or needs_review order. Paid orders void
// their issued tickets; orders with used tickets cannot be cancelled.
func (s *Store) CancelOrder(ctx context.Context, orderID, reason string) (Order, error) {
	var order Order

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		var (
			parsedID string
			status   string
		)
		if err := tx.QueryRow(ctx, `
			SELECT id::text, status FROM ticketing.orders WHERE id = $1::uuid FOR UPDATE
		`, orderID).Scan(&parsedID, &status); err != nil {
			if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
				return ErrOrderNotFound
			}
			return fmt.Errorf("lock order for cancel: %w", err)
		}

		if status == "cancelled" {
			var err error
			order, err = loadOrder(ctx, tx, parsedID)
			return err
		}
		if err := domain.TransitionOrder(domain.OrderStatus(status), domain.OrderCancelled); err != nil {
			return ErrOrderNotCancellable
		}

		if status == "paid" {
			var used int
			if err := tx.QueryRow(ctx, `
				SELECT count(*)::int FROM ticketing.tickets
				WHERE order_id = $1::uuid AND status = 'used'
			`, parsedID).Scan(&used); err != nil {
				return fmt.Errorf("count used tickets: %w", err)
			}
			if used > 0 {
				return ErrTicketsInUse
			}

			if _, err := tx.Exec(ctx, `
				UPDATE ticketing.tickets SET status = 'void'
				WHERE order_id = $1::uuid AND status = 'issued'
			`, parsedID); err != nil {
				return fmt.Errorf("void tickets: %w", err)
			}
		}

		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.show_seats
			SET status = 'available', order_id = NULL, held_until = NULL
			WHERE order_id = $1::uuid AND status IN ('held', 'sold')
		`, parsedID); err != nil {
			return fmt.Errorf("free cancelled seats: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE ticketing.orders
			SET status = 'cancelled',
			    notes = CASE
			        WHEN $2 = '' THEN notes
			        WHEN btrim(notes) = '' THEN $2
			        ELSE notes || E'\n' || $2
			    END
			WHERE id = $1::uuid
		`, parsedID, strings.TrimSpace(reason)); err != nil {
			return fmt.Errorf("cancel order: %w", err)
		}

		var err error
		order, err = loadOrder(ctx, tx, parsedID)
		return err
	})
	if err != nil {
		return Order{}, err
	}

	return order, nil
}

// ListShowOrders returns every order for a show, newest first.
func (s *Store) ListShowOrders(ctx context.Context, showID string) ([]Order, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM ticketing.shows WHERE id = $1)
	`, showID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check show exists: %w", err)
	}
	if !exists {
		return nil, ErrShowNotFound
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text FROM ticketing.orders
		WHERE show_id = $1
		ORDER BY created_at DESC, id
	`, showID)
	if err != nil {
		return nil, fmt.Errorf("select show orders: %w", err)
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan show order id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate show orders: %w", err)
	}
	rows.Close()

	orders := make([]Order, 0, len(ids))
	for _, id := range ids {
		order, err := loadOrder(ctx, s.pool, id)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *Store) orderByIdempotencyKey(ctx context.Context, showID, key string) (Order, bool, error) {
	var orderID string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text FROM ticketing.orders
		WHERE show_id = $1 AND idempotency_key = $2
	`, showID, key).Scan(&orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, false, nil
	}
	if err != nil {
		return Order{}, false, fmt.Errorf("select idempotent order: %w", err)
	}

	order, err := loadOrder(ctx, s.pool, orderID)
	if err != nil {
		return Order{}, false, err
	}

	return order, true, nil
}

type lockedOrder struct {
	ID              string
	ShowID          string
	Status          string
	ExpiresAt       time.Time
	accessTokenHash string
	idempotencyKey  *string
}

func lockOrderForToken(ctx context.Context, tx pgx.Tx, orderID, tokenHash string) (lockedOrder, error) {
	var locked lockedOrder
	if err := tx.QueryRow(ctx, `
		SELECT id::text, show_id, status, expires_at, access_token_hash, idempotency_key
		FROM ticketing.orders
		WHERE id = $1::uuid
		FOR UPDATE
	`, orderID).Scan(
		&locked.ID,
		&locked.ShowID,
		&locked.Status,
		&locked.ExpiresAt,
		&locked.accessTokenHash,
		&locked.idempotencyKey,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
			return lockedOrder{}, ErrOrderNotFound
		}
		return lockedOrder{}, fmt.Errorf("lock order: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(locked.accessTokenHash), []byte(tokenHash)) != 1 {
		return lockedOrder{}, ErrOrderTokenMismatch
	}

	return locked, nil
}

type orderQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func loadOrder(ctx context.Context, q orderQuerier, orderID string) (Order, error) {
	var order Order
	if err := q.QueryRow(ctx, `
		SELECT o.id::text, o.show_id, o.status, o.source, o.payment_method,
		       o.buyer_name, o.buyer_phone, o.buyer_email, o.total_cents,
		       s.currency, o.expires_at, o.paid_at, o.notes, o.created_at,
		       o.access_token_hash, o.idempotency_key
		FROM ticketing.orders o
		JOIN ticketing.shows s ON s.id = o.show_id
		WHERE o.id = $1::uuid
	`, orderID).Scan(
		&order.ID,
		&order.ShowID,
		&order.Status,
		&order.Source,
		&order.PaymentMethod,
		&order.BuyerName,
		&order.BuyerPhone,
		&order.BuyerEmail,
		&order.TotalCents,
		&order.Currency,
		&order.ExpiresAt,
		&order.PaidAt,
		&order.Notes,
		&order.CreatedAt,
		&order.accessTokenHash,
		&order.idempotencyKey,
	); err != nil {
		return Order{}, fmt.Errorf("read order: %w", err)
	}

	rows, err := q.Query(ctx, `
		SELECT oi.seat_code, ss.label, ss.tier_key, ss.tier_name, oi.price_cents
		FROM ticketing.order_items oi
		JOIN ticketing.show_seats ss
		  ON ss.show_id = oi.show_id AND ss.seat_code = oi.seat_code
		WHERE oi.order_id = $1::uuid
		ORDER BY oi.seat_code
	`, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("read order seats: %w", err)
	}
	defer rows.Close()

	order.Seats = make([]OrderSeat, 0)
	for rows.Next() {
		var seat OrderSeat
		if err := rows.Scan(&seat.Code, &seat.Label, &seat.TierKey, &seat.TierName, &seat.PriceCents); err != nil {
			return Order{}, fmt.Errorf("scan order seat: %w", err)
		}
		order.Seats = append(order.Seats, seat)
	}
	if err := rows.Err(); err != nil {
		return Order{}, fmt.Errorf("iterate order seats: %w", err)
	}
	rows.Close()

	ticketRows, err := q.Query(ctx, `
		SELECT id::text, seat_code, status, issued_at, used_at
		FROM ticketing.tickets
		WHERE order_id = $1::uuid
		ORDER BY seat_code
	`, orderID)
	if err != nil {
		return Order{}, fmt.Errorf("read order tickets: %w", err)
	}
	defer ticketRows.Close()

	order.Tickets = make([]Ticket, 0)
	for ticketRows.Next() {
		var ticket Ticket
		if err := ticketRows.Scan(&ticket.ID, &ticket.SeatCode, &ticket.Status, &ticket.IssuedAt, &ticket.UsedAt); err != nil {
			return Order{}, fmt.Errorf("scan order ticket: %w", err)
		}
		order.Tickets = append(order.Tickets, ticket)
	}
	if err := ticketRows.Err(); err != nil {
		return Order{}, fmt.Errorf("iterate order tickets: %w", err)
	}

	return order, nil
}

func sameSeatSet(seats []OrderSeat, codes []string) bool {
	if len(seats) != len(codes) {
		return false
	}

	have := make([]string, len(seats))
	for i, seat := range seats {
		have[i] = seat.Code
	}
	sort.Strings(have)

	want := uniqueSorted(codes)

	for i := range have {
		if have[i] != want[i] {
			return false
		}
	}

	return true
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)

	return out
}

func nullable(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
