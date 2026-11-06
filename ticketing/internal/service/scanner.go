package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

var (
	// ErrInvalidScannerToken is returned when a scanner token is missing or
	// does not belong to any show.
	ErrInvalidScannerToken = errors.New("service: invalid scanner token")
	// ErrTicketWrongShow is returned when a valid ticket belongs to another show.
	ErrTicketWrongShow = errors.New("service: ticket belongs to a different show")
	// ErrTicketNotIssued is returned when a ticket is already used or void.
	ErrTicketNotIssued = errors.New("service: ticket is not issued")
)

// TicketNotIssuedError describes a ticket that can no longer be admitted.
type TicketNotIssuedError struct {
	Status string
}

func (e *TicketNotIssuedError) Error() string {
	return fmt.Sprintf("service: ticket is %s", e.Status)
}

// Is lets errors.Is(err, ErrTicketNotIssued) match the typed error.
func (e *TicketNotIssuedError) Is(target error) bool {
	return target == ErrTicketNotIssued
}

// ScannerStore is the persistence the admit flow needs.
type ScannerStore interface {
	SetShowScannerToken(ctx context.Context, showID, tokenHash string) error
	RevokeShowScannerToken(ctx context.Context, showID string) error
	ShowIDForScannerToken(ctx context.Context, tokenHash string) (string, error)
	AdmitTicket(ctx context.Context, showID, ticketID string) (store.TicketInfo, error)
}

// Scanner implements per-show scanner credentials and ticket admission.
type Scanner struct {
	store  ScannerStore
	signer *qr.Signer
}

// NewScanner builds the scanner service.
func NewScanner(st ScannerStore, signer *qr.Signer) *Scanner {
	return &Scanner{store: st, signer: signer}
}

// ScannerToken is the one-time response when a scanner credential is created.
type ScannerToken struct {
	ShowID       string `json:"show_id"`
	ScannerToken string `json:"scanner_token"`
}

// IssueToken creates or rotates a show's scanner credential. The plaintext is
// returned once; only its hash is stored.
func (s *Scanner) IssueToken(ctx context.Context, showID string) (ScannerToken, error) {
	showID = strings.TrimSpace(showID)
	if showID == "" {
		return ScannerToken{}, domain.Invalid("show_id", "is required")
	}
	if len(showID) > maxIDLength {
		return ScannerToken{}, domain.Invalid("show_id", "must be at most %d characters", maxIDLength)
	}

	token, err := newSecretToken()
	if err != nil {
		return ScannerToken{}, err
	}
	if err := s.store.SetShowScannerToken(ctx, showID, hashSecretToken(token)); err != nil {
		return ScannerToken{}, err
	}

	return ScannerToken{ShowID: showID, ScannerToken: token}, nil
}

// RevokeToken removes a show's scanner credential.
func (s *Scanner) RevokeToken(ctx context.Context, showID string) error {
	showID = strings.TrimSpace(showID)
	if showID == "" {
		return domain.Invalid("show_id", "is required")
	}

	return s.store.RevokeShowScannerToken(ctx, showID)
}

// AdmittedTicket is the scanner-facing view of a ticket that was just used.
type AdmittedTicket struct {
	TicketID       string  `json:"ticket_id"`
	ShowID         string  `json:"show_id"`
	SeatCode       string  `json:"seat_code"`
	SeatLabel      *string `json:"seat_label"`
	TierName       string  `json:"tier_name"`
	Description    string  `json:"description"`
	OrderReference string  `json:"order_reference"`
	Status         string  `json:"status"`
	UsedAt         string  `json:"used_at"`
}

// AdmitResult is the successful admit response.
type AdmitResult struct {
	Result string         `json:"result"`
	Ticket AdmittedTicket `json:"ticket"`
}

// Admit verifies the per-show scanner token and the ticket QR token, then
// atomically marks the ticket used.
func (s *Scanner) Admit(ctx context.Context, scannerToken, qrToken string) (AdmitResult, error) {
	scannerToken = strings.TrimSpace(scannerToken)
	qrToken = strings.TrimSpace(qrToken)

	if scannerToken == "" {
		return AdmitResult{}, ErrInvalidScannerToken
	}
	if qrToken == "" {
		return AdmitResult{}, domain.Invalid("qr_token", "is required")
	}

	showID, err := s.store.ShowIDForScannerToken(ctx, hashSecretToken(scannerToken))
	if errors.Is(err, store.ErrScannerTokenNotFound) {
		return AdmitResult{}, ErrInvalidScannerToken
	}
	if err != nil {
		return AdmitResult{}, err
	}

	ticketID, err := s.signer.Verify(qrToken)
	if err != nil {
		return AdmitResult{}, ErrInvalidTicketToken
	}

	ticket, err := s.store.AdmitTicket(ctx, showID, ticketID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrTicketNotFound):
			return AdmitResult{}, ErrInvalidTicketToken
		case errors.Is(err, store.ErrTicketWrongShow):
			return AdmitResult{}, ErrTicketWrongShow
		case errors.Is(err, store.ErrTicketNotIssued):
			status := ""
			var notIssued *store.TicketNotIssuedError
			if errors.As(err, &notIssued) {
				status = notIssued.Status
			}
			return AdmitResult{}, &TicketNotIssuedError{Status: status}
		default:
			return AdmitResult{}, err
		}
	}

	label := ""
	if ticket.SeatLabel != nil {
		label = *ticket.SeatLabel
	}
	usedAt := ""
	if ticket.UsedAt != nil {
		usedAt = formatTime(*ticket.UsedAt)
	}

	return AdmitResult{
		Result: "admitted",
		Ticket: AdmittedTicket{
			TicketID:       ticket.ID,
			ShowID:         ticket.ShowID,
			SeatCode:       ticket.SeatCode,
			SeatLabel:      ticket.SeatLabel,
			TierName:       ticket.TierName,
			Description:    domain.TicketDescription(label, ticket.TierName),
			OrderReference: shortOrderReference(ticket.OrderID),
			Status:         ticket.Status,
			UsedAt:         usedAt,
		},
	}, nil
}
