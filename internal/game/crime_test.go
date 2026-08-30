package game

import (
	"testing"
	"time"
)

func TestResolveCrimeCleanPayout(t *testing.T) {
	// Roll above the catch chance pays the full gross: 1 h at 15 €/h, rested.
	out := ResolveCrime(CrimeParams{Duration: time.Hour, Mood: 100}, 0.5)
	if out.EarningsCents != CrimeRatePerHourCents {
		t.Fatalf("earnings: got %d, want %d", out.EarningsCents, CrimeRatePerHourCents)
	}
	if out.Caught {
		t.Fatal("clean run must not be flagged caught")
	}
	if out.MoodDelta != -5 { // moodCost(1h)
		t.Fatalf("mood delta: got %d, want -5", out.MoodDelta)
	}
}

func TestResolveCrimeCaught(t *testing.T) {
	// Roll below the catch chance: a fraction of the gross and double mood cost.
	out := ResolveCrime(CrimeParams{Duration: time.Hour, Mood: 100}, 0.1)
	if !out.Caught {
		t.Fatal("roll below catch chance must be caught")
	}
	want := int(float64(CrimeRatePerHourCents) * CrimeCatchPayoutFactor)
	if out.EarningsCents != want {
		t.Fatalf("caught earnings: got %d, want %d", out.EarningsCents, want)
	}
	if out.MoodDelta != -10 { // 2 * moodCost(1h)
		t.Fatalf("caught mood delta: got %d, want -10", out.MoodDelta)
	}
}

func TestResolveCrimeCatchBoundary(t *testing.T) {
	// Roll exactly at the catch chance still counts as caught.
	if !ResolveCrime(CrimeParams{Duration: time.Hour, Mood: 100}, CrimeCatchChance).Caught {
		t.Fatal("roll at the catch chance must count as caught")
	}
}

func TestResolveCrimeMoodScales(t *testing.T) {
	// Exhausted characters pull a smaller gross, like the other income actions.
	low := ResolveCrime(CrimeParams{Duration: time.Hour, Mood: 0}, 0.5)
	high := ResolveCrime(CrimeParams{Duration: time.Hour, Mood: 100}, 0.5)
	if low.EarningsCents >= high.EarningsCents {
		t.Fatalf("exhausted must earn less: low=%d high=%d", low.EarningsCents, high.EarningsCents)
	}
}
