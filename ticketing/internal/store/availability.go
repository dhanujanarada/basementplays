package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/dhanujanarada/basementplays/ticketing/internal/seatplan"
)

// PublicSeat is one seat as the booking site sees it. Expired holds are
// reported as available so correctness never depends on the sweeper.
type PublicSeat struct {
	Code       string
	Label      *string
	TierKey    string
	TierName   string
	PriceCents int
	SortIndex  int
	Status     string
}

// PublicShow is the availability read model for one show.
type PublicShow struct {
	ID             string
	ProductionSlug string
	Title          string
	StartsAt       time.Time
	Venue          string
	Status         string
	Currency       string
	Columns        int
	StageLabel     string
	Seats          []PublicSeat
}

// ShowAvailability loads a show and its seats from one consistent snapshot.
// It treats an overdue held seat as available, matching the claim query.
func (s *Store) ShowAvailability(ctx context.Context, showID string) (PublicShow, error) {
	var show PublicShow

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return PublicShow{}, fmt.Errorf("begin availability read: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var layoutJSON []byte
	err = tx.QueryRow(ctx, `
		SELECT s.id, s.production_slug, s.title, s.starts_at, s.venue,
		       s.status, s.currency, p.layout
		FROM ticketing.shows s
		JOIN ticketing.seat_plans p ON p.id = s.seat_plan_id
		WHERE s.id = $1
	`, showID).Scan(
		&show.ID,
		&show.ProductionSlug,
		&show.Title,
		&show.StartsAt,
		&show.Venue,
		&show.Status,
		&show.Currency,
		&layoutJSON,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicShow{}, ErrShowNotFound
	}
	if err != nil {
		return PublicShow{}, fmt.Errorf("read show availability: %w", err)
	}

	var layout seatplan.Layout
	if err := json.Unmarshal(layoutJSON, &layout); err != nil {
		return PublicShow{}, fmt.Errorf("decode show seat plan layout: %w", err)
	}
	show.Columns = layout.Columns
	show.StageLabel = layout.StageLabel

	rows, err := tx.Query(ctx, `
		SELECT seat_code, label, tier_key, tier_name, price_cents, sort_index,
		       CASE
		           WHEN status = 'held' AND held_until < now() THEN 'available'
		           ELSE status
		       END
		FROM ticketing.show_seats
		WHERE show_id = $1
		ORDER BY sort_index, seat_code
	`, showID)
	if err != nil {
		return PublicShow{}, fmt.Errorf("read show seats: %w", err)
	}
	defer rows.Close()

	show.Seats = make([]PublicSeat, 0)
	for rows.Next() {
		var seat PublicSeat
		if err := rows.Scan(
			&seat.Code,
			&seat.Label,
			&seat.TierKey,
			&seat.TierName,
			&seat.PriceCents,
			&seat.SortIndex,
			&seat.Status,
		); err != nil {
			return PublicShow{}, fmt.Errorf("scan show seat: %w", err)
		}
		show.Seats = append(show.Seats, seat)
	}
	if err := rows.Err(); err != nil {
		return PublicShow{}, fmt.Errorf("iterate show seats: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return PublicShow{}, fmt.Errorf("commit availability read: %w", err)
	}

	return show, nil
}
