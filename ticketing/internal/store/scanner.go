package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	// ErrScannerTokenNotFound is returned when no show owns the token hash.
	ErrScannerTokenNotFound = errors.New("store: scanner token not found")
	// ErrTicketWrongShow is returned when a ticket belongs to another show.
	ErrTicketWrongShow = errors.New("store: ticket belongs to a different show")
	// ErrTicketNotIssued is returned when a ticket is already used or void.
	ErrTicketNotIssued = errors.New("store: ticket is not issued")
)

// TicketNotIssuedError carries the terminal status of a ticket that can no
// longer be admitted.
type TicketNotIssuedError struct {
	Status string
	Ticket TicketInfo
}

func (e *TicketNotIssuedError) Error() string {
	return fmt.Sprintf("store: ticket is %s", e.Status)
}

// Is lets errors.Is(err, ErrTicketNotIssued) match the typed error.
func (e *TicketNotIssuedError) Is(target error) bool {
	return target == ErrTicketNotIssued
}

// SetShowScannerToken replaces a show's scanner credential. Only the hash is
// stored; the caller returns the plaintext once.
func (s *Store) SetShowScannerToken(ctx context.Context, showID, tokenHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE ticketing.shows
		SET scanner_token_hash = $2, scanner_token_set_at = now()
		WHERE id = $1
	`, showID, tokenHash)
	if err != nil {
		return fmt.Errorf("set scanner token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrShowNotFound
	}

	return nil
}

// RevokeShowScannerToken removes a show's scanner credential.
func (s *Store) RevokeShowScannerToken(ctx context.Context, showID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE ticketing.shows
		SET scanner_token_hash = NULL, scanner_token_set_at = NULL
		WHERE id = $1
	`, showID)
	if err != nil {
		return fmt.Errorf("revoke scanner token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrShowNotFound
	}

	return nil
}

// ShowIDForScannerToken resolves a scanner token hash to its show.
func (s *Store) ShowIDForScannerToken(ctx context.Context, tokenHash string) (string, error) {
	var showID string

	err := s.pool.QueryRow(ctx, `
		SELECT id FROM ticketing.shows WHERE scanner_token_hash = $1
	`, tokenHash).Scan(&showID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrScannerTokenNotFound
	}
	if err != nil {
		return "", fmt.Errorf("resolve scanner token: %w", err)
	}

	return showID, nil
}

// AdmitTicket atomically moves an issued ticket to used. The UPDATE is a single
// statement, so two scanners racing on the same ticket produce one success.
//
// A ticket that is already used or void returns a *TicketNotIssuedError with
// its current status; a ticket from another show returns ErrTicketWrongShow.
func (s *Store) AdmitTicket(ctx context.Context, showID, ticketID string) (TicketInfo, error) {
	var ticket TicketInfo

	err := s.pool.QueryRow(ctx, `
		WITH admitted AS (
			UPDATE ticketing.tickets
			SET status = 'used', used_at = now()
			WHERE id = $1::uuid AND show_id = $2 AND status = 'issued'
			RETURNING id, order_id, show_id, seat_code, issued_at, used_at
		)
		SELECT a.id::text, a.show_id,
		       sh.production_slug, sh.title, sh.starts_at, sh.venue,
		       a.seat_code, ss.label, ss.tier_key, ss.tier_name, ss.price_cents,
		       a.order_id::text, o.status, a.issued_at, a.used_at
		FROM admitted a
		JOIN ticketing.shows sh ON sh.id = a.show_id
		JOIN ticketing.orders o ON o.id = a.order_id
		JOIN ticketing.show_seats ss
		  ON ss.show_id = a.show_id AND ss.seat_code = a.seat_code
	`, ticketID, showID).Scan(
		&ticket.ID,
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
		return s.classifyAdmitFailure(ctx, showID, ticketID)
	}
	if err != nil {
		return TicketInfo{}, fmt.Errorf("admit ticket: %w", err)
	}

	ticket.Status = "used"

	return ticket, nil
}

func (s *Store) classifyAdmitFailure(ctx context.Context, showID, ticketID string) (TicketInfo, error) {
	ticket, err := s.TicketByID(ctx, ticketID)
	if err != nil {
		return TicketInfo{}, err
	}
	if ticket.ShowID != showID {
		return ticket, ErrTicketWrongShow
	}

	return ticket, &TicketNotIssuedError{Status: ticket.Status, Ticket: ticket}
}
