package game

import (
	"testing"
	"time"
)

func TestCoffeeBoostActive(t *testing.T) {
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	if CoffeeBoostActive(now.Add(time.Hour), now) != true {
		t.Fatal("boost should be active inside the window")
	}
	if CoffeeBoostActive(now, now) != false {
		t.Fatal("boost expired exactly at the boundary")
	}
	if CoffeeBoostActive(now.Add(-time.Minute), now) != false {
		t.Fatal("boost should be expired after the window")
	}
}

func TestBoostedSkillTrainingDuration(t *testing.T) {
	// Base = 120 min at Konzentration 0, Geisteswissenschaften (factor 1.0).
	// With the boost it should halve to 60.
	if got := BoostedSkillTrainingDuration("geschicklichkeit", 0, "Geisteswissenschaften", true); got != 60 {
		t.Fatalf("boosted duration: got %d, want 60", got)
	}
	if got := BoostedSkillTrainingDuration("geschicklichkeit", 0, "Geisteswissenschaften", false); got != 120 {
		t.Fatalf("unboosted duration: got %d, want 120", got)
	}
}
