package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestSeatTransitions(t *testing.T) {
	statuses := []SeatStatus{SeatAvailable, SeatHeld, SeatSold, SeatBlocked}

	allowed := map[[2]SeatStatus]bool{
		{SeatAvailable, SeatHeld}:    true,
		{SeatAvailable, SeatBlocked}: true,
		{SeatHeld, SeatSold}:         true,
		{SeatHeld, SeatAvailable}:    true,
		{SeatBlocked, SeatAvailable}: true,
		{SeatSold, SeatAvailable}:    true,
	}

	for _, from := range statuses {
		for _, to := range statuses {
			from, to := from, to
			name := string(from) + "_to_" + string(to)

			t.Run(name, func(t *testing.T) {
				want := allowed[[2]SeatStatus{from, to}]
				if got := from.CanTransitionTo(to); got != want {
					t.Fatalf("%s.CanTransitionTo(%s) = %v, want %v", from, to, got, want)
				}

				err := TransitionSeat(from, to)
				if want && err != nil {
					t.Fatalf("TransitionSeat(%s, %s) = %v, want nil", from, to, err)
				}
				if !want && err == nil {
					t.Fatalf("TransitionSeat(%s, %s) = nil, want error", from, to)
				}
			})
		}
	}
}

func TestOrderTransitions(t *testing.T) {
	statuses := []OrderStatus{
		OrderPending, OrderPaid, OrderExpired, OrderCancelled, OrderNeedsReview,
	}

	allowed := map[[2]OrderStatus]bool{
		{OrderPending, OrderPaid}:          true,
		{OrderPending, OrderExpired}:       true,
		{OrderPending, OrderCancelled}:     true,
		{OrderPending, OrderNeedsReview}:   true,
		{OrderExpired, OrderPaid}:          true,
		{OrderExpired, OrderNeedsReview}:   true,
		{OrderPaid, OrderCancelled}:        true,
		{OrderNeedsReview, OrderPaid}:      true,
		{OrderNeedsReview, OrderCancelled}: true,
	}

	for _, from := range statuses {
		for _, to := range statuses {
			from, to := from, to
			name := string(from) + "_to_" + string(to)

			t.Run(name, func(t *testing.T) {
				want := allowed[[2]OrderStatus{from, to}]
				if got := from.CanTransitionTo(to); got != want {
					t.Fatalf("%s.CanTransitionTo(%s) = %v, want %v", from, to, got, want)
				}

				err := TransitionOrder(from, to)
				if want && err != nil {
					t.Fatalf("TransitionOrder(%s, %s) = %v, want nil", from, to, err)
				}
				if !want && err == nil {
					t.Fatalf("TransitionOrder(%s, %s) = nil, want error", from, to)
				}
			})
		}
	}
}

func TestTicketTransitions(t *testing.T) {
	statuses := []TicketStatus{TicketIssued, TicketUsed, TicketVoid}

	allowed := map[[2]TicketStatus]bool{
		{TicketIssued, TicketUsed}: true,
		{TicketIssued, TicketVoid}: true,
	}

	for _, from := range statuses {
		for _, to := range statuses {
			from, to := from, to
			name := string(from) + "_to_" + string(to)

			t.Run(name, func(t *testing.T) {
				want := allowed[[2]TicketStatus{from, to}]
				if got := from.CanTransitionTo(to); got != want {
					t.Fatalf("%s.CanTransitionTo(%s) = %v, want %v", from, to, got, want)
				}

				err := TransitionTicket(from, to)
				if want && err != nil {
					t.Fatalf("TransitionTicket(%s, %s) = %v, want nil", from, to, err)
				}
				if !want && err == nil {
					t.Fatalf("TransitionTicket(%s, %s) = nil, want error", from, to)
				}
			})
		}
	}
}

func TestTransitionErrorDetails(t *testing.T) {
	err := TransitionOrder(OrderPaid, OrderPending)
	if err == nil {
		t.Fatal("TransitionOrder(paid, pending) = nil, want error")
	}
	if !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("errors.Is(err, ErrIllegalTransition) = false for %v", err)
	}

	var transition *TransitionError
	if !errors.As(err, &transition) {
		t.Fatalf("errors.As(%v, *TransitionError) = false", err)
	}
	if transition.Entity != "order" || transition.From != "paid" || transition.To != "pending" {
		t.Fatalf("TransitionError = %+v", transition)
	}
	for _, part := range []string{"order", "paid", "pending"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error %q does not mention %q", err, part)
		}
	}
}

func TestUnknownStatusesAreRejected(t *testing.T) {
	if SeatStatus("gone").Valid() {
		t.Error("SeatStatus(\"gone\").Valid() = true")
	}
	if OrderStatus("gone").Valid() {
		t.Error("OrderStatus(\"gone\").Valid() = true")
	}
	if TicketStatus("gone").Valid() {
		t.Error("TicketStatus(\"gone\").Valid() = true")
	}

	if err := TransitionSeat("gone", SeatHeld); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("TransitionSeat from unknown status = %v", err)
	}
	if err := TransitionOrder(OrderPending, "gone"); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("TransitionOrder to unknown status = %v", err)
	}
	if err := TransitionTicket("gone", TicketIssued); !errors.Is(err, ErrIllegalTransition) {
		t.Errorf("TransitionTicket from unknown status = %v", err)
	}
}
