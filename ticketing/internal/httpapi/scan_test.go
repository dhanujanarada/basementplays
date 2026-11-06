package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const testScannerToken = "scanner-token-for-tests"

type fakeScanStore struct {
	resolvedShow string
	resolveErr   error
	admitted     store.TicketInfo
	admitErr     error

	seenHash    string
	issuedShow  string
	revokedShow string
}

func (f *fakeScanStore) SetShowScannerToken(_ context.Context, showID, tokenHash string) error {
	f.issuedShow = showID
	f.seenHash = tokenHash
	return nil
}

func (f *fakeScanStore) RevokeShowScannerToken(_ context.Context, showID string) error {
	f.revokedShow = showID
	return nil
}

func (f *fakeScanStore) ShowIDForScannerToken(_ context.Context, tokenHash string) (string, error) {
	f.seenHash = tokenHash
	return f.resolvedShow, f.resolveErr
}

func (f *fakeScanStore) AdmitTicket(_ context.Context, showID, ticketID string) (store.TicketInfo, error) {
	if f.admitErr != nil {
		return store.TicketInfo{}, f.admitErr
	}

	ticket := f.admitted
	ticket.ID = ticketID
	ticket.ShowID = showID
	return ticket, nil
}

func scanHandler(st service.ScannerStore, signer *qr.Signer) http.Handler {
	return New(Options{
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Scanner: service.NewScanner(st, signer),
	})
}

func scanRequest(t *testing.T, token, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/scan/admit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set(scannerTokenHeader, token)
	}
	return req
}

