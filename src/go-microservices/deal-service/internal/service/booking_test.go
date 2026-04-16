package service

import (
	"testing"
	"time"
)

func TestValidateBookingDateRange(t *testing.T) {
	checkIn := time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC)

	if err := ValidateBookingDateRange(checkIn, checkOut); err != nil {
		t.Fatalf("expected valid range, got error: %v", err)
	}
	if err := ValidateBookingDateRange(checkIn, checkIn); err == nil {
		t.Fatalf("expected error for same start/end date")
	}
	if err := ValidateBookingDateRange(checkOut, checkIn); err == nil {
		t.Fatalf("expected error for reversed range")
	}
}

func TestDatesOverlap(t *testing.T) {
	aStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	aEnd := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)

	overlapStart := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	overlapEnd := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	if !DatesOverlap(aStart, aEnd, overlapStart, overlapEnd) {
		t.Fatalf("expected overlap to be true")
	}

	touchingStart := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	touchingEnd := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	if DatesOverlap(aStart, aEnd, touchingStart, touchingEnd) {
		t.Fatalf("expected touching ranges to be non-overlapping")
	}
}
