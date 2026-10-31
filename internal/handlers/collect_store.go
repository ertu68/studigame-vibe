package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"studigame/internal/game"
	"studigame/internal/web"
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
	DurationSeconds int     `json:"duration_seconds"`
	ContainerLevel  int     `json:"container_level"`
	Mood            int     `json:"mood"`
	SkillFactor     float64 `json:"skill_factor,omitempty"`
}

// jobPayload is the frozen input of a job shift.
type jobPayload struct {
	DurationSeconds int     `json:"duration_seconds"`
	Mood            int     `json:"mood"`
	SkillFactor     float64 `json:"skill_factor,omitempty"`
}

// crimePayload is the frozen input of a crime shift. Skill-neutral in the MVP.
type crimePayload struct {
	DurationSeconds int `json:"duration_seconds"`
	Mood            int `json:"mood"`
}

// ActionSummary is the resolved outcome shown to the player once, for any kind
// of action. Title and Detail describe the result in German; Donated is set
// only when the container overflowed into the charity pool.
type ActionSummary struct {
	Title       string
	Detail      string
	Donated     string
	HasDonation bool
	MoodDelta   int
}

// actionOutcome carries what a resolution produces before it is booked.
type actionOutcome struct {
	EarningsCents int
	ECTSCents     int // exam credit in hundredths of an ECTS point; never money
	MoodDelta     int
	Detail        string
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

	skillFactor, _, _, err := SkillFactorFor(db, char.ID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(collectPayload{
		DurationSeconds: int(d.Seconds()),
		ContainerLevel:  char.ContainerLevel,
		Mood:            char.Mood,
		SkillFactor:     skillFactor,
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

// StartJob books a working-student shift. Same contract as StartCollect: the
// inputs are frozen so a later container upgrade cannot boost a running shift,
// and the open-action index guarantees at most one shift per character.
func StartJob(db *sql.DB, char Character, d time.Duration, now time.Time) error {
	if !game.ValidCollectDuration(d) {
		return ErrInvalidShift
	}

	_, jobFactor, _, err := SkillFactorFor(db, char.ID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(jobPayload{
		DurationSeconds: int(d.Seconds()),
		Mood:            char.Mood,
		SkillFactor:     jobFactor,
	})
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO actions (character_id, kind, started_at, ends_at, payload_json)
		 VALUES (?, 'job', ?, ?, ?)`,
		char.ID, now, now.Add(d), string(payload),
	)
	if isUniqueViolation(err) {
		return ErrActionRunning
	}
	return err
}

// StartCrime books a "Nebenverdienst" shift. Same contract as StartJob: the
// inputs are frozen, and the open-action index guarantees one shift per
// character. No skill factor yet — crime is skill-neutral in the MVP.
func StartCrime(db *sql.DB, char Character, d time.Duration, now time.Time) error {
	if !game.ValidCollectDuration(d) {
		return ErrInvalidShift
	}

	payload, err := json.Marshal(crimePayload{
		DurationSeconds: int(d.Seconds()),
		Mood:            char.Mood,
	})
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO actions (character_id, kind, started_at, ends_at, payload_json)
		 VALUES (?, 'crime', ?, ?, ?)`,
		char.ID, now, now.Add(d), string(payload),
	)
	if isUniqueViolation(err) {
		return ErrActionRunning
	}
	return err
}

// ResolveDueAction settles a finished shift: it resolves the outcome by kind,
// moves the money and marks the action resolved — all in one transaction, so a
// reload mid-way cannot pay out twice. It returns nil if nothing was due.
func ResolveDueAction(db *sql.DB, char Character, now time.Time, roll float64) (*ActionSummary, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// The row lock is the WHERE clause: resolved_at is set inside this
	// transaction, so a concurrent request finds nothing to resolve.
	var actionID int64
	var kind, payloadJSON string
	err = tx.QueryRow(
		`SELECT id, kind, payload_json FROM actions
		 WHERE character_id = ? AND resolved_at IS NULL AND ends_at <= ?`,
		char.ID, now,
	).Scan(&actionID, &kind, &payloadJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	outcome, err := resolveOutcome(kind, payloadJSON, char, roll)
	if err != nil {
		return nil, err
	}
	summary, err := bookOutcome(tx, char, kind, actionID, outcome, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &summary, nil
}

// resolveOutcome computes the earned money and mood for a finished action,
// dispatching on its kind. A new action adds a case here — a switch, no
// resolver framework (plan §3b).
func resolveOutcome(kind, payloadJSON string, char Character, roll float64) (actionOutcome, error) {
	switch kind {
	case "collect":
		var p collectPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return actionOutcome{}, err
		}
		out := game.ResolveCollect(game.CollectParams{
			Duration:       time.Duration(p.DurationSeconds) * time.Second,
			ContainerLevel: p.ContainerLevel,
			Mood:           p.Mood,
			SkillFactor:    p.SkillFactor,
		}, roll)
		return actionOutcome{
			EarningsCents: out.EarningsCents,
			MoodDelta:     out.MoodDelta,
			Detail:        fmt.Sprintf("%d Flaschen, %s", out.Bottles, web.FormatCents(out.EarningsCents)),
		}, nil

	case "job":
		var p jobPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return actionOutcome{}, err
		}
		out := game.ResolveJob(game.JobParams{
			Duration:    time.Duration(p.DurationSeconds) * time.Second,
			Mood:        p.Mood,
			SkillFactor: p.SkillFactor,
		})
		return actionOutcome{
			EarningsCents: out.EarningsCents,
			MoodDelta:     out.MoodDelta,
			Detail:        web.FormatCents(out.EarningsCents) + " verdient",
		}, nil

	case "crime":
		var p crimePayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return actionOutcome{}, err
		}
		out := game.ResolveCrime(game.CrimeParams{
			Duration: time.Duration(p.DurationSeconds) * time.Second,
			Mood:     p.Mood,
		}, roll)
		if out.Caught {
			return actionOutcome{
				EarningsCents: out.EarningsCents,
				MoodDelta:     out.MoodDelta,
				Detail:        "Erwischt worden, nur " + web.FormatCents(out.EarningsCents) + " übrig",
			}, nil
		}
		return actionOutcome{
			EarningsCents: out.EarningsCents,
			MoodDelta:     out.MoodDelta,
			Detail:        "Abgeliefert, " + web.FormatCents(out.EarningsCents) + " verdient",
		}, nil

	case "klausur":
		var p klausurPayload
		if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
			return actionOutcome{}, err
		}
		out := game.ResolveKlausur(game.ECTSParams{
			Duration:    time.Duration(p.DurationSeconds) * time.Second,
			Mood:        p.Mood,
			SkillFactor: p.SkillFactor,
		})
		return actionOutcome{
			ECTSCents: int(math.Round(out.ECTSPoints * 100)),
			MoodDelta: out.MoodDelta,
			Detail:    fmt.Sprintf("%.2f ECTS-Punkte erreicht", out.ECTSPoints),
		}, nil

	default:
		return actionOutcome{}, fmt.Errorf("unknown action kind %q", kind)
	}
}

// bookOutcome moves a resolution's money and mood onto the character, books
// the ledger and updates the charity pool — the one booking path shared by all
// action kinds, so a new action can never fork the money handling.
func bookOutcome(tx *sql.Tx, char Character, kind string, actionID int64, out actionOutcome, now time.Time) (ActionSummary, error) {
	// Balance and mood are read here, not taken from char: the character was
	// loaded before the transaction and may be stale by now.
	var money, mood int
	if err := tx.QueryRow(
		`SELECT money_cents, mood FROM characters WHERE id = ?`, char.ID,
	).Scan(&money, &mood); err != nil {
		return ActionSummary{}, err
	}
	capacity, err := capacityOf(tx, char)
	if err != nil {
		return ActionSummary{}, err
	}
	easeBonus, err := shopEffect(tx, char.ID, "mood_bonus")
	if err != nil {
		return ActionSummary{}, err
	}
	out.MoodDelta = game.EaseExhaustion(out.MoodDelta, easeBonus)
	// Rest starts when the shift ends, not when the player comes back.
	var endsAt time.Time
	if err := tx.QueryRow(`SELECT ends_at FROM actions WHERE id = ?`, actionID).Scan(&endsAt); err != nil {
		return ActionSummary{}, err
	}
	kept, donated := game.SplitAtCapacity(money, out.EarningsCents, capacity)

	summary := ActionSummary{
		Title:       actionTitle(kind),
		Detail:      out.Detail,
		Donated:     web.FormatCents(donated),
		HasDonation: donated > 0,
		MoodDelta:   out.MoodDelta,
	}

	if _, err := tx.Exec(
		`UPDATE actions SET resolved_at = ?, result_json = ? WHERE id = ?`,
		now, out.Detail, actionID,
	); err != nil {
		return ActionSummary{}, err
	}
	if _, err := tx.Exec(
		`UPDATE characters SET mood = ?, mood_at = ?, ects_cents = ects_cents + ? WHERE id = ?`,
		game.ApplyMood(mood, out.MoodDelta), endsAt.UTC().Format(time.RFC3339Nano), out.ECTSCents, char.ID,
	); err != nil {
		return ActionSummary{}, err
	}
	if kept > 0 {
		if _, err := postMoney(tx, char.ID, kept, kind, actionID); err != nil {
			return ActionSummary{}, err
		}
	}

	if donated > 0 {
		if _, err := tx.Exec(
			`UPDATE charity_pool SET cents = cents + ? WHERE edition_id = ?`,
			donated, char.EditionID,
		); err != nil {
			return ActionSummary{}, err
		}
	}

	return summary, nil
}

// actionTitle is the German headline for a resolved action, per kind.
func actionTitle(kind string) string {
	switch kind {
	case "collect":
		return "Leergut eingesammelt"
	case "job":
		return "Werkstudenten-Job erledigt"
	case "crime":
		return "Nebenverdienst bezahlt"
	case "klausur":
		return "Klausur geschrieben"
	default:
		return "Aktion vorbei"
	}
}

// actionInProgressLabel is the present-tense counterpart to actionTitle, for
// the status bar's "gerade beschäftigt mit" badge while a shift still runs.
func actionInProgressLabel(kind string) string {
	switch kind {
	case "collect":
		return "Leergut sammeln"
	case "job":
		return "Werkstudenten-Job"
	case "crime":
		return "Nebenverdienst"
	case "klausur":
		return "Klausurvorbereitung"
	default:
		return "Aktion läuft"
	}
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
	tx, err := db.Begin()
	if err != nil {
		return char, err
	}
	defer tx.Rollback()

	// Level is read inside the transaction: a double click must not buy two
	// tiers at the price of one.
	var level int
	if err := tx.QueryRow(`SELECT container_level FROM characters WHERE id = ?`, char.ID).Scan(&level); err != nil {
		return char, err
	}
	if level >= game.MaxContainerLevel {
		return char, ErrContainerMax
	}

	cost := game.UpgradeCostCents(level, game.ContainerUpgradeBaseCents)
	balance, err := postMoney(tx, char.ID, -cost, "container_upgrade", nil)
	if err != nil {
		return char, err
	}
	if _, err := tx.Exec(
		`UPDATE characters SET container_level = ? WHERE id = ?`, level+1, char.ID,
	); err != nil {
		return char, err
	}
	if err := tx.Commit(); err != nil {
		return char, err
	}

	char.ContainerLevel = level + 1
	char.MoneyCents = balance
	return char, nil
}
