package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrTicketNotFound is returned for an unknown ticket id or invalid UUID.
var ErrTicketNotFound = errors.New("store: ticket not found")

// TicketInfo is everything the public ticket page needs, plus the production
// slug used to choose the site theme.
type TicketInfo struct {
	ID             string
	Status         string
	ShowID         string
	ProductionSlug string
	ShowTitle      string
	StartsAt       time.Time
	Venue          string
	SeatCode       string
	SeatLabel      *string
	TierKey        string
	TierName       string
	PriceCents     int
	OrderID        string
	OrderStatus    string
	IssuedAt       time.Time
	UsedAt         *time.Time
}

// TicketByID loads a ticket with its show, seat, and order context.
func (s *Store) TicketByID(ctx context.Context, ticketID string) (TicketInfo, error) {
	var ticket TicketInfo

	err := s.pool.QueryRow(ctx, `
		SELECT t.id::text, t.status, t.show_id,
		       s.production_slug, s.title, s.starts_at, s.venue,
		       t.seat_code, ss.label, ss.tier_key, ss.tier_name, ss.price_cents,
		       o.id::text, o.status, t.issued_at, t.used_at
		FROM ticketing.tickets t
		JOIN ticketing.orders o ON o.id = t.order_id
		JOIN ticketing.shows s ON s.id = t.show_id
		JOIN ticketing.show_seats ss
		  ON ss.show_id = t.show_id AND ss.seat_code = t.seat_code
		WHERE t.id = $1::uuid
	`, ticketID).Scan(
		&ticket.ID,
		&ticket.Status,
		&ticket.ShowID,
		&ticket.ProductionSlug,
		&ticket.ShowTitle,
		&ticket.StartsAt,
		&ticket.Venue,
		&ticket.SeatCode,
		&ticket.SeatLabel,
		&ticket.TierKey,
		&ticket.TierName,
		&ticket.PriceCents,
		&ticket.OrderID,
		&ticket.OrderStatus,
		&ticket.IssuedAt,
		&ticket.UsedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return TicketInfo{}, ErrTicketNotFound
	}
	if err != nil {
		return TicketInfo{}, fmt.Errorf("read ticket: %w", err)
	}

	return ticket, nil
}
