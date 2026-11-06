package service

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/seatplan"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const (
	defaultCurrency = "LKR"
	maxSlugLength   = 200
	maxTitleLength  = 200
	maxIDLength     = 200
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

var validShowStatuses = map[string]bool{
	"draft":     true,
	"on_sale":   true,
	"closed":    true,
	"cancelled": true,
}

// Shows implements show use cases.
type Shows struct {
	store *store.Store
}

// NewShows builds the show service.
func NewShows(st *store.Store) *Shows {
	return &Shows{store: st}
}

// UpsertShowInput is the internal PUT payload.
type UpsertShowInput struct {
	ProductionSlug string    `json:"production_slug"`
	Title          string    `json:"title"`
	StartsAt       time.Time `json:"starts_at"`
	Venue          string    `json:"venue"`
	SeatPlanKey    string    `json:"seat_plan_key"`
	Status         string    `json:"status"`
	Currency       string    `json:"currency"`
}

// Upsert validates the payload and idempotently creates or updates the show.
func (s *Shows) Upsert(ctx context.Context, id string, input UpsertShowInput) (store.Show, error) {
	input = normalizeShowInput(input)
	if err := validateUpsertShow(id, input); err != nil {
		return store.Show{}, err
	}

	return s.store.UpsertShow(ctx, store.ShowInput{
		ID:             id,
		ProductionSlug: input.ProductionSlug,
		Title:          input.Title,
		StartsAt:       input.StartsAt,
		Venue:          input.Venue,
		SeatPlanKey:    input.SeatPlanKey,
		Status:         input.Status,
		Currency:       input.Currency,
	})
}

func normalizeShowInput(input UpsertShowInput) UpsertShowInput {
	input.ProductionSlug = strings.TrimSpace(input.ProductionSlug)
	input.Title = strings.TrimSpace(input.Title)
	input.Venue = strings.TrimSpace(input.Venue)
	input.SeatPlanKey = strings.TrimSpace(input.SeatPlanKey)
	input.Status = strings.TrimSpace(input.Status)
	input.Currency = strings.TrimSpace(input.Currency)

	if input.Status == "" {
		input.Status = "draft"
	}
	if input.Currency == "" {
		input.Currency = defaultCurrency
	}

	return input
}

func validateUpsertShow(id string, input UpsertShowInput) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Invalid("id", "is required")
	}
	if len(id) > maxIDLength {
		return domain.Invalid("id", "must be at most %d characters", maxIDLength)
	}

	if input.ProductionSlug == "" {
		return domain.Invalid("production_slug", "is required")
	}
	if len(input.ProductionSlug) > maxSlugLength {
		return domain.Invalid("production_slug", "must be at most %d characters", maxSlugLength)
	}
	if input.Title == "" {
		return domain.Invalid("title", "is required")
	}
	if utf8.RuneCountInString(input.Title) > maxTitleLength {
		return domain.Invalid("title", "must be at most %d characters", maxTitleLength)
	}
	if input.StartsAt.IsZero() {
		return domain.Invalid("starts_at", "is required (RFC 3339)")
	}
	if input.SeatPlanKey == "" {
		return domain.Invalid("seat_plan_key", "is required")
	}
	if err := seatplan.ValidateKey(input.SeatPlanKey); err != nil {
		return err
	}
	if !validShowStatuses[input.Status] {
		return domain.Invalid("status", `must be one of "draft", "on_sale", "closed", "cancelled"`)
	}
	if !currencyPattern.MatchString(input.Currency) {
		return domain.Invalid("currency", "must be a three-letter uppercase code")
	}

	return nil
}
