//go:build integration

package store_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const testHoldTTL = 10 * time.Minute

func createShowWithSeats(t *testing.T, ctx context.Context, st *store.Store, seats int) string {
	t.Helper()

	suffix := uniqueSuffix()
	planKey := "hold-plan-" + suffix
	createPlan(t, ctx, st, planKey, seats)

	showID := "hold-show-" + suffix
	if _, err := st.UpsertShow(ctx, store.ShowInput{
		ID:             showID,
		ProductionSlug: "hold-test",
		Title:          "Hold test",
		StartsAt:       time.Now().Add(48 * time.Hour),
		Venue:          "Basement",
		SeatPlanKey:    planKey,
		Status:         "on_sale",
		Currency:       "LKR",
	}); err != nil {
		t.Fatalf("create show: %v", err)
	}

	return showID
}

func holdSeats(
	t *testing.T,
	ctx context.Context,
	st *store.Store,
	showID string,
	seats []string,
	idempotencyKey, tokenHash string,
) store.Order {
	t.Helper()

	order, _, err := st.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       seats,
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  idempotencyKey,
		AccessTokenHash: tokenHash,
	})
	if err != nil {
		t.Fatalf("CreateHold(%v): %v", seats, err)
	}

	return order
}

func seatStatus(t *testing.T, show store.PublicShow, code string) string {
	t.Helper()

	for _, seat := range show.Seats {
		if seat.Code == code {
			return seat.Status
		}
	}

	t.Fatalf("seat %q missing from show %q", code, show.ID)
	return ""
}

func TestHoldLifecycleAndRelease(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 3)

	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	if show.Status != "on_sale" || show.Currency != "LKR" {
		t.Errorf("show = %+v", show)
	}
	if show.Columns != 10 {
		t.Errorf("layout columns = %d, want 10", show.Columns)
	}
	if len(show.Seats) != 3 {
		t.Fatalf("seats = %d, want 3", len(show.Seats))
	}

	order := holdSeats(t, ctx, st, showID, []string{"S003", "S001"}, "lifecycle-key", "lifecycle-hash")

	if order.Status != "pending" || order.Source != "online" {
		t.Errorf("order = %+v", order)
	}
	if order.TotalCents != 500000 {
		t.Errorf("total = %d, want 500000", order.TotalCents)
	}
	if len(order.Seats) != 2 || order.Seats[0].Code != "S001" || order.Seats[1].Code != "S003" {
		t.Errorf("order seats = %+v", order.Seats)
	}
	if !order.ExpiresAt.After(time.Now()) {
		t.Errorf("expires_at = %v, want in the future", order.ExpiresAt)
	}

	show, err = st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability after hold: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "held" {
		t.Errorf("S001 = %q, want held", got)
	}
	if got := seatStatus(t, show, "S003"); got != "held" {
		t.Errorf("S003 = %q, want held", got)
	}
	if got := seatStatus(t, show, "S002"); got != "available" {
		t.Errorf("S002 = %q, want available", got)
	}

	if err := st.ReleaseHold(ctx, order.ID, "wrong-hash"); !errors.Is(err, store.ErrOrderTokenMismatch) {
		t.Fatalf("ReleaseHold(wrong token) = %v, want ErrOrderTokenMismatch", err)
	}

	if err := st.ReleaseHold(ctx, order.ID, "lifecycle-hash"); err != nil {
		t.Fatalf("ReleaseHold: %v", err)
	}

	show, err = st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability after release: %v", err)
	}
	for _, code := range []string{"S001", "S002", "S003"} {
		if got := seatStatus(t, show, code); got != "available" {
			t.Errorf("%s = %q, want available", code, got)
		}
	}

	// Releasing again is idempotent.
	if err := st.ReleaseHold(ctx, order.ID, "lifecycle-hash"); err != nil {
		t.Fatalf("second ReleaseHold: %v", err)
	}

	if err := st.ReleaseHold(ctx, "not-a-uuid", "lifecycle-hash"); !errors.Is(err, store.ErrOrderNotFound) {
		t.Fatalf("ReleaseHold(invalid id) = %v, want ErrOrderNotFound", err)
	}
	if err := st.ReleaseHold(ctx, "00000000-0000-0000-0000-000000000000", "lifecycle-hash"); !errors.Is(err, store.ErrOrderNotFound) {
		t.Fatalf("ReleaseHold(unknown id) = %v, want ErrOrderNotFound", err)
	}
}

