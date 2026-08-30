package game

import "time"

// The Campus-Kiosk (plan §3b: "Kaffee statt Promille — temporärer
// Konzentrations-Boost"). Coffee is a consumable money sink that speeds up
// skill training within a time window, the same timed-boost mechanic the
// original had for alcohol — here it stays wholesome. All numbers are
// provisional tuning knobs.
const (
	// CoffeeCostCents is what one coffee sets a student back.
	CoffeeCostCents = 200 // 2,00 €
	// CoffeeBoostWindow is how long one coffee's concentration lasts.
	CoffeeBoostWindow = 2 * time.Hour
	// CoffeeDurationFactor halves skill training while the boost is up.
	CoffeeDurationFactor = 0.5
)

// CoffeeBoostActive reports whether a student is currently caffeinated.
func CoffeeBoostActive(until time.Time, now time.Time) bool {
	return now.Before(until)
}

// BoostedSkillTrainingDuration is the training duration with a coffee boost
// applied: the base duration scaled by CoffeeDurationFactor, rounded down.
func BoostedSkillTrainingDuration(skill string, konzLevel int, subject string, boost bool) int {
	d := SkillTrainingDuration(skill, konzLevel, subject)
	if boost {
		d = int(float64(d) * CoffeeDurationFactor)
	}
	return d
}
