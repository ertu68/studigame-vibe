package game

import "time"

// Tuning knobs for the shop effects and mood. Provisional, see ROADMAP 🙋.
const (
	// MoodRegenPerHour is how much mood a rested student gains per hour.
	MoodRegenPerHour = 5
	// MaxMoodBonusPercent caps how much shop items can ease exhaustion.
	MaxMoodBonusPercent = 50
	// MaxPassiveHours caps how many hours of passive income are paid at once,
	// so a long absence is not a jackpot (same idea as MaxBafogPeriods).
	MaxPassiveHours = 24
)

// RegenMood is mood after resting for whole hours, capped at 100.
func RegenMood(mood, hours int) int {
	return clampInt(mood+MoodRegenPerHour*hours, 0, 100)
}

// EaseExhaustion shrinks a (negative) mood delta by the items' mood bonus in
// percent. Positive deltas pass through unchanged.
func EaseExhaustion(delta, bonusPercent int) int {
	if delta >= 0 {
		return delta
	}
	return delta * (100 - min(bonusPercent, MaxMoodBonusPercent)) / 100
}

// WholeHours is how many full hours lie between a stored anchor and now.
// Callers move the anchor by exactly this many hours, so no partial hour is lost
// and a reload cannot pay twice.
func WholeHours(since, now time.Time) int {
	if !now.After(since) {
		return 0
	}
	return int(now.Sub(since) / time.Hour)
}
