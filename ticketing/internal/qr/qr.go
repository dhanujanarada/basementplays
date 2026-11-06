// Package qr signs and verifies ticket QR tokens.
package qr

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const tokenVersion = "v1"

// ErrInvalidToken is returned when a token is malformed or its signature does
// not match.
var ErrInvalidToken = errors.New("qr: invalid ticket token")

// Signer issues and verifies ticket tokens with HMAC-SHA256.
type Signer struct {
	secret    []byte
	publicURL string
}

// NewSigner builds a signer. publicBaseURL is the site origin used to build
// the URL encoded in the QR code, for example https://basementplays.com.
func NewSigner(secret, publicBaseURL string) *Signer {
	return &Signer{
		secret:    []byte(secret),
		publicURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

// Sign returns v1.<ticket_uuid>.<base64url(hmac)> for a ticket id.
func (s *Signer) Sign(ticketID string) (string, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" || strings.Contains(ticketID, ".") {
		return "", fmt.Errorf("qr: invalid ticket id %q", ticketID)
	}

	return tokenVersion + "." + ticketID + "." + s.signature(ticketID), nil
}

// Verify checks the version and signature, then returns the ticket id.
func (s *Signer) Verify(token string) (string, error) {
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != tokenVersion || parts[1] == "" {
		return "", ErrInvalidToken
	}

	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", ErrInvalidToken
	}

	want, err := base64.RawURLEncoding.DecodeString(s.signature(parts[1]))
	if err != nil {
		return "", ErrInvalidToken
	}
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

// URL returns the public ticket-page URL encoded in the QR code.
func (s *Signer) URL(token string) string {
	return s.publicURL + "/t/" + token
}

func (s *Signer) signature(ticketID string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(tokenVersion + "|" + ticketID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
