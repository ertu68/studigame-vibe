package game

import "time"

// JobRatePerHourCents is the fixed hourly Werkstudent-Wage in cents — the
// single tuning knob for the job action. Steady income: no bottle luck, always
// the same per hour, unlike collecting. Lower than the peak collect yield, but
// reliable.
const JobRatePerHourCents = 800 // 8,00 €/h

// JobParams are the inputs to a finished job shift.
type JobParams struct {
	Duration time.Duration
	Mood     int // 0-100, tired characters earn less
}

// JobOutcome is what a finished job shift produced. Worth noting: unlike
// collecting there is no luck component — the wage is fixed per hour.
type JobOutcome struct {
	EarningsCents int
	MoodDelta     int
}

// ResolveJob computes the outcome of a finished job shift. Mood scales the
// wage between 60 % (exhausted) and 100 % (rested), so the stat gates all
// income, not just bottles.
func ResolveJob(p JobParams) JobOutcome {
	hours := p.Duration.Hours()
	moodFactor := 0.6 + 0.4*clamp01(float64(p.Mood)/100)
	earnings := int(JobRatePerHourCents * hours * moodFactor)
	return JobOutcome{
		EarningsCents: earnings,
		MoodDelta:     moodCost(p.Duration),
	}
}
