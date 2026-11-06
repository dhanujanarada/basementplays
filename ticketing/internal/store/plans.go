package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrPlanKeyExists is returned when a plan key is already taken.
var ErrPlanKeyExists = errors.New("store: seat plan key already exists")

// ErrPlanInUse is returned when a plan is (or just became) referenced by a show.
var ErrPlanInUse = errors.New("store: seat plan is in use")

// ErrPlanNotFound is returned when a show references an unknown plan key.
var ErrPlanNotFound = errors.New("store: seat plan not found")

// SeatPlan is a full plan row, including its stored layout.
type SeatPlan struct {
	ID        string          `json:"id"`
	Key       string          `json:"key"`
	Name      string          `json:"name"`
	Layout    json.RawMessage `json:"layout"`
	CreatedAt time.Time       `json:"created_at"`
}

// SeatPlanSummary is the list view: no layout, plus how many shows use it.
type SeatPlanSummary struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	ShowCount int       `json:"show_count"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePlan inserts a new plan. Layout must already be validated.
func (s *Store) CreatePlan(ctx context.Context, key, name string, layout json.RawMessage) (SeatPlan, error) {
	plan, err := scanPlan(s.pool.QueryRow(ctx, `
		INSERT INTO ticketing.seat_plans (key, name, layout)
		VALUES ($1, $2, $3)
		RETURNING id::text, key, name, layout, created_at
	`, key, name, layout))
	if err != nil {
		if isUniqueViolation(err) {
			return SeatPlan{}, ErrPlanKeyExists
		}
		return SeatPlan{}, fmt.Errorf("insert seat plan: %w", err)
	}

	return plan, nil
}

// GetPlan loads one plan by key.
func (s *Store) GetPlan(ctx context.Context, key string) (SeatPlan, error) {
	plan, err := scanPlan(s.pool.QueryRow(ctx, `
		SELECT id::text, key, name, layout, created_at
		FROM ticketing.seat_plans
		WHERE key = $1
	`, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return SeatPlan{}, ErrNotFound
	}
	if err != nil {
		return SeatPlan{}, fmt.Errorf("select seat plan: %w", err)
	}

	return plan, nil
}

// ListPlans returns plans in creation order with their show usage.
func (s *Store) ListPlans(ctx context.Context) ([]SeatPlanSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.key, p.name, count(s.id)::int, p.created_at
		FROM ticketing.seat_plans p
		LEFT JOIN ticketing.shows s ON s.seat_plan_id = p.id
		GROUP BY p.id
		ORDER BY p.created_at, p.key
	`)
	if err != nil {
		return nil, fmt.Errorf("select seat plans: %w", err)
	}
	defer rows.Close()

	plans := make([]SeatPlanSummary, 0)
	for rows.Next() {
		var plan SeatPlanSummary
		if err := rows.Scan(&plan.ID, &plan.Key, &plan.Name, &plan.ShowCount, &plan.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan seat plan: %w", err)
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seat plans: %w", err)
	}

	return plans, nil
}

// UpdatePlanLayout locks a plan, rejects the change while any show uses it, and
// applies transform inside the transaction. transform must return a validated
// plan.
func (s *Store) UpdatePlanLayout(
	ctx context.Context,
	key string,
	transform func(SeatPlan) (SeatPlan, error),
) (SeatPlan, error) {
	var updated SeatPlan

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		plan, err := planForUpdate(ctx, tx, key)
		if err != nil {
			return err
		}

		var used bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM ticketing.shows WHERE seat_plan_id = $1::uuid
			)
		`, plan.ID).Scan(&used); err != nil {
			return fmt.Errorf("check seat plan usage: %w", err)
		}
		if used {
			return ErrPlanInUse
		}

		next, err := transform(plan)
		if err != nil {
			return err
		}

		updated, err = scanPlan(tx.QueryRow(ctx, `
			UPDATE ticketing.seat_plans
			SET layout = $2
			WHERE key = $1
			RETURNING id::text, key, name, layout, created_at
		`, key, next.Layout))
		if err != nil {
			return fmt.Errorf("update seat plan: %w", err)
		}

		return nil
	})
	if err != nil {
		return SeatPlan{}, err
	}

	return updated, nil
}

// DuplicatePlan copies an existing layout under a new key and name.
func (s *Store) DuplicatePlan(ctx context.Context, sourceKey, newKey, newName string) (SeatPlan, error) {
	var created SeatPlan

	err := s.InTx(ctx, func(tx pgx.Tx) error {
		source, err := scanPlan(tx.QueryRow(ctx, `
			SELECT id::text, key, name, layout, created_at
			FROM ticketing.seat_plans
			WHERE key = $1
		`, sourceKey))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("select source seat plan: %w", err)
		}

		created, err = scanPlan(tx.QueryRow(ctx, `
			INSERT INTO ticketing.seat_plans (key, name, layout)
			VALUES ($1, $2, $3)
			RETURNING id::text, key, name, layout, created_at
		`, newKey, newName, source.Layout))
		if err != nil {
			if isUniqueViolation(err) {
				return ErrPlanKeyExists
			}
			return fmt.Errorf("insert duplicate seat plan: %w", err)
		}

		return nil
	})
	if err != nil {
		return SeatPlan{}, err
	}

	return created, nil
}

func scanPlan(row pgx.Row) (SeatPlan, error) {
	var (
		plan   SeatPlan
		layout []byte
	)
	if err := row.Scan(&plan.ID, &plan.Key, &plan.Name, &layout, &plan.CreatedAt); err != nil {
		return SeatPlan{}, err
	}
	plan.Layout = layout

	return plan, nil
}

func planForUpdate(ctx context.Context, tx pgx.Tx, key string) (SeatPlan, error) {
	plan, err := scanPlan(tx.QueryRow(ctx, `
		SELECT id::text, key, name, layout, created_at
		FROM ticketing.seat_plans
		WHERE key = $1
		FOR UPDATE
	`, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return SeatPlan{}, ErrNotFound
	}
	if err != nil {
		return SeatPlan{}, fmt.Errorf("lock seat plan: %w", err)
	}

	return plan, nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
