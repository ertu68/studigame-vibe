package game

import "time"

// BAföG is passive Grundförderung: after applying, a fixed sum lands in the
// container every BafogPeriod. It costs no mood and needs no shift — it just
// arrives, unlike collecting or the job. The amount is the single tuning knob.
const (
	BafogPayoutCents = 3000 // 30,00 € per period
	MaxBafogPeriods  = 4    // clamp so an abandoned account cannot bank a windfall
)

// BafogPeriod is how often the Förderung is paid out.
var BafogPeriod = 6 * time.Hour

// BafogPeriodsElapsed is how many full payouts are due between the last
// scheduled payout and now. It clamps at MaxBafogPeriods, so returning after
// weeks pays out a bounded amount instead of a jackpot.
func BafogPeriodsElapsed(nextAt, now time.Time) int {
	if !now.After(nextAt) {
		return 0
	}
	periods := int(now.Sub(nextAt) / BafogPeriod)
	if periods > MaxBafogPeriods {
		periods = MaxBafogPeriods
	}
	return periods
}

// BafogNextAfter advances the payout pointer by the given number of settled
// periods. nextAt stays aligned to the original cadence, never to now.
func BafogNextAfter(nextAt time.Time, periods int) time.Time {
	return nextAt.Add(time.Duration(periods) * BafogPeriod)
}