func TestHoldIdempotencyAndConflicts(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)

	first := holdSeats(t, ctx, st, showID, []string{"S001"}, "idem-key", "idem-hash")

	replay, replayed, err := st.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       []string{"S001"},
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  "idem-key",
		AccessTokenHash: "different-hash",
	})
	if err != nil {
		t.Fatalf("idempotent CreateHold: %v", err)
	}
	if !replayed {
		t.Fatal("replayed = false, want true")
	}
	if replay.ID != first.ID {
		t.Fatalf("replayed order = %q, want %q", replay.ID, first.ID)
	}
	if replay.TotalCents != first.TotalCents || len(replay.Seats) != 1 {
		t.Fatalf("replayed order = %+v", replay)
	}

	_, _, err = st.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       []string{"S002"},
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  "idem-key",
		AccessTokenHash: "idem-hash",
	})
	if !errors.Is(err, store.ErrIdempotencyMismatch) {
		t.Fatalf("mismatched replay = %v, want ErrIdempotencyMismatch", err)
	}

	_, _, err = st.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       []string{"S001"},
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  "other-key",
		AccessTokenHash: "other-hash",
	})
	if !errors.Is(err, store.ErrSeatTaken) {
		t.Fatalf("hold taken seat = %v, want ErrSeatTaken", err)
	}

	_, _, err = st.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       []string{"S999"},
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  "unknown-seat-key",
		AccessTokenHash: "unknown-seat-hash",
	})
	if !errors.Is(err, store.ErrUnknownSeat) {
		t.Fatalf("hold unknown seat = %v, want ErrUnknownSeat", err)
	}

	_, _, err = st.CreateHold(ctx, store.HoldInput{
		ShowID:          "missing-show-" + uniqueSuffix(),
		SeatCodes:       []string{"S001"},
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  "missing-show-key",
		AccessTokenHash: "missing-show-hash",
	})
	if !errors.Is(err, store.ErrShowNotFound) {
		t.Fatalf("hold missing show = %v, want ErrShowNotFound", err)
	}
}

func TestConcurrentIdempotentHoldsReturnOneOrder(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)

	const goroutines = 20

	type result struct {
		order    store.Order
		replayed bool
		err      error
	}

	start := make(chan struct{})
	results := make(chan result, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start

			order, replayed, err := st.CreateHold(ctx, store.HoldInput{
				ShowID:          showID,
				SeatCodes:       []string{"S001", "S002"},
				ExpiresAt:       time.Now().Add(testHoldTTL),
				IdempotencyKey:  "concurrent-idem",
				AccessTokenHash: "concurrent-idem-hash",
			})
			results <- result{order: order, replayed: replayed, err: err}
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var (
		created  int
		replayed int
		orderID  string
	)
	for res := range results {
		if res.err != nil {
			t.Fatalf("CreateHold: %v", res.err)
		}
		if orderID == "" {
			orderID = res.order.ID
		} else if res.order.ID != orderID {
			t.Fatalf("saw orders %q and %q, want one order", orderID, res.order.ID)
		}

		if res.replayed {
			replayed++
		} else {
			created++
		}
	}

	if created != 1 || replayed != goroutines-1 {
		t.Fatalf("created=%d replayed=%d, want 1 and %d", created, replayed, goroutines-1)
	}

	var orders int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM ticketing.orders
		WHERE show_id = $1 AND idempotency_key = 'concurrent-idem'
	`, showID).Scan(&orders); err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if orders != 1 {
		t.Fatalf("orders = %d, want 1", orders)
	}
}

func TestClosedShowRejectsHolds(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.shows SET status = 'closed' WHERE id = $1
	`, showID); err != nil {
		t.Fatalf("close show: %v", err)
	}

	_, _, err := st.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       []string{"S001"},
		ExpiresAt:       time.Now().Add(testHoldTTL),
		IdempotencyKey:  "closed-key",
		AccessTokenHash: "closed-hash",
	})
	if !errors.Is(err, store.ErrSalesClosed) {
		t.Fatalf("hold on closed show = %v, want ErrSalesClosed", err)
	}
}

