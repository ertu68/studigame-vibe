package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"studigame/internal/game"
)

var (
	ErrActionRunning  = errors.New("character is already busy")
	ErrInvalidShift   = errors.New("unknown shift length")
	ErrNoOpenAction   = errors.New("no open action")
	ErrNotEnoughMoney = errors.New("not enough money for the upgrade")
	ErrContainerMax   = errors.New("container already at max level")
)

// OpenAction is an action that has been started and not resolved yet.
type OpenAction struct {
	ID        int64
	Kind      string
	StartedAt time.Time
	EndsAt    time.Time
}

// Done reports whether the shift is over and can be resolved.
func (a OpenAction) Done(now time.Time) bool { return !now.Before(a.EndsAt) }

// Remaining is the time left, rounded to whole seconds, never negative.
func (a OpenAction) Remaining(now time.Time) time.Duration {
	left := a.EndsAt.Sub(now).Round(time.Second)
	if left < 0 {
		return 0
	}
	return left
}

type collectPayload struct {
	DurationSeconds int `json:"duration_seconds"`
	ContainerLevel  int `json:"container_level"`
	Mood            int `json:"mood"`
}

type collectResult struct {
	Bottles       int `json:"bottles"`
	EarningsCents int `json:"earnings_cents"`
	KeptCents     int `json:"kept_cents"`
	DonatedCents  int `json:"donated_cents"`
	MoodDelta     int `json:"mood_delta"`
}

// CollectSummary is the resolved outcome shown to the player once.
type CollectSummary struct {
	Bottles     int
	Earnings    string
	Donated     string
	HasDonation bool
	MoodDelta   int
}

// FindOpenAction returns the character's unresolved action, if any.
func FindOpenAction(db *sql.DB, characterID int64) (OpenAction, error) {
	var a OpenAction
	err := db.QueryRow(
		`SELECT id, kind, started_at, ends_at FROM actions
		 WHERE character_id = ? AND resolved_at IS NULL`, characterID,
	).Scan(&a.ID, &a.Kind, &a.StartedAt, &a.EndsAt)
	if err == sql.ErrNoRows {
		return OpenAction{}, ErrNoOpenAction
	}
	return a, err
}

// StartCollect books a collect shift. The inputs are frozen into the payload so
// a later equipment change cannot retroactively improve a running shift. The
// outcome itself is not decided here — see ResolveDueAction.
func StartCollect(db *sql.DB, char Character, d time.Duration, now time.Time) error {
	if !game.ValidCollectDuration(d) {
		return ErrInvalidShift
	}

	payload, err := json.Marshal(collectPayload{
		DurationSeconds: int(d.Seconds()),
		ContainerLevel:  char.ContainerLevel,
		Mood:            char.Mood,
	})
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO actions (character_id, kind, started_at, ends_at, payload_json)
		 VALUES (?, 'collect', ?, ?, ?)`,
		char.ID, now, now.Add(d), string(payload),
	)
	if isUniqueViolation(err) {
		return ErrActionRunning
	}
	return err
}

// ResolveDueAction settles a finished shift: it rolls the outcome, moves the
// money and marks the action resolved — all in one transaction, so a reload
// mid-way cannot pay out twice. It returns nil if nothing was due.
func ResolveDueAction(db *sql.DB, char Character, now time.Time, roll float64) (*CollectSummary, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// The row lock is the WHERE clause: resolved_at is set inside this
	// transaction, so a concurrent request finds nothing to resolve.
	var actionID int64
	var payloadJSON string
	err = tx.QueryRow(
		`SELECT id, payload_json FROM actions
		 WHERE character_id = ? AND resolved_at IS NULL AND ends_at <= ?`,
		char.ID, now,
	).Scan(&actionID, &payloadJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var payload collectPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return nil, err
	}

	outcome := game.ResolveCollect(game.CollectParams{
		Duration:       time.Duration(payload.DurationSeconds) * time.Second,
		ContainerLevel: payload.ContainerLevel,
		Mood:           payload.Mood,
	}, roll)

	capacity := game.ContainerCapacity(char.ContainerLevel)
	kept, donated := game.SplitAtCapacity(char.MoneyCents, outcome.EarningsCents, capacity)
	balance := char.MoneyCents + kept
	mood := game.ApplyMood(char.Mood, outcome.MoodDelta)

	result, err := json.Marshal(collectResult{
		Bottles:       outcome.Bottles,
		EarningsCents: outcome.EarningsCents,
		KeptCents:     kept,
		DonatedCents:  donated,
		MoodDelta:     outcome.MoodDelta,
	})
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(
		`UPDATE actions SET resolved_at = ?, result_json = ? WHERE id = ?`,
		now, string(result), actionID,
	); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(
		`UPDATE characters SET money_cents = ?, mood = ? WHERE id = ?`,
		balance, mood, char.ID,
	); err != nil {
		return nil, err
	}

	if kept > 0 {
		if _, err := tx.Exec(
			`INSERT INTO ledger (character_id, delta_cents, balance_after, reason, ref_id)
			 VALUES (?, ?, ?, 'collect', ?)`,
			char.ID, kept, balance, actionID,
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

	return &CollectSummary{
		Bottles:     outcome.Bottles,
		Earnings:    formatCents(kept),
		Donated:     formatCents(donated),
		HasDonation: donated > 0,
		MoodDelta:   outcome.MoodDelta,
	}, nil
}

// CharityPool is the donated total of an edition, for display.
func CharityPool(db *sql.DB, editionID int64) (int, error) {
	var cents int
	err := db.QueryRow(`SELECT cents FROM charity_pool WHERE edition_id = ?`, editionID).Scan(&cents)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return cents, err
}

// UpgradeContainer buys the next container tier. It checks the price and the
// hard max level, debits the ledger and bumps container_level — one transaction,
// so money and level can never disagree. Returns the refreshed character.
func UpgradeContainer(db *sql.DB, char Character) (Character, error) {
	if char.ContainerLevel >= game.MaxContainerLevel {
		return char, ErrContainerMax
	}

	cost := game.UpgradeCostCents(char.ContainerLevel, game.ContainerUpgradeBaseCents)
	if char.MoneyCents < cost {
		return char, ErrNotEnoughMoney
	}

	nextLevel := char.ContainerLevel + 1
	balance := char.MoneyCents - cost

	tx, err := db.Begin()
	if err != nil {
		return char, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE characters SET container_level = ?, money_cents = ? WHERE id = ?`,
		nextLevel, balance, char.ID,
	); err != nil {
		return char, err
	}

	if _, err := tx.Exec(
		`INSERT INTO ledger (character_id, delta_cents, balance_after, reason)
		 VALUES (?, ?, ?, 'container_upgrade')`,
		char.ID, -cost, balance,
	); err != nil {
		return char, err
	}

	if err := tx.Commit(); err != nil {
		return char, err
	}

	char.ContainerLevel = nextLevel
	char.MoneyCents = balance
	return char, nil
}
