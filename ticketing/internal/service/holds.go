package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const (
	maxSeatCodeLength       = 64
	maxIdempotencyKeyLength = 255
	maxOrderIDLength        = 64
)

// Hold is the public response for a newly created hold. OrderToken is only
// returned once, when the hold is created; it is never stored in plaintext.
type Hold struct {
	ID         string            `json:"order_id"`
	ShowID     string            `json:"show_id"`
	Status     string            `json:"status"`
	Source     string            `json:"source"`
	TotalCents int               `json:"total_cents"`
	Currency   string            `json:"currency"`
	ExpiresAt  string            `json:"expires_at"`
	Seats      []store.OrderSeat `json:"seats"`
	OrderToken string            `json:"order_token,omitempty"`
}

// Holds implements public hold and release use cases.
type Holds struct {
	store    *store.Store
	holdTTL  int // minutes
	maxSeats int
}

// NewHolds builds the hold service. holdMinutes is HOLD_MINUTES and maxSeats is
// MAX_SEATS_PER_ORDER.
func NewHolds(st *store.Store, holdMinutes, maxSeats int) *Holds {
	return &Holds{store: st, holdTTL: holdMinutes, maxSeats: maxSeats}
}

// Create claims seats and returns a pending order. The boolean reports an
// idempotent replay; in that case OrderToken is empty.
func (h *Holds) Create(
	ctx context.Context,
	showID string,
	seatCodes []string,
	idempotencyKey string,
) (Hold, bool, error) {
	showID = strings.TrimSpace(showID)
	if showID == "" {
		return Hold{}, false, domain.Invalid("show_id", "is required")
	}
	if len(showID) > 200 {
		return Hold{}, false, domain.Invalid("show_id", "must be at most 200 characters")
	}

	seats, err := normalizeSeats(seatCodes, h.maxSeats)
	if err != nil {
		return Hold{}, false, err
	}

	key, err := normalizeIdempotencyKey(idempotencyKey)
	if err != nil {
		return Hold{}, false, err
	}

	token, err := newSecretToken()
	if err != nil {
		return Hold{}, false, err
	}

	order, replayed, err := h.store.CreateHold(ctx, store.HoldInput{
		ShowID:          showID,
		SeatCodes:       seats,
		ExpiresAt:       holdExpiry(h.holdTTL),
		IdempotencyKey:  key,
		AccessTokenHash: hashSecretToken(token),
	})
	if err != nil {
		return Hold{}, false, err
	}

	hold := Hold{
		ID:         order.ID,
		ShowID:     order.ShowID,
		Status:     order.Status,
		Source:     order.Source,
		TotalCents: order.TotalCents,
		Currency:   order.Currency,
		ExpiresAt:  order.ExpiresAt.Format(timeFormat),
		Seats:      order.Seats,
	}
	if !replayed {
		hold.OrderToken = token
	}

	return hold, replayed, nil
}

// Release cancels a pending hold after checking the buyer's order token.
func (h *Holds) Release(ctx context.Context, orderID, orderToken string) error {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return domain.Invalid("order_id", "is required")
	}
	if len(orderID) > maxOrderIDLength {
		return domain.Invalid("order_id", "is invalid")
	}

	orderToken = strings.TrimSpace(orderToken)
	if orderToken == "" {
		return domain.Invalid("order_token", "is required")
	}

	return h.store.ReleaseHold(ctx, orderID, hashSecretToken(orderToken))
}

func normalizeSeats(codes []string, maxSeats int) ([]string, error) {
	if len(codes) == 0 {
		return nil, domain.Invalid("seats", "at least one seat is required")
	}

	seen := make(map[string]bool, len(codes))
	seats := make([]string, 0, len(codes))

	for i, raw := range codes {
		code := strings.TrimSpace(raw)
		field := fmt.Sprintf("seats[%d]", i)

		if code == "" {
			return nil, domain.Invalid(field, "must not be blank")
		}
		if len(code) > maxSeatCodeLength {
			return nil, domain.Invalid(field, "must be at most %d characters", maxSeatCodeLength)
		}
		if !utf8.ValidString(code) {
			return nil, domain.Invalid(field, "must be valid UTF-8")
		}
		if seen[code] {
			return nil, domain.Invalid(field, "duplicate seat code %q", code)
		}

		seen[code] = true
		seats = append(seats, code)
	}

	if len(seats) > maxSeats {
		return nil, domain.Invalid("seats", "at most %d seats are allowed per order", maxSeats)
	}

	sort.Strings(seats)

	return seats, nil
}

func normalizeIdempotencyKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", nil
	}
	if len(key) > maxIdempotencyKeyLength {
		return "", domain.Invalid(
			"Idempotency-Key",
			"must be at most %d characters",
			maxIdempotencyKeyLength,
		)
	}
	for _, r := range key {
		if unicode.IsControl(r) {
			return "", domain.Invalid("Idempotency-Key", "must not contain control characters")
		}
	}

	return key, nil
}

func holdExpiry(minutes int) time.Time {
	if minutes <= 0 {
		minutes = 10
	}

	return time.Now().Add(time.Duration(minutes) * time.Minute)
}

// newSecretToken returns a 256-bit URL-safe opaque token. Order access tokens
// and per-show scanner tokens share this format.
func newSecretToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate secret token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

// hashSecretToken returns the lowercase hex SHA-256 of a secret token. Only the
// hash is persisted.
func hashSecretToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
