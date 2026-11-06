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

// ErrCannotRelink is returned when a show already has orders or non-available
// seats, so its seat snapshot must not be recreated.
var ErrCannotRelink = errors.New("store: show has orders or unavailable seats")

// ShowInput is the idempotent PUT payload after validation.
type ShowInput struct {
	ID             string
	ProductionSlug string
	Title          string
	StartsAt       time.Time
	Venue          string
	SeatPlanKey    string
	Status         string
	Currency       string
}

// Show is the stored show plus its plan key and seat count.
type Show struct {
	ID             string    `json:"id"`
	ProductionSlug string    `json:"production_slug"`
	Title          string    `json:"title"`
	StartsAt       time.Time `json:"starts_at"`
	Venue          string    `json:"venue"`
	SeatPlanID     string    `json:"seat_plan_id"`
	SeatPlanKey    string    `json:"seat_plan_key"`
	Status         string    `json:"status"`
	Currency       string    `json:"currency"`
	UpdatedAt      time.Time `json:"updated_at"`
	SeatCount      int       `json:"seat_count"`
}

// UpsertShow creates or updates a show and snapshots its seat rows from the
// plan. It is safe to call repeatedly with the same payload.
//
// A show that already uses a different plan may only be relinked while it has
// no order history and every seat is still available.
func (s *Store) UpsertShow(ctx context.Context, input ShowInput) (Show, error) {
	var show Show

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		// Locking the plan serialises label edits against show linking.
		var (
			planID     string
			layoutJSON []byte
		)
		err := tx.QueryRow(ctx, `
			SELECT id::text, layout
			FROM ticketing.seat_plans
			WHERE key = $1
			FOR UPDATE
		`, input.SeatPlanKey).Scan(&planID, &layoutJSON)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPlanNotFound
		}
		if err != nil {
			return fmt.Errorf("lock seat plan: %w", err)
		}

		var layout seatplan.Layout
		if err := json.Unmarshal(layoutJSON, &layout); err != nil {
			return fmt.Errorf("decode seat plan layout: %w", err)
		}
		if err := seatplan.ValidateLayout(layout); err != nil {
			return err
		}

		var existingPlanID *string
		err = tx.QueryRow(ctx, `
			SELECT seat_plan_id::text
			FROM ticketing.shows
			WHERE id = $1
			FOR UPDATE
		`, input.ID).Scan(&existingPlanID)

		exists := true
		if errors.Is(err, pgx.ErrNoRows) {
			exists = false
		} else if err != nil {
			return fmt.Errorf("lock show: %w", err)
		}

		rebuildSeats := false

		if !exists {
			if _, err := tx.Exec(ctx, `
				INSERT INTO ticketing.shows (
					id, production_slug, title, starts_at, venue,
					seat_plan_id, status, currency
				)
				VALUES ($1, $2, $3, $4, $5, $6::uuid, $7, $8)
			`, input.ID, input.ProductionSlug, input.Title, input.StartsAt,
				input.Venue, planID, input.Status, input.Currency); err != nil {
				return fmt.Errorf("insert show: %w", err)
			}
			rebuildSeats = true
		} else {
			if existingPlanID == nil || *existingPlanID != planID {
				var inUse bool
				if err := tx.QueryRow(ctx, `
					SELECT EXISTS (
						SELECT 1 FROM ticketing.order_items WHERE show_id = $1
					) OR EXISTS (
						SELECT 1 FROM ticketing.show_seats
						WHERE show_id = $1 AND status <> 'available'
					)
				`, input.ID).Scan(&inUse); err != nil {
					return fmt.Errorf("check show usage: %w", err)
				}
				if inUse {
					return ErrCannotRelink
				}

				if _, err := tx.Exec(ctx, `
					DELETE FROM ticketing.show_seats WHERE show_id = $1
				`, input.ID); err != nil {
					return fmt.Errorf("delete old show seats: %w", err)
				}
				rebuildSeats = true
			}

			if _, err := tx.Exec(ctx, `
				UPDATE ticketing.shows
				SET production_slug = $2,
				    title = $3,
				    starts_at = $4,
				    venue = $5,
				    seat_plan_id = $6::uuid,
				    status = $7,
				    currency = $8,
				    updated_at = now()
				WHERE id = $1
			`, input.ID, input.ProductionSlug, input.Title, input.StartsAt,
				input.Venue, planID, input.Status, input.Currency); err != nil {
				return fmt.Errorf("update show: %w", err)
			}
		}

		if rebuildSeats {
			if err := insertShowSeats(ctx, tx, input.ID, layout); err != nil {
				return err
			}
		}

		show, err = scanShow(tx.QueryRow(ctx, `
			SELECT s.id, s.production_slug, s.title, s.starts_at, s.venue,
			       s.seat_plan_id::text, p.key, s.status, s.currency, s.updated_at,
			       (SELECT count(*)::int FROM ticketing.show_seats WHERE show_id = s.id)
			FROM ticketing.shows s
			JOIN ticketing.seat_plans p ON p.id = s.seat_plan_id
			WHERE s.id = $1
		`, input.ID))
		if err != nil {
			return fmt.Errorf("read show after upsert: %w", err)
		}

		return nil
	})
	if err != nil {
		return Show{}, err
	}

	return show, nil
}

func scanShow(row pgx.Row) (Show, error) {
	var show Show
	if err := row.Scan(
		&show.ID,
		&show.ProductionSlug,
		&show.Title,
		&show.StartsAt,
		&show.Venue,
		&show.SeatPlanID,
		&show.SeatPlanKey,
		&show.Status,
		&show.Currency,
		&show.UpdatedAt,
		&show.SeatCount,
	); err != nil {
		return Show{}, err
	}

	return show, nil
}

func insertShowSeats(ctx context.Context, tx pgx.Tx, showID string, layout seatplan.Layout) error {
	tierByKey := make(map[string]seatplan.Tier, len(layout.Tiers))
	for _, tier := range layout.Tiers {
		tierByKey[tier.Key] = tier
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"ticketing", "show_seats"},
		[]string{
			"show_id", "seat_code", "tier_key", "tier_name",
			"label", "price_cents", "sort_index",
		},
		pgx.CopyFromSlice(len(layout.Seats), func(i int) ([]any, error) {
			seat := layout.Seats[i]
			tier := tierByKey[seat.Tier]

			var label any
			if seat.Label != nil {
				label = *seat.Label
			}

			return []any{
				showID,
				seat.Code,
				tier.Key,
				tier.Name,
				label,
				tier.PriceCents,
				i,
			}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("insert show seats: %w", err)
	}

	return nil
}