func TestScanAdmit(t *testing.T) {
	signer := qr.NewSigner("secret", "https://basementplays.test")
	label := "12"
	usedAt := time.Date(2026, 12, 1, 20, 5, 0, 0, time.UTC)
	st := &fakeScanStore{
		resolvedShow: "show-1",
		admitted: store.TicketInfo{
			Status:      "used",
			SeatCode:    "S012",
			SeatLabel:   &label,
			TierName:    "Front",
			OrderID:     "12345678-1234-1234-1234-1234567890ab",
			OrderStatus: "paid",
			UsedAt:      &usedAt,
		},
	}

	ticketID := "3f8b2b6e-1c2d-4e5f-8a9b-0c1d2e3f4a5b"
	qrToken, err := signer.Sign(ticketID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	rec := httptest.NewRecorder()
	scanHandler(st, signer).ServeHTTP(rec, scanRequest(t, testScannerToken, `{"qr_token":"`+qrToken+`"}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}

	sum := sha256.Sum256([]byte(testScannerToken))
	if st.seenHash != hex.EncodeToString(sum[:]) {
		t.Errorf("lookup hash = %q, want the SHA-256 of the header token", st.seenHash)
	}

	var body struct {
		Result string `json:"result"`
		Ticket struct {
			TicketID       string  `json:"ticket_id"`
			ShowID         string  `json:"show_id"`
			Description    string  `json:"description"`
			OrderReference string  `json:"order_reference"`
			Status         string  `json:"status"`
			UsedAt         string  `json:"used_at"`
			SeatLabel      *string `json:"seat_label"`
		} `json:"ticket"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Result != "admitted" {
		t.Errorf("result = %q", body.Result)
	}
	if body.Ticket.TicketID != ticketID || body.Ticket.ShowID != "show-1" {
		t.Errorf("ticket = %+v", body.Ticket)
	}
	if body.Ticket.Description != "Seat 12 (Front)" {
		t.Errorf("description = %q", body.Ticket.Description)
	}
	if body.Ticket.OrderReference != "7890AB" || body.Ticket.Status != "used" {
		t.Errorf("order reference/status = %q / %q", body.Ticket.OrderReference, body.Ticket.Status)
	}
	if body.Ticket.UsedAt != "2026-12-01T20:05:00Z" {
		t.Errorf("used_at = %q", body.Ticket.UsedAt)
	}
}

func TestScanAdmitErrorMapping(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	validQR, err := signer.Sign("3f8b2b6e-1c2d-4e5f-8a9b-0c1d2e3f4a5b")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	tests := []struct {
		name       string
		header     string
		body       string
		store      *fakeScanStore
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing scanner token",
			header:     "",
			body:       `{"qr_token":"` + validQR + `"}`,
			store:      &fakeScanStore{resolvedShow: "show-1"},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "invalid_scanner_token",
		},
		{
			name:       "unknown scanner token",
			header:     "unknown",
			body:       `{"qr_token":"` + validQR + `"}`,
			store:      &fakeScanStore{resolveErr: store.ErrScannerTokenNotFound},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "invalid_scanner_token",
		},
		{
			name:       "bad qr token",
			header:     testScannerToken,
			body:       `{"qr_token":"not-a-token"}`,
			store:      &fakeScanStore{resolvedShow: "show-1"},
			wantStatus: http.StatusNotFound,
			wantCode:   "ticket_not_found",
		},
		{
			name:       "missing qr token",
			header:     testScannerToken,
			body:       `{}`,
			store:      &fakeScanStore{resolvedShow: "show-1"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
		},
		{
			name:       "wrong show",
			header:     testScannerToken,
			body:       `{"qr_token":"` + validQR + `"}`,
			store:      &fakeScanStore{resolvedShow: "show-1", admitErr: store.ErrTicketWrongShow},
			wantStatus: http.StatusForbidden,
			wantCode:   "wrong_show",
		},
		{
			name:       "already used",
			header:     testScannerToken,
			body:       `{"qr_token":"` + validQR + `"}`,
			store:      &fakeScanStore{resolvedShow: "show-1", admitErr: &store.TicketNotIssuedError{Status: "used"}},
			wantStatus: http.StatusConflict,
			wantCode:   "ticket_already_used",
		},
		{
			name:       "void",
			header:     testScannerToken,
			body:       `{"qr_token":"` + validQR + `"}`,
			store:      &fakeScanStore{resolvedShow: "show-1", admitErr: &store.TicketNotIssuedError{Status: "void"}},
			wantStatus: http.StatusConflict,
			wantCode:   "ticket_void",
		},
		{
			name:       "unknown ticket",
			header:     testScannerToken,
			body:       `{"qr_token":"` + validQR + `"}`,
			store:      &fakeScanStore{resolvedShow: "show-1", admitErr: store.ErrTicketNotFound},
			wantStatus: http.StatusNotFound,
			wantCode:   "ticket_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			scanHandler(tt.store, signer).ServeHTTP(rec, scanRequest(t, tt.header, tt.body))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if body := decodeBody(t, rec.Body.Bytes()); body["error"] != tt.wantCode {
				t.Errorf("error = %q, want %q", body["error"], tt.wantCode)
			}
		})
	}
}

func TestScanAdmitRejectsInvalidJSON(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	st := &fakeScanStore{resolvedShow: "show-1"}

	rec := httptest.NewRecorder()
	scanHandler(st, signer).ServeHTTP(
		rec,
		scanRequest(t, testScannerToken, `{"qr_token":"x","extra":true}`),
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestInternalScannerTokenRoutes(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	st := &fakeScanStore{}
	api := &internalAPI{
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		scanner: service.NewScanner(st, signer),
	}

	issue := httptest.NewRequest(http.MethodPost, "/shows/show-1/scanner-token", nil)
	issue.SetPathValue("id", "show-1")
	rec := httptest.NewRecorder()
	api.issueScannerToken(rec, issue)

	if rec.Code != http.StatusCreated {
		t.Fatalf("issue status = %d, want 201 (%s)", rec.Code, rec.Body.String())
	}

	var body struct {
		ShowID       string `json:"show_id"`
		ScannerToken string `json:"scanner_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.ShowID != "show-1" || len(body.ScannerToken) < 40 {
		t.Fatalf("body = %+v", body)
	}
	if st.issuedShow != "show-1" || st.seenHash == "" || st.seenHash == body.ScannerToken {
		t.Errorf("stored show/hash = %q / %q", st.issuedShow, st.seenHash)
	}

	revoke := httptest.NewRequest(http.MethodDelete, "/shows/show-1/scanner-token", nil)
	revoke.SetPathValue("id", "show-1")
	rec = httptest.NewRecorder()
	api.revokeScannerToken(rec, revoke)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("revoke status = %d, want 204", rec.Code)
	}
	if st.revokedShow != "show-1" {
		t.Errorf("revoked show = %q", st.revokedShow)
	}
}
