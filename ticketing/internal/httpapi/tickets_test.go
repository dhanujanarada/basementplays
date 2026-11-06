package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

type fakeTicketStore struct {
	ticket store.TicketInfo
}

func (f fakeTicketStore) TicketByID(context.Context, string) (store.TicketInfo, error) {
	return f.ticket, nil
}

func ticketAPI(t *testing.T) (*publicAPI, *qr.Signer) {
	t.Helper()

	label := "12"
	info := store.TicketInfo{
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

	signer := qr.NewSigner("secret", "https://basementplays.com")
	api := &publicAPI{
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		tickets: service.NewTickets(fakeTicketStore{ticket: info}, signer),
	}

	return api, signer
}

func TestGetTicket(t *testing.T) {
	api, signer := ticketAPI(t)
	token, err := signer.Sign("3f8b2b6e-1c2d-4e5f-8a9b-0c1d2e3f4a5b")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/public/tickets/"+token, nil)
	req.SetPathValue("token", token)
	rec := httptest.NewRecorder()
	api.getTicket(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["production_slug"] != "dear-daughter" {
		t.Errorf("production_slug = %v", body["production_slug"])
	}
	if body["description"] != "Seat 12 (Front)" {
		t.Errorf("description = %v", body["description"])
	}
	if body["order_reference"] != "7890AB" {
		t.Errorf("order_reference = %v", body["order_reference"])
	}
	if body["qr_url"] != "https://basementplays.com/t/"+token {
		t.Errorf("qr_url = %v", body["qr_url"])
	}
}

func TestGetTicketRejectsInvalidToken(t *testing.T) {
	api, _ := ticketAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/public/tickets/bad", nil)
	req.SetPathValue("token", "bad")
	rec := httptest.NewRecorder()
	api.getTicket(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (%s)", rec.Code, rec.Body.String())
	}
}
