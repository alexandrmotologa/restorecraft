package scheduler

import (
	"testing"
	"time"
)

func TestNextCronTime(t *testing.T) {
	from := time.Date(2026, 11, 6, 2, 0, 0, 0, time.UTC)

	// Test "0 3 * * *" -> should be 2026-11-06 03:00:00
	next, err := NextCronTime("0 3 * * *", from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2026, 11, 6, 3, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}

	// Test "@daily" from 23:30 -> next day 00:00
	fromLate := time.Date(2026, 11, 6, 23, 30, 0, 0, time.UTC)
	nextDaily, err := NextCronTime("@daily", fromLate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedDaily := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
	if !nextDaily.Equal(expectedDaily) {
		t.Fatalf("expected %v, got %v", expectedDaily, nextDaily)
	}
}
