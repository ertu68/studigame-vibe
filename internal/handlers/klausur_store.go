package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	"studigame/internal/game"
)

type klausurPayload struct {
	DurationSeconds int     `json:"duration_seconds"`
	Mood            int     `json:"mood"`
	SkillFactor     float64 `json:"skill_factor,omitempty"`
}

// StartKlausur books an exam prep shift. Same contract as other actions:
// inputs are frozen and only one action per character is allowed.
func StartKlausur(db *sql.DB, char Character, d time.Duration, now time.Time) error {
	if !game.ValidCollectDuration(d) {
		return ErrInvalidShift
	}

	_, _, konzentrFactor, err := SkillFactorFor(db, char.ID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(klausurPayload{
		DurationSeconds: int(d.Seconds()),
		Mood:            char.Mood,
		SkillFactor:     konzentrFactor,
	})
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO actions (character_id, kind, started_at, ends_at, payload_json)
		 VALUES (?, 'klausur', ?, ?, ?)`,
		char.ID, now, now.Add(d), string(payload),
	)
	if isUniqueViolation(err) {
		return ErrActionRunning
	}
	return err
}

// extend resolveOutcome in collect_store.go to handle klausur:
//
// case "klausur":
// 	var p klausurPayload
// 	if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
// 		return actionOutcome{}, err
// 	}
// 	out := game.ResolveKlausur(game.ECTSParams{
// 		Duration:    time.Duration(p.DurationSeconds) * time.Second,
// 		Mood:        p.Mood,
// 		SkillFactor: p.SkillFactor,
// 	})
// 	return actionOutcome{
//		EarningsCents: 0,  // No money earned
//		MoodDelta:     out.MoodDelta,
//		Detail:        fmt.Sprintf("%.2f ECTS-Punkte erreicht", out.ECTSPoints),
// 	}, nil

// extend bookOutcome in collect_store.go to handle ECTS booking:
//
// Replace the UPDATE characters line with:
// 	if _, err := tx.Exec(
// 		`UPDATE characters SET money_cents = ?, mood = ?, ects_cents = ects_cents + ? WHERE id = ?`,
// 		balance, mood, int(out.ECTSPoints*100), char.ID,
// 	); err != nil {
// 		return ActionSummary{}, err
// 	}
//
// And add to the ledger insertion:
// 	if kept > 0 || out.ECTSPoints > 0 {
// 		reason := kind
// 		if out.ECTSPoints > 0 {
// 			reason = "klausur"  // More specific reason for ECTS
// 		}
// 		if _, err := tx.Exec(
// 			`INSERT INTO ledger (character_id, delta_cents, balance_after, reason, ref_id)
// 			 VALUES (?, ?, ?, ?, ?)`,
// 			char.ID, kept, balance, reason, actionID,
// 		); err != nil {
// 			return ActionSummary{}, err
// 		}
// 	}