func TestExpiredHoldIsReclaimableWithoutSweeper(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	old := holdSeats(t, ctx, st, showID, []string{"S001"}, "expired-old", "expired-old-hash")

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.orders SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid
	`, old.ID); err != nil {
		t.Fatalf("expire order: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.show_seats SET held_until = now() - interval '1 minute' WHERE order_id = $1::uuid
	`, old.ID); err != nil {
		t.Fatalf("expire seat hold: %v", err)
	}

	// The sweeper has not run: availability must still treat the seat as free.
	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "available" {
		t.Fatalf("S001 = %q, want available before sweep", got)
	}

	reclaimed := holdSeats(t, ctx, st, showID, []string{"S001"}, "expired-new", "expired-new-hash")
	if reclaimed.ID == old.ID {
		t.Fatal("reclaimed order has the same id as the expired order")
	}

	show, err = st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability after reclaim: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "held" {
		t.Fatalf("S001 = %q, want held by the new order", got)
	}

	if _, err := st.ExpireHolds(ctx); err != nil {
		t.Fatalf("ExpireHolds: %v", err)
	}

	var oldStatus, newStatus, seatOrder, seatStatusValue string
	if err := pool.QueryRow(ctx, `
		SELECT status FROM ticketing.orders WHERE id = $1::uuid
	`, old.ID).Scan(&oldStatus); err != nil {
		t.Fatalf("read old order: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT status FROM ticketing.orders WHERE id = $1::uuid
	`, reclaimed.ID).Scan(&newStatus); err != nil {
		t.Fatalf("read new order: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT order_id::text, status FROM ticketing.show_seats WHERE show_id = $1 AND seat_code = 'S001'
	`, showID).Scan(&seatOrder, &seatStatusValue); err != nil {
		t.Fatalf("read seat: %v", err)
	}

	if oldStatus != "expired" {
		t.Errorf("old order status = %q, want expired", oldStatus)
	}
	if newStatus != "pending" {
		t.Errorf("new order status = %q, want pending", newStatus)
	}
	if seatOrder != reclaimed.ID || seatStatusValue != "held" {
		t.Errorf("seat order=%q status=%q, want %q held", seatOrder, seatStatusValue, reclaimed.ID)
	}
}

func TestSweeperExpiresHoldsAndFreesSeats(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)
	order := holdSeats(t, ctx, st, showID, []string{"S001", "S002"}, "sweep-key", "sweep-hash")

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.orders SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid
	`, order.ID); err != nil {
		t.Fatalf("expire order: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.show_seats SET held_until = now() - interval '1 minute' WHERE order_id = $1::uuid
	`, order.ID); err != nil {
		t.Fatalf("expire seats: %v", err)
	}

	if _, err := st.ExpireHolds(ctx); err != nil {
		t.Fatalf("ExpireHolds: %v", err)
	}

	var status string
	if err := pool.QueryRow(ctx, `
		SELECT status FROM ticketing.orders WHERE id = $1::uuid
	`, order.ID).Scan(&status); err != nil {
		t.Fatalf("read order: %v", err)
	}
	if status != "expired" {
		t.Fatalf("order status = %q, want expired", status)
	}

	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	for _, code := range []string{"S001", "S002"} {
		if got := seatStatus(t, show, code); got != "available" {
			t.Errorf("%s = %q, want available", code, got)
		}
	}

	// Nothing left to expire.
	if _, err := st.ExpireHolds(ctx); err != nil {
		t.Fatalf("second ExpireHolds: %v", err)
	}
}

