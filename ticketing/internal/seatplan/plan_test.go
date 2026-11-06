package seatplan

import (
	"errors"
	"strings"
	"testing"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
)

func validParams() Params {
	return Params{
		Name: "Sample 50",
		Key:  "sample-50",
		Tiers: []TierInput{
			{Name: "Front", PriceCents: 300000, SeatCount: 20},
			{Name: "Standard", PriceCents: 250000, SeatCount: 30},
		},
		Columns:    10,
		Numbering:  NumberingSequential,
		StageLabel: "Stage",
	}
}

func TestGeneratePlan(t *testing.T) {
	layout, err := GeneratePlan(validParams())
	if err != nil {
		t.Fatalf("GeneratePlan() error = %v", err)
	}

	if len(layout.Tiers) != 2 {
		t.Fatalf("tiers = %d, want 2", len(layout.Tiers))
	}
	if layout.Tiers[0].Key != "front" || layout.Tiers[0].Name != "Front" {
		t.Errorf("first tier = %+v", layout.Tiers[0])
	}
	if layout.Tiers[1].Key != "standard" || layout.Tiers[1].SeatCount != 30 {
		t.Errorf("second tier = %+v", layout.Tiers[1])
	}

	if len(layout.Seats) != 50 {
		t.Fatalf("seats = %d, want 50", len(layout.Seats))
	}
	if layout.Seats[0].Code != "S001" || layout.Seats[0].Tier != "front" {
		t.Errorf("first seat = %+v", layout.Seats[0])
	}
	if layout.Seats[19].Code != "S020" || layout.Seats[19].Tier != "front" {
		t.Errorf("seat 20 = %+v", layout.Seats[19])
	}
	if layout.Seats[20].Code != "S021" || layout.Seats[20].Tier != "standard" {
		t.Errorf("seat 21 = %+v", layout.Seats[20])
	}

	// Sequential numbering continues across tiers.
	if layout.Seats[0].Label == nil || *layout.Seats[0].Label != "1" {
		t.Errorf("first label = %v, want 1", layout.Seats[0].Label)
	}
	if layout.Seats[49].Label == nil || *layout.Seats[49].Label != "50" {
		t.Errorf("last label = %v, want 50", layout.Seats[49].Label)
	}

	if layout.Columns != 10 || layout.Numbering != NumberingSequential {
		t.Errorf("layout = %+v", layout)
	}
	if layout.StageLabel != "Stage" {
		t.Errorf("stage label = %q", layout.StageLabel)
	}

	if err := ValidateLayout(layout); err != nil {
		t.Errorf("ValidateLayout(generated) = %v", err)
	}
}

func TestGeneratePlanDefaults(t *testing.T) {
	params := Params{
		Name:  "No numbering",
		Key:   "no-numbering",
		Tiers: []TierInput{{Name: "General", PriceCents: 0, SeatCount: 3}},
	}

	layout, err := GeneratePlan(params)
	if err != nil {
		t.Fatalf("GeneratePlan() error = %v", err)
	}

	if layout.Columns != DefaultColumns {
		t.Errorf("columns = %d, want %d", layout.Columns, DefaultColumns)
	}
	if layout.Numbering != NumberingSequential {
		t.Errorf("numbering = %q, want sequential", layout.Numbering)
	}
	for _, seat := range layout.Seats {
		if seat.Label == nil || *seat.Label == "" {
			t.Errorf("seat %s has no label", seat.Code)
		}
	}
}

func TestGeneratePlanWithoutNumbering(t *testing.T) {
	params := validParams()
	params.Numbering = NumberingNone

	layout, err := GeneratePlan(params)
	if err != nil {
		t.Fatalf("GeneratePlan() error = %v", err)
	}

	for _, seat := range layout.Seats {
		if seat.Label != nil {
			t.Fatalf("seat %s label = %q, want nil", seat.Code, *seat.Label)
		}
	}
}

func TestGeneratePlanTierKeys(t *testing.T) {
	params := Params{
		Name: "Keys",
		Key:  "keys",
		Tiers: []TierInput{
			{Name: "Front Row", PriceCents: 100, SeatCount: 1},
			{Name: "Front-Row.", PriceCents: 100, SeatCount: 1},
			{Name: "!!!", PriceCents: 100, SeatCount: 1},
		},
	}

	layout, err := GeneratePlan(params)
	if err != nil {
		t.Fatalf("GeneratePlan() error = %v", err)
	}

	want := []string{"front-row", "front-row-2", "tier"}
	for i, key := range want {
		if layout.Tiers[i].Key != key {
			t.Errorf("tier %d key = %q, want %q", i, layout.Tiers[i].Key, key)
		}
	}
}

func TestGeneratePlanSeatCodeWidth(t *testing.T) {
	params := Params{
		Name:     "Large",
		Key:      "large",
		Tiers:    []TierInput{{Name: "Standard", PriceCents: 100, SeatCount: 1000}},
		MaxSeats: 1000,
	}

	layout, err := GeneratePlan(params)
	if err != nil {
		t.Fatalf("GeneratePlan() error = %v", err)
	}

	if layout.Seats[0].Code != "S0001" {
		t.Errorf("first code = %q, want S0001", layout.Seats[0].Code)
	}
	if layout.Seats[999].Code != "S1000" {
		t.Errorf("last code = %q, want S1000", layout.Seats[999].Code)
	}
}

func TestGeneratePlanValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Params)
		field  string
	}{
		{
			name:   "missing name",
			mutate: func(p *Params) { p.Name = " " },
			field:  "name",
		},
		{
			name:   "name too long",
			mutate: func(p *Params) { p.Name = strings.Repeat("x", MaxNameLength+1) },
			field:  "name",
		},
		{
			name:   "missing key",
			mutate: func(p *Params) { p.Key = "" },
			field:  "key",
		},
		{
			name:   "invalid key",
			mutate: func(p *Params) { p.Key = "Sample 50" },
			field:  "key",
		},
		{
			name:   "zero columns",
			mutate: func(p *Params) { p.Columns = -1 },
			field:  "columns",
		},
		{
			name:   "bad numbering",
			mutate: func(p *Params) { p.Numbering = "roman" },
			field:  "numbering",
		},
		{
			name:   "no tiers",
			mutate: func(p *Params) { p.Tiers = nil },
			field:  "tiers",
		},
		{
			name: "empty tier name",
			mutate: func(p *Params) {
				p.Tiers[0].Name = " "
			},
			field: "tiers[0].name",
		},
		{
			name: "negative price",
			mutate: func(p *Params) {
				p.Tiers[0].PriceCents = -1
			},
			field: "tiers[0].price_cents",
		},
		{
			name: "zero seats",
			mutate: func(p *Params) {
				p.Tiers[1].SeatCount = 0
			},
			field: "tiers[1].seat_count",
		},
		{
			name: "duplicate tier name",
			mutate: func(p *Params) {
				p.Tiers[1].Name = " front "
			},
			field: "tiers[1].name",
		},
		{
			name:   "too many seats",
			mutate: func(p *Params) { p.MaxSeats = 49 },
			field:  "tiers[1].seat_count",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			params := validParams()
			tt.mutate(&params)

			_, err := GeneratePlan(params)
			if err == nil {
				t.Fatal("GeneratePlan() = nil error")
			}

			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("error %v is not *domain.ValidationError", err)
			}
			if validation.Field != tt.field {
				t.Fatalf("field = %q, want %q (%v)", validation.Field, tt.field, err)
			}
		})
	}
}

func TestValidateLayoutErrors(t *testing.T) {
	base := func() Layout {
		label := "1"
		return Layout{
			Columns:   10,
			Numbering: NumberingSequential,
			Tiers: []Tier{
				{Key: "standard", Name: "Standard", PriceCents: 100, SeatCount: 1},
			},
			Seats: []Seat{{Code: "S001", Tier: "standard", Label: &label}},
		}
	}

	tests := []struct {
		name   string
		mutate func(*Layout)
		field  string
	}{
		{
			name:   "bad columns",
			mutate: func(l *Layout) { l.Columns = 0 },
			field:  "layout.columns",
		},
		{
			name:   "bad numbering",
			mutate: func(l *Layout) { l.Numbering = "roman" },
			field:  "layout.numbering",
		},
		{
			name:   "no tiers",
			mutate: func(l *Layout) { l.Tiers = nil; l.Seats = nil },
			field:  "layout.tiers",
		},
		{
			name:   "no seats",
			mutate: func(l *Layout) { l.Seats = nil },
			field:  "layout.seats",
		},
		{
			name:   "count mismatch",
			mutate: func(l *Layout) { l.Tiers[0].SeatCount = 2 },
			field:  "layout.seats",
		},
		{
			name:   "unknown tier",
			mutate: func(l *Layout) { l.Seats[0].Tier = "vip" },
			field:  "layout.seats[0].tier",
		},
		{
			name:   "duplicate seat",
			mutate: func(l *Layout) { l.Tiers[0].SeatCount = 2; l.Seats = append(l.Seats, l.Seats[0]) },
			field:  "layout.seats[1].code",
		},
		{
			name:   "negative price",
			mutate: func(l *Layout) { l.Tiers[0].PriceCents = -1 },
			field:  "layout.tiers[0].price_cents",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			layout := base()
			tt.mutate(&layout)

			err := ValidateLayout(layout)
			if err == nil {
				t.Fatal("ValidateLayout() = nil error")
			}

			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("error %v is not *domain.ValidationError", err)
			}
			if validation.Field != tt.field {
				t.Fatalf("field = %q, want %q (%v)", validation.Field, tt.field, err)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	valid := []string{"sample-50", "a", "front-row-2", strings.Repeat("a", maxKeyLength)}
	for _, key := range valid {
		if err := ValidateKey(key); err != nil {
			t.Errorf("ValidateKey(%q) = %v", key, err)
		}
	}

	invalid := []string{"", "Sample-50", "-front", "front-", "front--row", "front row", strings.Repeat("a", maxKeyLength+1)}
	for _, key := range invalid {
		if err := ValidateKey(key); err == nil {
			t.Errorf("ValidateKey(%q) = nil", key)
		}
	}
}
