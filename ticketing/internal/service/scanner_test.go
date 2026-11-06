package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

type fakeScannerStore struct {
	resolvedShow string
	resolveErr   error
	admitted     store.TicketInfo
	admitErr     error

	seenHash    string
	issuedShow  string
	revokedShow string
	setErr      error
}

func (f *fakeScannerStore) SetShowScannerToken(_ context.Context, showID, tokenHash string) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.issuedShow = showID
	f.seenHash = tokenHash
	return nil
}

func (f *fakeScannerStore) RevokeShowScannerToken(_ context.Context, showID string) error {
	f.revokedShow = showID
	return f.setErr
}

func (f *fakeScannerStore) ShowIDForScannerToken(_ context.Context, tokenHash string) (string, error) {
	f.seenHash = tokenHash
	return f.resolvedShow, f.resolveErr
}

func (f *fakeScannerStore) AdmitTicket(_ context.Context, showID, ticketID string) (store.TicketInfo, error) {
	if f.admitErr != nil {
		return store.TicketInfo{}, f.admitErr
	}

	ticket := f.admitted
	ticket.ShowID = showID
	ticket.ID = ticketID
	return ticket, nil
}

func sampleAdmittedTicket() store.TicketInfo {
	label := "12"
	usedAt := time.Date(2026, 12, 1, 20, 5, 0, 0, time.UTC)

	return store.TicketInfo{
		ID:          "3f8b2b6e-1c2d-4e5f-8a9b-0c1d2e3f4a5b",
		Status:      "used",
		ShowID:      "show-1",
		SeatCode:    "S012",
		SeatLabel:   &label,
		TierName:    "Front",
		OrderID:     "12345678-1234-1234-1234-1234567890ab",
		OrderStatus: "paid",
		UsedAt:      &usedAt,
	}
}

