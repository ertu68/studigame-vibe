package game

import "time"

// BAföG is passive Grundförderung: after applying, a fixed sum lands in the
// container every BafogPeriod. It costs no mood and needs no shift — it just
// arrives, unlike collecting or the job. The amount is the single tuning knob.
const (
	BafogPayoutCents = 2000 // 20,00 € per period (≈3,33 €/h, below the 8 €/h job)
	MaxBafogPeriods  = 4    // clamp so an abandoned account cannot bank a windfall
)

// BafogPeriod is how often the Förderung is paid out.
var BafogPeriod = 6 * time.Hour

// BafogFullPeriods is how many whole periods lie between the payout pointer
// and now, unclamped. The pointer advances by this many, so skipped periods are
// forfeited rather than caught up on the next reload.
func BafogFullPeriods(nextAt, now time.Time) int {
	if !now.After(nextAt) {
		return 0
	}
	return int(now.Sub(nextAt) / BafogPeriod)
}

// BafogPeriodsElapsed is how many payouts are due: the full periods, clamped at
// MaxBafogPeriods, so returning after weeks pays a bounded amount, not a jackpot.
func BafogPeriodsElapsed(nextAt, now time.Time) int {
	return min(BafogFullPeriods(nextAt, now), MaxBafogPeriods)
}

// BafogNextAfter advances the payout pointer by the given number of settled
// periods. nextAt stays aligned to the original cadence, never to now.
func BafogNextAfter(nextAt time.Time, periods int) time.Time {
	return nextAt.Add(time.Duration(periods) * BafogPeriod)
}
