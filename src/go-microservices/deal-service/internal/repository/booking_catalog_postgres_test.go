package repository

import "testing"

func TestHasCapacity(t *testing.T) {
	tests := []struct {
		name              string
		totalPlaces       int
		overlappingGuests int
		requestedGuests   int
		want              bool
	}{
		{name: "room remains for request", totalPlaces: 5, overlappingGuests: 2, requestedGuests: 2, want: true},
		{name: "request fills capacity exactly", totalPlaces: 5, overlappingGuests: 3, requestedGuests: 2, want: true},
		{name: "request exceeds capacity", totalPlaces: 5, overlappingGuests: 4, requestedGuests: 2, want: false},
		{name: "invalid request size", totalPlaces: 5, overlappingGuests: 0, requestedGuests: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasCapacity(tt.totalPlaces, tt.overlappingGuests, tt.requestedGuests); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
