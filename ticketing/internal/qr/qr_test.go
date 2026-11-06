package qr

import (
	"errors"
	"strings"
	"testing"
)

const testTicketID = "3f8b2b6e-1c2d-4e5f-8a9b-0c1d2e3f4a5b"

func TestSignAndVerify(t *testing.T) {
	signer := NewSigner("test-secret", "https://basementplays.com/")

	token, err := signer.Sign(testTicketID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if !strings.HasPrefix(token, "v1."+testTicketID+".") {
		t.Fatalf("token = %q", token)
	}
	if strings.Contains(token, "=") {
		t.Fatalf("token contains base64 padding: %q", token)
	}

	got, err := signer.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got != testTicketID {
		t.Fatalf("ticket id = %q, want %q", got, testTicketID)
	}

	if url := signer.URL(token); url != "https://basementplays.com/t/"+token {
		t.Fatalf("URL = %q", url)
	}
}

func TestVerifyRejectsTamperedTokens(t *testing.T) {
	signer := NewSigner("test-secret", "https://example.com")
	token, err := signer.Sign(testTicketID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	parts := strings.Split(token, ".")
	tests := map[string]string{
		"empty":          "",
		"one part":       parts[0],
		"two parts":      parts[0] + "." + parts[1],
		"four parts":     token + ".extra",
		"wrong version":  "v2." + parts[1] + "." + parts[2],
		"empty ticket":   "v1.." + parts[2],
		"ticket changed": parts[0] + ".11111111-1111-1111-1111-111111111111." + parts[2],
		"signature":      parts[0] + "." + parts[1] + "." + strings.Repeat("A", 43),
		"bad base64":     parts[0] + "." + parts[1] + ".!!!",
	}

	for name, bad := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := signer.Verify(bad); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("Verify(%q) = %v, want ErrInvalidToken", bad, err)
			}
		})
	}
}

func TestVerifyRejectsDifferentSecret(t *testing.T) {
	token, err := NewSigner("secret-one", "").Sign(testTicketID)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	if _, err := NewSigner("secret-two", "").Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify with another secret = %v, want ErrInvalidToken", err)
	}
}

func TestSignRejectsEmptyOrDottedID(t *testing.T) {
	signer := NewSigner("secret", "")

	for _, id := range []string{"", "  ", "ticket.1"} {
		if _, err := signer.Sign(id); err == nil {
			t.Errorf("Sign(%q) = nil", id)
		}
	}
}
