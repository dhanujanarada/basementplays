//go:build integration

package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

func TestScannerTokensAndAdmit(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)
	otherShowID := createShowWithSeats(t, ctx, st, 2)

	tokenHash := "scan-hash-" + uniqueSuffix()
	if err := st.SetShowScannerToken(ctx, showID, tokenHash); err != nil {
		t.Fatalf("SetShowScannerToken: %v", err)
	}

	resolved, err := st.ShowIDForScannerToken(ctx, tokenHash)
	if err != nil {
		t.Fatalf("ShowIDForScannerToken: %v", err)
	}
	if resolved != showID {
		t.Fatalf("resolved show = %q, want %q", resolved, showID)
	}

	rotatedHash := "scan-hash-rotated-" + uniqueSuffix()
	if err := st.SetShowScannerToken(ctx, showID, rotatedHash); err != nil {
		t.Fatalf("rotate scanner token: %v", err)
	}
	if _, err := st.ShowIDForScannerToken(ctx, tokenHash); !errors.Is(err, store.ErrScannerTokenNotFound) {
		t.Fatalf("old token = %v, want ErrScannerTokenNotFound", err)
	}
	if _, err := st.ShowIDForScannerToken(ctx, rotatedHash); err != nil {
		t.Fatalf("rotated token: %v", err)
	}
	if err := st.SetShowScannerToken(ctx, "missing-show", "scan-hash-"+uniqueSuffix()); !errors.Is(err, store.ErrShowNotFound) {
		t.Fatalf("unknown show = %v, want ErrShowNotFound", err)
	}

	method := "cash"
	name := "Scanner Test"
	phone := "0771234567"

	order, _, err := st.CreateOrder(ctx, store.CreateOrderInput{
		ShowID:          showID,
		SeatCodes:       []string{"S002"},
		Source:          "manual",
		PaymentMethod:   &method,
		BuyerName:       &name,
		BuyerPhone:      &phone,
		ExpiresAt:       time.Now().Add(testManualTTL),
		AccessTokenHash: "scanner-admit-hash",
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

	admitted, err := st.AdmitTicket(ctx, showID, ticketID)
	if err != nil {
		t.Fatalf("AdmitTicket: %v", err)
	}
	if admitted.ID != ticketID || admitted.Status != "used" || admitted.UsedAt == nil {
		t.Errorf("admitted ticket = %+v", admitted)
	}
	if admitted.ShowID != showID || admitted.SeatCode != "S002" {
		t.Errorf("admitted show/seat = %+v", admitted)
	}
	if admitted.SeatLabel == nil || *admitted.SeatLabel != "2" {
		t.Errorf("seat label = %v", admitted.SeatLabel)
	}
	if admitted.TierName != "Standard" || admitted.PriceCents != 250000 {
		t.Errorf("tier/price = %+v", admitted)
	}
	if admitted.OrderID != order.ID {
		t.Errorf("order = %q, want %q", admitted.OrderID, order.ID)
	}

	_, err = st.AdmitTicket(ctx, showID, ticketID)
	if !errors.Is(err, store.ErrTicketNotIssued) {
		t.Fatalf("second admit = %v, want ErrTicketNotIssued", err)
	}
	var notIssued *store.TicketNotIssuedError
	if !errors.As(err, &notIssued) || notIssued.Status != "used" {
		t.Fatalf("second admit status = %+v, want used", notIssued)
	}

	if _, err := st.AdmitTicket(ctx, otherShowID, ticketID); !errors.Is(err, store.ErrTicketWrongShow) {
		t.Fatalf("wrong-show admit = %v, want ErrTicketWrongShow", err)
	}

	for _, unknown := range []string{"00000000-0000-0000-0000-000000000000", "not-a-uuid"} {
		if _, err := st.AdmitTicket(ctx, showID, unknown); !errors.Is(err, store.ErrTicketNotFound) {
			t.Errorf("AdmitTicket(%q) = %v, want ErrTicketNotFound", unknown, err)
		}
	}

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.tickets SET status = 'void', used_at = NULL WHERE id = $1::uuid
	`, ticketID); err != nil {
		t.Fatalf("void ticket: %v", err)
	}
	if _, err := st.AdmitTicket(ctx, showID, ticketID); !errors.Is(err, store.ErrTicketNotIssued) {
		t.Fatalf("void admit = %v, want ErrTicketNotIssued", err)
	} else if !errors.As(err, &notIssued) || notIssued.Status != "void" {
		t.Fatalf("void admit status = %+v, want void", notIssued)
	}

	if err := st.RevokeShowScannerToken(ctx, showID); err != nil {
		t.Fatalf("RevokeShowScannerToken: %v", err)
	}
	if _, err := st.ShowIDForScannerToken(ctx, rotatedHash); !errors.Is(err, store.ErrScannerTokenNotFound) {
		t.Fatalf("revoked token = %v, want ErrScannerTokenNotFound", err)
	}
	if err := st.RevokeShowScannerToken(ctx, "missing-show"); !errors.Is(err, store.ErrShowNotFound) {
		t.Fatalf("revoke unknown show = %v, want ErrShowNotFound", err)
	}
}
