// Package game contains pure gameplay logic — no database, no clock, fully testable.
// Time and randomness are handed in, never fetched here.
package game

import (
	"math"
	"time"
)

// CollectDurations are the shift lengths a player can pick.
// Longer shifts pay more in total but lock the character up.
var CollectDurations = []time.Duration{
	15 * time.Minute,
	1 * time.Hour,
	4 * time.Hour,
	8 * time.Hour,
}

// ValidCollectDuration reports whether d is one of the offered shift lengths.
func ValidCollectDuration(d time.Duration) bool {
	for _, allowed := range CollectDurations {
		if d == allowed {
			return true
		}
	}
	return false
}

// CollectParams are the inputs to a collect shift.
type CollectParams struct {
	Duration       time.Duration
	ContainerLevel int     // 0-4, bigger container carries more
	Mood           int     // 0-100, tired characters collect less
	SkillFactor    float64 // Lerneffizienz bonus; 0 (unset) means 1.0
}

// CollectOutcome is what a finished shift produced.
type CollectOutcome struct {
	Bottles       int
	EarningsCents int
	MoodDelta     int
}

// baseBottlesPerHour is the yield of a rested, lucky character.
const baseBottlesPerHour = 40

// centsPerBottle is the German deposit on a single bottle.
const centsPerBottle = 25

// ResolveCollect computes the outcome of a finished shift. The roll is the
// random component and must be in [0,1); it is drawn when the action is
// resolved, never when it is started, so the result cannot be read from the
// database in advance.
func ResolveCollect(p CollectParams, roll float64) CollectOutcome {
	hours := p.Duration.Hours()

	// Mood scales the yield between 60 % (exhausted) and 100 % (rested).
	moodFactor := 0.6 + 0.4*clamp01(float64(p.Mood)/100)

	// Each container level adds 15 % capacity.
	containerFactor := 1 + 0.15*float64(p.ContainerLevel)

	// Luck swings the result by ±20 %.
	luckFactor := 0.8 + 0.4*clamp01(roll)

	skillFactor := gearRatio(p.SkillFactor)
	bottles := int(baseBottlesPerHour * hours * moodFactor * containerFactor * luckFactor * skillFactor)
	if bottles < 0 {
		bottles = 0
	}

	return CollectOutcome{
		Bottles:       bottles,
		EarningsCents: bottles * centsPerBottle,
		MoodDelta:     moodCost(p.Duration),
	}
}

// moodCost is the exhaustion a shift causes: 5 points per hour, at least 2.
func moodCost(d time.Duration) int {
	cost := int(5 * d.Hours())
	if cost < 2 {
		cost = 2
	}
	return -cost
}

// ApplyMood keeps mood inside 0-100 after a delta.
func ApplyMood(mood, delta int) int {
	return clampInt(mood+delta, 0, 100)
}

// MaxContainerLevel is the highest buyable container tier (Bausparvertrag).
// It is a hard ceiling — the upgrade function reports "not buyable" beyond it
// and the UI hides the button. After that the skill system takes over.
const MaxContainerLevel = 4

// ContainerUpgradeBaseCents is the single tuning knob for upgrade prices.
const ContainerUpgradeBaseCents = 2000

// ContainerCapacity is how many cents a character can carry at each level.
// Level 0 has to hold more than the 10 € starting capital, otherwise a new
// character is full before the first shift and everything earned would go
// straight to charity. Every level must also hold enough to afford the next
// upgrade (checked in TestContainerCapacityAffordsNextUpgrade) — level 1 was
// bumped from 20 € to 80 € because the old 50 € cap could never reach the
// ~70 € cost of the level 1→2 upgrade, a permanent softlock.
func ContainerCapacity(level int) int {
	capacities := []int{2000, 8000, 20000, 100000, 500000}
	if level < 0 || level >= len(capacities) {
		return 0
	}
	return capacities[level]
}

// ContainerLevelName is the flavour name of a container tier, for display.
func ContainerLevelName(level int) string {
	names := []string{
		"Portemonnaie",
		"Jutebeutel",
		"Rucksack",
		"IKEA-FRAKTA-Tasche",
		"Bausparvertrag",
	}
	if level < 0 || level >= len(names) {
		return ""
	}
	return names[level]
}

// UpgradeCostCents is what a container upgrade from level to level+1 costs:
// base × (level+1)^1.8, rounded to the nearest euro so the price never shows
// an odd amount like 69,64 €. The exponent > 1 makes later rungs
// overproportionally expensive (anti-snowball). It reports 0 at the max
// level, meaning "not buyable". With the standard base of 2000: 20 €, 70 €,
// 144 €, 243 €.
func UpgradeCostCents(level, base int) int {
	if level < 0 || level >= MaxContainerLevel {
		return 0
	}
	raw := float64(base) * math.Pow(float64(level+1), 1.8)
	return int(math.Round(raw/100)) * 100
}

// SplitAtCapacity divides an incoming amount into the part that still fits in
// the container and the overflow, which goes to the charity pool.
func SplitAtCapacity(balance, incoming, capacity int) (kept, overflow int) {
	room := capacity - balance
	if room <= 0 {
		return 0, incoming
	}
	if incoming <= room {
		return incoming, 0
	}
	return room, incoming - room
}

// gearRatio turns a stored skill factor into a multiplier, treating 0 (field
// not set) as 1.0 so older payloads still resolve at full yield.
func gearRatio(f float64) float64 {
	if f <= 0 {
		return 1
	}
	return f
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampInt(v, minimum, maximum int) int {
	if v < minimum {
		return minimum
	}
	if v > maximum {
		return maximum
	}
	return v
}
