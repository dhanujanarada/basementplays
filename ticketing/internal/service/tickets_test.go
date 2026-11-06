package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

type fakeTicketLookup struct {
	ticket store.TicketInfo
	err    error
	id     string
}

func (f *fakeTicketLookup) TicketByID(_ context.Context, ticketID string) (store.TicketInfo, error) {
	f.id = ticketID
	return f.ticket, f.err
}

func sampleTicketInfo() store.TicketInfo {
	label := "12"
	return store.TicketInfo{
		ID:             "3f8b2b6e-1c2d-4e5f-8a9b-0c1d2e3f4a5b",
		Status:         "issued",
		ShowID:         "show-1",
		ProductionSlug: "dear-daughter",
		ShowTitle:      "Dear Daughter",
		StartsAt:       time.Date(2026, 12, 1, 19, 30, 0, 0, time.UTC),
		Venue:          "Basement",
		SeatCode:       "S012",
		SeatLabel:      &label,
		TierKey:        "front",
		TierName:       "Front",
		PriceCents:     300000,
		OrderID:        "12345678-1234-1234-1234-1234567890ab",
		OrderStatus:    "paid",
		IssuedAt:       time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC),
	}
}

func TestTicketLookup(t *testing.T) {
	signer := qr.NewSigner("secret", "https://basementplays.com")
	lookup := &fakeTicketLookup{ticket: sampleTicketInfo()}
	tickets := NewTickets(lookup, signer)

	token, err := signer.Sign(lookup.ticket.ID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	got, err := tickets.Lookup(context.Background(), token)
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if lookup.id != lookup.ticket.ID {
		t.Fatalf("store id = %q, want %q", lookup.id, lookup.ticket.ID)
	}
	if got.TicketID != lookup.ticket.ID || got.Status != "issued" {
		t.Errorf("ticket = %+v", got)
	}
	if got.ProductionSlug != "dear-daughter" || got.ShowTitle != "Dear Daughter" {
		t.Errorf("production/show = %q / %q", got.ProductionSlug, got.ShowTitle)
	}
	if got.StartsAt != "2026-12-01T19:30:00Z" {
		t.Errorf("starts_at = %q", got.StartsAt)
	}
	if got.Description != "Seat 12 (Front)" {
		t.Errorf("description = %q", got.Description)
	}
	if got.OrderReference != "7890AB" {
		t.Errorf("order reference = %q, want 7890AB", got.OrderReference)
	}
	if got.QRToken != token {
		t.Errorf("qr token = %q, want %q", got.QRToken, token)
	}
	if got.QRURL != "https://basementplays.com/t/"+token {
		t.Errorf("qr url = %q", got.QRURL)
	}
}

func TestTicketLookupWithoutLabel(t *testing.T) {
	signer := qr.NewSigner("secret", "https://example.com")
	info := sampleTicketInfo()
	info.SeatLabel = nil
	lookup := &fakeTicketLookup{ticket: info}
	tickets := NewTickets(lookup, signer)

	token, err := signer.Sign(info.ID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	got, err := tickets.Lookup(context.Background(), token)
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Description != "Front admission" {
		t.Errorf("description = %q", got.Description)
	}
}

func TestTicketLookupRejectsBadTokens(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	tickets := NewTickets(&fakeTicketLookup{}, signer)

	for _, token := range []string{"", "not-a-token", "v1.ticket.AAAA"} {
		if _, err := tickets.Lookup(context.Background(), token); !errors.Is(err, ErrInvalidTicketToken) {
			t.Errorf("Lookup(%q) = %v, want ErrInvalidTicketToken", token, err)
		}
	}

	valid, err := signer.Sign(sampleTicketInfo().ID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	tampered := "v1.11111111-1111-1111-1111-111111111111." + valid[len("v1."+sampleTicketInfo().ID+"."):]
	if _, err := tickets.Lookup(context.Background(), tampered); !errors.Is(err, ErrInvalidTicketToken) {
		t.Fatalf("Lookup(tampered) = %v, want ErrInvalidTicketToken", err)
	}
}

func TestTicketLookupUnknownTicket(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	lookup := &fakeTicketLookup{err: store.ErrTicketNotFound}
	tickets := NewTickets(lookup, signer)

	token, err := signer.Sign(sampleTicketInfo().ID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	if _, err := tickets.Lookup(context.Background(), token); !errors.Is(err, ErrInvalidTicketToken) {
		t.Fatalf("Lookup = %v, want ErrInvalidTicketToken", err)
	}
}