func TestTwoHundredGoroutinesOneSeat(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)

	const goroutines = 200

	start := make(chan struct{})
	results := make(chan error, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start

			_, _, err := st.CreateHold(ctx, store.HoldInput{
				ShowID:          showID,
				SeatCodes:       []string{"S001"},
				ExpiresAt:       time.Now().Add(testHoldTTL),
				AccessTokenHash: fmt.Sprintf("race-hash-%d", i),
			})
			results <- err
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var (
		winners int
		taken   int
		others  []error
	)
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, store.ErrSeatTaken):
			taken++
		default:
			others = append(others, err)
		}
	}

	if len(others) > 0 {
		t.Fatalf("unexpected errors: %v", others)
	}
	if winners != 1 {
		t.Fatalf("winners = %d, want exactly 1", winners)
	}
	if taken != goroutines-1 {
		t.Fatalf("seat_taken = %d, want %d", taken, goroutines-1)
	}

	var held int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM ticketing.show_seats
		WHERE show_id = $1 AND status = 'held'
	`, showID).Scan(&held); err != nil {
		t.Fatalf("count held seats: %v", err)
	}
	if held != 1 {
		t.Fatalf("held seats = %d, want 1", held)
	}
}

func TestOverlappingMultiSeatHoldsDoNotDeadlockOrOversell(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 3)

	requests := [][]string{
		{"S001", "S002"},
		{"S002", "S003"},
	}

	const rounds = 30

	for round := 0; round < rounds; round++ {
		tokenHash := fmt.Sprintf("overlap-hash-%d", round)

		type result struct {
			order store.Order
			err   error
		}

		start := make(chan struct{})
		results := make(chan result, len(requests))

		var wg sync.WaitGroup
		wg.Add(len(requests))
		for _, seats := range requests {
			seats := seats
			go func() {
				defer wg.Done()
				<-start

				order, _, err := st.CreateHold(ctx, store.HoldInput{
					ShowID:          showID,
					SeatCodes:       seats,
					ExpiresAt:       time.Now().Add(testHoldTTL),
					AccessTokenHash: tokenHash,
				})
				results <- result{order: order, err: err}
			}()
		}

		close(start)
		wg.Wait()
		close(results)

		var (
			winners int
			winner  store.Order
		)
		for res := range results {
			if res.err == nil {
				winners++
				winner = res.order
				continue
			}
			if !errors.Is(res.err, store.ErrSeatTaken) {
				t.Fatalf("round %d: unexpected error: %v", round, res.err)
			}
		}

		if winners != 1 {
			t.Fatalf("round %d: winners = %d, want exactly 1", round, winners)
		}

		show, err := st.ShowAvailability(ctx, showID)
		if err != nil {
			t.Fatalf("round %d: ShowAvailability: %v", round, err)
		}

		available := 0
		for _, seat := range show.Seats {
			if seat.Status == "available" {
				available++
			}
		}
		if available != 1 {
			t.Fatalf("round %d: available = %d, want 1 (winner holds 2)", round, available)
		}

		var heldRows int
		if err := pool.QueryRow(ctx, `
			SELECT count(*)::int FROM ticketing.show_seats
			WHERE show_id = $1 AND status = 'held'
		`, showID).Scan(&heldRows); err != nil {
			t.Fatalf("round %d: count held: %v", round, err)
		}
		if heldRows != 2 {
			t.Fatalf("round %d: held rows = %d, want 2", round, heldRows)
		}

		if err := st.ReleaseHold(ctx, winner.ID, tokenHash); err != nil {
			t.Fatalf("round %d: ReleaseHold: %v", round, err)
		}
	}
}
