// Package seatplan generates and validates seat-plan layouts. It is a pure
// package: no database, HTTP, or clock dependencies.
package seatplan

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
)

const (
	DefaultColumns  = 10
	DefaultMaxSeats = 500

	// NumberingNone leaves every seat label NULL.
	NumberingNone Numbering = "none"
	// NumberingSequential labels seats 1..N across all tiers.
	NumberingSequential Numbering = "sequential"
)

// Numbering is the label mode for a plan.
type Numbering string

// Valid reports whether n is a supported numbering mode.
func (n Numbering) Valid() bool {
	return n == NumberingNone || n == NumberingSequential
}

// TierInput is one tier supplied by the plan builder.
type TierInput struct {
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
	SeatCount  int    `json:"seat_count"`
}

// Params are the generator inputs. MaxSeats defaults to DefaultMaxSeats.
type Params struct {
	Name       string
	Key        string
	Tiers      []TierInput
	Columns    int
	Numbering  Numbering
	StageLabel string
	MaxSeats   int
}

// Tier is the stored form of a tier inside a layout.
type Tier struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
	SeatCount  int    `json:"seat_count"`
}

// Seat is the stored form of one seat inside a layout. Label is NULL when the
// plan uses no numbering.
type Seat struct {
	Code  string  `json:"code"`
	Tier  string  `json:"tier"`
	Label *string `json:"label"`
}

// Layout is the jsonb document stored in ticketing.seat_plans.layout.
type Layout struct {
	StageLabel string    `json:"stage_label"`
	Columns    int       `json:"columns"`
	Numbering  Numbering `json:"numbering"`
	Tiers      []Tier    `json:"tiers"`
	Seats      []Seat    `json:"seats"`
}

// GeneratePlan validates params and builds a deterministic layout.
func GeneratePlan(params Params) (Layout, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return Layout{}, domain.Invalid("name", "is required")
	}
	if len(name) > MaxNameLength {
		return Layout{}, domain.Invalid("name", "must be at most %d characters", MaxNameLength)
	}
	if err := ValidateKey(params.Key); err != nil {
		return Layout{}, err
	}

	columns := params.Columns
	if columns == 0 {
		columns = DefaultColumns
	}
	if columns < 1 {
		return Layout{}, domain.Invalid("columns", "must be at least 1")
	}

	numbering := params.Numbering
	if numbering == "" {
		numbering = NumberingSequential
	}
	if !numbering.Valid() {
		return Layout{}, domain.Invalid("numbering", `must be "none" or "sequential"`)
	}

	maxSeats := params.MaxSeats
	if maxSeats <= 0 {
		maxSeats = DefaultMaxSeats
	}

	if len(params.Tiers) == 0 {
		return Layout{}, domain.Invalid("tiers", "at least one tier is required")
	}

	seenNames := make(map[string]bool, len(params.Tiers))
	usedKeys := make(map[string]bool, len(params.Tiers))
	tiers := make([]Tier, 0, len(params.Tiers))
	total := 0

	for i, input := range params.Tiers {
		field := fmt.Sprintf("tiers[%d]", i)

		tierName := strings.TrimSpace(input.Name)
		if tierName == "" {
			return Layout{}, domain.Invalid(field+".name", "is required")
		}
		if len(tierName) > MaxNameLength {
			return Layout{}, domain.Invalid(field+".name", "must be at most %d characters", MaxNameLength)
		}
		if input.PriceCents < 0 {
			return Layout{}, domain.Invalid(field+".price_cents", "must not be negative")
		}
		if input.SeatCount <= 0 {
			return Layout{}, domain.Invalid(field+".seat_count", "must be greater than zero")
		}

		normalized := strings.ToLower(tierName)
		if seenNames[normalized] {
			return Layout{}, domain.Invalid(field+".name", "duplicate tier name %q", tierName)
		}
		seenNames[normalized] = true

		total += input.SeatCount
		if total > maxSeats {
			return Layout{}, domain.Invalid(
				field+".seat_count",
				"total seats %d exceeds the maximum of %d",
				total, maxSeats,
			)
		}

		tiers = append(tiers, Tier{
			Key:        uniqueSlug(tierName, usedKeys),
			Name:       tierName,
			PriceCents: input.PriceCents,
			SeatCount:  input.SeatCount,
		})
	}

	width := len(strconv.Itoa(total))
	if width < 3 {
		width = 3
	}

	seats := make([]Seat, 0, total)
	number := 1
	for _, tier := range tiers {
		for i := 0; i < tier.SeatCount; i++ {
			seat := Seat{
				Code: fmt.Sprintf("S%0*d", width, number),
				Tier: tier.Key,
			}
			if numbering == NumberingSequential {
				label := strconv.Itoa(number)
				seat.Label = &label
			}
			seats = append(seats, seat)
			number++
		}
	}

	layout := Layout{
		StageLabel: strings.TrimSpace(params.StageLabel),
		Columns:    columns,
		Numbering:  numbering,
		Tiers:      tiers,
		Seats:      seats,
	}

	if err := ValidateLayout(layout); err != nil {
		return Layout{}, fmt.Errorf("generated layout is invalid: %w", err)
	}

	return layout, nil
}
