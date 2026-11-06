//go:build integration

package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/seatplan"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

func uniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func createPlan(t *testing.T, ctx context.Context, st *store.Store, key string, seats int) store.SeatPlan {
	t.Helper()

	layout, err := seatplan.GeneratePlan(seatplan.Params{
		Name:      "Test plan " + key,
		Key:       key,
		Tiers:     []seatplan.TierInput{{Name: "Standard", PriceCents: 250000, SeatCount: seats}},
		Columns:   10,
		Numbering: seatplan.NumberingSequential,
		MaxSeats:  500,
	})
	if err != nil {
		t.Fatalf("generate plan: %v", err)
	}

	raw, err := json.Marshal(layout)
	if err != nil {
		t.Fatalf("encode layout: %v", err)
	}

	plan, err := st.CreatePlan(ctx, key, "Test plan "+key, raw)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}

	return plan
}

func decodeLayout(t *testing.T, plan store.SeatPlan) seatplan.Layout {
	t.Helper()

	var layout seatplan.Layout
	if err := json.Unmarshal(plan.Layout, &layout); err != nil {
		t.Fatalf("decode layout: %v", err)
	}
	return layout
}

func stringPtr(s string) *string {
	return &s
}

func TestPlanLifecycle(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)
	suffix := uniqueSuffix()

	key := "lifecycle-" + suffix
	created := createPlan(t, ctx, st, key, 5)

	if created.Key != key || created.Name != "Test plan "+key {
		t.Fatalf("created plan = %+v", created)
	}
	if layout := decodeLayout(t, created); len(layout.Seats) != 5 {
		t.Fatalf("seats = %d, want 5", len(layout.Seats))
	}

	loaded, err := st.GetPlan(ctx, key)
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if loaded.ID != created.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, created.ID)
	}

	summaries, err := st.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	found := false
	for _, summary := range summaries {
		if summary.Key == key {
			found = true
			if summary.ShowCount != 0 {
				t.Errorf("show_count = %d, want 0", summary.ShowCount)
			}
		}
	}
	if !found {
		t.Fatalf("plan %q missing from list", key)
	}

	updated, err := st.UpdatePlanLayout(ctx, key, func(plan store.SeatPlan) (store.SeatPlan, error) {
		var layout seatplan.Layout
		if err := json.Unmarshal(plan.Layout, &layout); err != nil {
			return plan, err
		}
		layout.Seats[0].Label = stringPtr("101")
		layout.Seats[1].Label = nil
		raw, err := json.Marshal(layout)
		if err != nil {
			return plan, err
		}
		plan.Layout = raw
		return plan, nil
	})
	if err != nil {
		t.Fatalf("UpdatePlanLayout: %v", err)
	}
	layout := decodeLayout(t, updated)
	if layout.Seats[0].Label == nil || *layout.Seats[0].Label != "101" {
		t.Errorf("seat 0 label = %v", layout.Seats[0].Label)
	}
	if layout.Seats[1].Label != nil {
		t.Errorf("seat 1 label = %v, want nil", *layout.Seats[1].Label)
	}
	if layout.Seats[2].Label == nil || *layout.Seats[2].Label != "3" {
		t.Errorf("seat 2 label = %v, want 3", layout.Seats[2].Label)
	}

	copyKey := "copy-" + suffix
	copied, err := st.DuplicatePlan(ctx, key, copyKey, "Copy")
	if err != nil {
		t.Fatalf("DuplicatePlan: %v", err)
	}
	if copied.Key != copyKey || copied.Name != "Copy" {
		t.Fatalf("copied plan = %+v", copied)
	}
	if got := len(decodeLayout(t, copied).Seats); got != 5 {
		t.Fatalf("copied seats = %d, want 5", got)
	}

	if _, err := st.DuplicatePlan(ctx, key, copyKey, "Copy again"); !errors.Is(err, store.ErrPlanKeyExists) {
		t.Fatalf("duplicate key error = %v, want ErrPlanKeyExists", err)
	}
	if _, err := st.DuplicatePlan(ctx, "missing-"+suffix, "other-"+suffix, "Other"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing source error = %v, want ErrNotFound", err)
	}
	if _, err := st.CreatePlan(ctx, key, "Duplicate", created.Layout); !errors.Is(err, store.ErrPlanKeyExists) {
		t.Fatalf("create duplicate error = %v, want ErrPlanKeyExists", err)
	}
}

func TestPlanEditsAreBlockedWhileUsed(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)
	suffix := uniqueSuffix()

	key := "used-" + suffix
	createPlan(t, ctx, st, key, 3)

	showID := "used-show-" + suffix
	if _, err := st.UpsertShow(ctx, store.ShowInput{
		ID:             showID,
		ProductionSlug: "dear-daughter",
		Title:          "Used show",
		StartsAt:       time.Now().Add(24 * time.Hour),
		Venue:          "Basement",
		SeatPlanKey:    key,
		Status:         "on_sale",
		Currency:       "LKR",
	}); err != nil {
		t.Fatalf("UpsertShow: %v", err)
	}

	_, err := st.UpdatePlanLayout(ctx, key, func(plan store.SeatPlan) (store.SeatPlan, error) {
		return plan, nil
	})
	if !errors.Is(err, store.ErrPlanInUse) {
		t.Fatalf("UpdatePlanLayout while used = %v, want ErrPlanInUse", err)
	}

	// Duplicating a used plan is the supported way to change it.
	if _, err := st.DuplicatePlan(ctx, key, "used-copy-"+suffix, "Copy"); err != nil {
		t.Fatalf("DuplicatePlan used source: %v", err)
	}

	summaries, err := st.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	for _, summary := range summaries {
		if summary.Key == key && summary.ShowCount != 1 {
			t.Errorf("show_count = %d, want 1", summary.ShowCount)
		}
	}
}

