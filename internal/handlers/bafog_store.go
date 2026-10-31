package handlers

import (
	"database/sql"
	"time"

	"studigame/internal/game"
	"studigame/internal/web"
)

// parseNextAt turns a nullable DATETIME TEXT column into *time.Time,
// nil when unset.
func parseNextAt(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	// DATETIME is stored in UTC; parsing with time.RFC3339Nano recovers it.
	t, err := time.Parse(time.RFC3339Nano, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// CollectBafog pays out every due period of passive Förderung into the
// container, overflowing to charity like any other income. It returns nil when
// nothing is due (disabled or not yet reached a payout). Booking, ledger and
// the payout pointer move in one transaction so a reload cannot double-pay.
func CollectBafog(db *sql.DB, char Character, now time.Time) (*ActionSummary, error) {
	if char.BafogNextAt == nil {
		return nil, nil
	}
	periods := game.BafogPeriodsElapsed(*char.BafogNextAt, now)
	if periods == 0 {
		return nil, nil
	}

	payout := periods * game.BafogPayoutCents
	nextAt := game.BafogNextAfter(*char.BafogNextAt, game.BafogFullPeriods(*char.BafogNextAt, now))

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	capacity, err := capacityOf(tx, char)
	if err != nil {
		return nil, err
	}
	kept, donated := game.SplitAtCapacity(char.MoneyCents, payout, capacity)
	balance := char.MoneyCents + kept

	// Guard against double-payout: only settle if the payout pointer is still
	// where the caller read it. A concurrent request that already settled will
	// update zero rows (bafog_next_at has moved), so the rollback below returns
	// nil and no money is booked twice. The unconditional UPDATE would otherwise
	// let two overlapping requests both pay one due period.
	res, err := tx.Exec(
		`UPDATE characters SET money_cents = ?, bafog_next_at = ? WHERE id = ? AND bafog_next_at = ?`,
		balance, nextAt.UTC().Format(time.RFC3339Nano), char.ID, char.BafogNextAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		// Someone else settled the due periods first (or BAföG was cancelled
		// mid-flight); nothing to book here.
		return nil, nil
	}

	if kept > 0 {
		if _, err := tx.Exec(
			`INSERT INTO ledger (character_id, delta_cents, balance_after, reason)
			 VALUES (?, ?, ?, 'bafog')`,
			char.ID, kept, balance,
		); err != nil {
			return nil, err
		}
	}

	if donated > 0 {
		if _, err := tx.Exec(
			`UPDATE charity_pool SET cents = cents + ? WHERE edition_id = ?`,
			donated, char.EditionID,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	detail := web.FormatCents(payout) + " beantragt"
	if donated > 0 {
		detail += ", Behälter voll, Rest gespendet"
	}
	return &ActionSummary{
		Title:       "BAföG eingegangen",
		Detail:      detail,
		Donated:     web.FormatCents(donated),
		HasDonation: donated > 0,
	}, nil
}

// SetBafog toggles passive Förderung. Applying anchors the payout cadence at
// now, so the first payout lands one BafogPeriod later. Cancelling stops future
// payouts without clawing back what arrived.
func SetBafog(db *sql.DB, char Character, enabled bool, now time.Time) (Character, error) {
	if enabled == (char.BafogNextAt != nil) {
		return char, nil
	}

	var nextAt *time.Time
	if enabled {
		t := now
		nextAt = &t
	}
	_, err := db.Exec(
		`UPDATE characters SET bafog_next_at = ? WHERE id = ?`,
		nullableNext(nextAt), char.ID,
	)
	if err != nil {
		return char, err
	}
	char.BafogNextAt = nextAt
	return char, nil
}

// nullableNext encodes a possibly-nil pointer for the DATETIME column, in UTC.
func nullableNext(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}
