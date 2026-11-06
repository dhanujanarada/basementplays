// Package service contains application logic: validation, transactions, and
// mapping between HTTP inputs and the store.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/seatplan"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

// Plans implements seat-plan use cases.
type Plans struct {
	store    *store.Store
	maxSeats int
}

// NewPlans builds the plan service. maxSeats is MAX_PLAN_SEATS.
func NewPlans(st *store.Store, maxSeats int) *Plans {
	return &Plans{store: st, maxSeats: maxSeats}
}

// CreatePlanInput is the generator request from the Backstage builder.
type CreatePlanInput struct {
	Name       string               `json:"name"`
	Key        string               `json:"key"`
	Tiers      []seatplan.TierInput `json:"tiers"`
	Columns    int                  `json:"columns"`
	Numbering  seatplan.Numbering   `json:"numbering"`
	StageLabel string               `json:"stage_label"`
}

// DuplicatePlanInput names the copy. Name defaults to "<source> copy".
type DuplicatePlanInput struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Create generates and stores a new plan.
func (s *Plans) Create(ctx context.Context, input CreatePlanInput) (store.SeatPlan, error) {
	layout, err := seatplan.GeneratePlan(seatplan.Params{
		Name:       input.Name,
		Key:        input.Key,
		Tiers:      input.Tiers,
		Columns:    input.Columns,
		Numbering:  input.Numbering,
		StageLabel: input.StageLabel,
		MaxSeats:   s.maxSeats,
	})
	if err != nil {
		return store.SeatPlan{}, err
	}

	raw, err := json.Marshal(layout)
	if err != nil {
		return store.SeatPlan{}, fmt.Errorf("encode seat plan layout: %w", err)
	}

	return s.store.CreatePlan(ctx, strings.TrimSpace(input.Key), strings.TrimSpace(input.Name), raw)
}

// Get returns one plan.
func (s *Plans) Get(ctx context.Context, key string) (store.SeatPlan, error) {
	return s.store.GetPlan(ctx, strings.TrimSpace(key))
}

// List returns all plans with their show usage.
func (s *Plans) List(ctx context.Context) ([]store.SeatPlanSummary, error) {
	return s.store.ListPlans(ctx)
}

// UpdateLabels edits or clears individual seat labels. It fails with
// store.ErrPlanInUse once any show references the plan.
func (s *Plans) UpdateLabels(
	ctx context.Context,
	key string,
	labels map[string]*string,
) (store.SeatPlan, error) {
	return s.store.UpdatePlanLayout(ctx, strings.TrimSpace(key), func(plan store.SeatPlan) (store.SeatPlan, error) {
		var layout seatplan.Layout
		if err := json.Unmarshal(plan.Layout, &layout); err != nil {
			return plan, fmt.Errorf("decode seat plan layout: %w", err)
		}

		index := make(map[string]int, len(layout.Seats))
		for i, seat := range layout.Seats {
			index[seat.Code] = i
		}

		for code, label := range labels {
			i, ok := index[code]
			if !ok {
				return plan, domain.Invalid("labels."+code, "unknown seat code")
			}

			if label == nil {
				layout.Seats[i].Label = nil
				continue
			}
			trimmed := strings.TrimSpace(*label)
			if trimmed == "" {
				layout.Seats[i].Label = nil
				continue
			}
			layout.Seats[i].Label = &trimmed
		}

		if err := seatplan.ValidateLayout(layout); err != nil {
			return plan, err
		}

		raw, err := json.Marshal(layout)
		if err != nil {
			return plan, fmt.Errorf("encode seat plan layout: %w", err)
		}
		plan.Layout = raw

		return plan, nil
	})
}

// Duplicate copies a plan under a new key.
func (s *Plans) Duplicate(
	ctx context.Context,
	sourceKey string,
	input DuplicatePlanInput,
) (store.SeatPlan, error) {
	newKey := strings.TrimSpace(input.Key)
	if err := seatplan.ValidateKey(newKey); err != nil {
		return store.SeatPlan{}, err
	}

	newName := strings.TrimSpace(input.Name)
	if newName == "" {
		source, err := s.store.GetPlan(ctx, strings.TrimSpace(sourceKey))
		if err != nil {
			return store.SeatPlan{}, err
		}
		newName = strings.TrimSpace(source.Name + " copy")
	}
	if len(newName) > seatplan.MaxNameLength {
		return store.SeatPlan{}, domain.Invalid("name", "must be at most %d characters", seatplan.MaxNameLength)
	}

	return s.store.DuplicatePlan(ctx, strings.TrimSpace(sourceKey), newKey, newName)
}
