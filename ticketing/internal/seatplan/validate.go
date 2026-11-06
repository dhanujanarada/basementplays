package seatplan

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
)

const (
	maxKeyLength  = 100
	MaxNameLength = 200
)

var keyPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// ValidateKey checks a plan key (slug) supplied by the builder.
func ValidateKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return domain.Invalid("key", "is required")
	}
	if len(key) > maxKeyLength {
		return domain.Invalid("key", "must be at most %d characters", maxKeyLength)
	}
	if !keyPattern.MatchString(key) {
		return domain.Invalid("key", "must contain only lowercase letters, digits, and single hyphens")
	}
	return nil
}

// ValidateLayout checks a stored layout for internal consistency. It is used
// before creating show_seats, and again after label edits.
func ValidateLayout(layout Layout) error {
	if layout.Columns < 1 {
		return domain.Invalid("layout.columns", "must be at least 1")
	}
	if !layout.Numbering.Valid() {
		return domain.Invalid("layout.numbering", `must be "none" or "sequential"`)
	}
	if len(layout.Tiers) == 0 {
		return domain.Invalid("layout.tiers", "at least one tier is required")
	}
	if len(layout.Seats) == 0 {
		return domain.Invalid("layout.seats", "at least one seat is required")
	}

	tierByKey := make(map[string]Tier, len(layout.Tiers))
	total := 0

	for i, tier := range layout.Tiers {
		field := "layout.tiers[" + strconv.Itoa(i) + "]"

		key := strings.TrimSpace(tier.Key)
		if key == "" {
			return domain.Invalid(field+".key", "is required")
		}
		if strings.TrimSpace(tier.Name) == "" {
			return domain.Invalid(field+".name", "is required")
		}
		if tier.PriceCents < 0 {
			return domain.Invalid(field+".price_cents", "must not be negative")
		}
		if tier.SeatCount < 0 {
			return domain.Invalid(field+".seat_count", "must not be negative")
		}
		if _, exists := tierByKey[key]; exists {
			return domain.Invalid(field+".key", "duplicate tier key %q", key)
		}

		tierByKey[key] = tier
		total += tier.SeatCount
	}

	if total != len(layout.Seats) {
		return domain.Invalid(
			"layout.seats",
			"has %d rows but the tiers declare %d seats",
			len(layout.Seats), total,
		)
	}

	seenCodes := make(map[string]bool, len(layout.Seats))
	perTier := make(map[string]int, len(layout.Tiers))

	for i, seat := range layout.Seats {
		field := "layout.seats[" + strconv.Itoa(i) + "]"

		code := strings.TrimSpace(seat.Code)
		if code == "" {
			return domain.Invalid(field+".code", "is required")
		}
		if seenCodes[code] {
			return domain.Invalid(field+".code", "duplicate seat code %q", code)
		}
		if _, exists := tierByKey[seat.Tier]; !exists {
			return domain.Invalid(field+".tier", "unknown tier %q", seat.Tier)
		}

		seenCodes[code] = true
		perTier[seat.Tier]++
	}

	for _, tier := range layout.Tiers {
		if perTier[tier.Key] != tier.SeatCount {
			return domain.Invalid(
				"layout.tiers",
				"tier %q declares %d seats but has %d rows",
				tier.Key, tier.SeatCount, perTier[tier.Key],
			)
		}
	}

	return nil
}

// uniqueSlug builds a stable tier key from a display name, suffixing collisions.
func uniqueSlug(name string, used map[string]bool) string {
	base := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if base == "" {
		base = "tier"
	}

	candidate := base
	for i := 2; used[candidate]; i++ {
		candidate = base + "-" + strconv.Itoa(i)
	}
	used[candidate] = true

	return candidate
}
