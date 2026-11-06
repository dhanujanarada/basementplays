//go:build integration

package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const (
	testCheckoutTTL = 15 * time.Minute
	testManualTTL   = 24 * time.Hour
)

func checkoutInput(orderID, token string, paymentMethod, key string) store.CheckoutInput {
	return store.CheckoutInput{
		OrderID:        orderID,
		TokenHash:      token,
		BuyerName:      "Nimal Perera",
		BuyerPhone:     "+94771234567",
		BuyerEmail:     "nimal@example.com",
		PaymentMethod:  paymentMethod,
		Source:         "online",
		ExpiresAt:      time.Now().Add(testCheckoutTTL),
		IdempotencyKey: key,
	}
}

func TestCheckoutCapturesBuyerAndExtendsHold(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "checkout-hold", "checkout-hash")
	originalExpiry := order.ExpiresAt

	checkedOut, err := st.Checkout(ctx, checkoutInput(order.ID, "checkout-hash", "payhere", "checkout-key"))
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if checkedOut.Status != "pending" || checkedOut.Source != "online" {
		t.Errorf("order = %+v", checkedOut)
	}
	if checkedOut.BuyerName == nil || *checkedOut.BuyerName != "Nimal Perera" {
		t.Errorf("buyer name = %v", checkedOut.BuyerName)
	}
	if checkedOut.BuyerPhone == nil || *checkedOut.BuyerPhone != "+94771234567" {
		t.Errorf("buyer phone = %v", checkedOut.BuyerPhone)
	}
	if checkedOut.BuyerEmail == nil || *checkedOut.BuyerEmail != "nimal@example.com" {
		t.Errorf("buyer email = %v", checkedOut.BuyerEmail)
	}
	if checkedOut.PaymentMethod == nil || *checkedOut.PaymentMethod != "payhere" {
		t.Errorf("payment method = %v", checkedOut.PaymentMethod)
	}
	if !checkedOut.ExpiresAt.After(originalExpiry.Add(-time.Second)) {
		t.Errorf("expires_at = %v, want extended from %v", checkedOut.ExpiresAt, originalExpiry)
	}

	if _, err := st.Checkout(ctx, checkoutInput(order.ID, "wrong-hash", "payhere", "checkout-key")); !errors.Is(err, store.ErrOrderTokenMismatch) {
		t.Fatalf("Checkout(wrong token) = %v, want ErrOrderTokenMismatch", err)
	}

	got, err := st.GetOrderForToken(ctx, order.ID, "checkout-hash")
	if err != nil {
		t.Fatalf("GetOrderForToken: %v", err)
	}
	if got.ID != order.ID || len(got.Seats) != 1 || got.Seats[0].Code != "S001" {
		t.Errorf("loaded order = %+v", got)
	}
	if len(got.Tickets) != 0 {
		t.Errorf("pending order tickets = %+v, want none", got.Tickets)
	}

	if _, err := st.GetOrderForToken(ctx, order.ID, "wrong-hash"); !errors.Is(err, store.ErrOrderTokenMismatch) {
		t.Fatalf("GetOrderForToken(wrong token) = %v, want ErrOrderTokenMismatch", err)
	}
}

func TestCheckoutBankTransferBecomesManualHold(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "bank-hold", "bank-hash")

	input := checkoutInput(order.ID, "bank-hash", "bank_transfer", "bank-key")
	input.Source = "manual"
	input.ExpiresAt = time.Now().Add(testManualTTL)

	checkedOut, err := st.Checkout(ctx, input)
	if err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if checkedOut.Source != "manual" {
		t.Errorf("source = %q, want manual", checkedOut.Source)
	}
	if !checkedOut.ExpiresAt.After(time.Now().Add(23 * time.Hour)) {
		t.Errorf("expires_at = %v, want roughly 24 hours", checkedOut.ExpiresAt)
	}
}

func TestCheckoutOnClosedStates(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "closed-hold", "closed-hash")

	if _, err := st.Checkout(ctx, checkoutInput(order.ID, "closed-hash", "payhere", "closed-key")); err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	if _, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID}); err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}

	paid, err := st.Checkout(ctx, checkoutInput(order.ID, "closed-hash", "payhere", "closed-key"))
	if err != nil {
		t.Fatalf("Checkout(paid): %v", err)
	}
	if paid.Status != "paid" {
		t.Errorf("status = %q, want paid", paid.Status)
	}

	if _, err := st.CancelOrder(ctx, order.ID, "changed mind"); err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if _, err := st.Checkout(ctx, checkoutInput(order.ID, "closed-hash", "payhere", "closed-key")); !errors.Is(err, store.ErrOrderNotCheckoutable) {
		t.Fatalf("Checkout(cancelled) = %v, want ErrOrderNotCheckoutable", err)
	}
}

