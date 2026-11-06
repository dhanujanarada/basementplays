package service

import (
	"context"
	"errors"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

// ErrInvalidTicketToken is returned when a ticket QR token is malformed, has a
// bad signature, or names a ticket that no longer exists.
var ErrInvalidTicketToken = errors.New("service: invalid ticket token")

// TicketLookup is the store operation the ticket service needs.
type TicketLookup interface {
	TicketByID(ctx context.Context, ticketID string) (store.TicketInfo, error)
}

// Ticket is the public ticket-page payload. It intentionally exposes only the
// short order reference, never buyer details or the full order id.
type Ticket struct {
	TicketID       string  `json:"ticket_id"`
	Status         string  `json:"status"`
	ProductionSlug string  `json:"production_slug"`
	ShowID         string  `json:"show_id"`
	ShowTitle      string  `json:"show_title"`
	StartsAt       string  `json:"starts_at"`
	Venue          string  `json:"venue"`
	SeatCode       string  `json:"seat_code"`
	SeatLabel      *string `json:"seat_label"`
	TierKey        string  `json:"tier_key"`
	TierName       string  `json:"tier_name"`
	Description    string  `json:"description"`
	PriceCents     int     `json:"price_cents"`
	OrderReference string  `json:"order_reference"`
	QRToken        string  `json:"qr_token"`
	QRURL          string  `json:"qr_url"`
}

// Tickets resolves signed QR tokens to ticket details.
type Tickets struct {
	store  TicketLookup
	signer *qr.Signer
}

// NewTickets builds the ticket lookup service.
func NewTickets(st TicketLookup, signer *qr.Signer) *Tickets {
	return &Tickets{store: st, signer: signer}
}

// Lookup verifies the token before loading the ticket.
func (t *Tickets) Lookup(ctx context.Context, token string) (Ticket, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Ticket{}, ErrInvalidTicketToken
	}

	ticketID, err := t.signer.Verify(token)
	if err != nil {
		return Ticket{}, ErrInvalidTicketToken
	}

	info, err := t.store.TicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, store.ErrTicketNotFound) {
			return Ticket{}, ErrInvalidTicketToken
		}
		return Ticket{}, err
	}

	canonical, err := t.signer.Sign(info.ID)
	if err != nil {
		return Ticket{}, err
	}

	label := ""
	if info.SeatLabel != nil {
		label = *info.SeatLabel
	}

	return Ticket{
		TicketID:       info.ID,
		Status:         info.Status,
		ProductionSlug: info.ProductionSlug,
		ShowID:         info.ShowID,
		ShowTitle:      info.ShowTitle,
		StartsAt:       formatTime(info.StartsAt),
		Venue:          info.Venue,
		SeatCode:       info.SeatCode,
		SeatLabel:      info.SeatLabel,
		TierKey:        info.TierKey,
		TierName:       info.TierName,
		Description:    domain.TicketDescription(label, info.TierName),
		PriceCents:     info.PriceCents,
		OrderReference: shortOrderReference(info.OrderID),
		QRToken:        canonical,
		QRURL:          t.signer.URL(canonical),
	}, nil
}

func shortOrderReference(orderID string) string {
	orderID = strings.ToUpper(strings.TrimSpace(orderID))
	if len(orderID) <= 6 {
		return orderID
	}
	return orderID[len(orderID)-6:]
}
