package service

import (
	"errors"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
)

func validShowInput() UpsertShowInput {
	return UpsertShowInput{
		ProductionSlug: "dear-daughter",
		Title:          "Dear Daughter",
		StartsAt:       time.Date(2026, 12, 1, 19, 30, 0, 0, time.UTC),
		Venue:          "Basement",
		SeatPlanKey:    "sample-50",
		Status:         "on_sale",
		Currency:       "LKR",
	}
}

func TestNormalizeShowInput(t *testing.T) {
	input := normalizeShowInput(UpsertShowInput{
		ProductionSlug: "  dear-daughter  ",
		Title:          "  Dear Daughter ",
		Venue:          "  Basement  ",
		SeatPlanKey:    "  sample-50 ",
	})

	if input.ProductionSlug != "dear-daughter" {
		t.Errorf("ProductionSlug = %q", input.ProductionSlug)
	}
	if input.Title != "Dear Daughter" {
		t.Errorf("Title = %q", input.Title)
	}
	if input.Venue != "Basement" {
		t.Errorf("Venue = %q", input.Venue)
	}
	if input.SeatPlanKey != "sample-50" {
		t.Errorf("SeatPlanKey = %q", input.SeatPlanKey)
	}
	if input.Status != "draft" {
		t.Errorf("Status = %q, want draft", input.Status)
	}
	if input.Currency != "LKR" {
		t.Errorf("Currency = %q, want LKR", input.Currency)
	}
}

func TestValidateUpsertShow(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		mutate func(*UpsertShowInput)
		field  string
	}{
		{
			name:  "missing id",
			id:    " ",
			field: "id",
		},
		{
			name:   "missing slug",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.ProductionSlug = "" },
			field:  "production_slug",
		},
		{
			name:   "missing title",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.Title = "" },
			field:  "title",
		},
		{
			name:   "missing starts_at",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.StartsAt = time.Time{} },
			field:  "starts_at",
		},
		{
			name:   "missing plan key",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.SeatPlanKey = "" },
			field:  "seat_plan_key",
		},
		{
			name:   "invalid plan key",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.SeatPlanKey = "Sample 50" },
			field:  "key",
		},
		{
			name:   "invalid status",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.Status = "sold_out" },
			field:  "status",
		},
		{
			name:   "invalid currency",
			id:     "show-1",
			mutate: func(in *UpsertShowInput) { in.Currency = "lkr" },
			field:  "currency",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			input := validShowInput()
			if tt.mutate != nil {
				tt.mutate(&input)
			}

			err := validateUpsertShow(tt.id, input)
			if err == nil {
				t.Fatal("validateUpsertShow() = nil")
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

	if err := validateUpsertShow("show-1", validShowInput()); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
}
