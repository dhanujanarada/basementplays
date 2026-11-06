package domain

import (
	"fmt"
	"strings"
)

// TicketDescription renders the public wording for a seat on a ticket.
// A non-empty label wins: "Seat 12 (Front)". Otherwise the tier is used:
// "Front admission".
func TicketDescription(label, tierName string) string {
	if trimmed := strings.TrimSpace(label); trimmed != "" {
		return fmt.Sprintf("Seat %s (%s)", trimmed, tierName)
	}

	return tierName + " admission"
}
