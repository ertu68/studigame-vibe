package game

import (
	"testing"
	"time"
)

func TestBafogPeriodsElapsed(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		now  time.Time
		want int
	}{
		{"before first payout", base.Add(5 * time.Hour), 0},
		{"exactly one period", base.Add(6 * time.Hour), 1},
		{"just under two", base.Add(11*time.Hour + 59*time.Minute), 1},
		{"two full periods", base.Add(12 * time.Hour), 2},
		{"clamped to max", base.Add(30 * 24 * time.Hour), MaxBafogPeriods},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BafogPeriodsElapsed(base, tc.now); got != tc.want {
				t.Fatalf("BafogPeriodsElapsed(%v) = %d, want %d", tc.now, got, tc.want)
			}
		})
	}
}

func TestBafogNextAfter(t *testing.T) {
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	next := BafogNextAfter(base, 3)
	want := base.Add(18 * time.Hour)
	if !next.Equal(want) {
		t.Fatalf("BafogNextAfter = %v, want %v", next, want)
	}
}