func TestScannerAdmit(t *testing.T) {
	signer := qr.NewSigner("secret", "https://basementplays.test")
	info := sampleAdmittedTicket()
	st := &fakeScannerStore{resolvedShow: "show-1", admitted: info}
	scanner := NewScanner(st, signer)

	qrToken, err := signer.Sign(info.ID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	result, err := scanner.Admit(context.Background(), "scanner-token", qrToken)
	if err != nil {
		t.Fatalf("Admit: %v", err)
	}
	if st.seenHash != hashSecretToken("scanner-token") {
		t.Errorf("scanner token was not hashed before lookup")
	}
	if result.Result != "admitted" {
		t.Errorf("result = %q, want admitted", result.Result)
	}
	if result.Ticket.TicketID != info.ID || result.Ticket.Status != "used" {
		t.Errorf("ticket = %+v", result.Ticket)
	}
	if result.Ticket.Description != "Seat 12 (Front)" {
		t.Errorf("description = %q", result.Ticket.Description)
	}
	if result.Ticket.OrderReference != "7890AB" {
		t.Errorf("order reference = %q", result.Ticket.OrderReference)
	}
	if result.Ticket.UsedAt != "2026-12-01T20:05:00Z" {
		t.Errorf("used_at = %q", result.Ticket.UsedAt)
	}
}

func TestScannerRejectsBadCredentialsAndTokens(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	ticketID := sampleAdmittedTicket().ID

	validQR, err := signer.Sign(ticketID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	tests := []struct {
		name           string
		scannerToken   string
		qrToken        string
		store          *fakeScannerStore
		want           error
		wantValidation bool
	}{
		{
			name:         "missing scanner token",
			scannerToken: "",
			qrToken:      validQR,
			store:        &fakeScannerStore{resolvedShow: "show-1"},
			want:         ErrInvalidScannerToken,
		},
		{
			name:         "unknown scanner token",
			scannerToken: "unknown",
			qrToken:      validQR,
			store:        &fakeScannerStore{resolveErr: store.ErrScannerTokenNotFound},
			want:         ErrInvalidScannerToken,
		},
		{
			name:         "bad qr token",
			scannerToken: "scanner-token",
			qrToken:      "not-a-token",
			store:        &fakeScannerStore{resolvedShow: "show-1"},
			want:         ErrInvalidTicketToken,
		},
		{
			name:           "missing qr token",
			scannerToken:   "scanner-token",
			qrToken:        "",
			store:          &fakeScannerStore{resolvedShow: "show-1"},
			wantValidation: true,
		},
		{
			name:         "unknown ticket",
			scannerToken: "scanner-token",
			qrToken:      validQR,
			store: &fakeScannerStore{
				resolvedShow: "show-1",
				admitErr:     store.ErrTicketNotFound,
			},
			want: ErrInvalidTicketToken,
		},
		{
			name:         "wrong show",
			scannerToken: "scanner-token",
			qrToken:      validQR,
			store: &fakeScannerStore{
				resolvedShow: "show-1",
				admitErr:     store.ErrTicketWrongShow,
			},
			want: ErrTicketWrongShow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewScanner(tt.store, signer).Admit(context.Background(), tt.scannerToken, tt.qrToken)
			if tt.wantValidation {
				var validation *domain.ValidationError
				if !errors.As(err, &validation) {
					t.Fatalf("Admit = %v, want a validation error", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("Admit = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestScannerRejectsUsedAndVoidTickets(t *testing.T) {
	signer := qr.NewSigner("secret", "")
	st := &fakeScannerStore{
		resolvedShow: "show-1",
		admitErr:     &store.TicketNotIssuedError{Status: "used"},
	}
	scanner := NewScanner(st, signer)

	token, err := signer.Sign(sampleAdmittedTicket().ID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	_, err = scanner.Admit(context.Background(), "scanner-token", token)
	if !errors.Is(err, ErrTicketNotIssued) {
		t.Fatalf("Admit = %v, want ErrTicketNotIssued", err)
	}
	var notIssued *TicketNotIssuedError
	if !errors.As(err, &notIssued) || notIssued.Status != "used" {
		t.Fatalf("status = %+v, want used", notIssued)
	}

	st.admitErr = &store.TicketNotIssuedError{Status: "void"}
	_, err = scanner.Admit(context.Background(), "scanner-token", token)
	if !errors.Is(err, ErrTicketNotIssued) {
		t.Fatalf("Admit(void) = %v, want ErrTicketNotIssued", err)
	}
}

func TestScannerIssueAndRevokeToken(t *testing.T) {
	st := &fakeScannerStore{}
	scanner := NewScanner(st, qr.NewSigner("secret", ""))

	issued, err := scanner.IssueToken(context.Background(), "show-1")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if issued.ShowID != "show-1" || len(issued.ScannerToken) < 40 {
		t.Fatalf("token = %+v", issued)
	}
	if st.issuedShow != "show-1" || st.seenHash != hashSecretToken(issued.ScannerToken) {
		t.Errorf("stored show/hash = %q / %q", st.issuedShow, st.seenHash)
	}
	if st.seenHash == issued.ScannerToken {
		t.Error("scanner token was stored in plaintext")
	}

	second, err := scanner.IssueToken(context.Background(), "show-1")
	if err != nil {
		t.Fatalf("IssueToken(rotate): %v", err)
	}
	if second.ScannerToken == issued.ScannerToken {
		t.Error("rotation returned the same token")
	}

	if err := scanner.RevokeToken(context.Background(), "show-1"); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if st.revokedShow != "show-1" {
		t.Errorf("revoked show = %q", st.revokedShow)
	}

	if _, err := scanner.IssueToken(context.Background(), "  "); err == nil {
		t.Error("IssueToken with a blank show id allowed")
	}
	if err := scanner.RevokeToken(context.Background(), " "); err == nil {
		t.Error("RevokeToken with a blank show id allowed")
	}
}

func TestScannerIssueTokenUnknownShow(t *testing.T) {
	st := &fakeScannerStore{setErr: store.ErrShowNotFound}
	scanner := NewScanner(st, qr.NewSigner("secret", ""))

	if _, err := scanner.IssueToken(context.Background(), "missing"); !errors.Is(err, store.ErrShowNotFound) {
		t.Fatalf("IssueToken = %v, want ErrShowNotFound", err)
	}
}