func TestCheckoutAfterHoldExpired(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "expired-checkout", "expired-checkout-hash")

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.orders SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid
	`, order.ID); err != nil {
		t.Fatalf("expire order: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.show_seats SET held_until = now() - interval '1 minute' WHERE order_id = $1::uuid
	`, order.ID); err != nil {
		t.Fatalf("expire seat: %v", err)
	}

	if _, err := st.Checkout(ctx, checkoutInput(order.ID, "expired-checkout-hash", "payhere", "expired-checkout-key")); !errors.Is(err, store.ErrHoldExpired) {
		t.Fatalf("Checkout(expired) = %v, want ErrHoldExpired", err)
	}
}

func TestManualOrderAndIdempotentMarkPaid(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 2)
	method := "cash"
	name := "Kasun Silva"
	phone := "0771234567"

	order, created, err := st.CreateOrder(ctx, store.CreateOrderInput{
		ShowID:          showID,
		SeatCodes:       []string{"S002", "S001"},
		Source:          "manual",
		PaymentMethod:   &method,
		BuyerName:       &name,
		BuyerPhone:      &phone,
		Notes:           "paid at the door",
		ExpiresAt:       time.Now().Add(testManualTTL),
		AccessTokenHash: "manual-hash",
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if created {
		t.Fatal("created = true, want false")
	}
	if order.Status != "pending" || order.Source != "manual" || order.TotalCents != 500000 {
		t.Errorf("order = %+v", order)
	}
	if order.BuyerName == nil || *order.BuyerName != name || order.Notes != "paid at the door" {
		t.Errorf("buyer/notes = %+v", order)
	}

	first, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID, PaymentMethod: &method})
	if err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}
	if first.NeedsReview {
		t.Fatal("NeedsReview = true, want false")
	}
	if first.Order.Status != "paid" || first.Order.PaidAt == nil {
		t.Errorf("paid order = %+v", first.Order)
	}
	if len(first.Order.Tickets) != 2 {
		t.Fatalf("tickets = %d, want 2", len(first.Order.Tickets))
	}
	for _, ticket := range first.Order.Tickets {
		if ticket.Status != "issued" {
			t.Errorf("ticket %s status = %q, want issued", ticket.SeatCode, ticket.Status)
		}
	}

	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	for _, code := range []string{"S001", "S002"} {
		if got := seatStatus(t, show, code); got != "sold" {
			t.Errorf("%s = %q, want sold", code, got)
		}
	}

	second, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID, PaymentMethod: &method})
	if err != nil {
		t.Fatalf("second MarkPaid: %v", err)
	}
	if second.Order.Status != "paid" || len(second.Order.Tickets) != 2 {
		t.Errorf("second MarkPaid order = %+v", second.Order)
	}

	var tickets int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)::int FROM ticketing.tickets WHERE order_id = $1::uuid
	`, order.ID).Scan(&tickets); err != nil {
		t.Fatalf("count tickets: %v", err)
	}
	if tickets != 2 {
		t.Fatalf("tickets = %d, want 2", tickets)
	}

	orders, err := st.ListShowOrders(ctx, showID)
	if err != nil {
		t.Fatalf("ListShowOrders: %v", err)
	}
	if len(orders) != 1 || orders[0].ID != order.ID {
		t.Fatalf("orders = %+v, want the manual order", orders)
	}
}

func TestLatePaymentReclaimsFreedSeats(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "late-hold", "late-hash")

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.orders SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid
	`, order.ID); err != nil {
		t.Fatalf("expire order: %v", err)
	}
	if _, err := st.ExpireHolds(ctx); err != nil {
		t.Fatalf("ExpireHolds: %v", err)
	}

	paid, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID})
	if err != nil {
		t.Fatalf("MarkPaid(late): %v", err)
	}
	if paid.NeedsReview {
		t.Fatal("NeedsReview = true, want false")
	}
	if paid.Order.Status != "paid" || len(paid.Order.Tickets) != 1 {
		t.Fatalf("order = %+v, want paid with one ticket", paid.Order)
	}

	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "sold" {
		t.Errorf("S001 = %q, want sold", got)
	}
}

