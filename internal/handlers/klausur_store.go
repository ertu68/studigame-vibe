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
