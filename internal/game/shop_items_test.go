package game

import (
	"testing"
	"time"
)

func TestRegenMood(t *testing.T) {
	cases := []struct{ mood, hours, want int }{
		{50, 0, 50}, {50, 2, 60}, {95, 3, 100}, {0, 100, 100},
	}
	for _, tc := range cases {
		if got := RegenMood(tc.mood, tc.hours); got != tc.want {
			t.Errorf("RegenMood(%d, %d) = %d, want %d", tc.mood, tc.hours, got, tc.want)
		}
	}
}

func TestEaseExhaustion(t *testing.T) {
	cases := []struct{ delta, bonus, want int }{
		{-20, 0, -20}, {-20, 10, -18}, {-20, 90, -10}, // capped at 50 %
		{5, 10, 5}, // gains are untouched
	}
	for _, tc := range cases {
		if got := EaseExhaustion(tc.delta, tc.bonus); got != tc.want {
			t.Errorf("EaseExhaustion(%d, %d) = %d, want %d", tc.delta, tc.bonus, got, tc.want)
		}
	}
}

func TestWholeHours(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if got := WholeHours(base, base.Add(150*time.Minute)); got != 2 {
		t.Errorf("got %d, want 2", got)
	}
	if got := WholeHours(base, base.Add(-time.Hour)); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}