func TestShowUpsertIsIdempotent(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)
	suffix := uniqueSuffix()

	key := "idempotent-" + suffix
	createPlan(t, ctx, st, key, 5)

	showID := "idempotent-show-" + suffix
	input := store.ShowInput{
		ID:             showID,
		ProductionSlug: "dear-daughter",
		Title:          "First title",
		StartsAt:       time.Now().Add(48 * time.Hour),
		Venue:          "Basement",
		SeatPlanKey:    key,
		Status:         "on_sale",
		Currency:       "LKR",
	}

	first, err := st.UpsertShow(ctx, input)
	if err != nil {
		t.Fatalf("first UpsertShow: %v", err)
	}
	if first.SeatCount != 5 || first.SeatPlanKey != key || first.Title != "First title" {
		t.Fatalf("first show = %+v", first)
	}

	input.Title = "Second title"
	input.Venue = "Main hall"
	second, err := st.UpsertShow(ctx, input)
	if err != nil {
		t.Fatalf("second UpsertShow: %v", err)
	}
	if second.Title != "Second title" || second.Venue != "Main hall" {
		t.Fatalf("second show = %+v", second)
	}
	if second.SeatCount != 5 {
		t.Fatalf("seat count after upsert = %d, want 5", second.SeatCount)
	}

	var seatRows int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM ticketing.show_seats WHERE show_id = $1
	`, showID).Scan(&seatRows); err != nil {
		t.Fatalf("count seats: %v", err)
	}
	if seatRows != 5 {
		t.Fatalf("show_seats rows = %d, want 5", seatRows)
	}

	rows, err := pool.Query(ctx, `
		SELECT seat_code, tier_name, price_cents, label, status
		FROM ticketing.show_seats
		WHERE show_id = $1
		ORDER BY sort_index
	`, showID)
	if err != nil {
		t.Fatalf("select seats: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			code, tierName, status string
			price                  int
			label                  *string
		)
		if err := rows.Scan(&code, &tierName, &price, &label, &status); err != nil {
			t.Fatalf("scan seat: %v", err)
		}
		if code != fmt.Sprintf("S%03d", count+1) {
			t.Errorf("seat %d code = %q", count, code)
		}
		if tierName != "Standard" || price != 250000 || status != "available" {
			t.Errorf("seat %s = tier %q price %d status %q", code, tierName, price, status)
		}
		if label == nil {
			t.Errorf("seat %s has nil label", code)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate seats: %v", err)
	}
	if count != 5 {
		t.Fatalf("scanned %d seats, want 5", count)
	}

	// A second show may share the same plan; it gets its own snapshot.
	otherShowID := "idempotent-other-" + suffix
	input.ID = otherShowID
	input.Title = "Other show"
	other, err := st.UpsertShow(ctx, input)
	if err != nil {
		t.Fatalf("UpsertShow for second show: %v", err)
	}
	if other.SeatCount != 5 {
		t.Fatalf("second show seat count = %d, want 5", other.SeatCount)
	}

	if _, err := st.UpsertShow(ctx, store.ShowInput{
		ID:             "missing-plan-" + suffix,
		ProductionSlug: "dear-daughter",
		Title:          "Missing plan",
		StartsAt:       time.Now().Add(72 * time.Hour),
		Venue:          "Basement",
		SeatPlanKey:    "does-not-exist-" + suffix,
		Status:         "on_sale",
		Currency:       "LKR",
	}); !errors.Is(err, store.ErrPlanNotFound) {
		t.Fatalf("unknown plan error = %v, want ErrPlanNotFound", err)
	}
}

func TestShowRelink(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)
	suffix := uniqueSuffix()

	planA := "relink-a-" + suffix
	planB := "relink-b-" + suffix
	createPlan(t, ctx, st, planA, 3)
	createPlan(t, ctx, st, planB, 4)

	showID := "relink-show-" + suffix
	input := store.ShowInput{
		ID:             showID,
		ProductionSlug: "dear-daughter",
		Title:          "Relink show",
		StartsAt:       time.Now().Add(96 * time.Hour),
		Venue:          "Basement",
		SeatPlanKey:    planA,
		Status:         "on_sale",
		Currency:       "LKR",
	}

	if _, err := st.UpsertShow(ctx, input); err != nil {
		t.Fatalf("first UpsertShow: %v", err)
	}

	input.SeatPlanKey = planB
	relinked, err := st.UpsertShow(ctx, input)
	if err != nil {
		t.Fatalf("relink UpsertShow: %v", err)
	}
	if relinked.SeatCount != 4 || relinked.SeatPlanKey != planB {
		t.Fatalf("relinked show = %+v", relinked)
	}

	var codes int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM ticketing.show_seats
		WHERE show_id = $1 AND seat_code LIKE 'S00%'
	`, showID).Scan(&codes); err != nil {
		t.Fatalf("count relinked seats: %v", err)
	}
	if codes != 4 {
		t.Fatalf("relinked seat codes = %d, want 4", codes)
	}

	// Sell one seat, then the show must not be relinked again.
	var orderID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO ticketing.orders (show_id, source, access_token_hash, expires_at)
		VALUES ($1, 'manual', 'relink-token-hash', now() + interval '24 hours')
		RETURNING id::text
	`, showID).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.show_seats
		SET status = 'sold', order_id = $2::uuid, held_until = NULL
		WHERE show_id = $1 AND seat_code = 'S001'
	`, showID, orderID); err != nil {
		t.Fatalf("sell seat: %v", err)
	}

	input.SeatPlanKey = planA
	if _, err := st.UpsertShow(ctx, input); !errors.Is(err, store.ErrCannotRelink) {
		t.Fatalf("relink with sold seat = %v, want ErrCannotRelink", err)
	}
}
