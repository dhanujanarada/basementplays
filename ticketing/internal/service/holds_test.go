package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
)

func TestNormalizeSeats(t *testing.T) {
	seats, err := normalizeSeats([]string{" S003 ", "S001", "S002"}, 6)
	if err != nil {
		t.Fatalf("normalizeSeats: %v", err)
	}

	want := []string{"S001", "S002", "S003"}
	for i := range want {
		if seats[i] != want[i] {
			t.Errorf("seat %d = %q, want %q", i, seats[i], want[i])
		}
	}

	tests := []struct {
		name  string
		seats []string
		field string
	}{
		{
			name:  "none",
			seats: nil,
			field: "seats",
		},
		{
			name:  "blank",
			seats: []string{"S001", "  "},
			field: "seats[1]",
		},
		{
			name:  "too long",
			seats: []string{strings.Repeat("S", maxSeatCodeLength+1)},
			field: "seats[0]",
		},
		{
			name:  "duplicate",
			seats: []string{"S001", "S001"},
			field: "seats[1]",
		},
		{
			name:  "too many",
			seats: []string{"S001", "S002", "S003"},
			field: "seats",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeSeats(tt.seats, 2)
			if err == nil {
				t.Fatal("normalizeSeats() = nil")
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
}

func TestNormalizeIdempotencyKey(t *testing.T) {
	key, err := normalizeIdempotencyKey("  key-123  ")
	if err != nil {
		t.Fatalf("normalizeIdempotencyKey: %v", err)
	}
	if key != "key-123" {
		t.Fatalf("key = %q, want key-123", key)
	}

	empty, err := normalizeIdempotencyKey("   ")
	if err != nil || empty != "" {
		t.Fatalf("blank key = %q, %v; want empty, nil", empty, err)
	}

	for _, bad := range []string{
		strings.Repeat("k", maxIdempotencyKeyLength+1),
		"key\ninjection",
	} {
		if _, err := normalizeIdempotencyKey(bad); err == nil {
			t.Errorf("normalizeIdempotencyKey(%q) = nil", bad)
		}
	}
}

func TestOrderToken(t *testing.T) {
	first, err := newSecretToken()
	if err != nil {
		t.Fatalf("newSecretToken: %v", err)
	}
	second, err := newSecretToken()
	if err != nil {
		t.Fatalf("newSecretToken: %v", err)
	}

	if first == second {
		t.Fatal("two order tokens are identical")
	}
	if len(first) < 40 {
		t.Fatalf("token length = %d, want at least 40", len(first))
	}

	want := sha256.Sum256([]byte("known-token"))
	if got := hashSecretToken("known-token"); got != hex.EncodeToString(want[:]) {
		t.Fatalf("hashSecretToken = %q", got)
	}
	if hashSecretToken(first) == hashSecretToken(second) {
		t.Fatal("distinct tokens hash to the same value")
	}
}

func TestHoldExpiry(t *testing.T) {
	before := time.Now()
	expires := holdExpiry(10)
	after := time.Now()

	if expires.Before(before.Add(9*time.Minute)) || expires.After(after.Add(11*time.Minute)) {
		t.Fatalf("holdExpiry(10) = %v, want about 10 minutes", expires)
	}

	if got := holdExpiry(0); got.Before(before.Add(9 * time.Minute)) {
		t.Fatalf("holdExpiry(0) = %v, want the default TTL", got)
	}
}
