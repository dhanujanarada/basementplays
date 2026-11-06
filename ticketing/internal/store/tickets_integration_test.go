//go:build integration

package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

func TestTicketByID(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)
	method := "cash"
	name := "Ticket Buyer"
	phone := "0771234567"

	order, _, err := st.CreateOrder(ctx, store.CreateOrderInput{
		ShowID:          showID,
		SeatCodes:       []string{"S002"},
		Source:          "manual",
		PaymentMethod:   &method,
		BuyerName:       &name,
		BuyerPhone:      &phone,
		ExpiresAt:       time.Now().Add(testManualTTL),
		AccessTokenHash: "ticket-lookup-hash",
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	paid, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID, PaymentMethod: &method})
	if err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}
	if len(paid.Order.Tickets) != 1 {
		t.Fatalf("tickets = %d, want 1", len(paid.Order.Tickets))
	}
	ticketID := paid.Order.Tickets[0].ID

	info, err := st.TicketByID(ctx, ticketID)
	if err != nil {
		t.Fatalf("TicketByID: %v", err)
	}
	if info.ID != ticketID || info.Status != "issued" {
		t.Errorf("ticket = %+v", info)
	}
	if info.ShowID != showID || info.ProductionSlug != "hold-test" || info.ShowTitle != "Hold test" {
		t.Errorf("show context = %+v", info)
	}
	if info.Venue != "Basement" {
		t.Errorf("venue = %q", info.Venue)
	}
	if info.SeatCode != "S002" || info.SeatLabel == nil || *info.SeatLabel != "2" {
		t.Errorf("seat context = %+v", info)
	}
	if info.TierKey != "standard" || info.TierName != "Standard" || info.PriceCents != 250000 {
		t.Errorf("tier context = %+v", info)
	}
	if info.OrderID != order.ID || info.OrderStatus != "paid" {
		t.Errorf("order context = %+v", info)
	}
	if info.UsedAt != nil {
		t.Errorf("used_at = %v, want nil", info.UsedAt)
	}

	if _, err := st.TicketByID(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, store.ErrTicketNotFound) {
		t.Fatalf("unknown ticket = %v, want ErrTicketNotFound", err)
	}
	if _, err := st.TicketByID(ctx, "not-a-uuid"); !errors.Is(err, store.ErrTicketNotFound) {
		t.Fatalf("invalid ticket = %v, want ErrTicketNotFound", err)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.tickets SET status = 'used', used_at = now() WHERE id = $1::uuid
	`, ticketID); err != nil {
		t.Fatalf("mark ticket used: %v", err)
	}

	used, err := st.TicketByID(ctx, ticketID)
	if err != nil {
		t.Fatalf("TicketByID(used): %v", err)
	}
	if used.Status != "used" || used.UsedAt == nil {
		t.Fatalf("used ticket = %+v", used)
	}
}
