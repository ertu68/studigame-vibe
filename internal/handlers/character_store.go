package handlers

import (
	"database/sql"
	"errors"
	"strings"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrCharacterExists   = errors.New("character already exists for this edition")
	ErrCharacterNotFound = errors.New("character not found")
	ErrEditionNotFound   = errors.New("edition not found or inactive")
	ErrSubjectNotFound   = errors.New("subject not found in this edition")
	ErrNameTaken         = errors.New("name already taken in this edition")
)

// ActiveEditions returns all editions currently open for play.
func ActiveEditions(db *sql.DB) ([]Edition, error) {
	rows, err := db.Query(`SELECT id, slug, name, active FROM editions WHERE active = TRUE ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Edition
	for rows.Next() {
		var e Edition
		if err := rows.Scan(&e.ID, &e.Slug, &e.Name, &e.Active); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ActiveSubjects returns the Studienfächer of an edition for the picker.
func ActiveSubjects(db *sql.DB, editionID int64) ([]Subject, error) {
	rows, err := db.Query(`SELECT id, edition_id, name FROM subjects WHERE edition_id = ? ORDER BY id`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Subject
	for rows.Next() {
		var s Subject
		if err := rows.Scan(&s.ID, &s.EditionID, &s.Name); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// SubjectByID loads a subject, ensuring it belongs to the given edition.
func SubjectByID(db *sql.DB, id, editionID int64) (Subject, error) {
	var s Subject
	err := db.QueryRow(
		`SELECT id, edition_id, name FROM subjects WHERE id = ? AND edition_id = ?`, id, editionID,
	).Scan(&s.ID, &s.EditionID, &s.Name)
	if err == sql.ErrNoRows {
		return Subject{}, ErrSubjectNotFound
	}
	return s, err
}

// CharacterForUser loads the character a user plays in a given edition.
func CharacterForUser(db *sql.DB, userID, editionID int64) (Character, error) {
	var c Character
	var next sql.NullString
	var caff sql.NullString
	err := db.QueryRow(
		`SELECT c.id, c.user_id, c.edition_id, e.slug, c.name, c.subject_id, s.name,
		        c.money_cents, c.container_level, c.mood, c.bafog_next_at, c.caffeine_until
		 FROM characters c
		 JOIN editions e ON e.id = c.edition_id
		 JOIN subjects s ON s.id = c.subject_id
		 WHERE c.user_id = ? AND c.edition_id = ?`, userID, editionID,
	).Scan(&c.ID, &c.UserID, &c.EditionID, &c.EditionSlug, &c.Name,
		&c.SubjectID, &c.SubjectName, &c.MoneyCents, &c.ContainerLevel, &c.Mood, &next, &caff)
	if err == sql.ErrNoRows {
		return Character{}, ErrCharacterNotFound
	}
	c.BafogNextAt = parseNextAt(next)
	c.CaffeineUntil = parseNextAt(caff)
	return c, err
}

// AnyCharacterForUser returns the user's first character across all editions,
// so the dashboard knows whether onboarding is still needed.
func AnyCharacterForUser(db *sql.DB, userID int64) (Character, error) {
	var c Character
	var next sql.NullString
	var caff sql.NullString
	err := db.QueryRow(
		`SELECT c.id, c.user_id, c.edition_id, e.slug, c.name, c.subject_id, s.name,
		        c.money_cents, c.container_level, c.mood, c.bafog_next_at, c.caffeine_until
		 FROM characters c
		 JOIN editions e ON e.id = c.edition_id
		 JOIN subjects s ON s.id = c.subject_id
		 WHERE c.user_id = ?
		 ORDER BY c.created_at LIMIT 1`, userID,
	).Scan(&c.ID, &c.UserID, &c.EditionID, &c.EditionSlug, &c.Name,
		&c.SubjectID, &c.SubjectName, &c.MoneyCents, &c.ContainerLevel, &c.Mood, &next, &caff)
	if err == sql.ErrNoRows {
		return Character{}, ErrCharacterNotFound
	}
	c.BafogNextAt = parseNextAt(next)
	c.CaffeineUntil = parseNextAt(caff)
	return c, err
}

// CreateCharacter creates a character and books its starting capital in the
// ledger. Both happen in one transaction so money never exists unrecorded.
func CreateCharacter(db *sql.DB, userID int64, editionSlug string, subjectID int64, name string) (Character, error) {
	name = strings.TrimSpace(name)

	var editionID int64
	err := db.QueryRow(`SELECT id FROM editions WHERE slug = ? AND active = TRUE`, editionSlug).Scan(&editionID)
	if err == sql.ErrNoRows {
		return Character{}, ErrEditionNotFound
	}
	if err != nil {
		return Character{}, err
	}

	subject, err := SubjectByID(db, subjectID, editionID)
	if err != nil {
		return Character{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Character{}, err
	}
	defer tx.Rollback()

	const startCapital = 1000 // 10 € — mirrors the column default
	res, err := tx.Exec(
		`INSERT INTO characters (user_id, edition_id, name, subject_id, money_cents)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, editionID, name, subject.ID, startCapital,
	)
	if err != nil {
		return Character{}, translateCharacterConstraint(err)
	}
	charID, err := res.LastInsertId()
	if err != nil {
		return Character{}, err
	}

	if _, err := tx.Exec(
		`INSERT INTO ledger (character_id, delta_cents, balance_after, reason)
		 VALUES (?, ?, ?, 'start_capital')`,
		charID, startCapital, startCapital,
	); err != nil {
		return Character{}, err
	}

	if err := tx.Commit(); err != nil {
		return Character{}, err
	}

	return Character{
		ID:          charID,
		UserID:      userID,
		EditionID:   editionID,
		EditionSlug: editionSlug,
		Name:        name,
		SubjectID:   subject.ID,
		SubjectName: subject.Name,
		MoneyCents:  startCapital,
		Mood:        100,
	}, nil
}

// translateCharacterConstraint maps the two UNIQUE constraints on characters
// onto errors the handler can show, and passes anything else through untouched.
func translateCharacterConstraint(err error) error {
	if !isUniqueViolation(err) {
		return err
	}
	if strings.Contains(err.Error(), "characters.name") {
		return ErrNameTaken
	}
	return ErrCharacterExists
}

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
func isUniqueViolation(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}
