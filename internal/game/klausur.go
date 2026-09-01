package game

import "time"

// ECTSParams are the inputs to a finished exam prep shift.
type ECTSParams struct {
	Duration    time.Duration
	Mood        int     // 0-100, tired characters earn fewer points
	SkillFactor float64 // Konzentration bonus; 0 (unset) means 1.0
}

// ECTSOutcome is what a finished exam prep shift produced.
type ECTSOutcome struct {
	ECTSPoints float64 // Decimal ECTS (e.g. 1.5 for half credit)
	MoodDelta  int
}

// BaseECTSPerHour is the fixed ECTS production rate for a rested character.
// Realistically: 1 ECTS ≈ 25–30 hours of work. But in Studigame, faster
// pacing demands higher hourly rates. Value chosen for early game viability:
// 0.25 ECTS/h means 4 hours = 1 full credit, allowing semester completion
// within days, not weeks.
const BaseECTSPerHour = 0.25

// ResolveKlausur computes the outcome of a finished exam prep shift.
// Mood scales between 60% (exhausted) and 100% (rested) productivity.
// Skill bonus increases yield multiplicatively (like other actions).
func ResolveKlausur(p ECTSParams) ECTSOutcome {
	hours := p.Duration.Hours()
	moodFactor := 0.6 + 0.4*clamp01(float64(p.Mood)/100)
	skillFactor := gearRatio(p.SkillFactor)
	points := BaseECTSPerHour * hours * moodFactor * skillFactor
	return ECTSOutcome{
		ECTSPoints: points,
		MoodDelta:  moodCost(p.Duration),
	}
}