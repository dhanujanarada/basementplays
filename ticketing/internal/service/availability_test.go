package service

import (
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

func TestSummarizeSeats(t *testing.T) {
	label := "1"
	seats := []store.PublicSeat{
		{Code: "S001", Label: &label, TierKey: "front", TierName: "Front", PriceCents: 300000, Status: "available"},
		{Code: "S002", TierKey: "front", TierName: "Front", PriceCents: 300000, Status: "held"},
		{Code: "S003", TierKey: "front", TierName: "Front", PriceCents: 300000, Status: "sold"},
		{Code: "S004", TierKey: "standard", TierName: "Standard", PriceCents: 250000, Status: "available"},
		{Code: "S005", TierKey: "standard", TierName: "Standard", PriceCents: 250000, Status: "blocked"},
	}

	tiers, total, remaining := summarizeSeats(seats)

	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if remaining != 2 {
		t.Errorf("remaining = %d, want 2", remaining)
	}
	if len(tiers) != 2 {
		t.Fatalf("tiers = %d, want 2", len(tiers))
	}

	if tiers[0].Key != "front" || tiers[0].SeatsTotal != 3 || tiers[0].SeatsRemaining != 1 {
		t.Errorf("front tier = %+v", tiers[0])
	}
	if tiers[1].Key != "standard" || tiers[1].SeatsTotal != 2 || tiers[1].SeatsRemaining != 1 {
		t.Errorf("standard tier = %+v", tiers[1])
	}
}

func TestSummarizeSeatsEmpty(t *testing.T) {
	tiers, total, remaining := summarizeSeats(nil)

	if len(tiers) != 0 || total != 0 || remaining != 0 {
		t.Fatalf("tiers=%v total=%d remaining=%d", tiers, total, remaining)
	}
	if got := priceFrom(tiers); got != 0 {
		t.Fatalf("priceFrom(empty) = %d, want 0", got)
	}
}

func TestPriceFrom(t *testing.T) {
	tiers := []TierSummary{
		{PriceCents: 300000},
		{PriceCents: 250000},
		{PriceCents: 500000},
	}

	if got := priceFrom(tiers); got != 250000 {
		t.Fatalf("priceFrom = %d, want 250000", got)
	}
}

func TestSummaryStartsAtFormat(t *testing.T) {
	when := time.Date(2026, 12, 1, 19, 30, 0, 0, time.UTC)
	if got := when.Format(timeFormat); got != "2026-12-01T19:30:00Z" {
		t.Fatalf("formatted time = %q", got)
	}
}