func TestMarkPaidNeedsReviewAndRecovery(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	old := holdSeats(t, ctx, st, showID, []string{"S001"}, "review-old", "review-old-hash")

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.orders SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid
	`, old.ID); err != nil {
		t.Fatalf("expire old order: %v", err)
	}
	if _, err := st.ExpireHolds(ctx); err != nil {
		t.Fatalf("ExpireHolds: %v", err)
	}

	newer := holdSeats(t, ctx, st, showID, []string{"S001"}, "review-new", "review-new-hash")

	review, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: old.ID})
	if err != nil {
		t.Fatalf("MarkPaid(taken): %v", err)
	}
	if !review.NeedsReview {
		t.Fatal("NeedsReview = false, want true")
	}
	if review.Order.Status != "needs_review" {
		t.Fatalf("status = %q, want needs_review", review.Order.Status)
	}
	if len(review.Order.Tickets) != 0 {
		t.Errorf("tickets = %+v, want none", review.Order.Tickets)
	}

	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "held" {
		t.Errorf("S001 = %q, want held by the newer order", got)
	}

	if err := st.ReleaseHold(ctx, newer.ID, "review-new-hash"); err != nil {
		t.Fatalf("ReleaseHold: %v", err)
	}

	recovered, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: old.ID})
	if err != nil {
		t.Fatalf("MarkPaid(recovered): %v", err)
	}
	if recovered.NeedsReview {
		t.Fatal("NeedsReview = true after recovery, want false")
	}
	if recovered.Order.Status != "paid" || len(recovered.Order.Tickets) != 1 {
		t.Fatalf("recovered order = %+v", recovered.Order)
	}
}

func TestCancelOrders(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)

	pending := holdSeats(t, ctx, st, showID, []string{"S001"}, "cancel-pending", "cancel-pending-hash")
	cancelled, err := st.CancelOrder(ctx, pending.ID, "buyer cancelled")
	if err != nil {
		t.Fatalf("CancelOrder(pending): %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", cancelled.Status)
	}
	if cancelled.Notes != "buyer cancelled" {
		t.Errorf("notes = %q, want reason", cancelled.Notes)
	}

	show, err := st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "available" {
		t.Errorf("S001 = %q, want available", got)
	}

	second := holdSeats(t, ctx, st, showID, []string{"S001"}, "cancel-paid", "cancel-paid-hash")
	if _, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: second.ID}); err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}

	paidCancel, err := st.CancelOrder(ctx, second.ID, "refunded")
	if err != nil {
		t.Fatalf("CancelOrder(paid): %v", err)
	}
	if paidCancel.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", paidCancel.Status)
	}
	for _, ticket := range paidCancel.Tickets {
		if ticket.Status != "void" {
			t.Errorf("ticket %s status = %q, want void", ticket.SeatCode, ticket.Status)
		}
	}

	show, err = st.ShowAvailability(ctx, showID)
	if err != nil {
		t.Fatalf("ShowAvailability after paid cancel: %v", err)
	}
	if got := seatStatus(t, show, "S001"); got != "available" {
		t.Errorf("S001 = %q, want available", got)
	}
}

func TestCancelRejectsUsedTicketsAndExpiredOrders(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "cancel-used", "cancel-used-hash")
	if _, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID}); err != nil {
		t.Fatalf("MarkPaid: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.tickets SET status = 'used', used_at = now() WHERE order_id = $1::uuid
	`, order.ID); err != nil {
		t.Fatalf("mark ticket used: %v", err)
	}

	if _, err := st.CancelOrder(ctx, order.ID, "too late"); !errors.Is(err, store.ErrTicketsInUse) {
		t.Fatalf("CancelOrder(used) = %v, want ErrTicketsInUse", err)
	}

	otherShow := createShowWithSeats(t, ctx, st, 1)
	expired := holdSeats(t, ctx, st, otherShow, []string{"S001"}, "cancel-expired", "cancel-expired-hash")
	if _, err := pool.Exec(ctx, `
		UPDATE ticketing.orders SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid
	`, expired.ID); err != nil {
		t.Fatalf("expire order: %v", err)
	}
	if _, err := st.ExpireHolds(ctx); err != nil {
		t.Fatalf("ExpireHolds: %v", err)
	}
	if _, err := st.CancelOrder(ctx, expired.ID, "too late"); !errors.Is(err, store.ErrOrderNotCancellable) {
		t.Fatalf("CancelOrder(expired) = %v, want ErrOrderNotCancellable", err)
	}
}

func TestMarkPaidRejectsCancelledOrder(t *testing.T) {
	pool, ctx := schemaTestPool(t)
	st := store.New(pool)

	showID := createShowWithSeats(t, ctx, st, 1)
	order := holdSeats(t, ctx, st, showID, []string{"S001"}, "pay-cancelled", "pay-cancelled-hash")
	if _, err := st.CancelOrder(ctx, order.ID, "cancelled"); err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}

	if _, err := st.MarkPaid(ctx, store.MarkPaidInput{OrderID: order.ID}); !errors.Is(err, store.ErrOrderNotPayable) {
		t.Fatalf("MarkPaid(cancelled) = %v, want ErrOrderNotPayable", err)
	}
}
