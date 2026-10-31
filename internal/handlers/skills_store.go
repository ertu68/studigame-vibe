package handlers

import (
	"database/sql"
	"errors"
	"time"

	"studigame/internal/game"
)

var (
	ErrSkillTrainingRunning = errors.New("a skill is already being trained")
	ErrSkillAtMax           = errors.New("skill already at max level")
	ErrSkillUnknown         = errors.New("unknown skill")
	ErrNoSkillTraining      = errors.New("no active skill training")
)

// SkillTraining is an in-progress level-up.
type SkillTraining struct {
	ID        int64
	Skill     string
	Level     int
	StartedAt time.Time
	EndsAt    time.Time
	CostCents int
}

// SkillLevels returns the current level of every known skill for a character.
// Unknown skills simply have no row and read as 0.
func SkillLevels(db *sql.DB, characterID int64) (map[string]int, error) {
	levels := make(map[string]int, len(game.SkillCatalog))
	for _, s := range game.SkillCatalog {
		levels[s.Slug] = 0
	}
	rows, err := db.Query(`SELECT skill, level FROM skills WHERE character_id = ?`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var slug string
		var level int
		if err := rows.Scan(&slug, &level); err != nil {
			return nil, err
		}
		levels[slug] = level
	}
	return levels, rows.Err()
}

// SkillFactorFor returns the current skill multipliers of a character, for
// freezing into a new action's payload at start.
func SkillFactorFor(db *sql.DB, characterID int64) (collect, job, konzentr float64, err error) {
	levels, err := SkillLevels(db, characterID)
	if err != nil {
		return 0, 0, 0, err
	}
	return game.CollectSkillFactor(levels["geschicklichkeit"]),
		game.JobSkillFactor(levels["auftreten"]),
		KonzentrSkillFactor(levels["konzentration"]), nil
}

// FindSkillTraining returns the character's active training, if any.
func FindSkillTraining(db *sql.DB, characterID int64) (SkillTraining, error) {
	var t SkillTraining
	err := db.QueryRow(
		`SELECT id, skill, level, started_at, ends_at, cost_cents
		 FROM skill_trainings WHERE character_id = ?`, characterID,
	).Scan(&t.ID, &t.Skill, &t.Level, &t.StartedAt, &t.EndsAt, &t.CostCents)
	if err == sql.ErrNoRows {
		return SkillTraining{}, ErrNoSkillTraining
	}
	return t, err
}

// KonzentrSkillFactor is a wrapper around game.KlausurSkillFactor
func KonzentrSkillFactor(konzentrationLevel int) float64 {
	return game.KlausurSkillFactor(konzentrationLevel)
}

// StartSkillTraining pays for and books one level-up of a skill, debiting the
// cost up front and recording the run — one transaction, one active training
// per character, so money and level can never disagree. Returns the number of
// minutes the run lasts.
func StartSkillTraining(db *sql.DB, char Character, slug string, now time.Time) (int, error) {
	skill := game.SkillBySlug(slug)
	if skill == nil {
		return 0, ErrSkillUnknown
	}

	// A due training is settled first, so the player cannot stack two runs
	// and the level read below already includes it.
	if err := ResolveSkillTrainings(db, char.ID, now); err != nil {
		return 0, err
	}
	levels, err := SkillLevels(db, char.ID)
	if err != nil {
		return 0, err
	}
	level := levels[slug]
	if level >= skill.MaxLevel {
		return 0, ErrSkillAtMax
	}

	base := game.SkillCostCents(*skill, level)
	konz := levels["konzentration"]
	boost := CoffeeBoostActive(char, now)
	duration := game.BoostedSkillTrainingDuration(slug, konz, char.SubjectName, boost)
	cost := int(float64(base) * game.SubjectSkillCostFactor(char.SubjectName, slug))
	target := level + 1

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO skill_trainings (character_id, skill, level, started_at, ends_at, cost_cents)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		char.ID, slug, target, now, now.Add(time.Duration(duration)*time.Minute), cost,
	); err != nil {
		if isUniqueViolation(err) {
			return 0, ErrSkillTrainingRunning
		}
		return 0, err
	}
	if _, err := postMoney(tx, char.ID, -cost, "skill_train", char.ID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return duration, nil
}

// ResolveSkillTrainings bumps the level of every training that is due (money
// is already paid, so completion only writes the level). Called before showing
// the skills page or starting a new run.
func ResolveSkillTrainings(db *sql.DB, characterID int64, now time.Time) error {
	rows, err := db.Query(
		`SELECT skill, level FROM skill_trainings WHERE character_id = ? AND ends_at <= ?`,
		characterID, now,
	)
	if err != nil {
		return err
	}
	var due []struct {
		skill string
		level int
	}
	for rows.Next() {
		var d struct {
			skill string
			level int
		}
		if err := rows.Scan(&d.skill, &d.level); err != nil {
			rows.Close()
			return err
		}
		due = append(due, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(due) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, d := range due {
		if _, err := tx.Exec(
			`INSERT INTO skills (character_id, skill, level) VALUES (?, ?, ?)
			 ON CONFLICT(character_id, skill) DO UPDATE SET level = excluded.level`,
			characterID, d.skill, d.level,
		); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM skill_trainings WHERE character_id = ? AND ends_at <= ?`, characterID, now); err != nil {
		return err
	}
	return tx.Commit()
}

// CancelSkillTraining aborts a run and refunds 80 % (plan §1.3) of its cost.
// It returns the refunded amount.
func CancelSkillTraining(db *sql.DB, char Character, id int64) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// DELETE ... RETURNING is the claim: a second cancel finds no row and
	// cannot refund twice.
	var cost int
	err = tx.QueryRow(
		`DELETE FROM skill_trainings WHERE id = ? AND character_id = ? RETURNING cost_cents`,
		id, char.ID,
	).Scan(&cost)
	if err == sql.ErrNoRows {
		return 0, ErrNoSkillTraining
	}
	if err != nil {
		return 0, err
	}

	refund := cost * 80 / 100
	if _, err := postMoney(tx, char.ID, refund, "skill_refund", id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return refund, nil
}
