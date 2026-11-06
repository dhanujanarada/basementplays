package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

func TestValidateBuyer(t *testing.T) {
	if err := validateBuyer("Nimal", "0771234567", ""); err != nil {
		t.Fatalf("validateBuyer(valid, no email): %v", err)
	}
	if err := validateBuyer("Nimal", "0771234567", "nimal@example.com"); err != nil {
		t.Fatalf("validateBuyer(valid): %v", err)
	}

	tests := []struct {
		name  string
		nameV string
		phone string
		email string
		field string
	}{
		{"missing name", "  ", "0771234567", "", "buyer_name"},
		{"long name", strings.Repeat("n", maxBuyerNameLength+1), "0771234567", "", "buyer_name"},
		{"control in name", "Ni\nmal", "0771234567", "", "buyer_name"},
		{"missing phone", "Nimal", "  ", "", "buyer_phone"},
		{"long phone", "Nimal", strings.Repeat("0", maxBuyerPhoneLength+1), "", "buyer_phone"},
		{"bad email", "Nimal", "0771234567", "not-an-email", "buyer_email"},
		{"long email", "Nimal", "0771234567", strings.Repeat("a", maxBuyerEmailLength) + "@x.com", "buyer_email"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := validateBuyer(tt.nameV, tt.phone, tt.email)
			if err == nil {
				t.Fatal("validateBuyer() = nil")
			}

			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("error %v is not *domain.ValidationError", err)
			}
			if validation.Field != tt.field {
				t.Fatalf("field = %q, want %q (%v)", validation.Field, tt.field, err)
			}
		})
	}
}

func TestCheckoutValidation(t *testing.T) {
	orders := NewOrders(nil, 15, 24, 6)

	valid := CheckoutInput{
		BuyerName:     "Nimal",
		BuyerPhone:    "0771234567",
		PaymentMethod: "payhere",
	}

	tests := []struct {
		name  string
		input CheckoutInput
		field string
	}{
		{
			name:  "missing name",
			input: CheckoutInput{BuyerPhone: "0771234567", PaymentMethod: "payhere"},
			field: "buyer_name",
		},
		{
			name:  "missing phone",
			input: CheckoutInput{BuyerName: "Nimal", PaymentMethod: "payhere"},
			field: "buyer_phone",
		},
		{
			name: "bad payment method",
			input: CheckoutInput{
				BuyerName:     "Nimal",
				BuyerPhone:    "0771234567",
				PaymentMethod: "cash",
			},
			field: "payment_method",
		},
		{
			name: "bad idempotency key",
			input: CheckoutInput{
				BuyerName:     "Nimal",
				BuyerPhone:    "0771234567",
				PaymentMethod: "payhere",
			},
			field: "Idempotency-Key",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			key := ""
			if tt.field == "Idempotency-Key" {
				key = strings.Repeat("k", maxIdempotencyKeyLength+1)
			}

			_, err := orders.Checkout(context.Background(), "order-1", "token", tt.input, key)
			if err == nil {
				t.Fatal("Checkout() = nil")
			}

			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("error %v is not *domain.ValidationError", err)
			}
			if validation.Field != tt.field {
				t.Fatalf("field = %q, want %q (%v)", validation.Field, tt.field, err)
			}
		})
	}

	// Empty token is rejected before touching the store.
	if _, err := orders.Checkout(context.Background(), "order-1", "  ", valid, ""); err == nil {
		t.Fatal("Checkout(no token) = nil")
	}
}

func TestMarkPaidAndCancelValidation(t *testing.T) {
	orders := NewOrders(nil, 15, 24, 6)

	badMethod := "cheque"
	if _, err := orders.MarkPaid(context.Background(), "order-1", &badMethod); err == nil {
		t.Fatal("MarkPaid(bad method) = nil")
	} else {
		var validation *domain.ValidationError
		if !errors.As(err, &validation) || validation.Field != "payment_method" {
			t.Fatalf("MarkPaid(bad method) = %v, want payment_method validation", err)
		}
	}

	if _, err := orders.Cancel(context.Background(), "order-1", strings.Repeat("r", maxCancelReason+1)); err == nil {
		t.Fatal("Cancel(long reason) = nil")
	}
}

func TestToPublicOrderOmitsBuyerDetails(t *testing.T) {
	paid := time.Date(2026, 12, 1, 19, 30, 0, 0, time.UTC)
	buyer := "Nimal"
	phone := "0771234567"
	email := "nimal@example.com"
	method := "cash"
	used := time.Date(2026, 12, 1, 20, 0, 0, 0, time.UTC)

	order := store.Order{
		ID:            "order-1",
		ShowID:        "show-1",
		Status:        "paid",
		Source:        "manual",
		PaymentMethod: &method,
		BuyerName:     &buyer,
		BuyerPhone:    &phone,
		BuyerEmail:    &email,
		TotalCents:    250000,
		Currency:      "LKR",
		ExpiresAt:     paid,
		PaidAt:        &paid,
		Notes:         "internal note",
		CreatedAt:     paid,
		Seats:         []store.OrderSeat{{Code: "S001", TierKey: "standard", TierName: "Standard", PriceCents: 250000}},
		Tickets: []store.Ticket{{
			ID:       "ticket-1",
			SeatCode: "S001",
			Status:   "used",
			IssuedAt: paid,
			UsedAt:   &used,
		}},
	}

	public := toPublicOrder(order)
	if public.ID != "order-1" || public.Status != "paid" || public.TotalCents != 250000 {
		t.Errorf("public order = %+v", public)
	}
	if len(public.Seats) != 1 || len(public.Tickets) != 1 {
		t.Fatalf("seats/tickets = %+v / %+v", public.Seats, public.Tickets)
	}
	if public.PaidAt == nil || *public.PaidAt != "2026-12-01T19:30:00Z" {
		t.Errorf("paid_at = %v", public.PaidAt)
	}
	if public.Tickets[0].UsedAt == nil || *public.Tickets[0].UsedAt != "2026-12-01T20:00:00Z" {
		t.Errorf("used_at = %v", public.Tickets[0].UsedAt)
	}

	internal := toInternalOrder(order)
	if internal.BuyerName == nil || *internal.BuyerName != buyer {
		t.Errorf("internal buyer name = %v", internal.BuyerName)
	}
	if internal.Notes != "internal note" {
		t.Errorf("internal notes = %q", internal.Notes)
	}
}
