package game

import (
	"testing"
	"time"
)

func TestValidCollectDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want bool
	}{
		{15 * time.Minute, true},
		{time.Hour, true},
		{8 * time.Hour, true},
		{30 * time.Minute, false},
		{24 * time.Hour, false},
		{0, false},
		{-time.Hour, false},
	}
	for _, tt := range tests {
		if got := ValidCollectDuration(tt.d); got != tt.want {
			t.Errorf("ValidCollectDuration(%v) = %v, want %v", tt.d, got, tt.want)
		}
	}
}

func TestResolveCollect(t *testing.T) {
	base := CollectParams{
		Duration:       time.Hour,
		ContainerLevel: 0,
		Mood:           100,
	}

	tests := []struct {
		name        string
		params      CollectParams
		roll        float64
		wantBottles int
		wantMood    int
	}{
		{"rested, average luck", base, 0.5, 40, -5},
		{"bad luck pays 20 % less", base, 0.0, 32, -5},
		{"good luck pays 20 % more", base, 0.999, 47, -5},
		{
			name:        "exhausted character collects 60 %",
			params:      CollectParams{Duration: time.Hour, Mood: 0},
			roll:        0.5,
			wantBottles: 24,
			wantMood:    -5,
		},
		{
			name:        "container level 4 adds 60 %",
			params:      CollectParams{Duration: time.Hour, ContainerLevel: 4, Mood: 100},
			roll:        0.5,
			wantBottles: 64,
			wantMood:    -5,
		},
		{
			name:        "short shift still costs some mood",
			params:      CollectParams{Duration: 15 * time.Minute, Mood: 100},
			roll:        0.5,
			wantBottles: 10,
			wantMood:    -2,
		},
		{
			name:        "long shift scales linearly",
			params:      CollectParams{Duration: 8 * time.Hour, Mood: 100},
			roll:        0.5,
			wantBottles: 320,
			wantMood:    -40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveCollect(tt.params, tt.roll)
			if got.Bottles != tt.wantBottles {
				t.Errorf("bottles: got %d, want %d", got.Bottles, tt.wantBottles)
			}
			if got.EarningsCents != got.Bottles*centsPerBottle {
				t.Errorf("earnings %d do not match %d bottles", got.EarningsCents, got.Bottles)
			}
			if got.MoodDelta != tt.wantMood {
				t.Errorf("mood delta: got %d, want %d", got.MoodDelta, tt.wantMood)
			}
		})
	}
}

// A roll outside [0,1) must not produce a better result than a maximum roll,
// otherwise a bad random source would turn into an exploit.
func TestResolveCollectClampsRoll(t *testing.T) {
	p := CollectParams{Duration: time.Hour, Mood: 100}
	high := ResolveCollect(p, 5)
	max := ResolveCollect(p, 1)
	if high.Bottles != max.Bottles {
		t.Fatalf("roll above 1 gave %d bottles, want %d", high.Bottles, max.Bottles)
	}
	low := ResolveCollect(p, -5)
	min := ResolveCollect(p, 0)
	if low.Bottles != min.Bottles {
		t.Fatalf("roll below 0 gave %d bottles, want %d", low.Bottles, min.Bottles)
	}
}

func TestApplyMood(t *testing.T) {
	tests := []struct {
		mood, delta, want int
	}{
		{100, -5, 95},
		{3, -10, 0},   // cannot fall below 0
		{98, 10, 100}, // cannot rise above 100
		{0, 0, 0},
	}
	for _, tt := range tests {
		if got := ApplyMood(tt.mood, tt.delta); got != tt.want {
			t.Errorf("ApplyMood(%d, %d) = %d, want %d", tt.mood, tt.delta, got, tt.want)
		}
	}
}

func TestContainerCapacity(t *testing.T) {
	tests := []struct {
		level, want int
	}{
		{0, 2000},
		{1, 5000},
		{2, 20000},
		{3, 100000},
		{4, 500000},
		{5, 0},
		{-1, 0},
	}
	for _, tt := range tests {
		if got := ContainerCapacity(tt.level); got != tt.want {
			t.Errorf("ContainerCapacity(%d) = %d, want %d", tt.level, got, tt.want)
		}
	}
}

// A fresh character must have room left, otherwise the first shift earns nothing.
func TestStartingCapitalFitsInFirstContainer(t *testing.T) {
	const startCapital = 1000
	if ContainerCapacity(0) <= startCapital {
		t.Fatalf("container level 0 holds %d, must exceed the %d start capital",
			ContainerCapacity(0), startCapital)
	}
}

func TestUpgradeCostCents(t *testing.T) {
	const base = 2000 // Mirrors the production base value.
	tests := []struct {
		level, want int
	}{
		{0, 2000},  // 20 €
		{1, 6964},  // ~70 €
		{2, 14449}, // ~144 €
		{3, 24251}, // ~243 €
		{4, 0},     // max: not buyable
		{5, 0},     // beyond max: not buyable
		{-1, 0},    // invalid: not buyable
	}
	for _, tt := range tests {
		if got := UpgradeCostCents(tt.level, base); got != tt.want {
			t.Errorf("UpgradeCostCents(%d, %d) = %d, want %d", tt.level, base, got, tt.want)
		}
	}
}

func TestUpgradeCostDefaults(t *testing.T) {
	if ContainerUpgradeBaseCents != 2000 {
		t.Errorf("ContainerUpgradeBaseCents = %d, want 2000", ContainerUpgradeBaseCents)
	}
	// The production cost for the first rung must match the plan's 20 €.
	if got := UpgradeCostCents(0, ContainerUpgradeBaseCents); got != 2000 {
		t.Errorf("first upgrade = %d, want 2000", got)
	}
}

// The exponent > 1 must make every rung strictly pricier than the one before.
func TestUpgradeCostStrictlyIncreasing(t *testing.T) {
	const base = 2000
	saw := 0
	for level := 0; level < MaxContainerLevel; level++ {
		cost := UpgradeCostCents(level, base)
		if level > 0 && cost <= saw {
			t.Fatalf("level %d costs %d, not above previous %d", level, cost, saw)
		}
		saw = cost
	}
}

func TestContainerLevelName(t *testing.T) {
	tests := []struct {
		level int
		want  string
	}{
		{0, "Portemonnaie"},
		{1, "Jutebeutel"},
		{2, "Rucksack"},
		{3, "IKEA-FRAKTA-Tasche"},
		{4, "Bausparvertrag"},
		{5, ""},
		{-1, ""},
	}
	for _, tt := range tests {
		if got := ContainerLevelName(tt.level); got != tt.want {
			t.Errorf("ContainerLevelName(%d) = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestSplitAtCapacity(t *testing.T) {
	tests := []struct {
		name                        string
		balance, incoming, capacity int
		wantKept, wantOverflow      int
	}{
		{"fits completely", 0, 300, 500, 300, 0},
		{"fills up exactly", 200, 300, 500, 300, 0},
		{"partly overflows", 400, 300, 500, 100, 200},
		{"already full", 500, 300, 500, 0, 300},
		{"over capacity already", 900, 300, 500, 0, 300},
		{"no container", 0, 300, 0, 0, 300},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kept, overflow := SplitAtCapacity(tt.balance, tt.incoming, tt.capacity)
			if kept != tt.wantKept || overflow != tt.wantOverflow {
				t.Errorf("got (%d, %d), want (%d, %d)", kept, overflow, tt.wantKept, tt.wantOverflow)
			}
			if kept+overflow != tt.incoming {
				t.Errorf("money vanished: %d + %d != %d", kept, overflow, tt.incoming)
			}
		})
	}
}
