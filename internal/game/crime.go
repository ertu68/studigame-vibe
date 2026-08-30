package game

import "time"

// CrimeRatePerHourCents is the base hourly rate for the "Nebenverdienst" in
// cents — higher than the Werkstudenten-Job (15 €/h vs 8 €/h), because it
// carries risk.
const CrimeRatePerHourCents = 1500

// CrimeCatchChance is the chance a shift ends in a bust (roll <= this value),
// which slashes the payout and hits mood twice as hard.
const CrimeCatchChance = 0.30

// CrimeCatchPayoutFactor is how much of the gross a busted shift still pays
// when the buyer flakes late or nobody looks too closely.
const CrimeCatchPayoutFactor = 0.25

// CrimeParams are the inputs to a finished crime shift. Skill-neutral in the
// MVP — the deferred Taschendiebstahl and Stressresistenz skills (see plan §3)
// would govern risk and mood cost once they return.
type CrimeParams struct {
	Duration time.Duration
	Mood     int
}

// CrimeOutcome is what a finished crime shift produced.
type CrimeOutcome struct {
	EarningsCents int
	MoodDelta     int
	Caught        bool
}

// ResolveCrime computes the outcome of a finished "Nebenverdienst" shift:
// ghostwriting a Hausarbeit or selling a Spickzettel. The roll is the random
// component, drawn when the action is resolved, never when it is started. A
// caught shift pays a fraction and costs double mood.
func ResolveCrime(p CrimeParams, roll float64) CrimeOutcome {
	hours := p.Duration.Hours()
	moodFactor := 0.6 + 0.4*clamp01(float64(p.Mood)/100)
	gross := int(CrimeRatePerHourCents * hours * moodFactor)

	caught := roll <= CrimeCatchChance
	var earnings int
	var moodDelta int
	if caught {
		earnings = int(float64(gross) * CrimeCatchPayoutFactor)
		moodDelta = 2 * moodCost(p.Duration)
	} else {
		earnings = gross
		moodDelta = moodCost(p.Duration)
	}
	return CrimeOutcome{EarningsCents: earnings, MoodDelta: moodDelta, Caught: caught}
}
