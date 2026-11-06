package domain

import "testing"

func TestTicketDescription(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		tierName string
		want     string
	}{
		{
			name:     "labelled seat",
			label:    "12",
			tierName: "Front",
			want:     "Seat 12 (Front)",
		},
		{
			name:     "label is trimmed",
			label:    "  A4  ",
			tierName: "Balcony",
			want:     "Seat A4 (Balcony)",
		},
		{
			name:     "empty label falls back to the tier",
			label:    "",
			tierName: "Standard",
			want:     "Standard admission",
		},
		{
			name:     "blank label falls back to the tier",
			label:    "   ",
			tierName: "Standard",
			want:     "Standard admission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TicketDescription(tt.label, tt.tierName); got != tt.want {
				t.Fatalf("TicketDescription(%q, %q) = %q, want %q", tt.label, tt.tierName, got, tt.want)
			}
		})
	}
}
