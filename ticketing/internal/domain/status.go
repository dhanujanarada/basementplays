// Package domain holds the ticketing state machines and other pure business rules.
// It has no database or HTTP dependencies.
package domain

import (
	"errors"
	"fmt"
)

// ErrIllegalTransition is the sentinel returned by every Transition* function.
// Callers can test it with errors.Is and inspect the details with errors.As.
var ErrIllegalTransition = errors.New("illegal state transition")

// TransitionError describes a rejected state change.
type TransitionError struct {
	Entity string
	From   string
	To     string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("%s: cannot transition from %q to %q", e.Entity, e.From, e.To)
}

// Is lets errors.Is(err, ErrIllegalTransition) match any TransitionError.
func (e *TransitionError) Is(target error) bool {
	return target == ErrIllegalTransition
}

// SeatStatus is a row state from ticketing.show_seats.
type SeatStatus string

const (
	SeatAvailable SeatStatus = "available"
	SeatHeld      SeatStatus = "held"
	SeatSold      SeatStatus = "sold"
	SeatBlocked   SeatStatus = "blocked"
)

// Valid reports whether s is a known seat status.
func (s SeatStatus) Valid() bool {
	switch s {
	case SeatAvailable, SeatHeld, SeatSold, SeatBlocked:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether the seat state machine allows from -> next.
//
//	available -> held | blocked
//	held      -> sold | available   (sold on payment, available on release/expiry)
//	blocked   -> available          (admin)
//	sold      -> available          (admin cancel of a paid order)
func (s SeatStatus) CanTransitionTo(next SeatStatus) bool {
	switch s {
	case SeatAvailable:
		return next == SeatHeld || next == SeatBlocked
	case SeatHeld:
		return next == SeatSold || next == SeatAvailable
	case SeatBlocked:
		return next == SeatAvailable
	case SeatSold:
		return next == SeatAvailable
	default:
		return false
	}
}

// TransitionSeat validates a seat state change.
func TransitionSeat(from, to SeatStatus) error {
	if from.CanTransitionTo(to) {
		return nil
	}
	return &TransitionError{Entity: "seat", From: string(from), To: string(to)}
}

// OrderStatus is a row state from ticketing.orders.
type OrderStatus string

const (
	OrderPending     OrderStatus = "pending"
	OrderPaid        OrderStatus = "paid"
	OrderExpired     OrderStatus = "expired"
	OrderCancelled   OrderStatus = "cancelled"
	OrderNeedsReview OrderStatus = "needs_review"
)

// Valid reports whether s is a known order status.
func (s OrderStatus) Valid() bool {
	switch s {
	case OrderPending, OrderPaid, OrderExpired, OrderCancelled, OrderNeedsReview:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether the order state machine allows from -> next.
//
//	pending      -> paid | expired | cancelled | needs_review
//	expired      -> paid | needs_review (a late payment may be re-claimed)
//	paid         -> cancelled
//	needs_review -> paid | cancelled
//	cancelled is terminal.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	switch s {
	case OrderPending:
		return next == OrderPaid ||
			next == OrderExpired ||
			next == OrderCancelled ||
			next == OrderNeedsReview
	case OrderExpired:
		return next == OrderPaid || next == OrderNeedsReview
	case OrderPaid:
		return next == OrderCancelled
	case OrderNeedsReview:
		return next == OrderPaid || next == OrderCancelled
	default:
		return false
	}
}

// TransitionOrder validates an order state change.
func TransitionOrder(from, to OrderStatus) error {
	if from.CanTransitionTo(to) {
		return nil
	}
	return &TransitionError{Entity: "order", From: string(from), To: string(to)}
}

// TicketStatus is a row state from ticketing.tickets.
type TicketStatus string

const (
	TicketIssued TicketStatus = "issued"
	TicketUsed   TicketStatus = "used"
	TicketVoid   TicketStatus = "void"
)

// Valid reports whether s is a known ticket status.
func (s TicketStatus) Valid() bool {
	switch s {
	case TicketIssued, TicketUsed, TicketVoid:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether the ticket state machine allows from -> next.
// used and void are terminal.
func (s TicketStatus) CanTransitionTo(next TicketStatus) bool {
	switch s {
	case TicketIssued:
		return next == TicketUsed || next == TicketVoid
	default:
		return false
	}
}

// TransitionTicket validates a ticket state change.
func TransitionTicket(from, to TicketStatus) error {
	if from.CanTransitionTo(to) {
		return nil
	}
	return &TransitionError{Entity: "ticket", From: string(from), To: string(to)}
}
