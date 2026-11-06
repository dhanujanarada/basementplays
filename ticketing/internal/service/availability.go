package service

import (
	"context"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

// TierSummary is one price tier with live availability.
type TierSummary struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	PriceCents     int    `json:"price_cents"`
	SeatsTotal     int    `json:"seats_total"`
	SeatsRemaining int    `json:"seats_remaining"`
}

// ShowSummary is the lightweight availability badge for a show.
type ShowSummary struct {
	ID             string        `json:"id"`
	ProductionSlug string        `json:"production_slug"`
	Title          string        `json:"title"`
	StartsAt       string        `json:"starts_at"`
	Venue          string        `json:"venue"`
	Status         string        `json:"status"`
	Currency       string        `json:"currency"`
	Columns        int           `json:"columns"`
	StageLabel     string        `json:"stage_label"`
	TotalSeats     int           `json:"total_seats"`
	SeatsRemaining int           `json:"seats_remaining"`
	PriceFrom      int           `json:"price_from"`
	Tiers          []TierSummary `json:"tiers"`
}

// Seat is one seat in the public seat map.
type Seat struct {
	Code       string  `json:"code"`
	Label      *string `json:"label"`
	TierKey    string  `json:"tier_key"`
	TierName   string  `json:"tier_name"`
	PriceCents int     `json:"price_cents"`
	Status     string  `json:"status"`
	SortIndex  int     `json:"sort_index"`
}

// SeatMap is the ETag-able seat layout for a show.
type SeatMap struct {
	ShowID         string        `json:"show_id"`
	Status         string        `json:"status"`
	Currency       string        `json:"currency"`
	Columns        int           `json:"columns"`
	StageLabel     string        `json:"stage_label"`
	TotalSeats     int           `json:"total_seats"`
	SeatsRemaining int           `json:"seats_remaining"`
	Tiers          []TierSummary `json:"tiers"`
	Seats          []Seat        `json:"seats"`
}

// Availability builds public read models from the store snapshot.
type Availability struct {
	store *store.Store
}

// NewAvailability builds the availability service.
func NewAvailability(st *store.Store) *Availability {
	return &Availability{store: st}
}

// Summary returns the show badge and per-tier remaining counts.
func (a *Availability) Summary(ctx context.Context, showID string) (ShowSummary, error) {
	show, err := a.store.ShowAvailability(ctx, strings.TrimSpace(showID))
	if err != nil {
		return ShowSummary{}, err
	}

	tiers, total, remaining := summarizeSeats(show.Seats)

	return ShowSummary{
		ID:             show.ID,
		ProductionSlug: show.ProductionSlug,
		Title:          show.Title,
		StartsAt:       show.StartsAt.Format(timeFormat),
		Venue:          show.Venue,
		Status:         show.Status,
		Currency:       show.Currency,
		Columns:        show.Columns,
		StageLabel:     show.StageLabel,
		TotalSeats:     total,
		SeatsRemaining: remaining,
		PriceFrom:      priceFrom(tiers),
		Tiers:          tiers,
	}, nil
}

// SeatMap returns every seat with its public status.
func (a *Availability) SeatMap(ctx context.Context, showID string) (SeatMap, error) {
	show, err := a.store.ShowAvailability(ctx, strings.TrimSpace(showID))
	if err != nil {
		return SeatMap{}, err
	}

	tiers, total, remaining := summarizeSeats(show.Seats)

	seats := make([]Seat, 0, len(show.Seats))
	for _, seat := range show.Seats {
		seats = append(seats, Seat{
			Code:       seat.Code,
			Label:      seat.Label,
			TierKey:    seat.TierKey,
			TierName:   seat.TierName,
			PriceCents: seat.PriceCents,
			Status:     seat.Status,
			SortIndex:  seat.SortIndex,
		})
	}

	return SeatMap{
		ShowID:         show.ID,
		Status:         show.Status,
		Currency:       show.Currency,
		Columns:        show.Columns,
		StageLabel:     show.StageLabel,
		TotalSeats:     total,
		SeatsRemaining: remaining,
		Tiers:          tiers,
		Seats:          seats,
	}, nil
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

// summarizeSeats returns tiers in seat order plus totals. Blocked, held, and
// sold seats are not remaining.
func summarizeSeats(seats []store.PublicSeat) ([]TierSummary, int, int) {
	tiers := make([]TierSummary, 0)
	index := make(map[string]int)
	remaining := 0

	for _, seat := range seats {
		i, ok := index[seat.TierKey]
		if !ok {
			i = len(tiers)
			index[seat.TierKey] = i
			tiers = append(tiers, TierSummary{
				Key:        seat.TierKey,
				Name:       seat.TierName,
				PriceCents: seat.PriceCents,
			})
		}

		tiers[i].SeatsTotal++
		if seat.Status == "available" {
			tiers[i].SeatsRemaining++
			remaining++
		}
	}

	return tiers, len(seats), remaining
}

func priceFrom(tiers []TierSummary) int {
	if len(tiers) == 0 {
		return 0
	}

	min := tiers[0].PriceCents
	for _, tier := range tiers[1:] {
		if tier.PriceCents < min {
			min = tier.PriceCents
		}
	}

	return min
}
